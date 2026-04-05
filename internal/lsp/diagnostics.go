package lsp

// DiagnosticsStore aggregates diagnostics from all language servers.
type DiagnosticsStore struct {
	byURI map[DocumentURI][]Diagnostic
}

// NewDiagnosticsStore creates an empty diagnostics store.
func NewDiagnosticsStore() *DiagnosticsStore {
	return &DiagnosticsStore{
		byURI: make(map[DocumentURI][]Diagnostic),
	}
}

// Update replaces all diagnostics for the given URI.
func (ds *DiagnosticsStore) Update(uri DocumentURI, diags []Diagnostic) {
	if len(diags) == 0 {
		delete(ds.byURI, uri)
	} else {
		ds.byURI[uri] = diags
	}
}

// ForURI returns diagnostics for the given URI.
func (ds *DiagnosticsStore) ForURI(uri DocumentURI) []Diagnostic {
	return ds.byURI[uri]
}

// CountsByURI returns the error and warning counts for a URI.
func (ds *DiagnosticsStore) CountsByURI(uri DocumentURI) (errors, warnings int) {
	for _, d := range ds.byURI[uri] {
		switch d.Severity {
		case SeverityError:
			errors++
		case SeverityWarning:
			warnings++
		}
	}
	return
}

// Clear removes all diagnostics.
func (ds *DiagnosticsStore) Clear() {
	ds.byURI = make(map[DocumentURI][]Diagnostic)
}
