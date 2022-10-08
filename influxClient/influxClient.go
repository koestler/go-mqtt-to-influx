package influxClient

import (
	"context"
	"github.com/influxdata/influxdb-client-go/v2"
	influxdb2Api "github.com/influxdata/influxdb-client-go/v2/api"
	influxdbHttp2 "github.com/influxdata/influxdb-client-go/v2/api/http"
	influxdb2Write "github.com/influxdata/influxdb-client-go/v2/api/write"
	"log"
	"strings"
	"time"
)

type Client struct {
	config           Config
	client           influxdb2.Client
	writeApi         influxdb2Api.WriteAPI
	blockingWriteApi influxdb2Api.WriteAPIBlocking

	auxiliaryTags []AuxiliaryTag
	localDb       LocalDb
	statistics    Statistics

	lastTransmission time.Time
	errorRetryDelay  time.Duration

	ctx    context.Context
	cancel context.CancelFunc

	retryWorkerShutdown chan struct{}
	retryWorkerClosed   chan struct{}
}

type Config interface {
	Name() string
	Url() string
	Token() string
	Org() string
	Bucket() string
	WriteInterval() time.Duration
	RetryInterval() time.Duration
	TimePrecision() time.Duration
	LogDebug() bool
}

type AuxiliaryTag interface {
	Tag() string
	MatchString(value string) bool
	TagValues() map[string]string
}

type Point interface {
	Measurement() string
	Tags() map[string]string
	Fields() map[string]interface{}
	Time() time.Time
}

type LocalDb interface {
	Enabled() bool
	InfluxBacklogAdd(client, batch string) error
	InfluxBacklogSize(client string) (err error, numbBatches, numbLines int)
	InfluxBacklogGet(client string) (err error, id int, batch string)
	InfluxBacklogDelete(id int) error
	InfluxAggregateBacklog(client string, batchSize int) error
}

type Statistics interface {
	IncrementOne(module, name, field string)
}

const batchSize = 1000
const retryQueueLimit = 60

func RunClient(config Config, auxiliaryTags []AuxiliaryTag, localDb LocalDb, statistics Statistics) *Client {
	// Create a new HTTPClient
	opts := influxdb2.DefaultOptions().SetUseGZip(true)
	opts = opts.SetFlushInterval(uint(config.WriteInterval().Milliseconds()))
	opts = opts.SetRetryInterval(uint(
		minD(config.RetryInterval(), (retryQueueLimit-1)*config.WriteInterval()).Milliseconds()),
	)
	opts = opts.SetBatchSize(batchSize).SetRetryBufferLimit(retryQueueLimit * batchSize)
	opts = opts.SetPrecision(config.TimePrecision())
	if config.LogDebug() {
		opts = opts.SetLogLevel(3)
	}
	dbClient := influxdb2.NewClientWithOptions(
		config.Url(),
		config.Token(),
		opts,
	)

	// create asynchronous, auto-retry write api instance
	writeApi := dbClient.WriteAPI(config.Org(), config.Bucket())
	if localDb.Enabled() && config.RetryInterval() > 0 {
		writeApi.SetWriteFailedCallback(failedCallbackHandler(config.Name(), localDb))
	} else {
		writeApi.SetWriteFailedCallback(func(batch string, error influxdbHttp2.Error, retryAttempts uint) bool {
			// log and retry until buffer is full
			log.Printf("influxClient[%s]: write failed, retryAttempts=%d", config.Name(), retryAttempts)
			return true
		})
	}

	// create synchronous api to write the backlog
	blockingWriteApi := dbClient.WriteAPIBlocking(config.Org(), config.Bucket())

	// create main context
	ctx, cancel := context.WithCancel(context.Background())

	// ping the api
	if ok, err := dbClient.Ping(ctx); ok {
		log.Printf("influxClient[%s]: ping successful", config.Name())
	} else {
		log.Printf("influxClient[%s]: ping failed: %s", config.Name(), err)
	}

	c := Client{
		config:           config,
		client:           dbClient,
		writeApi:         writeApi,
		blockingWriteApi: blockingWriteApi,

		auxiliaryTags: auxiliaryTags,
		localDb:       localDb,
		statistics:    statistics,

		ctx:    ctx,
		cancel: cancel,

		retryWorkerShutdown: make(chan struct{}),
		retryWorkerClosed:   make(chan struct{}),
	}

	if localDb.Enabled() {
		go c.retryWorker()
	}

	// create client object
	return &c
}

func minD(a, b time.Duration) time.Duration {
	if a <= b {
		return a
	}
	return b
}

func (ic *Client) Shutdown() {
	if ic.localDb.Enabled() {
		// send remaining points
		close(ic.retryWorkerShutdown)
		// wait for retryWorker to shut down
		<-ic.retryWorkerClosed
	}

	ic.writeApi.Flush()
	ic.client.Close()

	// cancel main context
	ic.cancel()

	log.Printf("influxClient[%s]: closed", ic.Name())
}

func (ic Client) Name() string {
	return ic.config.Name()
}

func (ic Client) WritePoint(point Point) {
	p := ToInfluxPoint(point)

	// add auxiliary tags to influx point
	for _, at := range ic.auxiliaryTags {
		if value, ok := point.Tags()[at.Tag()]; ok && at.MatchString(value) {
			for key, value := range at.TagValues() {
				p.AddTag(key, value)
			}
		}
	}

	ic.writeApi.WritePoint(p)

	// statistics
	line := influxdb2Write.PointToLineProtocol(p, time.Second)
	measurement := strings.Fields(line)[0]
	ic.statistics.IncrementOne("influx", ic.Name(), measurement)
}

func (ic Client) retryWorker() {
	defer close(ic.retryWorkerClosed)

	aggregateTicker := time.Tick(10 * ic.config.WriteInterval())
	retryTicker := time.Tick(ic.config.RetryInterval())
	retryChan := make(chan struct{}, 1)
	triggerRetryHandler := func() {
		select {
		case retryChan <- struct{}{}:
		default:
		}
	}

	for {
		select {
		case <-ic.retryWorkerShutdown:
			return // shutdown
		case <-aggregateTicker:
			if err := ic.localDb.InfluxAggregateBacklog(ic.Name(), batchSize); err != nil {
				log.Printf("influxClient[%s]: aggregate failed: %s", ic.Name(), err)
			}
		case <-retryTicker:
			triggerRetryHandler()
		case <-retryChan:
			if ic.retryHandler() {
				triggerRetryHandler()
			}
		}
	}
}

func failedCallbackHandler(client string, localDb LocalDb) func(batch string, error influxdbHttp2.Error, retryAttempts uint) bool {
	return func(batch string, error influxdbHttp2.Error, retryAttempts uint) bool {
		// write to backlog
		if err := localDb.InfluxBacklogAdd(client, batch); err != nil {
			// cannot write to backlog, enable retry
			log.Printf("influxClient[%s]: write failed, cannot write backlog, keep retrying, err=%s", client, err)
			return true
		} else {
			log.Printf("influxClient[%s]: write failed, added to backlog", client)
			return false
		}
	}
}

func (ic Client) retryHandler() (success bool) {
	// while there is something on the backlog, send it synchronously and remove it on success
	if err, numbBatches, numbLines := ic.localDb.InfluxBacklogSize(ic.Name()); err != nil {
		log.Printf("influxClient[%s]: retryHandler: cannot access backlog: err=%s", ic.Name(), err)
		return false
	} else if numbBatches < 1 {
		return false
	} else {
		log.Printf(
			"influxClient[%s]: retryHandler: backlog is not empty: numbBatches=%d, numbLines=%d",
			ic.Name(), numbBatches, numbLines,
		)
	}

	err, id, batch := ic.localDb.InfluxBacklogGet(ic.Name())
	if err != nil {
		return false
	}

	// try to write to influxdb synchronously
	if err = ic.blockingWriteApi.WriteRecord(ic.ctx, batch); err != nil {
		log.Printf("influxClient[%s]: retryHandler: error while writing batch, err=%s", ic.Name(), err)
		return false
	} else {
		log.Printf("influxClient[%s]: retryHandler: batch written to influxdb, id=%d", ic.Name(), id)
		if err = ic.localDb.InfluxBacklogDelete(id); err != nil {
			log.Printf("influxClient[%s]: retryHandler: cannot remove entry from backlog, id=%d, err=%s", ic.Name(), id, err)
			return false
		}
		// successfully send some points and wrote that to the db
		return true
	}

}
