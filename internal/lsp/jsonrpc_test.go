package lsp

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"testing"
	"time"
)

func TestWriteMessage(t *testing.T) {
	var buf bytes.Buffer
	c := NewClient(strings.NewReader(""), &buf)

	err := c.writeMessage(map[string]string{"hello": "world"})
	if err != nil {
		t.Fatalf("writeMessage: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "Content-Length:") {
		t.Error("missing Content-Length header")
	}
	if !strings.Contains(output, `"hello":"world"`) {
		t.Errorf("unexpected body: %s", output)
	}
}

func TestReadMessage(t *testing.T) {
	body := `{"jsonrpc":"2.0","id":1,"result":{"ok":true}}`
	input := fmt.Sprintf("Content-Length: %d\r\n\r\n%s", len(body), body)

	c := NewClient(strings.NewReader(input), io.Discard)
	msg, err := c.readMessage()
	if err != nil {
		t.Fatalf("readMessage: %v", err)
	}

	var id int64
	if err := json.Unmarshal(msg.ID, &id); err != nil {
		t.Fatalf("unmarshal ID: %v", err)
	}
	if id != 1 {
		t.Errorf("expected ID 1, got %d", id)
	}
}

func TestCallAndResponse(t *testing.T) {
	// Create a pipe to simulate server stdin/stdout.
	serverReader, clientWriter := io.Pipe()
	clientReader, serverWriter := io.Pipe()

	c := NewClient(clientReader, clientWriter)
	c.StartReadLoop()
	defer c.Close()

	// Simulate server: read request, send response.
	go func() {
		tempClient := NewClient(serverReader, serverWriter)
		msg, err := tempClient.readMessage()
		if err != nil {
			return
		}
		// Send response with the same ID.
		resp := fmt.Sprintf(`{"jsonrpc":"2.0","id":%s,"result":{"status":"ok"}}`, string(msg.ID))
		header := fmt.Sprintf("Content-Length: %d\r\n\r\n%s", len(resp), resp)
		serverWriter.Write([]byte(header))
	}()

	_, ch := c.Call("test/method", map[string]string{"key": "value"})

	select {
	case resp := <-ch:
		if resp.Error != nil {
			t.Fatalf("unexpected error: %v", resp.Error)
		}
		var result map[string]string
		if err := json.Unmarshal(resp.Result, &result); err != nil {
			t.Fatalf("unmarshal result: %v", err)
		}
		if result["status"] != "ok" {
			t.Errorf("expected status ok, got %s", result["status"])
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for response")
	}
}

func TestNotification(t *testing.T) {
	var buf bytes.Buffer
	c := NewClient(strings.NewReader(""), &buf)

	err := c.Notify("test/notify", map[string]int{"x": 1})
	if err != nil {
		t.Fatalf("Notify: %v", err)
	}

	output := buf.String()
	if strings.Contains(output, `"id"`) {
		t.Error("notification should not have an ID field")
	}
	if !strings.Contains(output, `"method":"test/notify"`) {
		t.Errorf("unexpected output: %s", output)
	}
}

func TestServerNotification(t *testing.T) {
	// Build a server notification message.
	body := `{"jsonrpc":"2.0","method":"textDocument/publishDiagnostics","params":{"uri":"file:///test.go"}}`
	input := fmt.Sprintf("Content-Length: %d\r\n\r\n%s", len(body), body)

	notifCh := make(chan string, 1)
	c := NewClient(strings.NewReader(input), io.Discard)
	c.OnNotification = func(method string, params json.RawMessage) {
		notifCh <- method
	}
	c.StartReadLoop()
	defer c.Close()

	select {
	case method := <-notifCh:
		if method != "textDocument/publishDiagnostics" {
			t.Errorf("expected publishDiagnostics, got %s", method)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for notification")
	}
}

func TestByteOffsetToUTF16_ASCII(t *testing.T) {
	s := "hello world"
	got := byteOffsetToUTF16(s, 5)
	if got != 5 {
		t.Errorf("expected 5, got %d", got)
	}
}

func TestByteOffsetToUTF16_Emoji(t *testing.T) {
	// 😀 is U+1F600, which is 4 bytes in UTF-8 and 2 UTF-16 code units (surrogate pair).
	s := "a😀b"
	// byte offsets: a=0, 😀=1..4, b=5
	// UTF-16 offsets: a=0, 😀=1..2, b=3

	// Byte offset 1 (start of emoji) -> UTF-16 offset 1.
	got := byteOffsetToUTF16(s, 1)
	if got != 1 {
		t.Errorf("expected 1 for byte offset 1, got %d", got)
	}

	// Byte offset 5 (start of 'b') -> UTF-16 offset 3.
	got = byteOffsetToUTF16(s, 5)
	if got != 3 {
		t.Errorf("expected 3 for byte offset 5, got %d", got)
	}
}

func TestUTF16ToByteOffset_ASCII(t *testing.T) {
	s := "hello"
	got := utf16ToByteOffset(s, 3)
	if got != 3 {
		t.Errorf("expected 3, got %d", got)
	}
}

func TestUTF16ToByteOffset_Emoji(t *testing.T) {
	s := "a😀b"
	// UTF-16 offset 3 (start of 'b') -> byte offset 5.
	got := utf16ToByteOffset(s, 3)
	if got != 5 {
		t.Errorf("expected 5 for UTF-16 offset 3, got %d", got)
	}
}

func TestURIRoundtrip(t *testing.T) {
	path := "/tmp/test/file.go"
	uri := URIFromPath(path)
	got, err := PathFromURI(uri)
	if err != nil {
		t.Fatalf("PathFromURI: %v", err)
	}
	if got != path {
		t.Errorf("expected %q, got %q", path, got)
	}
}
