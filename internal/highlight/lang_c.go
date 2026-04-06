package highlight

import "regexp"

// LangC returns the language definition for C.
func LangC() *LanguageDef {
	return &LanguageDef{
		ID:                "c",
		Extensions:        []string{".c", ".h"},
		FileNames:         []string{"Makefile"},
		LineComment:       "//",
		BlockCommentOpen:  "/*",
		BlockCommentClose: "*/",
		Patterns: []PatternRule{
			{TokenComment, regexp.MustCompile(`//.*`)},
			{TokenPreproc, regexp.MustCompile(`^\s*#\s*(?:include|define|undef|if|ifdef|ifndef|elif|else|endif|pragma|error|warning|line)\b.*`)},
			{TokenString, regexp.MustCompile(`"(?:[^"\\]|\\.)*"`)},
			{TokenString, regexp.MustCompile(`'(?:[^'\\]|\\.)*'`)},
			{TokenKeyword, regexp.MustCompile(`\b(?:auto|break|case|const|continue|default|do|else|enum|extern|for|goto|if|inline|register|restrict|return|sizeof|static|struct|switch|typedef|union|volatile|while|_Alignas|_Alignof|_Atomic|_Bool|_Complex|_Generic|_Imaginary|_Noreturn|_Static_assert|_Thread_local)\b`)},
			{TokenTypeName, regexp.MustCompile(`\b(?:void|char|short|int|long|float|double|signed|unsigned|size_t|ssize_t|ptrdiff_t|intptr_t|uintptr_t|int8_t|int16_t|int32_t|int64_t|uint8_t|uint16_t|uint32_t|uint64_t|bool|FILE)\b`)},
			{TokenConstant, regexp.MustCompile(`\b(?:NULL|true|false|EOF|stdin|stdout|stderr|EXIT_SUCCESS|EXIT_FAILURE)\b`)},
			{TokenBuiltin, regexp.MustCompile(`\b(?:printf|fprintf|sprintf|snprintf|scanf|sscanf|malloc|calloc|realloc|free|memcpy|memset|memmove|strcmp|strncmp|strlen|strcpy|strncpy|strcat|strncat|fopen|fclose|fread|fwrite|fgets|fputs|puts|getchar|putchar|exit|abort|assert)\b`)},
			{TokenNumber, regexp.MustCompile(`\b(?:0[xX][0-9a-fA-F]+[uUlL]*|0[0-7]+[uUlL]*|0[bB][01]+[uUlL]*|[0-9]+(?:\.[0-9]+)?(?:[eE][+-]?[0-9]+)?[fFlLuU]*)\b`)},
			{TokenFunction, regexp.MustCompile(`\b([a-zA-Z_]\w*)\s*\(`)},
			{TokenOperator, regexp.MustCompile(`(?:->|&&|\|\||<<|>>|[+\-*/%&|^<>=!]=?|~)`)},
		},
	}
}

// LangCPP returns the language definition for C++.
func LangCPP() *LanguageDef {
	cpp := LangC()
	cpp.ID = "cpp"
	cpp.Extensions = []string{".cpp", ".cc", ".cxx", ".hpp", ".hh", ".hxx", ".h++"}
	cppKeywords := PatternRule{
		TokenKeyword,
		regexp.MustCompile(`\b(?:alignas|alignof|and|and_eq|asm|bitand|bitor|catch|class|compl|concept|consteval|constexpr|constinit|co_await|co_return|co_yield|decltype|delete|dynamic_cast|explicit|export|friend|module|mutable|namespace|new|noexcept|not|not_eq|operator|or|or_eq|override|private|protected|public|reinterpret_cast|requires|static_assert|static_cast|template|this|throw|try|typeid|typename|using|virtual|xor|xor_eq)\b`),
	}
	cppTypes := PatternRule{
		TokenTypeName,
		regexp.MustCompile(`\b(?:auto|char8_t|char16_t|char32_t|wchar_t|nullptr_t|string|vector|map|set|unordered_map|unordered_set|shared_ptr|unique_ptr|weak_ptr|optional|variant|tuple|array|span|string_view)\b`),
	}
	cppConsts := PatternRule{
		TokenConstant,
		regexp.MustCompile(`\b(?:nullptr|this)\b`),
	}
	cpp.Patterns = append([]PatternRule{cppKeywords, cppTypes, cppConsts}, cpp.Patterns...)
	return cpp
}
