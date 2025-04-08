package websockets

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"sema/models/delta"
)

func startTestServer(t *testing.T) (*httptest.Server, *websocket.Dialer, string) {
	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}

	var connLock sync.Mutex
	var testConns []*websocket.Conn

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			t.Fatal(err)
		}
		connLock.Lock()
		testConns = append(testConns, conn)
		connLock.Unlock()
	}))

	return server, &websocket.Dialer{}, "ws" + server.URL[4:]
}

func TestOpenConnection(t *testing.T) {
	server, dialer, url := startTestServer(t)
	defer server.Close()

	manager := SpawnWebSocketManager()
	sectionID := "open-section"

	conn, _, err := dialer.Dial(url, nil)
	assert.NoError(t, err)

	manager.OpenConnection(sectionID, conn)
	assert.Equal(t, 1, manager.GetNumofConns(sectionID))
}

func TestCloseConnection(t *testing.T) {
	server, dialer, url := startTestServer(t)
	defer server.Close()

	manager := SpawnWebSocketManager()
	sectionID := "close-section"

	conn, _, err := dialer.Dial(url, nil)
	assert.NoError(t, err)

	manager.OpenConnection(sectionID, conn)
	assert.Equal(t, 1, manager.GetNumofConns(sectionID))

	manager.CloseConnection(sectionID, conn)
	assert.Equal(t, 0, manager.GetNumofConns(sectionID))
}

func TestSendToIDExpectConn(t *testing.T) {
	server, dialer, url := startTestServer(t)
	defer server.Close()

	manager := SpawnWebSocketManager()
	sectionID := "send-section"

	conn1, _, err := dialer.Dial(url, nil)
	assert.NoError(t, err)
	conn2, _, err := dialer.Dial(url, nil)
	assert.NoError(t, err)

	manager.OpenConnection(sectionID, conn1)
	manager.OpenConnection(sectionID, conn2)

	// Send a delta message from conn1, expect it NOT to receive it
	msg := delta.Delta{}
	manager.SendToIDExpectConn(sectionID, msg, conn1)
	assert.Equal(t, 2, manager.GetNumofConns(sectionID))
}

func TestRequestSectionContents(t *testing.T) {
	server, dialer, url := startTestServer(t)
	defer server.Close()

	manager := SpawnWebSocketManager()
	sectionID := "request-section"

	conn1, _, err := dialer.Dial(url, nil)
	assert.NoError(t, err)
	conn2, _, err := dialer.Dial(url, nil)
	assert.NoError(t, err)

	manager.OpenConnection(sectionID, conn1)
	manager.OpenConnection(sectionID, conn2)

	manager.RequestSectionContents(sectionID, conn1)
	assert.Equal(t, 2, manager.GetNumofConns(sectionID))
}

func TestGetNumofConns(t *testing.T) {
	server, dialer, url := startTestServer(t)
	defer server.Close()

	manager := SpawnWebSocketManager()
	sectionID := "count-section"

	assert.Equal(t, 0, manager.GetNumofConns(sectionID))

	conn, _, err := dialer.Dial(url, nil)
	assert.NoError(t, err)

	manager.OpenConnection(sectionID, conn)
	assert.Equal(t, 1, manager.GetNumofConns(sectionID))
}
