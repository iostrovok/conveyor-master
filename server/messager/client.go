package messager

type IHttpClient interface {
	ID() int
	Close()
	addHttpRequest(*HttpMessage)
	ReadHttpRequest() chan *HttpMessage
}

type HttpClient struct {
	id         int
	fromNodes  chan *HttpMessage
	topMessage *Message
}

// Serve starts server
func NewClient(id int, topMessage *Message) IHttpClient {
	out := &HttpClient{
		id:         id,
		topMessage: topMessage,
		fromNodes:  make(chan *HttpMessage, 1000),
	}

	return out
}

// ID
func (c *HttpClient) ID() int {
	return c.id
}

// ID
func (c *HttpClient) Close() {
	close(c.fromNodes)
	c = nil
}

// ReadHttpRequest
func (c *HttpClient) ReadHttpRequest() chan *HttpMessage {
	return c.fromNodes
}

// addHttpRequest
func (c *HttpClient) addHttpRequest(request *HttpMessage) {
	c.fromNodes <- request
}
