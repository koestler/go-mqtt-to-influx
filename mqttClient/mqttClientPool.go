package mqttClient

import (
	"sync"
)

type ClientPool struct {
	clients      map[string]Client
	clientsMutex sync.RWMutex
}

func RunPool() (pool *ClientPool) {
	pool = &ClientPool{
		clients: make(map[string]Client),
	}
	return
}

func (p *ClientPool) RunClients() {
	p.clientsMutex.RLock()
	defer p.clientsMutex.RUnlock()
	for _, c := range p.clients {
		c.Run()
	}
}

func (p *ClientPool) Shutdown() {
	p.clientsMutex.RLock()
	defer p.clientsMutex.RUnlock()
	for _, c := range p.clients {
		c.Shutdown()
	}
}

func (p *ClientPool) AddClient(client Client) {
	p.clientsMutex.Lock()
	defer p.clientsMutex.Unlock()
	p.clients[client.Name()] = client
}

func (p *ClientPool) RemoveClient(client Client) {
	p.clientsMutex.Lock()
	defer p.clientsMutex.Unlock()
	delete(p.clients, client.Name())
}

func (p *ClientPool) GetClientsByNames(clientNames []string) (clients []Client) {
	p.clientsMutex.RLock()
	defer p.clientsMutex.RUnlock()

	if len(clientNames) < 1 {
		clients = make([]Client, len(p.clients))
		i := 0
		for _, c := range p.clients {
			clients[i] = c
			i++
		}
	} else {
		clients = make([]Client, 0, len(clientNames))
		for _, clientName := range clientNames {
			if c, ok := p.clients[clientName]; ok {
				clients = append(clients, c)
			}
		}
	}

	return
}
