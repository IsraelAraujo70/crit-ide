package lsp

import (
	"fmt"
	"net/url"
	"path/filepath"
	"runtime"
	"strings"
	"unicode/utf16"
	"unicode/utf8"
)

// EditorToLSPPosition converts an editor byte-offset position to an LSP position.
// LSP uses 0-based line and UTF-16 character offset.
func EditorToLSPPosition(line, byteCol int, lineContent string) Position {
	utf16Col := byteOffsetToUTF16(lineContent, byteCol)
	return Position{Line: line, Character: utf16Col}
}

// LSPToEditorPosition converts an LSP position to an editor byte offset.
func LSPToEditorPosition(pos Position, lineContent string) (line, byteCol int) {
	return pos.Line, utf16ToByteOffset(lineContent, pos.Character)
}

// byteOffsetToUTF16 converts a byte offset within a string to a UTF-16 code unit count.
func byteOffsetToUTF16(s string, byteOff int) int {
	if byteOff <= 0 {
		return 0
	}
	if byteOff >= len(s) {
		byteOff = len(s)
	}

	utf16Units := 0
	i := 0
	for i < byteOff {
		r, size := utf8.DecodeRuneInString(s[i:])
		if r == utf8.RuneError && size <= 1 {
			// Invalid UTF-8 byte.
			utf16Units++
			i++
			continue
		}
		// Count UTF-16 code units for this rune.
		if r >= 0x10000 {
			utf16Units += 2 // Surrogate pair.
		} else {
			utf16Units++
		}
		i += size
	}
	return utf16Units
}

// utf16ToByteOffset converts a UTF-16 code unit offset to a byte offset.
func utf16ToByteOffset(s string, utf16Off int) int {
	if utf16Off <= 0 {
		return 0
	}

	units := 0
	i := 0
	for i < len(s) && units < utf16Off {
		r, size := utf8.DecodeRuneInString(s[i:])
		if r == utf8.RuneError && size <= 1 {
			units++
			i++
			continue
		}
		if r >= 0x10000 {
			units += 2
		} else {
			units++
		}
		i += size
	}
	return i
}

// UTF16Len returns the number of UTF-16 code units in a rune.
func UTF16Len(r rune) int {
	if r >= 0x10000 {
		return len(utf16.Encode([]rune{r}))
	}
	return 1
}

// URIFromPath converts a file path to a DocumentURI.
func URIFromPath(path string) DocumentURI {
	absPath, err := filepath.Abs(path)
	if err != nil {
		absPath = path
	}
	// Normalize path separators.
	absPath = filepath.ToSlash(absPath)
	if runtime.GOOS == "windows" {
		// Windows paths need a leading / for file URI.
		if !strings.HasPrefix(absPath, "/") {
			absPath = "/" + absPath
		}
	}
	return DocumentURI("file://" + absPath)
}

// PathFromURI converts a DocumentURI to a file path.
func PathFromURI(uri DocumentURI) (string, error) {
	u, err := url.Parse(string(uri))
	if err != nil {
		return "", fmt.Errorf("parse URI: %w", err)
	}
	if u.Scheme != "file" {
		return "", fmt.Errorf("unsupported URI scheme: %s", u.Scheme)
	}
	path := u.Path
	if runtime.GOOS == "windows" && len(path) > 0 && path[0] == '/' {
		path = path[1:] // Remove leading / on Windows.
	}
	return filepath.FromSlash(path), nil
}
