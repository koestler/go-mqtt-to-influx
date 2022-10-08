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
	InfluxBacklogAdd(client, batch string) error
	InfluxBacklogGet(client string) (err error, id int, batch string)
	InfluxBacklogDelete(id int) error
}

type Statistics interface {
	IncrementOne(module, name, field string)
}

func RunClient(config Config, auxiliaryTags []AuxiliaryTag, localDb LocalDb, statistics Statistics) *Client {
	// Create a new HTTPClient
	opts := influxdb2.DefaultOptions().SetUseGZip(true)
	opts = opts.SetFlushInterval(uint(config.WriteInterval().Milliseconds()))
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
	writeApi.SetWriteFailedCallback(failedCallbackHandler(config.Name(), localDb))

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

		shutdown: make(chan struct{}),
		closed:   make(chan struct{}),
	}

	go c.worker()

	// create client object
	return &c
}

func (ic *Client) Shutdown() {
	// send remaining points
	close(ic.shutdown)
	// wait for worker to shut down
	<-ic.closed

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

	ticker := time.Tick(60 * time.Second)
	for {
		select {
		case <-ic.shutdown:
			return // shutdown
		case <-ticker:
			ic.retryHandler()
		}
		// todo: handle retry
	}
}

func failedCallbackHandler(client string, localDb LocalDb) func(batch string, error influxdbHttp2.Error, retryAttempts uint) bool {
	return func(batch string, error influxdbHttp2.Error, retryAttempts uint) bool {
		// write to backlog
		if err := localDb.InfluxBacklogAdd(client, batch); err != nil {
			// cannot write to backlog, retry up to 3 times
			log.Printf("influxClient[%s]: write failed, cannot write backlog, keep retrying, err=%s", client, err)
			return true
		} else {
			log.Printf("influxClient[%s]: write failed, added to backlog", client)
			return false
		}
	}
}

func (ic Client) retryHandler() {
	// while there is something on the backlog, send it synchronously and remove it on success
	for {
		log.Printf("influxClient[%s]: retryHandler", ic.Name())

		err, id, batch := ic.localDb.InfluxBacklogGet(ic.Name())
		if err != nil {
			break
		}

		// try to write to influxdb synchronously
		if err = ic.blockingWriteApi.WriteRecord(ic.ctx, batch); err != nil {
			log.Printf("influxClient[%s]: retryHandler: error while writing batch, err=%s", ic.Name(), err)
			break
		}

		log.Printf("influxClient[%s]: retryHandler: backlog written to influxdb, id=%d", ic.Name(), id)

		if err = ic.localDb.InfluxBacklogDelete(id); err != nil {
			log.Printf("influxClient[%s]: retryHandler: cannot remove entry from backlog, id=%d, err=%s", ic.Name(), id, err)
		}
	}

}
