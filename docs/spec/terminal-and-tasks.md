# Embedded Terminal and Tasks

## Terminal

### Requirements

- Integrated terminal as a pane
- Multiple sessions
- Configurable shell
- Copy/paste support
- Correct resize handling
- Scrollback buffer
- Send command via action
- Integration with task runner

### Use Cases

- Run build
- Run tests
- Run REPL
- Execute Git commands

### Planned Action IDs

- `terminal.toggle`
- `terminal.new`
- `terminal.next`

## Tasks and Automation

### Requirements

- Define tasks per project
- Run build/test/lint
- Capture output
- Click on error → go to file:line (problem matcher)
- Tasks configurable per language

### Configuration Example

```toml
[tasks.build]
cmd = "go build ./..."

[tasks.test]
cmd = "go test ./..."

[tasks.lint]
cmd = "golangci-lint run"
```

### Problem Matcher

The task runner should parse output for patterns like:

```
file.go:42:10: error message
```

And create navigable diagnostics that open the file at the correct line/column.
