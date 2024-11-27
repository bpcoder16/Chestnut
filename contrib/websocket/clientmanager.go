package websocket

import "sync"

type ClientManager struct {
	clients *sync.Map
}

func NewClientManager() *ClientManager {
	return &ClientManager{
		clients: new(sync.Map),
	}
}

func (m *ClientManager) Len() int {
	cnt := 0
	m.clients.Range(func(k, v interface{}) bool {
		cnt++
		return true
	})
	return cnt
}

func (m *ClientManager) Store(uuid string, c *Client) {
	m.clients.Store(uuid, c)
}

func (m *ClientManager) Delete(uuid string) {
	m.clients.Delete(uuid)
}

func (m *ClientManager) IsExist(uuid string) bool {
	_, ok := m.clients.Load(uuid)
	return ok
}

func (m *ClientManager) Range(f func(key, value any) bool) {
	m.clients.Range(f)
}
