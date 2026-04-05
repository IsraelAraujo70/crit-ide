package lsp

import (
	"fmt"

	"github.com/israelcorrea/crit-ide/internal/events"
)

// LangServerConfig describes how to start a language server.
type LangServerConfig struct {
	Command string
	Args    []string
}

// Manager manages language server instances per language.
type Manager struct {
	servers  map[string]*Server
	configs  map[string]LangServerConfig
	bus      *events.Bus
	rootPath string
}

// NewManager creates a new LSP manager with default server configurations.
func NewManager(bus *events.Bus, rootPath string) *Manager {
	m := &Manager{
		servers:  make(map[string]*Server),
		configs:  defaultServerConfigs(),
		bus:      bus,
		rootPath: rootPath,
	}
	return m
}

// defaultServerConfigs returns the built-in language server configurations.
func defaultServerConfigs() map[string]LangServerConfig {
	return map[string]LangServerConfig{
		"go":         {Command: "gopls", Args: []string{}},
		"python":     {Command: "pylsp", Args: []string{}},
		"rust":       {Command: "rust-analyzer", Args: []string{}},
		"javascript": {Command: "typescript-language-server", Args: []string{"--stdio"}},
		"typescript": {Command: "typescript-language-server", Args: []string{"--stdio"}},
		"c":          {Command: "clangd", Args: []string{}},
		"cpp":        {Command: "clangd", Args: []string{}},
		"css":        {Command: "vscode-css-language-server", Args: []string{"--stdio"}},
		"html":       {Command: "vscode-html-language-server", Args: []string{"--stdio"}},
		"json":       {Command: "vscode-json-language-server", Args: []string{"--stdio"}},
		"yaml":       {Command: "yaml-language-server", Args: []string{"--stdio"}},
		"toml":       {Command: "taplo", Args: []string{"lsp", "stdio"}},
	}
}

// EnsureServer starts a server for the given language if not already running.
// Returns the server, or an error if the language has no config or fails to start.
func (m *Manager) EnsureServer(langID string) (*Server, error) {
	if srv, ok := m.servers[langID]; ok && srv.State() == StateReady {
		return srv, nil
	}

	cfg, ok := m.configs[langID]
	if !ok {
		return nil, fmt.Errorf("no server configured for language: %s", langID)
	}

	srv := NewServer(langID, cfg.Command, cfg.Args, m.bus, m.rootPath)
	if err := srv.Start(); err != nil {
		return nil, fmt.Errorf("start %s server: %w", langID, err)
	}

	if err := srv.Initialize(); err != nil {
		_ = srv.Stop()
		return nil, fmt.Errorf("initialize %s server: %w", langID, err)
	}

	m.servers[langID] = srv
	return srv, nil
}

// ServerFor returns the running server for a language, or nil.
func (m *Manager) ServerFor(langID string) *Server {
	if srv, ok := m.servers[langID]; ok && srv.State() == StateReady {
		return srv
	}
	return nil
}

// StopAll shuts down all running servers.
func (m *Manager) StopAll() {
	for id, srv := range m.servers {
		_ = srv.Stop()
		delete(m.servers, id)
	}
}
