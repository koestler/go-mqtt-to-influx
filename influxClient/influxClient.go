package influxClient

import (
	"context"
	"github.com/influxdata/influxdb-client-go/v2"
	influxdb2Api "github.com/influxdata/influxdb-client-go/v2/api"
	influxdbHttp2 "github.com/influxdata/influxdb-client-go/v2/api/http"
	influxdb2Write "github.com/influxdata/influxdb-client-go/v2/api/write"
	"github.com/pkg/errors"
	"log"
	"net"
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

	backlogChan chan string

	shutdown chan struct{}
	closed   chan struct{}
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

const batchSize = 5000
const retryQueueLimit = 20

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

	backlogChan := make(chan string, 8)

	// create asynchronous, auto-retry write api instance
	writeApi := dbClient.WriteAPI(config.Org(), config.Bucket())
	if localDb.Enabled() && config.RetryInterval() > 0 {
		writeApi.SetWriteFailedCallback(failedCallbackHandler(config.Name(), backlogChan))
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

		backlogChan: backlogChan,
		shutdown:    make(chan struct{}),
		closed:      make(chan struct{}),
	}

	go c.worker()

	// create client object
	return &c
}

func minD(a, b time.Duration) time.Duration {
	if a <= b {
		return a
	}
	return b
}

func (ic Client) Shutdown() {
	if ic.localDb.Enabled() {
		// send remaining points
		close(ic.shutdown)
		// wait for retryWorker to shut down
		<-ic.closed
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

func (ic Client) worker() {
	defer close(ic.closed)

	aggregateTicker := time.NewTicker(10 * ic.config.WriteInterval())
	retryTicker := time.NewTicker(ic.config.RetryInterval())
	if !ic.localDb.Enabled() || ic.config.RetryInterval() <= 0 {
		aggregateTicker.Stop()
		retryTicker.Stop()
	}

	retryChan := make(chan struct{}, 1)
	triggerRetryHandler := func() {
		select {
		case retryChan <- struct{}{}:
		default:
		}
	}

	for {
		select {
		case <-ic.shutdown:
			return // shutdown
		case batch := <-ic.backlogChan:
			if err := ic.localDb.InfluxBacklogAdd(ic.Name(), batch); err != nil {
				log.Printf("influxClient[%s]: add failed: %s", ic.Name(), err)
			}
		case <-aggregateTicker.C:
			if err := ic.localDb.InfluxAggregateBacklog(ic.Name(), batchSize); err != nil {
				log.Printf("influxClient[%s]: aggregate failed: %s", ic.Name(), err)
			}
		case <-retryTicker.C:
			triggerRetryHandler()
		case <-retryChan:
			if ic.retryHandler() {
				triggerRetryHandler()
			}
		}
	}
}

func failedCallbackHandler(client string, retryBatchChan chan string) func(batch string, error influxdbHttp2.Error, retryAttempts uint) bool {
	return func(batch string, error influxdbHttp2.Error, retryAttempts uint) bool {
		// write to backlog
		select {
		case retryBatchChan <- batch:
			log.Printf("influxClient[%s]: write failed, added to backlog", client)
			return false
		default:
			log.Printf("influxClient[%s]: write failed, backlog chan full, keep retrying", client)
			return true
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
	var removeFromDb bool
	if err = ic.blockingWriteApi.WriteRecord(ic.ctx, batch); err == nil {
		log.Printf("influxClient[%s]: retryHandler: batch written to influxdb, id=%d", ic.Name(), id)
		removeFromDb = true
	} else {
		removeFromDb = isPermanentError(err)
		log.Printf("influxClient[%s]: retryHandler: error while writing batch, isPermanentError=%t, err=%s", ic.Name(), removeFromDb, err)
	}

	if removeFromDb {
		if err = ic.localDb.InfluxBacklogDelete(id); err == nil {
			// successfully send some points and wrote that to the db
			return true
		} else {
			log.Printf("influxClient[%s]: retryHandler: cannot remove entry from backlog, id=%d, err=%s", ic.Name(), id, err)
		}
	}
	return false
}

func isPermanentError(err error) bool {
	var nerr net.Error
	if errors.As(err, &nerr) {
		// is network error -> non permanent
		return false
	}
	return true
}
