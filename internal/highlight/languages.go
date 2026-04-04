package highlight

import "regexp"

func init() {
	Register(langGo())
	Register(langPython())
	Register(langJavaScript())
	Register(langTypeScript())
	Register(langRust())
	Register(langC())
	Register(langJava())
	Register(langRuby())
	Register(langJSON())
	Register(langYAML())
	Register(langMarkdown())
	Register(langHTML())
	Register(langCSS())
	Register(langBash())
	Register(langSQL())
}

// funcCall matches `word(` and trims the trailing `(` so only the name is highlighted.
func funcCall() Rule {
	return Rule{regexp.MustCompile(`\w+\(`), Function, 1}
}

func langGo() *Language {
	return &Language{
		Name:          "Go",
		Extensions:    []string{".go"},
		LineComment:   "//",
		BlockComStart: "/*",
		BlockComEnd:   "*/",
		Rules: []Rule{
			{regexp.MustCompile(`\b(break|case|chan|const|continue|default|defer|else|fallthrough|for|func|go|goto|if|import|interface|map|package|range|return|select|struct|switch|type|var)\b`), Keyword, 0},
			{regexp.MustCompile(`\b(bool|byte|complex64|complex128|error|float32|float64|int|int8|int16|int32|int64|rune|string|uint|uint8|uint16|uint32|uint64|uintptr|any)\b`), Type, 0},
			{regexp.MustCompile(`\b(append|cap|close|complex|copy|delete|imag|len|make|new|panic|print|println|real|recover)\b`), Builtin, 0},
			{regexp.MustCompile(`\b(true|false|nil|iota)\b`), Builtin, 0},
			{regexp.MustCompile(`\b[A-Z]\w*\b`), Type, 0},
			{regexp.MustCompile(`\b\d[\d_]*(\.[\d_]+)?([eE][+-]?\d+)?\b`), Number, 0},
			{regexp.MustCompile(`0x[0-9a-fA-F_]+`), Number, 0},
			funcCall(),
			{regexp.MustCompile(`[+\-*/%&|^~<>=!:]+`), Operator, 0},
		},
	}
}

func langPython() *Language {
	return &Language{
		Name:        "Python",
		Extensions:  []string{".py", ".pyi"},
		LineComment: "#",
		Rules: []Rule{
			{regexp.MustCompile(`\b(and|as|assert|async|await|break|class|continue|def|del|elif|else|except|finally|for|from|global|if|import|in|is|lambda|nonlocal|not|or|pass|raise|return|try|while|with|yield)\b`), Keyword, 0},
			{regexp.MustCompile(`\b(int|float|str|bool|list|dict|set|tuple|bytes|bytearray|complex|frozenset|memoryview|range|type|object|None)\b`), Type, 0},
			{regexp.MustCompile(`\b(print|len|range|input|open|map|filter|zip|enumerate|sorted|reversed|sum|min|max|abs|any|all|isinstance|issubclass|hasattr|getattr|setattr|delattr|super|property|staticmethod|classmethod)\b`), Builtin, 0},
			{regexp.MustCompile(`\b(True|False|None)\b`), Builtin, 0},
			{regexp.MustCompile(`\b\d[\d_]*(\.[\d_]+)?([eE][+-]?\d+)?\b`), Number, 0},
			{regexp.MustCompile(`0x[0-9a-fA-F_]+`), Number, 0},
			funcCall(),
			{regexp.MustCompile(`@\w+`), Builtin, 0},
			{regexp.MustCompile(`[+\-*/%&|^~<>=!:]+`), Operator, 0},
		},
	}
}

func langJavaScript() *Language {
	return &Language{
		Name:          "JavaScript",
		Extensions:    []string{".js", ".jsx", ".mjs", ".cjs"},
		LineComment:   "//",
		BlockComStart: "/*",
		BlockComEnd:   "*/",
		Rules: []Rule{
			{regexp.MustCompile(`\b(async|await|break|case|catch|class|const|continue|debugger|default|delete|do|else|export|extends|finally|for|from|function|if|import|in|instanceof|let|new|of|return|static|super|switch|this|throw|try|typeof|var|void|while|with|yield)\b`), Keyword, 0},
			{regexp.MustCompile(`\b(Array|Boolean|Date|Error|Function|Map|Number|Object|Promise|RegExp|Set|String|Symbol|WeakMap|WeakSet)\b`), Type, 0},
			{regexp.MustCompile(`\b(console|document|window|global|module|exports|require|process)\b`), Builtin, 0},
			{regexp.MustCompile(`\b(true|false|null|undefined|NaN|Infinity)\b`), Builtin, 0},
			{regexp.MustCompile(`\b\d[\d_]*(\.[\d_]+)?([eE][+-]?\d+)?\b`), Number, 0},
			{regexp.MustCompile(`0x[0-9a-fA-F_]+`), Number, 0},
			funcCall(),
			{regexp.MustCompile(`=>|[+\-*/%&|^~<>=!?:]+`), Operator, 0},
		},
	}
}

func langTypeScript() *Language {
	ts := langJavaScript()
	ts.Name = "TypeScript"
	ts.Extensions = []string{".ts", ".tsx"}
	ts.Rules = append(ts.Rules,
		Rule{regexp.MustCompile(`\b(interface|type|enum|namespace|declare|abstract|implements|readonly|keyof|infer|never|unknown)\b`), Keyword, 0},
	)
	return ts
}

func langRust() *Language {
	return &Language{
		Name:          "Rust",
		Extensions:    []string{".rs"},
		LineComment:   "//",
		BlockComStart: "/*",
		BlockComEnd:   "*/",
		Rules: []Rule{
			{regexp.MustCompile(`\b(as|async|await|break|const|continue|crate|dyn|else|enum|extern|fn|for|if|impl|in|let|loop|match|mod|move|mut|pub|ref|return|self|Self|static|struct|super|trait|type|unsafe|use|where|while|yield)\b`), Keyword, 0},
			{regexp.MustCompile(`\b(bool|char|f32|f64|i8|i16|i32|i64|i128|isize|str|u8|u16|u32|u64|u128|usize|String|Vec|Box|Option|Result)\b`), Type, 0},
			{regexp.MustCompile(`\b(true|false|None|Some|Ok|Err)\b`), Builtin, 0},
			{regexp.MustCompile(`\b\d[\d_]*(\.[\d_]+)?([eE][+-]?\d+)?\b`), Number, 0},
			{regexp.MustCompile(`0x[0-9a-fA-F_]+`), Number, 0},
			funcCall(),
			{regexp.MustCompile(`\w+!`), Builtin, 0},
			{regexp.MustCompile(`[+\-*/%&|^~<>=!:]+`), Operator, 0},
		},
	}
}

func langC() *Language {
	return &Language{
		Name:          "C/C++",
		Extensions:    []string{".c", ".h", ".cpp", ".hpp", ".cc", ".cxx", ".hh"},
		LineComment:   "//",
		BlockComStart: "/*",
		BlockComEnd:   "*/",
		Rules: []Rule{
			{regexp.MustCompile(`\b(auto|break|case|const|continue|default|do|else|enum|extern|for|goto|if|inline|register|restrict|return|sizeof|static|struct|switch|typedef|union|volatile|while|class|namespace|template|typename|using|virtual|public|private|protected|new|delete|throw|try|catch|override|final|nullptr)\b`), Keyword, 0},
			{regexp.MustCompile(`\b(void|char|short|int|long|float|double|signed|unsigned|bool|size_t|int8_t|int16_t|int32_t|int64_t|uint8_t|uint16_t|uint32_t|uint64_t)\b`), Type, 0},
			{regexp.MustCompile(`\b(true|false|NULL|nullptr)\b`), Builtin, 0},
			{regexp.MustCompile(`#\s*\w+`), Builtin, 0},
			{regexp.MustCompile(`\b\d[\d_]*(\.[\d_]+)?([eE][+-]?\d+)?[fFlLuU]*\b`), Number, 0},
			{regexp.MustCompile(`0x[0-9a-fA-F_]+[uUlL]*`), Number, 0},
			funcCall(),
			{regexp.MustCompile(`[+\-*/%&|^~<>=!:]+`), Operator, 0},
		},
	}
}

func langJava() *Language {
	return &Language{
		Name:          "Java",
		Extensions:    []string{".java"},
		LineComment:   "//",
		BlockComStart: "/*",
		BlockComEnd:   "*/",
		Rules: []Rule{
			{regexp.MustCompile(`\b(abstract|assert|break|case|catch|class|const|continue|default|do|else|enum|extends|final|finally|for|goto|if|implements|import|instanceof|interface|native|new|package|private|protected|public|return|static|strictfp|super|switch|synchronized|this|throw|throws|transient|try|volatile|while)\b`), Keyword, 0},
			{regexp.MustCompile(`\b(boolean|byte|char|double|float|int|long|short|void|String|Integer|Long|Double|Float|Boolean|Character|Object|List|Map|Set|Array)\b`), Type, 0},
			{regexp.MustCompile(`\b(true|false|null)\b`), Builtin, 0},
			{regexp.MustCompile(`@\w+`), Builtin, 0},
			{regexp.MustCompile(`\b\d[\d_]*(\.[\d_]+)?([eE][+-]?\d+)?[fFdDlL]?\b`), Number, 0},
			{regexp.MustCompile(`0x[0-9a-fA-F_]+[lL]?`), Number, 0},
			funcCall(),
			{regexp.MustCompile(`[+\-*/%&|^~<>=!:]+`), Operator, 0},
		},
	}
}

func langRuby() *Language {
	return &Language{
		Name:        "Ruby",
		Extensions:  []string{".rb", "Gemfile", "Rakefile"},
		LineComment: "#",
		Rules: []Rule{
			{regexp.MustCompile(`\b(alias|and|begin|break|case|class|def|defined\?|do|else|elsif|end|ensure|for|if|in|module|next|not|or|redo|rescue|retry|return|self|super|then|undef|unless|until|when|while|yield)\b`), Keyword, 0},
			{regexp.MustCompile(`\b(true|false|nil)\b`), Builtin, 0},
			{regexp.MustCompile(`\b(puts|print|require|require_relative|include|extend|attr_accessor|attr_reader|attr_writer|raise)\b`), Builtin, 0},
			{regexp.MustCompile(`:\w+`), String, 0},
			{regexp.MustCompile(`@{1,2}\w+`), Type, 0},
			{regexp.MustCompile(`\$\w+`), Type, 0},
			{regexp.MustCompile(`\b\d[\d_]*(\.[\d_]+)?([eE][+-]?\d+)?\b`), Number, 0},
			funcCall(),
			{regexp.MustCompile(`[+\-*/%&|^~<>=!]+`), Operator, 0},
		},
	}
}

func langJSON() *Language {
	return &Language{
		Name:       "JSON",
		Extensions: []string{".json"},
		Rules: []Rule{
			{regexp.MustCompile(`\b(true|false|null)\b`), Builtin, 0},
			{regexp.MustCompile(`\b-?\d+(\.\d+)?([eE][+-]?\d+)?\b`), Number, 0},
			{regexp.MustCompile(`[{}\[\]:,]`), Punctuation, 0},
		},
	}
}

func langYAML() *Language {
	return &Language{
		Name:        "YAML",
		Extensions:  []string{".yaml", ".yml"},
		LineComment: "#",
		Rules: []Rule{
			{regexp.MustCompile(`^[\w.-]+\s*:`), Keyword, 1},
			{regexp.MustCompile(`\b(true|false|null|yes|no|on|off)\b`), Builtin, 0},
			{regexp.MustCompile(`\b-?\d+(\.\d+)?([eE][+-]?\d+)?\b`), Number, 0},
		},
	}
}

func langMarkdown() *Language {
	return &Language{
		Name:       "Markdown",
		Extensions: []string{".md", ".markdown"},
		Rules: []Rule{
			{regexp.MustCompile(`^#{1,6}\s.*`), Keyword, 0},
			{regexp.MustCompile(`\*\*[^*]+\*\*`), Keyword, 0},
			{regexp.MustCompile(`\*[^*]+\*`), Type, 0},
			{regexp.MustCompile("`.+?`"), String, 0},
			{regexp.MustCompile(`\[.+?\]\(.+?\)`), Builtin, 0},
			{regexp.MustCompile(`^[-*+]\s`), Operator, 0},
			{regexp.MustCompile(`^\d+\.\s`), Operator, 0},
		},
	}
}

func langHTML() *Language {
	return &Language{
		Name:          "HTML",
		Extensions:    []string{".html", ".htm"},
		BlockComStart: "<!--",
		BlockComEnd:   "-->",
		Rules: []Rule{
			{regexp.MustCompile(`</?[a-zA-Z][\w-]*`), Keyword, 0},
			{regexp.MustCompile(`/?>`), Keyword, 0},
			{regexp.MustCompile(`\b[a-zA-Z][\w-]*=`), Type, 1},
			{regexp.MustCompile(`&\w+;`), Builtin, 0},
		},
	}
}

func langCSS() *Language {
	return &Language{
		Name:          "CSS",
		Extensions:    []string{".css", ".scss", ".less"},
		LineComment:   "//",
		BlockComStart: "/*",
		BlockComEnd:   "*/",
		Rules: []Rule{
			{regexp.MustCompile(`[.#][\w-]+`), Keyword, 0},
			{regexp.MustCompile(`[\w-]+\s*:`), Type, 1},
			{regexp.MustCompile(`\b\d+(px|em|rem|%|vh|vw|pt|cm|mm|in|s|ms)?\b`), Number, 0},
			{regexp.MustCompile(`#[0-9a-fA-F]{3,8}\b`), Number, 0},
			{regexp.MustCompile(`@\w+`), Builtin, 0},
			{regexp.MustCompile(`[{}();:,]`), Punctuation, 0},
		},
	}
}

func langBash() *Language {
	return &Language{
		Name:        "Shell",
		Extensions:  []string{".sh", ".bash", ".zsh", ".fish", "Makefile", "Dockerfile"},
		LineComment: "#",
		Rules: []Rule{
			{regexp.MustCompile(`\b(if|then|else|elif|fi|for|while|do|done|case|esac|in|function|return|exit|break|continue|export|source|local|readonly|declare|typeset|unset|shift|select|until|trap)\b`), Keyword, 0},
			{regexp.MustCompile(`\$\{?\w+\}?`), Type, 0},
			{regexp.MustCompile(`\b(echo|cd|ls|grep|sed|awk|cat|mkdir|rm|cp|mv|chmod|chown|curl|wget|git|docker|make|npm|pip|apt|brew)\b`), Builtin, 0},
			{regexp.MustCompile(`\b\d+\b`), Number, 0},
			{regexp.MustCompile(`[|&;<>]+`), Operator, 0},
		},
	}
}

func langSQL() *Language {
	return &Language{
		Name:        "SQL",
		Extensions:  []string{".sql"},
		LineComment: "--",
		Rules: []Rule{
			{regexp.MustCompile(`(?i)\b(SELECT|FROM|WHERE|INSERT|UPDATE|DELETE|CREATE|DROP|ALTER|TABLE|INDEX|VIEW|INTO|VALUES|SET|JOIN|LEFT|RIGHT|INNER|OUTER|ON|AND|OR|NOT|IN|LIKE|BETWEEN|IS|NULL|AS|ORDER|BY|GROUP|HAVING|LIMIT|OFFSET|UNION|ALL|DISTINCT|EXISTS|CASE|WHEN|THEN|ELSE|END|BEGIN|COMMIT|ROLLBACK|GRANT|REVOKE|PRIMARY|FOREIGN|KEY|REFERENCES|CONSTRAINT|DEFAULT|CHECK|UNIQUE)\b`), Keyword, 0},
			{regexp.MustCompile(`(?i)\b(INT|INTEGER|BIGINT|SMALLINT|TINYINT|FLOAT|DOUBLE|DECIMAL|NUMERIC|CHAR|VARCHAR|TEXT|BLOB|DATE|TIME|DATETIME|TIMESTAMP|BOOLEAN|SERIAL)\b`), Type, 0},
			{regexp.MustCompile(`(?i)\b(COUNT|SUM|AVG|MIN|MAX|COALESCE|IFNULL|NULLIF|CAST|CONVERT)\b`), Function, 0},
			{regexp.MustCompile(`\b\d+(\.\d+)?\b`), Number, 0},
			{regexp.MustCompile(`[=<>!]+`), Operator, 0},
		},
	}
}
