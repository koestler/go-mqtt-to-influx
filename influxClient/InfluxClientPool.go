package influxClient

import (
	"github.com/influxdata/influxdb-client-go/v2"
	influxdb2Write "github.com/influxdata/influxdb-client-go/v2/api/write"
	"sync"
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

func (p *ClientPool) Shutdown() {
	p.clientsMutex.RLock()
	defer p.clientsMutex.RUnlock()
	for _, c := range p.clients {
		c.Shutdown()
	}
}

func (p *ClientPool) AddClient(client *Client) {
	p.clientsMutex.Lock()
	defer p.clientsMutex.Unlock()
	p.clients[client.Name()] = client
}

func (p *ClientPool) RemoveClient(client *Client) {
	p.clientsMutex.Lock()
	defer p.clientsMutex.Unlock()
	delete(p.clients, client.Name())
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
	} else {
		receivers = make([]*Client, 0, len(receiversNames))
		for _, receiverName := range receiversNames {
			if receiver, ok := p.clients[receiverName]; ok {
				receivers = append(receivers, receiver)
			}
		}
	}
	return
}

func (p *ClientPool) GetReceiverClientsNames(receiverNames []string) (ret []string) {
	receivers := p.getReceiverClients(receiverNames)
	ret = make([]string, len(receivers))
	for i, r := range receivers {
		ret[i] = r.Name()
	}
	return
}

func ToInfluxPoint(point Point) *influxdb2Write.Point {
	return influxdb2.NewPoint(point.Measurement(), point.Tags(), point.Fields(), point.Time())
}

func (p *ClientPool) WritePoint(point Point, receiverNames []string) {
	for _, receiver := range p.getReceiverClients(receiverNames) {
		receiver.WritePoint(point)
	}
}
