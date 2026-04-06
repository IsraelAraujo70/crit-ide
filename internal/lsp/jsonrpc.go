package lsp

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
)

// JSONRPCMessage represents a JSON-RPC 2.0 message.
type JSONRPCMessage struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id,omitempty"`
	Method  string          `json:"method,omitempty"`
	Params  json.RawMessage `json:"params,omitempty"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *JSONRPCError   `json:"error,omitempty"`
}

// JSONRPCError represents a JSON-RPC error.
type JSONRPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (e *JSONRPCError) Error() string {
	return fmt.Sprintf("JSON-RPC error %d: %s", e.Code, e.Message)
}

// Response is a JSON-RPC response received from the server.
type Response struct {
	Result json.RawMessage
	Error  *JSONRPCError
}

// Client is a JSON-RPC 2.0 client communicating over stdin/stdout.
type Client struct {
	writer  io.Writer
	reader  *bufio.Reader
	nextID  atomic.Int64
	pending map[int64]chan *Response
	mu      sync.Mutex
	closed  chan struct{}

	// OnNotification is called for server notifications (no ID).
	OnNotification func(method string, params json.RawMessage)
}

// NewClient creates a JSON-RPC client from the given reader/writer pair.
// Typically stdin/stdout of a language server process.
func NewClient(r io.Reader, w io.Writer) *Client {
	c := &Client{
		writer:  w,
		reader:  bufio.NewReaderSize(r, 64*1024),
		pending: make(map[int64]chan *Response),
		closed:  make(chan struct{}),
	}
	return c
}

// StartReadLoop starts the goroutine that reads JSON-RPC messages.
// Must be called once after creating the client.
func (c *Client) StartReadLoop() {
	go c.readLoop()
}

// Close signals the read loop to stop.
func (c *Client) Close() {
	select {
	case <-c.closed:
	default:
		close(c.closed)
	}
}

// Call sends a JSON-RPC request and returns a channel for the response.
func (c *Client) Call(method string, params any) (int64, <-chan *Response) {
	id := c.nextID.Add(1)
	ch := make(chan *Response, 1)

	c.mu.Lock()
	c.pending[id] = ch
	c.mu.Unlock()

	msg := struct {
		JSONRPC string `json:"jsonrpc"`
		ID      int64  `json:"id"`
		Method  string `json:"method"`
		Params  any    `json:"params,omitempty"`
	}{
		JSONRPC: "2.0",
		ID:      id,
		Method:  method,
		Params:  params,
	}

	if err := c.writeMessage(msg); err != nil {
		c.mu.Lock()
		delete(c.pending, id)
		c.mu.Unlock()
		ch <- &Response{Error: &JSONRPCError{Code: -1, Message: err.Error()}}
	}

	return id, ch
}

// Notify sends a JSON-RPC notification (no response expected).
func (c *Client) Notify(method string, params any) error {
	msg := struct {
		JSONRPC string `json:"jsonrpc"`
		Method  string `json:"method"`
		Params  any    `json:"params,omitempty"`
	}{
		JSONRPC: "2.0",
		Method:  method,
		Params:  params,
	}
	return c.writeMessage(msg)
}

// writeMessage serializes a message and writes it with Content-Length header.
func (c *Client) writeMessage(msg any) error {
	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("marshal: %w", err)
	}

	header := fmt.Sprintf("Content-Length: %d\r\n\r\n", len(data))
	_, err = fmt.Fprint(c.writer, header)
	if err != nil {
		return fmt.Errorf("write header: %w", err)
	}
	_, err = c.writer.Write(data)
	if err != nil {
		return fmt.Errorf("write body: %w", err)
	}
	return nil
}

// readLoop reads JSON-RPC messages from the server and dispatches them.
func (c *Client) readLoop() {
	for {
		select {
		case <-c.closed:
			return
		default:
		}

		msg, err := c.readMessage()
		if err != nil {
			// Connection closed or error — cancel all pending.
			c.mu.Lock()
			for id, ch := range c.pending {
				ch <- &Response{Error: &JSONRPCError{Code: -1, Message: "connection closed"}}
				delete(c.pending, id)
			}
			c.mu.Unlock()
			return
		}

		c.dispatchMessage(msg)
	}
}

// readMessage reads a single JSON-RPC message with Content-Length framing.
func (c *Client) readMessage() (*JSONRPCMessage, error) {
	// Read headers until empty line.
	contentLength := -1
	for {
		line, err := c.reader.ReadString('\n')
		if err != nil {
			return nil, fmt.Errorf("read header: %w", err)
		}
		line = strings.TrimSpace(line)
		if line == "" {
			break // End of headers.
		}
		if strings.HasPrefix(line, "Content-Length:") {
			val := strings.TrimSpace(strings.TrimPrefix(line, "Content-Length:"))
			n, err := strconv.Atoi(val)
			if err != nil {
				return nil, fmt.Errorf("invalid Content-Length: %w", err)
			}
			contentLength = n
		}
	}

	if contentLength < 0 {
		return nil, fmt.Errorf("missing Content-Length header")
	}

	// Read body.
	body := make([]byte, contentLength)
	_, err := io.ReadFull(c.reader, body)
	if err != nil {
		return nil, fmt.Errorf("read body: %w", err)
	}

	var msg JSONRPCMessage
	if err := json.Unmarshal(body, &msg); err != nil {
		return nil, fmt.Errorf("unmarshal: %w", err)
	}
	return &msg, nil
}

// dispatchMessage routes a message to the appropriate handler.
func (c *Client) dispatchMessage(msg *JSONRPCMessage) {
	// If it has an ID, it's a response to one of our requests.
	if msg.ID != nil && msg.Method == "" {
		var id int64
		if err := json.Unmarshal(msg.ID, &id); err != nil {
			return
		}

		c.mu.Lock()
		ch, ok := c.pending[id]
		if ok {
			delete(c.pending, id)
		}
		c.mu.Unlock()

		if ok {
			ch <- &Response{Result: msg.Result, Error: msg.Error}
		}
		return
	}

	// Otherwise it's a notification from the server.
	if c.OnNotification != nil && msg.Method != "" {
		c.OnNotification(msg.Method, msg.Params)
	}
}
