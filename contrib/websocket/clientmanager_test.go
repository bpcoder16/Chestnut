package websocket

import (
	"context"
	"testing"
)

func TestClientManagerIndexesByUUIDAndUserID(t *testing.T) {
	m := NewClientManager()
	c1 := NewClient(nil, "uuid-1", "user-1")
	c2 := NewClient(nil, "uuid-2", "user-1")
	c3 := NewClient(nil, "uuid-3", "user-2")

	m.StoreClient(c1)
	m.StoreClient(c2)
	m.StoreClient(c3)

	if got := m.Len(); got != 3 {
		t.Fatalf("Len() = %d, want 3", got)
	}
	if got := m.CountByUserID("user-1"); got != 2 {
		t.Fatalf("CountByUserID(user-1) = %d, want 2", got)
	}
	if !m.IsUserOnline("user-2") {
		t.Fatal("IsUserOnline(user-2) = false, want true")
	}
	if got, ok := m.GetByUUID("uuid-1"); !ok || got != c1 {
		t.Fatalf("GetByUUID(uuid-1) = %v, %v; want c1, true", got, ok)
	}

	m.DeleteClient(c1)
	if m.ExistsByUUID("uuid-1") {
		t.Fatal("uuid-1 still exists after DeleteClient")
	}
	if got := m.CountByUserID("user-1"); got != 1 {
		t.Fatalf("CountByUserID(user-1) after delete = %d, want 1", got)
	}

	replacement := NewClient(nil, "uuid-2", "user-3")
	m.StoreClient(replacement)
	if got := m.CountByUserID("user-1"); got != 0 {
		t.Fatalf("old user index count = %d, want 0", got)
	}
	if got := m.CountByUserID("user-3"); got != 1 {
		t.Fatalf("new user index count = %d, want 1", got)
	}
}

func TestClientManagerNoopAndEmptyUserID(t *testing.T) {
	m := NewClientManager()

	m.StoreClient(nil)
	m.DeleteClient(nil)
	m.DeleteByUUID("missing")

	if got := m.Len(); got != 0 {
		t.Fatalf("Len() after no-op calls = %d, want 0", got)
	}

	client := NewClient(nil, "uuid-1", "")
	m.StoreClient(client)
	if got := m.Len(); got != 1 {
		t.Fatalf("Len() = %d, want 1", got)
	}
	if got := m.CountByUserID(""); got != 0 {
		t.Fatalf("CountByUserID(empty) = %d, want 0", got)
	}
	if m.IsUserOnline("") {
		t.Fatal("IsUserOnline(empty) = true, want false")
	}
}

func TestClientManagerDeleteByUUIDAndMissingLookups(t *testing.T) {
	m := NewClientManager()
	client := NewClient(nil, "uuid-1", "user-1")
	m.StoreClient(client)

	if got, ok := m.GetByUUID("missing"); ok || got != nil {
		t.Fatalf("GetByUUID(missing) = %v, %v; want nil, false", got, ok)
	}
	if m.ExistsByUUID("missing") {
		t.Fatal("ExistsByUUID(missing) = true, want false")
	}

	m.DeleteByUUID("uuid-1")
	if got := m.Len(); got != 0 {
		t.Fatalf("Len() after DeleteByUUID = %d, want 0", got)
	}
	if got := m.CountByUserID("user-1"); got != 0 {
		t.Fatalf("CountByUserID(user-1) after DeleteByUUID = %d, want 0", got)
	}
}

func TestClientManagerRangeStopsWhenCallbackReturnsFalse(t *testing.T) {
	m := NewClientManager()
	m.StoreClient(NewClient(nil, "uuid-1", "user-1"))
	m.StoreClient(NewClient(nil, "uuid-2", "user-1"))
	m.StoreClient(NewClient(nil, "uuid-3", "user-2"))

	calls := 0
	m.Range(func(uuid string, client *Client) bool {
		if uuid == "" || client == nil {
			t.Fatalf("Range callback got uuid=%q client=%v, want populated values", uuid, client)
		}
		calls++
		return false
	})

	if calls != 1 {
		t.Fatalf("Range callback calls = %d, want 1", calls)
	}
	if got := m.Len(); got != 3 {
		t.Fatalf("Len() after stopped Range = %d, want 3", got)
	}
}

func TestClientManagerRangeAllowsCloseInCallback(t *testing.T) {
	m := NewClientManager()
	ws := &WebSocket{clientManager: m}
	for _, client := range []*Client{
		NewClient(nil, "uuid-1", "user-1"),
		NewClient(nil, "uuid-2", "user-1"),
		NewClient(nil, "uuid-3", "user-2"),
	} {
		client.ws = ws
		m.StoreClient(client)
	}

	closed := 0
	m.Range(func(_ string, client *Client) bool {
		client.Close(context.Background())
		closed++
		return true
	})

	if closed != 3 {
		t.Fatalf("closed = %d, want 3", closed)
	}
	if got := m.Len(); got != 0 {
		t.Fatalf("Len() = %d, want 0", got)
	}
}

func TestClientManagerRangeByUserIDAllowsCloseInCallback(t *testing.T) {
	m := NewClientManager()
	ws := &WebSocket{clientManager: m}
	c1 := NewClient(nil, "uuid-1", "user-1")
	c2 := NewClient(nil, "uuid-2", "user-1")
	c1.ws = ws
	c2.ws = ws
	m.StoreClient(c1)
	m.StoreClient(c2)

	closed := 0
	m.RangeByUserID("user-1", func(client *Client) bool {
		client.Close(context.Background())
		closed++
		return true
	})

	if closed != 2 {
		t.Fatalf("closed = %d, want 2", closed)
	}
	if got := m.CountByUserID("user-1"); got != 0 {
		t.Fatalf("CountByUserID(user-1) = %d, want 0", got)
	}
	if got := m.Len(); got != 0 {
		t.Fatalf("Len() = %d, want 0", got)
	}
}

func TestClientManagerRangeByUserIDStopsAndIgnoresMissingUser(t *testing.T) {
	m := NewClientManager()
	m.StoreClient(NewClient(nil, "uuid-1", "user-1"))
	m.StoreClient(NewClient(nil, "uuid-2", "user-1"))
	m.StoreClient(NewClient(nil, "uuid-3", "user-2"))

	missingCalls := 0
	m.RangeByUserID("missing", func(client *Client) bool {
		missingCalls++
		return true
	})
	if missingCalls != 0 {
		t.Fatalf("RangeByUserID(missing) callback calls = %d, want 0", missingCalls)
	}

	calls := 0
	m.RangeByUserID("user-1", func(client *Client) bool {
		if client == nil || client.UserID() != "user-1" {
			t.Fatalf("RangeByUserID callback got %v, want user-1 client", client)
		}
		calls++
		return false
	})
	if calls != 1 {
		t.Fatalf("RangeByUserID callback calls = %d, want 1", calls)
	}
	if got := m.CountByUserID("user-1"); got != 2 {
		t.Fatalf("CountByUserID(user-1) after stopped RangeByUserID = %d, want 2", got)
	}
}

func TestClientManagerCloseByUserID(t *testing.T) {
	m := NewClientManager()
	ws := &WebSocket{clientManager: m}
	c1 := NewClient(nil, "uuid-1", "user-1")
	c2 := NewClient(nil, "uuid-2", "user-1")
	c3 := NewClient(nil, "uuid-3", "user-2")
	c1.ws = ws
	c2.ws = ws
	c3.ws = ws
	m.StoreClient(c1)
	m.StoreClient(c2)
	m.StoreClient(c3)

	m.CloseByUserID(context.Background(), "user-1")

	if !c1.IsClosed() || !c2.IsClosed() {
		t.Fatal("user-1 clients are not both closed")
	}
	if c3.IsClosed() {
		t.Fatal("user-2 client closed, want open")
	}
	if got := m.CountByUserID("user-1"); got != 0 {
		t.Fatalf("CountByUserID(user-1) = %d, want 0", got)
	}
	if got := m.CountByUserID("user-2"); got != 1 {
		t.Fatalf("CountByUserID(user-2) = %d, want 1", got)
	}
	if got := m.Len(); got != 1 {
		t.Fatalf("Len() = %d, want 1", got)
	}
}
