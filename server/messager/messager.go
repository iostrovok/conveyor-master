package messager

import (
	"sync"

	"github.com/iostrovok/conveyor/protobuf/go/nodes"
)

type HttpMessage struct {
	ClusterID        string      `json:"ClusterID"`
	NodeID           string      `json:"NodeID"`
	AllNodes         bool        `json:"AllNodes"`
	Error            error       `json:"Error"`
	ManagerData      interface{} `json:"ManagerData"`
	ErrorManagerData interface{} `json:"ErrorManagerData"`
}

type IMessage interface {
	HttpClient() IHttpClient
	DeleteClient(client IHttpClient)
	AddHttpRequest(request *HttpMessage)
	AddGrpcRequest(request *nodes.SlaveNodeInfoRequest)
	ReadGrpcRequest() *nodes.SlaveNodeInfoRequest
}

type Message struct {
	mxHttp sync.RWMutex
	mxNode sync.RWMutex

	clientCount int

	fromHttp chan *nodes.SimpleResult
	clients  []IHttpClient
}

// Serve starts server
func New() IMessage {
	out := &Message{
		mxHttp:   sync.RWMutex{},
		mxNode:   sync.RWMutex{},
		fromHttp: make(chan *nodes.SimpleResult, 1000),
		clients:  make([]IHttpClient, 0),
	}

	return out
}

// AddGrpcRequest
func (m *Message) AddGrpcRequest(request *nodes.SlaveNodeInfoRequest) {
	m.mxNode.Lock()
	defer m.mxNode.Unlock()

	mess := &HttpMessage{
		ClusterID:        request.GetClusterID(),
		NodeID:           request.GetNodeID(),
		AllNodes:         false,
		ManagerData:      request.ManagerData,
		ErrorManagerData: request.ErrorManagerData,
	}

	for _, c := range m.clients {
		c.addHttpRequest(mess)
	}
}

func (m *Message) AddHttpRequest(request *HttpMessage) {
	m.mxHttp.Lock()
	defer m.mxHttp.Unlock()

	mess := &nodes.SimpleResult{
		OK: true,
	}

	m.fromHttp <- mess
}

// ReadGrpcRequest
func (m *Message) ReadGrpcRequest() *nodes.SlaveNodeInfoRequest {
	m.mxNode.RLock()
	defer m.mxNode.RUnlock()

	return &nodes.SlaveNodeInfoRequest{}
}

// ReadGrpcRequest
func (m *Message) HttpClient() IHttpClient {
	m.mxNode.RLock()
	defer m.mxNode.RUnlock()

	m.clientCount++

	client := NewClient(m.clientCount, m)
	m.clients = append(m.clients, client)
	return client
}

// DeleteClient
func (m *Message) DeleteClient(client IHttpClient) {
	m.mxNode.RLock()
	defer m.mxNode.RUnlock()

	if len(m.clients) < 2 {
		client.Close()
		m.fromHttp = make(chan *nodes.SimpleResult, 1000)
		m.clients = make([]IHttpClient, 0)
		return
	}

	for i := range m.clients {
		if client.ID() == m.clients[i].ID() {
			if i == 0 {
				client.Close()
				if len(m.clients)-1 == i {
					m.clients = m.clients[:len(m.clients)]
				} else if i == 0 {
					m.clients = m.clients[1:]
				} else {
					m.clients = append(m.clients[0:i], m.clients[i:]...)
				}
			}
		}
	}
}
