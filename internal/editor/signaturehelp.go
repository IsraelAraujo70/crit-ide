package editor

// SignatureHelpState holds the state of the signature help popup.
type SignatureHelpState struct {
	Label           string // Full signature label (e.g., "func Foo(a int, b string) error").
	Parameters      []SignatureParam
	ActiveParameter int
	CursorRow       int // Editor row where popup was triggered.
	CursorCol       int // Editor col where popup was triggered.
}

// SignatureParam represents a single parameter in a signature.
type SignatureParam struct {
	Label string // Parameter label (e.g., "a int").
	Start int    // Start offset in signature label.
	End   int    // End offset in signature label.
}
