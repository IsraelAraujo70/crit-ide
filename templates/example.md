# Crit-IDE Markdown Example

## Overview

This is a **terminal IDE** written in _Go_ that supports syntax highlighting for **multiple languages**.

## Features

- Regex-based syntax highlighting (V1)
- **LSP integration** with diagnostics
- _Configurable themes_

### Code Examples

Inline code: `fmt.Println("hello")`

```go
func main() {
    fmt.Println("Hello from crit-ide!")
}
```

## Links

- [Project Repository](https://github.com/example/crit-ide)
- Documentation: https://docs.example.com/crit-ide

## Task List

1. Implement Tree-sitter for V2
2. Add file explorer panel
3. Project-wide search

* Bullet point one
* Bullet point two
* Bullet point three

---

> This is a blockquote.
> It can span multiple lines.

## Table

| Language   | Extension | LSP Server               |
|------------|-----------|--------------------------|
| Go         | .go       | gopls                    |
| Python     | .py       | pylsp                    |
| Rust       | .rs       | rust-analyzer            |
| TypeScript | .ts       | typescript-language-server|
