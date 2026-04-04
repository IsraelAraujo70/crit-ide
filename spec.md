# SPEC — IDE de Terminal Completa em Go

## 1. Visão geral

Construir uma IDE de terminal, escrita em Go, com foco em:

* experiência moderna de edição
* arquitetura extensível por plugins
* integração forte com IA local
* integração nativa com Git
* suporte completo a LSP
* suporte a mouse dentro do terminal
* keybindings totalmente configuráveis
* alta performance em arquivos grandes
* UX inspirada em Vim em flexibilidade e eficiência, **sem** modal editing obrigatório

O produto não deve ser um clone de Vim, Neovim, Helix ou Emacs. Ele deve herdar apenas os pontos fortes:

* terminal-first
* configuração poderosa
* automação por keybinds/comandos
* sistema de plugins
* composição de painéis, buffers e ações

Mas deve rejeitar explicitamente:

* normal mode / insert mode como conceito central
* dependência de macros obscuras para tarefas comuns
* UX hostil para mouse
* configuração excessivamente fragmentada

---

## 2. Objetivo do produto

Entregar uma IDE que rode inteiramente no terminal e seja útil como editor principal para desenvolvimento profissional, com suporte a:

* edição de código
* navegação de projeto
* busca global
* refatoração via LSP
* autocomplete com IA local
* execução de tarefas
* debugging futuro
* code review com diff viewer
* fluxo Git completo

A IDE deve permitir que um usuário:

1. abra um projeto grande
2. navegue rapidamente pelos arquivos
3. edite múltiplos buffers
4. veja erros e warnings em tempo real
5. faça autocomplete via LSP + IA local
6. use Git sem sair da IDE
7. personalize keybindings, layout e plugins

---

## 3. Princípios de design

### 3.1 Terminal-first

Toda a aplicação deve funcionar integralmente em TTY moderno.

### 3.2 Non-modal by default

A edição deve ser direta, como em IDEs gráficas e editores modernos.

### 3.3 Keyboard-first, mouse-enabled

Tudo deve funcionar por teclado, mas o mouse deve ser cidadão de primeira classe.

### 3.4 Extensibilidade real

O core precisa expor APIs claras para plugins, ações, painéis e fontes de autocomplete.

### 3.5 Performance agressiva

Abrir arquivos grandes, fazer scrolling e redrawing deve ser rápido.

### 3.6 Configuração simples

Configurar keybinds, tema, plugins e comportamento não pode virar um inferno.

### 3.7 IA como recurso nativo

IA não deve ser um plugin improvisado. Deve fazer parte da arquitetura desde o começo.

---

## 4. Escopo do produto

## 4.1 Em escopo

* editor de texto multi-buffer
* splits horizontais e verticais
* tabs/workspaces
* file tree / project explorer
* fuzzy finder
* busca e replace em projeto
* syntax highlighting
* LSP
* diagnostics
* autocomplete híbrido
* integração com Git
* diff viewer
* painel de terminal embutido
* sistema de comandos
* sistema de keymaps
* mouse support
* sistema de plugins
* configuração por arquivo
* tema/color scheme
* statusline / panels / popups

## 4.2 Fora de escopo na V1

* debugger completo estilo DAP avançado
* colaboração em tempo real
* UI gráfica fora do terminal
* refatoração cross-project proprietária avançada sem LSP
* indexador semântico gigante distribuído

## 4.3 Pós-V1

* DAP/debugging
* profiler integration
* remote development
* pair programming
* semantic search cross-repo
* AI agent workflows
* notebook-like panes

---

## 5. Público-alvo

### 5.1 Usuário principal

Desenvolvedor que:

* trabalha muito no terminal
* quer algo mais integrado que Vim cru
* quer menos modalismo e menos atrito
* quer IA local e privacidade
* quer customização de alto nível

### 5.2 Usuário secundário

* usuário de VS Code que quer migrar para terminal
* usuário de Neovim que cansou de configuração espalhada
* usuário de JetBrains que quer leveza e terminal nativo

---

## 6. Diferenciais do produto

1. **IDE de terminal de verdade**, não só editor.
2. **Non-modal by default**, com keymaps configuráveis.
3. **IA local integrada** para autocomplete e assistência.
4. **Git e diff viewer nativos**, sem gambiarra.
5. **Plugin system bem definido** desde o início.
6. **Mouse no terminal funcionando direito**.
7. **Arquitetura orientada a ações**, não só a comandos de texto.

---

## 7. Requisitos funcionais

## 7.1 Core de edição

### Requisitos

* abrir, criar, salvar, salvar como, reload
* múltiplos buffers simultâneos
* detecção de arquivo alterado externamente
* undo/redo multi-step
* seleção por teclado e mouse
* copy/cut/paste
* múltiplos cursores no roadmap
* indentação automática
* comment/uncomment
* auto-pair de delimitadores
* line numbers
* soft wrap / hard wrap
* highlight de seleção, palavra atual e matching bracket

### Estrutura sugerida

* `EditorState`
* `Buffer`
* `View`
* `Cursor`
* `Selection`
* `DocumentSnapshot`
* `UndoManager`

---

## 7.2 Buffers

### Conceito

Buffer representa documento aberto, independente de quantas views o exibem.

### Requisitos

* buffer pode existir sem estar visível
* múltiplas views do mesmo buffer
* dirty tracking
* readonly mode
* scratch buffers
* output buffers
* diff buffers
* terminal buffers

### Tipos de buffer

* file buffer
* ephemeral buffer
* search result buffer
* diff buffer
* log buffer
* terminal session buffer

---

## 7.3 Layout e janelas

### Requisitos

* split horizontal
* split vertical
* resize por mouse e teclado
* troca de foco entre painéis
* abas por workspace
* painéis fixos e flutuantes
* popups/modals leves

### Painéis iniciais

* editor pane
* file explorer
* terminal pane
* diagnostics pane
* git pane
* command palette
* autocomplete popup
* hover popup

---

## 7.4 Input e keymaps

### Filosofia

Tudo que o usuário faz é uma **Action**. Keybinds disparam Actions.

### Requisitos

* keymaps customizáveis
* namespaces por contexto
* chord mappings
* leader key opcional
* binding por modo de foco, não por modal editing
* mouse events configuráveis
* fallback default keymap

### Exemplos

* `Ctrl+P` → OpenFileFinder
* `Ctrl+Shift+F` → GlobalSearch
* `Leader G D` → OpenDiffView
* `F2` → RenameSymbol
* `Ctrl+Click` → GoToDefinition

### Contextos de input

* editor
* file tree
* terminal
* popup
* diff viewer
* global

### Exemplo de configuração

```toml
[keymap.global]
"ctrl+p" = "file.find"
"ctrl+shift+p" = "command.palette"

[keymap.editor]
"f12" = "lsp.definition"
"f2" = "lsp.rename"
"leader g d" = "git.diff.open"
```

---

## 7.5 Mouse support

### Requisitos

* posicionar cursor
* selecionar texto
* clicar para focar painel
* scroll vertical e horizontal
* resize de splits
* clicar em diagnostics
* clicar em itens de autocomplete
* clicar em arquivos no explorer
* suporte a double click e drag

### Observação

Mouse precisa funcionar bem em terminals modernos com escape sequences apropriadas.

---

## 7.6 Fuzzy finder e navegação

### Requisitos

* open file by fuzzy match
* switch buffer
* goto symbol
* goto workspace symbol
* recent files
* command palette
* search in project
* grep results navigation

### Fontes

* arquivos
* buffers
* símbolos LSP
* comandos
* branches Git
* commits
* arquivos modificados

---

## 7.7 Busca e replace

### Requisitos

* busca local no buffer
* replace no buffer
* busca global no projeto
* replace no projeto com preview
* regex opcional
* case-sensitive opcional
* whole word opcional

### Implementação sugerida

Usar engine interna com integração futura com `ripgrep` quando disponível.

---

## 7.8 Syntax highlighting e parsing

### Requisitos

* highlighting por linguagem
* folding futuro
* identificação de comentários, strings, keywords
* matching brackets
* incremental parsing idealmente

### Recomendação

Usar Tree-sitter, se viável via bindings estáveis em Go.

Se Tree-sitter complicar demais no começo:

* V1 com highlighting simples
* V2 com parser incremental real

---

## 7.9 LSP

### Requisitos

* iniciar/parar language servers
* diagnostics em tempo real
* hover
* go to definition
* go to references
* rename symbol
* completion
* signature help
* formatting
* code actions
* document symbols
* workspace symbols
* semantic tokens se suportado

### Arquitetura sugerida

Componente dedicado:

* `LSPManager`
* `LSPClient`
* `LanguageRegistry`
* `DiagnosticsStore`

### Fluxo

1. buffer abre
2. linguagem detectada
3. servidor correspondente inicia
4. eventos de edição geram sync incremental
5. respostas atualizam diagnostics/completion/navigation

---

## 7.10 IA local e autocomplete

### Objetivo

Autocomplete híbrido com priorização de:

1. sugestões do buffer/projeto
2. sugestões do LSP
3. sugestões do modelo local

### Casos de uso

* completar linha
* completar bloco
* sugerir boilerplate
* explicar erro
* gerar teste
* refatorar trecho selecionado

### Requisitos

* rodar modelo local
* latência baixa
* streaming de sugestões
* cancelamento de inferência
* fallback quando modelo estiver offline
* contexto de arquivo atual
* contexto de arquivos relevantes próximos
* limite de tokens/contexto configurável

### Estratégia recomendada

Separar em dois níveis:

#### Nível 1 — Completion inline

* ghost text
* aceita com `Tab` ou binding customizável
* cancela ao continuar digitando

#### Nível 2 — Assist panel

* painel para prompts rápidos
* ações contextuais sobre seleção
* explain, refactor, docstring, tests

### Backend sugerido

Criar uma abstração:

* `AIProvider`
* `CompletionProvider`
* `ContextBuilder`

E suportar engines locais como:

* Ollama
* llama.cpp server
* vLLM local futuramente

### Interface sugerida

```go
type CompletionProvider interface {
    Complete(ctx context.Context, req CompletionRequest) (<-chan CompletionChunk, error)
}

type AIProvider interface {
    CompleteInline(ctx context.Context, req InlineCompletionRequest) (<-chan InlineSuggestion, error)
    RunAction(ctx context.Context, req AIActionRequest) (<-chan AIActionChunk, error)
}
```

### Política de ranking

Combinar score por:

* proximidade textual
* contexto sintático
* origem da sugestão
* confiança do LSP
* confiança do modelo
* histórico de aceitação do usuário

---

## 7.11 Git integration

### Requisitos

* branch atual na statusline
* arquivos modificados
* stage/unstage
* discard changes
* commit
* amend commit
* checkout branch
* create branch
* pull/push
* blame futuro
* hunk actions
* diff viewer

### Diff viewer

Deve suportar:

* side-by-side
* inline diff
* navegação por hunk
* stage hunk
* revert hunk
* comparar worktree vs index
* comparar index vs HEAD
* comparar branches/commits

### Componentes sugeridos

* `GitService`
* `RepoState`
* `DiffEngine`
* `GitPanel`

### Implementação

No começo, shell out para `git` com parsing robusto.
Depois, avaliar biblioteca Go se fizer sentido.

---

## 7.12 Terminal embutido

### Requisitos

* terminal integrado como painel
* múltiplas sessões
* shell configurável
* copy/paste
* resize correto
* scrollback
* envio de comando por ação
* integração com tasks

### Casos de uso

* rodar build
* rodar testes
* rodar REPL
* executar comando Git

---

## 7.13 Tasks e automação

### Requisitos

* definir tasks por projeto
* rodar build/test/lint
* capturar saída
* clicar em erro e ir para arquivo/linha
* tasks configuráveis por linguagem

### Exemplo

```toml
[tasks.build]
cmd = "go build ./..."

[tasks.test]
cmd = "go test ./..."

[tasks.lint]
cmd = "golangci-lint run"
```

---

## 7.14 Sistema de comandos

### Requisitos

* command palette
* comandos internos registrados por ID
* comandos expostos a plugins
* argumentos opcionais
* comandos invocáveis por keybind ou menu

### Exemplo

* `file.open`
* `buffer.next`
* `lsp.definition`
* `git.diff.open`
* `ai.explain.selection`

---

## 7.15 Sistema de plugins

### Objetivo

Permitir extensão sem transformar o core em caos.

### Requisitos

* plugins podem registrar comandos
* plugins podem registrar keymaps padrão
* plugins podem adicionar painéis
* plugins podem escutar eventos
* plugins podem contribuir completion sources
* plugins podem contribuir language features

### Opções de arquitetura

#### Opção A — Plugins como processos externos

Vantagens:

* isolamento
* menos risco de crashar o core
* linguagem agnóstica

Desvantagens:

* IPC mais complexo
* mais latência

#### Opção B — Plugins Go compilados

Vantagens:

* integração forte
* rápida

Desvantagens:

* compatibilidade delicada
* deployment pior

#### Opção C — Runtime embarcado (Lua, JS, Starlark)

Vantagens:

* configuração scriptável
* comunidade tende a gostar

Desvantagens:

* mais superfície técnica

### Recomendação

Para V1:

* **config + comandos + extensões por processo externo via RPC leve**

Para V2:

* adicionar scripting embutido

### API mínima de plugin

* register command
* register event handler
* register completion source
* register panel
* register settings schema

### Eventos importantes

* on startup
* on buffer open
* on buffer save
* on diagnostics update
* on git state change
* on command executed

---

## 7.16 Configuração

### Requisitos

* arquivo global do usuário
* arquivo por projeto
* override local
* reload sem reiniciar
* schema validável

### Áreas configuráveis

* tema
* keymaps
* fontes
* comportamento do editor
* LSPs
* IA
* Git
* plugins
* tasks

### Formato sugerido

TOML.

Porque:

* mais limpo que JSON
* mais previsível que YAML
* ótimo para config

---

## 7.17 Observabilidade

### Requisitos

* logs internos
* painel de logs/debug
* tracing opcional
* medição de latência de redraw
* medição de latência de autocomplete
* captura de falhas de plugins

---

## 8. Requisitos não funcionais

## 8.1 Performance

* startup rápido
* scroll suave
* baixo consumo de memória relativo
* suporte razoável a arquivos grandes
* autocomplete não pode travar UI

## 8.2 Confiabilidade

* crash de plugin não derruba tudo
* autosave opcional
* recovery de sessão futuro

## 8.3 Portabilidade

* Linux prioritário
* macOS suportado
* Windows depois, se arquitetura de terminal permitir

## 8.4 Segurança

* plugins externos com permissões claras futuramente
* IA local sem enviar código para fora por padrão

---

## 9. Arquitetura proposta

## 9.1 Visão em camadas

### Camada 1 — Core

* estado global
* buffers
* layout
* actions
* input router
* render scheduler

### Camada 2 — Serviços

* LSP
* Git
* AI
* Search
* Tasks
* Config
* PluginHost

### Camada 3 — UI TUI

* panes
* widgets
* menus
* popup system
* statusline
* mouse handling

### Camada 4 — Integrações externas

* git
* language servers
* AI backend local
* rg/fd opcionais
* shells

---

## 9.2 Modelo de estado

```text
AppState
 ├── Config
 ├── WorkspaceManager
 ├── BufferManager
 ├── LayoutTree
 ├── ActionRegistry
 ├── CommandRegistry
 ├── LSPManager
 ├── GitService
 ├── AIService
 ├── PluginHost
 ├── TaskRunner
 └── UIState
```

---

## 9.3 Event-driven architecture

Quase tudo deve conversar por eventos e ações:

* input gera action
* action altera state ou chama service
* service publica evento
* UI reage ao estado novo

### Eventos exemplo

* BufferOpened
* BufferChanged
* DiagnosticsUpdated
* GitStatusUpdated
* CompletionRequested
* CompletionReceived
* LayoutChanged
* PluginCrashed

Isso reduz acoplamento e facilita plugin system.

---

## 9.4 Renderização

### Requisitos

* render incremental
* diff de tela quando possível
* evitar redraw completo em toda tecla
* scheduler desacoplado do processamento pesado

### Recomendação

Separar:

* input loop
* event loop
* render loop

Nunca deixar inferência de IA ou resposta de LSP bloquear render.

---

## 10. Tecnologias sugeridas

## 10.1 Linguagem

* Go

## 10.2 UI terminal

Opções para avaliar:

* Bubble Tea
* tcell
* termbox-like libs
* implementação própria sobre ANSI + input parser

### Recomendação pragmática

Se quiser acelerar:

* usar **tcell** ou **Bubble Tea + componentes próprios**

Se quiser controle real de IDE e mouse/layout/render:

* **tcell tende a ser mais adequada**

Bubble Tea é boa, mas pode ficar engessada para uma IDE muito rica se você não controlar muito bem a arquitetura.

## 10.3 Parsing/highlighting

* Tree-sitter bindings em Go

## 10.4 LSP

* cliente próprio via stdio/json-rpc

## 10.5 Git

* shell out para binário `git`

## 10.6 IA local

* Ollama API local
* llama.cpp server compatível com HTTP

---

## 11. Módulos principais

## 11.1 editor/

* buffer
* cursor
* selection
* undo
* text operations

## 11.2 ui/

* panes
* widgets
* layout
* renderer
* theme

## 11.3 input/

* key parser
* mouse parser
* keymap resolver
* action dispatcher

## 11.4 lsp/

* process management
* rpc client
* completion
* diagnostics

## 11.5 git/

* repo detection
* status
* diff
* stage/commit ops

## 11.6 ai/

* provider abstraction
* context builder
* inline completion
* assist actions

## 11.7 plugins/

* plugin host
* rpc transport
* manifests
* lifecycle

## 11.8 config/

* load
* merge
* validate
* watch/reload

## 11.9 search/

* fuzzy
* grep
* indexing opcional

## 11.10 terminal/

* pty integration
* terminal pane
* session management

---

## 12. Fluxos principais

## 12.1 Abrir projeto

1. usuário abre diretório
2. workspace inicializa
3. Git é detectado
4. config local do projeto é carregada
5. file tree indexa estrutura mínima
6. serviços de linguagem ficam prontos sob demanda

## 12.2 Digitar código

1. usuário digita
2. buffer atualiza
3. parser/highlighter reage
4. LSP recebe sync incremental
5. completion pipeline prepara sugestões
6. IA local é acionada quando apropriado
7. ghost text ou popup aparece

## 12.3 Abrir diff

1. usuário aciona `git.diff.open`
2. GitService obtém diff
3. DiffEngine processa hunks
4. DiffView abre em painel dedicado
5. usuário navega por mouse ou keybinds

---

## 13. UX mínima da V1

Ao abrir a IDE, o usuário deve ter:

* explorer à esquerda
* editor principal ao centro
* statusline abaixo
* terminal opcional embaixo
* command palette e autocomplete em popup

### Barra de status deve mostrar

* arquivo atual
* linguagem
* posição cursor
* encoding/line endings
* branch Git
* diagnostics
* status LSP
* status AI local

---

## 14. Roadmap técnico

## Fase 0 — Prova de conceito

* abrir arquivo
* editar buffer
* salvar
* keymaps básicos
* splits simples
* mouse click + scroll

## Fase 1 — Core editor utilizável

* múltiplos buffers
* undo/redo
* file explorer
* fuzzy open
* syntax highlighting básico
* command system

## Fase 2 — LSP

* diagnostics
* definition
* hover
* rename
* completion

## Fase 3 — Git

* repo status
* diff viewer
* stage/unstage
* commit básico

## Fase 4 — IA local

* inline completion
* explain selection
* geração simples por prompt

## Fase 5 — Plugins

* commands
* events
* settings
* completion providers

## Fase 6 — Refino pesado

* performance
* arquivos grandes
* UX polish
* recovery
* tasks

---

## 15. Riscos técnicos

## 15.1 Renderização no terminal

Fazer UI rica, rápida e estável no terminal é difícil.

### Mitigação

* manter render loop simples
* usar diff rendering
* testar em vários terminals

## 15.2 Tree-sitter em Go

Bindings e integração podem dar trabalho.

### Mitigação

* começar simples
* isolar parser atrás de interface

## 15.3 Plugins cedo demais

Plugin system mal definido estraga o projeto.

### Mitigação

* primeiro estabilizar actions, commands e events
* depois abrir API pública

## 15.4 IA local com baixa latência

Pode ficar pesada demais.

### Mitigação

* completar só quando contexto fizer sentido
* cancelamento agressivo
* modelos pequenos por padrão

## 15.5 Terminal cross-platform

Windows tende a dar mais dor.

### Mitigação

* Linux/macOS primeiro

---

## 16. Decisões arquiteturais recomendadas

### Decisão 1

**Non-modal por padrão.**
Sem normal/insert mode como base.

### Decisão 2

**Tudo é Action.**
Keybind, clique, comando e plugin disparam actions.

### Decisão 3

**Serviços externos isolados.**
LSP, Git, AI e plugins ficam atrás de interfaces.

### Decisão 4

**Core mínimo e consistente antes de plugins.**
Não abrir API pública cedo demais.

### Decisão 5

**Git e IA são features nativas.**
Não tratar como penduricalhos.

---

## 17. Estrutura inicial do repositório

```text
/cmd/ide
/internal/app
/internal/editor
/internal/ui
/internal/input
/internal/layout
/internal/render
/internal/config
/internal/commands
/internal/actions
/internal/events
/internal/lsp
/internal/git
/internal/ai
/internal/search
/internal/terminal
/internal/plugins
/internal/theme
/internal/logging
/pkg/api
/assets/themes
/docs
```

---

## 18. Exemplo de manifesto de plugin

```toml
name = "git-tools"
version = "0.1.0"
entry = "./git-tools-plugin"

[contributes.commands]
"git.openHistory" = "Open Git History"

[contributes.keymap.editor]
"leader g h" = "git.openHistory"
```

---

## 19. Exemplo de ações centrais

* `file.open`
* `file.save`
* `buffer.close`
* `buffer.next`
* `pane.split.horizontal`
* `pane.split.vertical`
* `pane.focus.left`
* `search.project`
* `command.palette`
* `lsp.definition`
* `lsp.references`
* `lsp.rename`
* `git.status.open`
* `git.diff.open`
* `git.stage.hunk`
* `ai.complete.inline`
* `ai.explain.selection`
* `terminal.toggle`

---

## 20. Critérios de sucesso da V1

A V1 será considerada boa se:

* abrir projetos reais sem travar
* editar código confortavelmente
* suportar mouse bem
* fornecer LSP útil
* mostrar diffs Git de forma sólida
* permitir autocomplete híbrido funcional
* aceitar configuração de keymaps simples
* ter arquitetura pronta para plugins

---

## 21. Opinião direta sobre a estratégia

Você está tentando construir algo grande de verdade. Então o erro fatal seria começar pelo “visual bonito” ou pelo plugin system antes de estabilizar o core.

A ordem certa é:

1. **editor + buffer + input + layout**
2. **ações/comandos/keymaps**
3. **LSP**
4. **Git/diff**
5. **IA local**
6. **plugins**

Se inverter isso, vira projeto eterno.

---

## 22. Stack recomendada para começar agora

### Base

* Go
* tcell
* JSON-RPC próprio para LSP
* shell out para git
* TOML para config

### IA local

* Ollama primeiro
* abstração para trocar backend depois

### Parsing

* highlighting simples primeiro
* Tree-sitter depois

---

## 23. MVP realista de 8 entregas

1. abrir/salvar arquivos
2. múltiplos buffers
3. splits + mouse
4. keymaps configuráveis
5. fuzzy file finder
6. LSP básico
7. Git status + diff viewer
8. autocomplete inline via IA local

---

## 24. Nome interno da arquitetura

Sugestão conceitual:

**Action-driven Terminal IDE Core**

Porque o centro não é “modo”, é **ação + estado + layout + serviços**.

---

## 25. Próximos passos imediatos

1. definir o modelo de `AppState`
2. definir `ActionRegistry`
3. implementar `Buffer` e `EditorView`
4. implementar `LayoutTree`
5. ligar input → actions → render
6. só então plugar LSP

---

## 26. Resumo final

A IDE deve ser:

* terminal-first
* non-modal
* keyboard-first com mouse de verdade
* configurável por actions/keymaps
* extensível por plugins
* com LSP e Git nativos
* com IA local para autocomplete e assistência
* escrita em Go com arquitetura orientada a eventos e serviços

Esse é um projeto bom. Mas só fica viável se você tratar como produto com arquitetura séria, e não como editorzinho com feature solta.

---

## 27. Apêndice técnico — arquitetura interna mais detalhada

## 27.1 Modelo de dados do editor

### Buffer

```go
type BufferID string

type Buffer struct {
    ID           BufferID
    Path         string
    Kind         BufferKind
    Rope         TextRope
    Version      int64
    Dirty        bool
    ReadOnly     bool
    LanguageID   string
    LineEndings  LineEnding
    Encoding     string
    CursorState  []Cursor
    Selections   []Selection
    Undo         *UndoManager
    Metadata     map[string]any
}
```

### View

```go
type ViewID string

type EditorView struct {
    ID              ViewID
    BufferID        BufferID
    ScrollX         int
    ScrollY         int
    Width           int
    Height          int
    ShowLineNumbers bool
    Wrap            bool
    Focused         bool
}
```

### Layout tree

```go
type NodeType int

const (
    NodeLeaf NodeType = iota
    NodeSplit
)

type LayoutNode struct {
    ID         string
    Type       NodeType
    Direction  SplitDirection
    Ratio      float64
    Pane       Pane
    Children   []*LayoutNode
}
```

### Observação importante

Não use `string` puro para representar o texto inteiro do buffer em edição séria. Isso degrada rápido.

### Estruturas candidatas

* piece table
* rope
* gap buffer

### Recomendação

Para essa IDE, a melhor aposta é:

* **piece table** ou **rope**

Porque:

* undo/redo fica mais natural
* edição em arquivos grandes escala melhor
* múltiplas operações de inserção/remoção ficam menos ruins que string pura

Se quiser pragmatismo máximo no começo:

* V0/V1 com estrutura simples por linhas
* mas já escondida atrás de interface

```go
type TextStore interface {
    Insert(pos Position, text string) error
    Delete(r Range) error
    Slice(r Range) string
    Line(n int) string
    LineCount() int
}
```

---

## 27.2 Pipeline de input

O ideal é o input passar por estágios claros:

1. terminal event
2. parser normaliza evento
3. input context resolver decide contexto ativo
4. keymap matcher tenta resolver chord/binding
5. action dispatcher executa ação
6. estado muda
7. render scheduler dispara repaint parcial

### Exemplo

```text
Ctrl+P
→ KeyEvent{Ctrl:true, Key:"p"}
→ Context: editor
→ Resolve binding: file.find
→ Dispatch action
→ Open fuzzy finder popup
→ Re-render popup + statusline
```

### Interface sugerida

```go
type Action interface {
    ID() string
    Run(ctx *ActionContext) error
}

type ActionContext struct {
    App      *AppState
    UI       UIAdapter
    Services *Services
    Args     map[string]any
}
```

---

## 27.3 Scheduler e concorrência

Aqui tem uma armadilha clássica em Go: sair criando goroutine para tudo e perder controle do estado.

### Regra prática

Estado de UI e editor deve ser alterado de forma serializada.

### Estratégia boa

* 1 loop principal de estado/eventos
* workers separados para tarefas pesadas
* resultados voltam como eventos para o loop principal

### Tarefas que podem rodar fora

* LSP requests
* Git status refresh
* grep global
* inferência IA
* parsing pesado

### Tarefas que não devem mutar estado direto

* qualquer worker paralelo

### Padrão recomendado

```text
worker paralelo
→ produz Event
→ event loop principal consome
→ atualiza AppState
→ agenda render
```

Isso evita race condition e UI maluca.

---

## 27.4 Completion pipeline

Seu autocomplete não deve ser uma coisa única. Deve ser um agregador de providers.

### Providers sugeridos

* buffer words
* snippets
* file path
* LSP completion
* AI inline completion
* AI list completion opcional

### Interface

```go
type Suggestion struct {
    Label         string
    InsertText    string
    Detail        string
    Kind          string
    Score         float64
    Source        string
    IsInlineGhost bool
}

type SuggestionProvider interface {
    Name() string
    Suggestions(ctx context.Context, req SuggestionRequest) ([]Suggestion, error)
}
```

### Agregador

```go
type CompletionEngine struct {
    Providers []SuggestionProvider
}
```

### Ranking inicial simples

```text
score final =
    sourceWeight
  + prefixMatchWeight
  + syntaxContextWeight
  + recencyWeight
  + userAcceptanceWeight
```

### Regra de UX importante

Não misturar ghost text com popup agressivo ao mesmo tempo toda hora.

Boa abordagem:

* inline ghost text quando a confiança for alta
* popup quando houver múltiplas sugestões boas

---

## 27.5 Como plugar IA local sem destruir a UX

Esse ponto decide se a IDE vai parecer boa ou travada.

### Regras obrigatórias

* requisição cancelável
* debounce curto
* timeout
* não bloquear digitação
* contexto pequeno e relevante
* cache de sugestões recentes

### Fluxo recomendado

1. usuário pausa 100–250 ms
2. contexto atual é capturado
3. request assíncrona é enviada
4. se o usuário continuar digitando, request anterior é cancelada
5. resposta chega em streaming
6. ghost text aparece se ainda fizer sentido

### Context builder mínimo

* prefixo atual da linha
* sufixo próximo
* função/bloco atual
* imports próximos
* nome do arquivo
* linguagem
* símbolos do LSP se disponíveis

### Não faça isso

* mandar arquivo inteiro sempre
* mandar projeto inteiro sempre
* esperar resposta completa para só então atualizar UI

---

## 27.6 Git — modelo interno

Não trate Git como “só executar comando”. Você precisa de estado derivado consistente.

### Estruturas sugeridas

```go
type RepoState struct {
    RootPath        string
    Branch          string
    HeadSHA         string
    HasUncommitted  bool
    StagedFiles     []GitFileStatus
    UnstagedFiles   []GitFileStatus
    UntrackedFiles  []GitFileStatus
    Conflicts       []GitFileStatus
}

type GitFileStatus struct {
    Path      string
    X         string
    Y         string
    RenamedFrom string
}
```

### Comandos base úteis

* `git status --porcelain=v1 -b`
* `git diff -- ...`
* `git diff --cached -- ...`
* `git show`
* `git log --oneline --decorate`
* `git blame`

### Diff engine

Você pode começar mostrando diff textual pronto do Git.
Depois evoluir para parser de hunk com estrutura própria:

```go
type DiffFile struct {
    OldPath string
    NewPath string
    Hunks   []DiffHunk
}

type DiffHunk struct {
    Header string
    Lines  []DiffLine
}

type DiffLine struct {
    Kind    DiffLineKind
    OldNo   int
    NewNo   int
    Text    string
}
```

Isso destrava viewer side-by-side e ações por hunk.

---

## 27.7 LSP — detalhes práticos

### Recomendação direta

Implemente seu próprio client pequeno. Não inventa framework gigante cedo.

### Responsabilidades do client

* spawn do processo
* stdin/stdout JSON-RPC
* controle de request ID
* pending requests map
* notificações assíncronas
* capabilities negotiation

### Estados do servidor

* stopped
* starting
* ready
* degraded
* crashed

### Estratégia por workspace

* 1 server por workspace/language quando fizer sentido
* ou multiplexar buffers por servidor

### Casos chatos que você já deve prever

* servidor travando
* resposta chegando fora de ordem
* cancelamento de request
* capability não suportada
* diagnostics stale

### Solução boa

Todo response precisa validar:

* buffer ainda existe?
* versão ainda é atual?
* pane ainda está visível?

Se não, descarta.

---

## 27.8 Sistema de tema

No terminal, tema ruim mata a experiência.

### Deve ser configurável

* foreground/background
* keyword
* string
* comment
* function
* type
* statusline normal/warning/error
* diff add/remove/header
* diagnostic colors
* popup borders
* selection/cursor line

### Exemplo

```toml
[theme]
name = "midnight"
background = "#0b1020"
foreground = "#d6deeb"

[theme.syntax]
keyword = "#c792ea"
string = "#ecc48d"
comment = "#637777"
function = "#82aaff"
```

---

## 27.9 Plugin RPC mínimo

Se você for de plugin externo por processo, faça um protocolo pequeno.

### Ciclo

1. IDE lê manifesto
2. sobe processo plugin
3. handshake de capabilities
4. registra comandos/eventos
5. troca mensagens JSON-RPC simples

### Mensagens mínimas

* initialize
* shutdown
* registerCommands
* executeCommand
* eventNotification
* requestCompletions
* providePanelData

### Segurança pragmática para V1

* plugin é confiável por instalação explícita
* log de crash
* timeout em request
* kill/restart do processo se travar

---

## 27.10 Persistência e sessão

Mesmo que você não entregue “session restore” na primeira versão, já deixe espaço para isso.

### Persistir no futuro

* arquivos abertos
* layout
* buffers recentes
* cursor por arquivo
* folds depois
* histórico de comandos

### Cache local útil

* índice de arquivos recente
* histórico de aceitação de autocomplete
* estado do explorer expandido

---

## 28. Decisões de UX que valem ouro

## 28.1 Sem modal editing, mas com eficiência

Você não quer Vim modes, mas quer velocidade.

Então substitua “modos” por:

* actions rápidas
* chords
* command palette
* seleção inteligente
* jump commands

### Exemplo

* `Ctrl+D` seleciona próxima ocorrência
* `Alt+↑/↓` move linha
* `Ctrl+/` comenta linha
* `Ctrl+Click` vai para definição
* `Alt+Enter` abre code actions

---

## 28.2 Descobribilidade

Editor poderoso sem descobribilidade vira nicho travado.

### Tem que ter

* command palette
* hint de keybind ao lado das ações
* painel de atalhos pesquisável
* ajuda contextual em popups

---

## 28.3 Statusline inteligente

A statusline não é decoração. Ela é painel de telemetria do usuário.

### Deve indicar rapidamente

* branch
* erros/warnings
* se arquivo está dirty
* encoding
* posição
* servidor LSP ativo
* modelo IA ativo/inativo
* task em execução

---

## 28.4 Empty states decentes

Quando não houver projeto aberto, a UI deve mostrar algo útil:

* abrir pasta
* arquivos recentes
* projetos recentes
* comandos mais usados

---

## 29. MVP técnico mais concreto

Se você quiser sair do papel sem se perder, faça nessa ordem exata:

### Sprint 1

* boot da app
* loop de eventos
* renderer básico
* abrir e salvar arquivo
* cursor, scroll, seleção simples

### Sprint 2

* BufferManager
* múltiplas views
* splits
* statusline
* mouse click/scroll

### Sprint 3

* command registry
* keymap engine
* command palette
* fuzzy open file

### Sprint 4

* syntax highlight simples
* file tree
* busca no projeto

### Sprint 5

* LSP definition/hover/diagnostics
* diagnostics panel
* completion popup

### Sprint 6

* Git status panel
* diff viewer inicial
* stage/unstage file

### Sprint 7

* terminal pane
* tasks
* problem matcher

### Sprint 8

* AI inline completion
* explain selection
* config recarregável

Depois disso você pensa em plugin system aberto.

---

## 30. Coisas que você não deve fazer cedo

* debugger complexo
* parser próprio de linguagem
* plugin API pública gigante
* sistema de tema excessivamente sofisticado
* integração com 10 backends de IA de cara
* abstração exagerada antes do fluxo real existir

---

## 31. Escolhas técnicas recomendadas, sem rodeio

### UI terminal

* **tcell**

### Config

* **TOML**

### LSP

* **cliente próprio leve**

### Git

* **shell out para git**

### AI local

* **Ollama primeiro**

### Estrutura de texto

* **piece table ou rope atrás de interface**

### Plugin system V1

* **processo externo + RPC simples**

### Arquitetura

* **event loop principal + workers assíncronos**


