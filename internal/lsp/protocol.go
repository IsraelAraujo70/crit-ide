package lsp

import "encoding/json"

// DocumentURI is a file URI (e.g., "file:///path/to/file.go").
type DocumentURI string

// Position in a text document (LSP uses 0-based line, 0-based UTF-16 character offset).
type Position struct {
	Line      int `json:"line"`
	Character int `json:"character"`
}

// Range in a text document.
type Range struct {
	Start Position `json:"start"`
	End   Position `json:"end"`
}

// Location represents a location in a document.
type Location struct {
	URI   DocumentURI `json:"uri"`
	Range Range       `json:"range"`
}

// TextDocumentIdentifier identifies a text document.
type TextDocumentIdentifier struct {
	URI DocumentURI `json:"uri"`
}

// VersionedTextDocumentIdentifier identifies a specific version of a text document.
type VersionedTextDocumentIdentifier struct {
	URI     DocumentURI `json:"uri"`
	Version int         `json:"version"`
}

// TextDocumentItem describes a text document.
type TextDocumentItem struct {
	URI        DocumentURI `json:"uri"`
	LanguageID string      `json:"languageId"`
	Version    int         `json:"version"`
	Text       string      `json:"text"`
}

// TextDocumentContentChangeEvent describes a content change.
type TextDocumentContentChangeEvent struct {
	Range *Range `json:"range,omitempty"`
	Text  string `json:"text"`
}

// TextDocumentPositionParams is used for hover, definition, etc.
type TextDocumentPositionParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
	Position     Position               `json:"position"`
}

// DiagnosticSeverity indicates the severity of a diagnostic.
type DiagnosticSeverity int

const (
	SeverityError   DiagnosticSeverity = 1
	SeverityWarning DiagnosticSeverity = 2
	SeverityInfo    DiagnosticSeverity = 3
	SeverityHint    DiagnosticSeverity = 4
)

// Diagnostic represents a diagnostic message.
type Diagnostic struct {
	Range    Range              `json:"range"`
	Severity DiagnosticSeverity `json:"severity,omitempty"`
	Code     json.RawMessage    `json:"code,omitempty"`
	Source   string             `json:"source,omitempty"`
	Message  string             `json:"message"`
}

// PublishDiagnosticsParams is the params for textDocument/publishDiagnostics.
type PublishDiagnosticsParams struct {
	URI         DocumentURI  `json:"uri"`
	Version     int          `json:"version,omitempty"`
	Diagnostics []Diagnostic `json:"diagnostics"`
}

// MarkupContent represents human-readable content.
type MarkupContent struct {
	Kind  string `json:"kind"`
	Value string `json:"value"`
}

// Hover is the result of a hover request.
type Hover struct {
	Contents MarkupContent `json:"contents"`
	Range    *Range        `json:"range,omitempty"`
}

// TextEdit represents a text edit operation.
type TextEdit struct {
	Range   Range  `json:"range"`
	NewText string `json:"newText"`
}

// --- Initialize ---

// InitializeParams is sent by the client to the server on startup.
type InitializeParams struct {
	ProcessID    int            `json:"processId"`
	RootURI      DocumentURI    `json:"rootUri"`
	Capabilities ClientCapabilities `json:"capabilities"`
}

// ClientCapabilities represents client capabilities (minimal for now).
type ClientCapabilities struct {
	TextDocument TextDocumentClientCapabilities `json:"textDocument,omitempty"`
}

// TextDocumentClientCapabilities represents text document capabilities.
type TextDocumentClientCapabilities struct {
	Synchronization    *TextDocumentSyncClientCapabilities   `json:"synchronization,omitempty"`
	Hover              *HoverClientCapabilities              `json:"hover,omitempty"`
	Completion         *CompletionClientCapabilities         `json:"completion,omitempty"`
	PublishDiagnostics *PublishDiagnosticsClientCapabilities  `json:"publishDiagnostics,omitempty"`
}

// TextDocumentSyncClientCapabilities represents sync capabilities.
type TextDocumentSyncClientCapabilities struct {
	DynamicRegistration bool `json:"dynamicRegistration,omitempty"`
	DidSave             bool `json:"didSave,omitempty"`
}

// HoverClientCapabilities represents hover capabilities.
type HoverClientCapabilities struct {
	ContentFormat []string `json:"contentFormat,omitempty"`
}

// PublishDiagnosticsClientCapabilities represents diagnostics capabilities.
type PublishDiagnosticsClientCapabilities struct {
	VersionSupport bool `json:"versionSupport,omitempty"`
}

// InitializeResult is the server's response to initialize.
type InitializeResult struct {
	Capabilities ServerCapabilities `json:"capabilities"`
}

// ServerCapabilities describes the server's capabilities.
type ServerCapabilities struct {
	TextDocumentSync           interface{}           `json:"textDocumentSync,omitempty"`
	HoverProvider              bool                  `json:"hoverProvider,omitempty"`
	DefinitionProvider         bool                  `json:"definitionProvider,omitempty"`
	CompletionProvider         *CompletionOptions    `json:"completionProvider,omitempty"`
	DocumentFormattingProvider bool                  `json:"documentFormattingProvider,omitempty"`
	RenameProvider             interface{}           `json:"renameProvider,omitempty"`  // bool or RenameOptions
	CodeActionProvider         interface{}           `json:"codeActionProvider,omitempty"` // bool or CodeActionOptions
	SignatureHelpProvider      *SignatureHelpOptions `json:"signatureHelpProvider,omitempty"`
}

// CompletionOptions represents completion server capabilities.
type CompletionOptions struct {
	TriggerCharacters []string `json:"triggerCharacters,omitempty"`
}

// CompletionItemKind identifies the kind of a completion item.
type CompletionItemKind int

const (
	CompletionKindText          CompletionItemKind = 1
	CompletionKindMethod        CompletionItemKind = 2
	CompletionKindFunction      CompletionItemKind = 3
	CompletionKindConstructor   CompletionItemKind = 4
	CompletionKindField         CompletionItemKind = 5
	CompletionKindVariable      CompletionItemKind = 6
	CompletionKindClass         CompletionItemKind = 7
	CompletionKindInterface     CompletionItemKind = 8
	CompletionKindModule        CompletionItemKind = 9
	CompletionKindProperty      CompletionItemKind = 10
	CompletionKindUnit          CompletionItemKind = 11
	CompletionKindValue         CompletionItemKind = 12
	CompletionKindEnum          CompletionItemKind = 13
	CompletionKindKeyword       CompletionItemKind = 14
	CompletionKindSnippet       CompletionItemKind = 15
	CompletionKindColor         CompletionItemKind = 16
	CompletionKindFile          CompletionItemKind = 17
	CompletionKindReference     CompletionItemKind = 18
	CompletionKindFolder        CompletionItemKind = 19
	CompletionKindEnumMember    CompletionItemKind = 20
	CompletionKindConstant      CompletionItemKind = 21
	CompletionKindStruct        CompletionItemKind = 22
	CompletionKindEvent         CompletionItemKind = 23
	CompletionKindOperator      CompletionItemKind = 24
	CompletionKindTypeParameter CompletionItemKind = 25
)

// CompletionItem represents a single completion suggestion.
type CompletionItem struct {
	Label         string             `json:"label"`
	Kind          CompletionItemKind `json:"kind,omitempty"`
	Detail        string             `json:"detail,omitempty"`
	Documentation json.RawMessage    `json:"documentation,omitempty"`
	InsertText    string             `json:"insertText,omitempty"`
	FilterText    string             `json:"filterText,omitempty"`
	SortText      string             `json:"sortText,omitempty"`
}

// CompletionList is the result of a completion request.
type CompletionList struct {
	IsIncomplete bool             `json:"isIncomplete"`
	Items        []CompletionItem `json:"items"`
}

// CompletionParams is sent for textDocument/completion.
type CompletionParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
	Position     Position               `json:"position"`
	Context      *CompletionContext     `json:"context,omitempty"`
}

// CompletionTriggerKind describes how a completion was triggered.
type CompletionTriggerKind int

const (
	CompletionTriggerInvoked                         CompletionTriggerKind = 1
	CompletionTriggerTriggerCharacter                CompletionTriggerKind = 2
	CompletionTriggerTriggerForIncompleteCompletions CompletionTriggerKind = 3
)

// CompletionContext contains additional information about the context in which
// a completion request was triggered.
type CompletionContext struct {
	TriggerKind      CompletionTriggerKind `json:"triggerKind"`
	TriggerCharacter string                `json:"triggerCharacter,omitempty"`
}

// CompletionClientCapabilities represents client completion capabilities.
type CompletionClientCapabilities struct {
	DynamicRegistration bool `json:"dynamicRegistration,omitempty"`
	CompletionItem      *struct {
		SnippetSupport bool `json:"snippetSupport,omitempty"`
	} `json:"completionItem,omitempty"`
}

// TextDocumentSyncKind defines how text documents are synced.
type TextDocumentSyncKind int

const (
	SyncNone        TextDocumentSyncKind = 0
	SyncFull        TextDocumentSyncKind = 1
	SyncIncremental TextDocumentSyncKind = 2
)

// TextDocumentSyncOptions describes text document sync options.
type TextDocumentSyncOptions struct {
	OpenClose bool                 `json:"openClose,omitempty"`
	Change    TextDocumentSyncKind `json:"change,omitempty"`
	Save      *SaveOptions         `json:"save,omitempty"`
}

// SaveOptions controls save notifications.
type SaveOptions struct {
	IncludeText bool `json:"includeText,omitempty"`
}

// --- Document operations params ---

// DidOpenTextDocumentParams is sent when a document is opened.
type DidOpenTextDocumentParams struct {
	TextDocument TextDocumentItem `json:"textDocument"`
}

// DidChangeTextDocumentParams is sent when a document changes.
type DidChangeTextDocumentParams struct {
	TextDocument   VersionedTextDocumentIdentifier  `json:"textDocument"`
	ContentChanges []TextDocumentContentChangeEvent `json:"contentChanges"`
}

// DidSaveTextDocumentParams is sent when a document is saved.
type DidSaveTextDocumentParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
	Text         *string                `json:"text,omitempty"`
}

// DidCloseTextDocumentParams is sent when a document is closed.
type DidCloseTextDocumentParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
}

// DocumentFormattingParams is sent for formatting requests.
type DocumentFormattingParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
	Options      FormattingOptions      `json:"options"`
}

// FormattingOptions describes formatting options.
type FormattingOptions struct {
	TabSize      int  `json:"tabSize"`
	InsertSpaces bool `json:"insertSpaces"`
}

// --- Rename ---

// RenameParams is sent for textDocument/rename.
type RenameParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
	Position     Position               `json:"position"`
	NewName      string                 `json:"newName"`
}

// WorkspaceEdit represents changes to multiple resources.
type WorkspaceEdit struct {
	Changes map[DocumentURI][]TextEdit `json:"changes,omitempty"`
}

// --- Code Action ---

// CodeActionParams is sent for textDocument/codeAction.
type CodeActionParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
	Range        Range                  `json:"range"`
	Context      CodeActionContext      `json:"context"`
}

// CodeActionContext carries additional information about the context.
type CodeActionContext struct {
	Diagnostics []Diagnostic `json:"diagnostics"`
}

// CodeActionKind categorizes a code action.
type CodeActionKind string

const (
	CodeActionQuickFix              CodeActionKind = "quickfix"
	CodeActionRefactor              CodeActionKind = "refactor"
	CodeActionRefactorExtract       CodeActionKind = "refactor.extract"
	CodeActionRefactorInline        CodeActionKind = "refactor.inline"
	CodeActionRefactorRewrite       CodeActionKind = "refactor.rewrite"
	CodeActionSource                CodeActionKind = "source"
	CodeActionSourceOrganizeImports CodeActionKind = "source.organizeImports"
)

// CodeAction represents a suggested edit or command.
type CodeAction struct {
	Title       string          `json:"title"`
	Kind        CodeActionKind  `json:"kind,omitempty"`
	Diagnostics []Diagnostic    `json:"diagnostics,omitempty"`
	Edit        *WorkspaceEdit  `json:"edit,omitempty"`
	Command     *Command        `json:"command,omitempty"`
}

// Command represents a reference to a command.
type Command struct {
	Title     string        `json:"title"`
	Command   string        `json:"command"`
	Arguments []interface{} `json:"arguments,omitempty"`
}

// CodeActionOptions describes code action server capabilities.
type CodeActionOptions struct {
	CodeActionKinds []CodeActionKind `json:"codeActionKinds,omitempty"`
}

// --- Signature Help ---

// SignatureHelpParams is sent for textDocument/signatureHelp.
type SignatureHelpParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
	Position     Position               `json:"position"`
}

// SignatureHelp represents the signature of a callable.
type SignatureHelp struct {
	Signatures      []SignatureInformation `json:"signatures"`
	ActiveSignature int                    `json:"activeSignature,omitempty"`
	ActiveParameter int                    `json:"activeParameter,omitempty"`
}

// SignatureInformation represents a single signature.
type SignatureInformation struct {
	Label         string                 `json:"label"`
	Documentation json.RawMessage        `json:"documentation,omitempty"`
	Parameters    []ParameterInformation `json:"parameters,omitempty"`
}

// ParameterInformation represents a single parameter.
type ParameterInformation struct {
	Label         json.RawMessage `json:"label"` // string or [int, int]
	Documentation json.RawMessage `json:"documentation,omitempty"`
}

// SignatureHelpOptions describes signature help server capabilities.
type SignatureHelpOptions struct {
	TriggerCharacters   []string `json:"triggerCharacters,omitempty"`
	RetriggerCharacters []string `json:"retriggerCharacters,omitempty"`
}
