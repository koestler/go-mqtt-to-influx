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
	config   Config
	client   influxdb2.Client
	writeApi influxdb2Api.WriteAPI

	auxiliaryTags []AuxiliaryTag
	localDb       LocalDb
	statistics    Statistics

	lastTransmission time.Time
	errorRetryDelay  time.Duration

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

	// ping the api
	if ok, err := dbClient.Ping(context.Background()); ok {
		log.Printf("influxClient[%s]: ping successful", config.Name())
	} else {
		log.Printf("influxClient[%s]: ping failed: %s", config.Name(), err)
	}

	c := Client{
		config:   config,
		client:   dbClient,
		writeApi: writeApi,

		auxiliaryTags: auxiliaryTags,
		localDb:       localDb,
		statistics:    statistics,

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

func (ic *Client) worker() {
	defer close(ic.closed)

	for {
		select {
		case <-ic.shutdown:
			return // shutdown
		}
		// todo: handle retry
	}
}

func failedCallbackHandler(client string, localDb LocalDb) func(batch string, error influxdbHttp2.Error, retryAttempts uint) bool {
	return func(batch string, error influxdbHttp2.Error, retryAttempts uint) bool {
		// retry once, and then write to backlog
		if retryAttempts < 1 {
			log.Printf("influxClient[%s]: write failed, retry", client)
			return true
		}

		// write to backlog
		if err := localDb.InfluxBacklogAdd(client, batch); err != nil {
			// cannot write to backlog, retry up to 3 times
			retry := retryAttempts < 3
			log.Printf("influxClient[%s]: cannot write backlog, retry=%t, err=%s", client, retry, err)
			return retry
		} else {
			log.Printf("influxClient[%s]: write failed, added to backlog", client)
		}

		return false
	}
}
