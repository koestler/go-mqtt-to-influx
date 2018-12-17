package influxDbClient

import (
	influxClient "github.com/influxdata/influxdb/client/v2"
	"log"
	"sync"
	"time"
)

type InfluxDbClientPool struct {
	clients      map[string]*InfluxDbClient
	clientsMutex sync.RWMutex
}

func RunPool() (pool *InfluxDbClientPool) {
	pool = &InfluxDbClientPool{
		clients: make(map[string]*InfluxDbClient),
	}

	return
}

func (p *InfluxDbClientPool) Stop() {
	p.clientsMutex.RLock()
	defer p.clientsMutex.RUnlock()
	for _, c := range p.clients {
		c.Stop()
	}
}

func (p *InfluxDbClientPool) AddClient(client *InfluxDbClient) {
	p.clientsMutex.Lock()
	defer p.clientsMutex.Unlock()
	p.clients[client.GetName()] = client
}

func (p *InfluxDbClientPool) RemoveClient(client *InfluxDbClient) {
	p.clientsMutex.Lock()
	defer p.clientsMutex.Unlock()
	delete(p.clients, client.GetName())
}

func (p *InfluxDbClientPool) getReceiverClients(receiversNames []string) (receivers []*InfluxDbClient) {
	p.clientsMutex.RLock()
	defer p.clientsMutex.RUnlock()

	if len(receiversNames) < 1 {
		receivers = make([]*InfluxDbClient, len(p.clients))
		i := 0
		for _, c := range p.clients {
			receivers[i] = c
			i += 1
		}
		return
	}

	for _, receiverName := range receiversNames {
		if receiver, ok := p.clients[receiverName]; ok {
			receivers = append(receivers, receiver)
		}
	}

	return
}

func (p *InfluxDbClientPool) WritePoints(
	measurement string,
	points Points,
	time time.Time,
	receiverNames []string,
) {
	p.WriteRawPoints(points.ToRaw(measurement, time), receiverNames)
}

func (p *InfluxDbClientPool) WriteRawPoints(rawPoints []RawPoint, receiverNames []string) {
	for _, point := range rawPoints {
		pt, err := influxClient.NewPoint(point.Measurement, point.Tags, point.Fields, point.Time)
		if err != nil {
			log.Printf("InfluxDbClientPool: error creating a point: %s", err)
			continue
		}

		for _, receiver := range p.getReceiverClients(receiverNames) {
			receiver.pointToSendChannel <- pt
		}
	}
}
