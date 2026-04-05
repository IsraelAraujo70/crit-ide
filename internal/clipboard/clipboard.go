package clipboard

import atotto "github.com/atotto/clipboard"

// Provider abstracts clipboard read/write for testability.
type Provider interface {
	Read() (string, error)
	Write(text string) error
}

// SystemClipboard uses the OS clipboard via atotto/clipboard.
type SystemClipboard struct{}

// Read returns the current clipboard content.
func (c *SystemClipboard) Read() (string, error) { return atotto.ReadAll() }

// Write sets the clipboard content.
func (c *SystemClipboard) Write(text string) error { return atotto.WriteAll(text) }
