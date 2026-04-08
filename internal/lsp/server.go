package lsp

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"

	"github.com/israelcorrea/crit-ide/internal/events"
	"github.com/israelcorrea/crit-ide/internal/logger"
)

// Server manages a single language server process.
type Server struct {
	langID       string
	command      string
	args         []string
	cmd          *exec.Cmd
	client       *Client
	state        ServerState
	capabilities ServerCapabilities
	bus          *events.Bus
	rootURI      DocumentURI
	rootPath     string
	docVersions  map[DocumentURI]int
}

// NewServer creates a new server instance (not yet started).
func NewServer(langID, command string, args []string, bus *events.Bus, rootPath string) *Server {
	return &Server{
		langID:      langID,
		command:     command,
		args:        args,
		bus:         bus,
		rootPath:    rootPath,
		rootURI:     URIFromPath(rootPath),
		state:       StateStopped,
		docVersions: make(map[DocumentURI]int),
	}
}

// State returns the current server state.
func (s *Server) State() ServerState {
	return s.state
}

// Start spawns the language server process and initializes the JSON-RPC client.
func (s *Server) Start() error {
	s.state = StateStarting
	s.notifyState("")

	s.cmd = exec.Command(s.command, s.args...)
	// Redirect server stderr to the debug log file when --debug is active,
	// otherwise discard it to prevent corrupting the terminal UI.
	if logger.Enabled() {
		s.cmd.Stderr = logger.Writer()
	}

	logger.Info("lsp: starting %s %v (root=%s)", s.command, s.args, s.rootPath)

	stdin, err := s.cmd.StdinPipe()
	if err != nil {
		s.state = StateCrashed
		s.notifyState(err.Error())
		return fmt.Errorf("stdin pipe: %w", err)
	}

	stdout, err := s.cmd.StdoutPipe()
	if err != nil {
		s.state = StateCrashed
		s.notifyState(err.Error())
		return fmt.Errorf("stdout pipe: %w", err)
	}

	if err := s.cmd.Start(); err != nil {
		s.state = StateCrashed
		s.notifyState(err.Error())
		logger.Error("lsp: failed to start %s: %v", s.command, err)
		return fmt.Errorf("start server: %w", err)
	}

	logger.Info("lsp: process started (pid=%d)", s.cmd.Process.Pid)

	s.client = NewClient(stdout, stdin)
	s.client.OnNotification = s.handleNotification
	s.client.StartReadLoop()

	return nil
}

// Initialize sends the LSP initialize and initialized handshake.
func (s *Server) Initialize() error {
	params := InitializeParams{
		ProcessID: os.Getpid(),
		RootURI:   s.rootURI,
		Capabilities: ClientCapabilities{
			TextDocument: TextDocumentClientCapabilities{
				Synchronization: &TextDocumentSyncClientCapabilities{
					DidSave: true,
				},
				Hover: &HoverClientCapabilities{
					ContentFormat: []string{"plaintext", "markdown"},
				},
				Completion: &CompletionClientCapabilities{
					DynamicRegistration: false,
				},
				PublishDiagnostics: &PublishDiagnosticsClientCapabilities{
					VersionSupport: true,
				},
			},
		},
	}

	_, ch := s.client.Call("initialize", params)
	resp := <-ch
	if resp.Error != nil {
		s.state = StateCrashed
		s.notifyState(resp.Error.Error())
		return resp.Error
	}

	var result InitializeResult
	if err := json.Unmarshal(resp.Result, &result); err != nil {
		s.state = StateCrashed
		s.notifyState(err.Error())
		return fmt.Errorf("unmarshal initialize result: %w", err)
	}
	s.capabilities = result.Capabilities

	// Send initialized notification.
	if err := s.client.Notify("initialized", struct{}{}); err != nil {
		return fmt.Errorf("initialized notify: %w", err)
	}

	s.state = StateReady
	s.notifyState("")
	logger.Info("lsp: %s initialized (hover=%v, definition=%v, format=%v, completion=%v, rename=%v, codeAction=%v, sigHelp=%v)",
		s.langID, s.capabilities.HoverProvider, s.capabilities.DefinitionProvider,
		s.capabilities.DocumentFormattingProvider, s.capabilities.CompletionProvider != nil,
		s.capabilities.RenameProvider != nil, s.capabilities.CodeActionProvider != nil,
		s.capabilities.SignatureHelpProvider != nil)
	return nil
}

// Stop gracefully shuts down the language server.
func (s *Server) Stop() error {
	if s.client == nil {
		return nil
	}

	// Send shutdown request.
	_, ch := s.client.Call("shutdown", nil)
	<-ch // Wait for response (ignore errors).

	// Send exit notification.
	_ = s.client.Notify("exit", nil)
	s.client.Close()

	if s.cmd != nil && s.cmd.Process != nil {
		_ = s.cmd.Process.Kill()
		_ = s.cmd.Wait()
	}

	s.state = StateStopped
	s.notifyState("")
	logger.Info("lsp: %s stopped", s.langID)
	return nil
}

// DidOpen notifies the server that a document was opened.
func (s *Server) DidOpen(uri DocumentURI, langID, content string) {
	if s.state != StateReady {
		return
	}
	s.docVersions[uri] = 1
	_ = s.client.Notify("textDocument/didOpen", DidOpenTextDocumentParams{
		TextDocument: TextDocumentItem{
			URI:        uri,
			LanguageID: langID,
			Version:    1,
			Text:       content,
		},
	})
	logger.Debug("lsp: didOpen %s", uri)
}

// DidChange notifies the server of a document change (full sync).
func (s *Server) DidChange(uri DocumentURI, content string) {
	if s.state != StateReady {
		return
	}
	ver := s.docVersions[uri] + 1
	s.docVersions[uri] = ver
	_ = s.client.Notify("textDocument/didChange", DidChangeTextDocumentParams{
		TextDocument: VersionedTextDocumentIdentifier{URI: uri, Version: ver},
		ContentChanges: []TextDocumentContentChangeEvent{
			{Text: content}, // Full sync: no range, just full content.
		},
	})
}

// DidSave notifies the server that a document was saved.
func (s *Server) DidSave(uri DocumentURI) {
	if s.state != StateReady {
		return
	}
	_ = s.client.Notify("textDocument/didSave", DidSaveTextDocumentParams{
		TextDocument: TextDocumentIdentifier{URI: uri},
	})
}

// DidClose notifies the server that a document was closed.
func (s *Server) DidClose(uri DocumentURI) {
	if s.state != StateReady {
		return
	}
	_ = s.client.Notify("textDocument/didClose", DidCloseTextDocumentParams{
		TextDocument: TextDocumentIdentifier{URI: uri},
	})
	delete(s.docVersions, uri)
}

// Definition requests go-to-definition. Result arrives as EventLSPDefinition.
func (s *Server) Definition(uri DocumentURI, pos Position) {
	if s.state != StateReady || !s.capabilities.DefinitionProvider {
		return
	}
	_, ch := s.client.Call("textDocument/definition", TextDocumentPositionParams{
		TextDocument: TextDocumentIdentifier{URI: uri},
		Position:     pos,
	})
	go func() {
		resp := <-ch
		if resp.Error != nil {
			return
		}
		var locations []Location
		// Definition can return Location, []Location, or LocationLink[].
		// Try []Location first.
		if err := json.Unmarshal(resp.Result, &locations); err != nil {
			// Try single Location.
			var loc Location
			if err2 := json.Unmarshal(resp.Result, &loc); err2 == nil {
				locations = []Location{loc}
			}
		}
		s.bus.Send(events.Event{
			Type:    events.EventLSPDefinition,
			Payload: &DefinitionPayload{Locations: locations},
		})
	}()
}

// HoverRequest requests hover information. Result arrives as EventLSPHover.
func (s *Server) HoverRequest(uri DocumentURI, pos Position) {
	if s.state != StateReady || !s.capabilities.HoverProvider {
		return
	}
	_, ch := s.client.Call("textDocument/hover", TextDocumentPositionParams{
		TextDocument: TextDocumentIdentifier{URI: uri},
		Position:     pos,
	})
	go func() {
		resp := <-ch
		if resp.Error != nil {
			return
		}
		if resp.Result == nil || string(resp.Result) == "null" {
			return
		}
		var hover Hover
		if err := json.Unmarshal(resp.Result, &hover); err != nil {
			return
		}
		s.bus.Send(events.Event{
			Type:    events.EventLSPHover,
			Payload: &HoverPayload{Contents: hover.Contents, Range: hover.Range},
		})
	}()
}

// Format requests document formatting. Result arrives as EventLSPFormat.
func (s *Server) Format(uri DocumentURI) {
	if s.state != StateReady || !s.capabilities.DocumentFormattingProvider {
		return
	}
	_, ch := s.client.Call("textDocument/formatting", DocumentFormattingParams{
		TextDocument: TextDocumentIdentifier{URI: uri},
		Options:      FormattingOptions{TabSize: 4, InsertSpaces: false},
	})
	go func() {
		resp := <-ch
		if resp.Error != nil {
			return
		}
		var edits []TextEdit
		if err := json.Unmarshal(resp.Result, &edits); err != nil {
			return
		}
		s.bus.Send(events.Event{
			Type:    events.EventLSPFormat,
			Payload: &FormatPayload{URI: uri, Edits: edits},
		})
	}()
}

// Completion requests completion items. Result arrives as EventLSPCompletion.
func (s *Server) Completion(uri DocumentURI, pos Position, triggerKind CompletionTriggerKind, triggerChar string) {
	if s.state != StateReady || s.capabilities.CompletionProvider == nil {
		return
	}
	params := CompletionParams{
		TextDocument: TextDocumentIdentifier{URI: uri},
		Position:     pos,
	}
	if triggerKind != 0 {
		params.Context = &CompletionContext{
			TriggerKind:      triggerKind,
			TriggerCharacter: triggerChar,
		}
	}
	_, ch := s.client.Call("textDocument/completion", params)
	go func() {
		resp := <-ch
		if resp.Error != nil {
			logger.Debug("lsp: completion error: %v", resp.Error)
			return
		}
		if resp.Result == nil || string(resp.Result) == "null" {
			return
		}
		// Completion can return CompletionList or []CompletionItem.
		var list CompletionList
		if err := json.Unmarshal(resp.Result, &list); err != nil {
			// Try []CompletionItem directly.
			var items []CompletionItem
			if err2 := json.Unmarshal(resp.Result, &items); err2 != nil {
				logger.Debug("lsp: unmarshal completion: %v / %v", err, err2)
				return
			}
			list.Items = items
		}
		s.bus.Send(events.Event{
			Type:    events.EventLSPCompletion,
			Payload: &CompletionPayload{Items: list.Items, IsIncomplete: list.IsIncomplete},
		})
	}()
}

// RequestRename requests a rename operation. Result arrives as EventLSPRename.
func (s *Server) RequestRename(uri DocumentURI, pos Position, newName string) {
	if s.state != StateReady || s.capabilities.RenameProvider == nil {
		return
	}
	_, ch := s.client.Call("textDocument/rename", RenameParams{
		TextDocument: TextDocumentIdentifier{URI: uri},
		Position:     pos,
		NewName:      newName,
	})
	go func() {
		resp := <-ch
		if resp.Error != nil {
			logger.Debug("lsp: rename error: %v", resp.Error)
			return
		}
		if resp.Result == nil || string(resp.Result) == "null" {
			return
		}
		var edit WorkspaceEdit
		if err := json.Unmarshal(resp.Result, &edit); err != nil {
			logger.Debug("lsp: unmarshal rename: %v", err)
			return
		}
		s.bus.Send(events.Event{
			Type:    events.EventLSPRename,
			Payload: &RenamePayload{Edit: &edit},
		})
	}()
}

// RequestCodeAction requests code actions. Result arrives as EventLSPCodeAction.
func (s *Server) RequestCodeAction(uri DocumentURI, rng Range, diagnostics []Diagnostic) {
	if s.state != StateReady || s.capabilities.CodeActionProvider == nil {
		return
	}
	_, ch := s.client.Call("textDocument/codeAction", CodeActionParams{
		TextDocument: TextDocumentIdentifier{URI: uri},
		Range:        rng,
		Context:      CodeActionContext{Diagnostics: diagnostics},
	})
	go func() {
		resp := <-ch
		if resp.Error != nil {
			logger.Debug("lsp: codeAction error: %v", resp.Error)
			return
		}
		if resp.Result == nil || string(resp.Result) == "null" {
			return
		}
		var actions []CodeAction
		if err := json.Unmarshal(resp.Result, &actions); err != nil {
			logger.Debug("lsp: unmarshal codeAction: %v", err)
			return
		}
		s.bus.Send(events.Event{
			Type:    events.EventLSPCodeAction,
			Payload: &CodeActionPayload{Actions: actions},
		})
	}()
}

// RequestSignatureHelp requests signature help. Result arrives as EventLSPSignatureHelp.
func (s *Server) RequestSignatureHelp(uri DocumentURI, pos Position) {
	if s.state != StateReady || s.capabilities.SignatureHelpProvider == nil {
		return
	}
	_, ch := s.client.Call("textDocument/signatureHelp", SignatureHelpParams{
		TextDocument: TextDocumentIdentifier{URI: uri},
		Position:     pos,
	})
	go func() {
		resp := <-ch
		if resp.Error != nil {
			logger.Debug("lsp: signatureHelp error: %v", resp.Error)
			return
		}
		if resp.Result == nil || string(resp.Result) == "null" {
			return
		}
		var help SignatureHelp
		if err := json.Unmarshal(resp.Result, &help); err != nil {
			logger.Debug("lsp: unmarshal signatureHelp: %v", err)
			return
		}
		if len(help.Signatures) == 0 {
			return
		}
		s.bus.Send(events.Event{
			Type: events.EventLSPSignatureHelp,
			Payload: &SignatureHelpPayload{
				Signatures:      help.Signatures,
				ActiveSignature: help.ActiveSignature,
				ActiveParameter: help.ActiveParameter,
			},
		})
	}()
}

// SignatureHelpTriggerCharacters returns the server's configured signature help trigger characters.
func (s *Server) SignatureHelpTriggerCharacters() []string {
	if s.capabilities.SignatureHelpProvider == nil {
		return nil
	}
	return s.capabilities.SignatureHelpProvider.TriggerCharacters
}

// HasRenameProvider returns true if the server supports rename.
func (s *Server) HasRenameProvider() bool {
	return s.capabilities.RenameProvider != nil
}

// HasCodeActionProvider returns true if the server supports code actions.
func (s *Server) HasCodeActionProvider() bool {
	return s.capabilities.CodeActionProvider != nil
}

// TriggerCharacters returns the server's configured completion trigger characters.
func (s *Server) TriggerCharacters() []string {
	if s.capabilities.CompletionProvider == nil {
		return nil
	}
	return s.capabilities.CompletionProvider.TriggerCharacters
}

// handleNotification processes server notifications.
func (s *Server) handleNotification(method string, params json.RawMessage) {
	logger.Debug("lsp: notification %s", method)
	switch method {
	case "textDocument/publishDiagnostics":
		var p PublishDiagnosticsParams
		if err := json.Unmarshal(params, &p); err != nil {
			logger.Error("lsp: unmarshal diagnostics: %v", err)
			return
		}
		logger.Debug("lsp: %d diagnostics for %s", len(p.Diagnostics), p.URI)
		s.bus.Send(events.Event{
			Type:    events.EventLSPDiagnostics,
			Payload: &DiagnosticsPayload{URI: p.URI, Diagnostics: p.Diagnostics},
		})
	}
}

// notifyState sends a server state change event to the bus.
func (s *Server) notifyState(errMsg string) {
	s.bus.Send(events.Event{
		Type: events.EventLSPServerState,
		Payload: &ServerStatePayload{
			LangID: s.langID,
			State:  s.state,
			Error:  errMsg,
		},
	})
}
