package highlight

import "regexp"

// LangShell returns the language definition for Shell/Bash.
func LangShell() *LanguageDef {
	return &LanguageDef{
		ID:          "shell",
		Extensions:  []string{".sh", ".bash", ".zsh", ".fish"},
		FileNames:   []string{".bashrc", ".zshrc", ".profile", ".bash_profile"},
		LineComment: "#",
		Patterns: []PatternRule{
			{TokenComment, regexp.MustCompile(`#.*`)},
			{TokenString, regexp.MustCompile(`"(?:[^"\\]|\\.)*"`)},
			{TokenString, regexp.MustCompile(`'[^']*'`)},
			{TokenVariable, regexp.MustCompile(`\$\{[^}]+\}|\$[A-Za-z_]\w*|\$[0-9@#?$!*-]`)},
			{TokenKeyword, regexp.MustCompile(`\b(?:if|then|else|elif|fi|for|while|until|do|done|case|esac|in|function|select|time|coproc)\b`)},
			{TokenBuiltin, regexp.MustCompile(`\b(?:echo|printf|read|cd|pwd|ls|cat|grep|sed|awk|find|sort|uniq|wc|head|tail|cut|tr|tee|xargs|test|export|source|alias|unalias|set|unset|shift|eval|exec|trap|wait|kill|exit|return|local|declare|typeset|readonly)\b`)},
			{TokenOperator, regexp.MustCompile(`(?:&&|\|\||;;|[|&;><]=?|<<|>>|2>&1)`)},
			{TokenNumber, regexp.MustCompile(`\b[0-9]+\b`)},
		},
	}
}
