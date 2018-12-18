package influxDbClient

import (
	influxClient "github.com/influxdata/influxdb/client/v2"
	"log"
	"sync"
	"time"
)

type ClientPool struct {
	clients      map[string]*Client
	clientsMutex sync.RWMutex
}

func RunPool() (pool *ClientPool) {
	pool = &ClientPool{
		clients: make(map[string]*Client),
	}

	return
}

func (p *ClientPool) Stop() {
	p.clientsMutex.RLock()
	defer p.clientsMutex.RUnlock()
	for _, c := range p.clients {
		c.Stop()
	}
}

func (p *ClientPool) AddClient(client *Client) {
	p.clientsMutex.Lock()
	defer p.clientsMutex.Unlock()
	p.clients[client.GetName()] = client
}

func (p *ClientPool) RemoveClient(client *Client) {
	p.clientsMutex.Lock()
	defer p.clientsMutex.Unlock()
	delete(p.clients, client.GetName())
}

func (p *ClientPool) getReceiverClients(receiversNames []string) (receivers []*Client) {
	p.clientsMutex.RLock()
	defer p.clientsMutex.RUnlock()

	if len(receiversNames) < 1 {
		receivers = make([]*Client, len(p.clients))
		i := 0
		for _, c := range p.clients {
			receivers[i] = c
			i++
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

func (p *ClientPool) WritePoints(
	measurement string,
	points Points,
	time time.Time,
	receiverNames []string,
) {
	p.WriteRawPoints(points.ToRaw(measurement, time), receiverNames)
}

func (p *ClientPool) WriteRawPoints(rawPoints []RawPoint, receiverNames []string) {
	for _, point := range rawPoints {
		pt, err := influxClient.NewPoint(point.Measurement, point.Tags, point.Fields, point.Time)
		if err != nil {
			log.Printf("ClientPool: error creating a point: %s", err)
			continue
		}

		for _, receiver := range p.getReceiverClients(receiverNames) {
			receiver.pointToSendChannel <- pt
		}
	}
}
