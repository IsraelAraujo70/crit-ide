package lsp

import (
	"encoding/json"
	"testing"
)

func TestWorkspaceEditMarshal(t *testing.T) {
	edit := WorkspaceEdit{
		Changes: map[DocumentURI][]TextEdit{
			"file:///tmp/foo.go": {
				{
					Range:   Range{Start: Position{Line: 1, Character: 0}, End: Position{Line: 1, Character: 5}},
					NewText: "newName",
				},
			},
		},
	}
	data, err := json.Marshal(edit)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var got WorkspaceEdit
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	edits, ok := got.Changes["file:///tmp/foo.go"]
	if !ok {
		t.Fatal("expected edits for file:///tmp/foo.go")
	}
	if len(edits) != 1 {
		t.Fatalf("expected 1 edit, got %d", len(edits))
	}
	if edits[0].NewText != "newName" {
		t.Fatalf("expected 'newName', got %q", edits[0].NewText)
	}
}

func TestCodeActionMarshal(t *testing.T) {
	action := CodeAction{
		Title: "Organize Imports",
		Kind:  CodeActionSourceOrganizeImports,
		Edit: &WorkspaceEdit{
			Changes: map[DocumentURI][]TextEdit{
				"file:///tmp/foo.go": {
					{
						Range:   Range{Start: Position{Line: 0, Character: 0}, End: Position{Line: 2, Character: 0}},
						NewText: "import \"fmt\"\n",
					},
				},
			},
		},
	}
	data, err := json.Marshal(action)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var got CodeAction
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if got.Title != "Organize Imports" {
		t.Fatalf("expected 'Organize Imports', got %q", got.Title)
	}
	if got.Kind != CodeActionSourceOrganizeImports {
		t.Fatalf("expected source.organizeImports, got %q", got.Kind)
	}
	if got.Edit == nil {
		t.Fatal("expected non-nil edit")
	}
}

func TestSignatureHelpMarshal(t *testing.T) {
	help := SignatureHelp{
		Signatures: []SignatureInformation{
			{
				Label: "func Foo(a int, b string) error",
				Parameters: []ParameterInformation{
					{Label: json.RawMessage(`"a int"`)},
					{Label: json.RawMessage(`[16, 24]`)},
				},
			},
		},
		ActiveSignature: 0,
		ActiveParameter: 1,
	}
	data, err := json.Marshal(help)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var got SignatureHelp
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(got.Signatures) != 1 {
		t.Fatalf("expected 1 signature, got %d", len(got.Signatures))
	}
	if got.ActiveParameter != 1 {
		t.Fatalf("expected active param 1, got %d", got.ActiveParameter)
	}
}

func TestRenameParamsMarshal(t *testing.T) {
	params := RenameParams{
		TextDocument: TextDocumentIdentifier{URI: "file:///tmp/foo.go"},
		Position:     Position{Line: 5, Character: 10},
		NewName:      "newIdentifier",
	}
	data, err := json.Marshal(params)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var got RenameParams
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if got.NewName != "newIdentifier" {
		t.Fatalf("expected 'newIdentifier', got %q", got.NewName)
	}
	if got.Position.Line != 5 || got.Position.Character != 10 {
		t.Fatalf("unexpected position: %+v", got.Position)
	}
}
