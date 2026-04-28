package websocket

import (
	"context"
	"sync"
)

type ClientManager struct {
	mu            sync.RWMutex
	clientsByUUID map[string]*Client
	uuidsByUserID map[string]map[string]struct{}
}

func NewClientManager() *ClientManager {
	return &ClientManager{
		clientsByUUID: make(map[string]*Client),
		uuidsByUserID: make(map[string]map[string]struct{}),
	}
}

func (m *ClientManager) Len() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.clientsByUUID)
}

func (m *ClientManager) StoreClient(c *Client) {
	if c == nil {
		return
	}
	m.mu.Lock()
	defer m.mu.Unlock()

	uuid := c.UUID()
	if oldClient, ok := m.clientsByUUID[uuid]; ok {
		m.deleteUserIDIndexLocked(oldClient.UserID(), uuid)
	}
	m.clientsByUUID[uuid] = c
	if c.UserID() == "" {
		return
	}
	if _, ok := m.uuidsByUserID[c.UserID()]; !ok {
		m.uuidsByUserID[c.UserID()] = make(map[string]struct{})
	}
	m.uuidsByUserID[c.UserID()][uuid] = struct{}{}
}

func (m *ClientManager) DeleteClient(c *Client) {
	if c == nil {
		return
	}
	m.DeleteByUUID(c.UUID())
}

func (m *ClientManager) DeleteByUUID(uuid string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	c, ok := m.clientsByUUID[uuid]
	if !ok {
		return
	}
	delete(m.clientsByUUID, uuid)
	m.deleteUserIDIndexLocked(c.UserID(), uuid)
}

func (m *ClientManager) ExistsByUUID(uuid string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	_, ok := m.clientsByUUID[uuid]
	return ok
}

func (m *ClientManager) GetByUUID(uuid string) (*Client, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	c, ok := m.clientsByUUID[uuid]
	return c, ok
}

func (m *ClientManager) Range(f func(uuid string, client *Client) bool) {
	items := m.snapshot()
	for _, item := range items {
		if !f(item.uuid, item.client) {
			return
		}
	}
}

func (m *ClientManager) RangeByUserID(userID string, f func(client *Client) bool) {
	clients := m.snapshotByUserID(userID)
	for _, c := range clients {
		if !f(c) {
			return
		}
	}
}

func (m *ClientManager) CloseByUserID(ctx context.Context, userID string) {
	m.RangeByUserID(userID, func(client *Client) bool {
		client.Close(ctx)
		return true
	})
}

func (m *ClientManager) CountByUserID(userID string) int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.uuidsByUserID[userID])
}

func (m *ClientManager) IsUserOnline(userID string) bool {
	return m.CountByUserID(userID) > 0
}

func (m *ClientManager) deleteUserIDIndexLocked(userID, uuid string) {
	if userID == "" {
		return
	}
	uuids, ok := m.uuidsByUserID[userID]
	if !ok {
		return
	}
	delete(uuids, uuid)
	if len(uuids) == 0 {
		delete(m.uuidsByUserID, userID)
	}
}

type clientManagerItem struct {
	uuid   string
	client *Client
}

func (m *ClientManager) snapshot() []clientManagerItem {
	m.mu.RLock()
	defer m.mu.RUnlock()

	items := make([]clientManagerItem, 0, len(m.clientsByUUID))
	for uuid, c := range m.clientsByUUID {
		items = append(items, clientManagerItem{uuid: uuid, client: c})
	}
	return items
}

func (m *ClientManager) snapshotByUserID(userID string) []*Client {
	m.mu.RLock()
	defer m.mu.RUnlock()

	uuids := m.uuidsByUserID[userID]
	clients := make([]*Client, 0, len(uuids))
	for uuid := range uuids {
		if c, ok := m.clientsByUUID[uuid]; ok {
			clients = append(clients, c)
		}
	}
	return clients
}
