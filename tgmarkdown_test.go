package tgmarkdown

import (
	"strings"
	"testing"

	"github.com/sshturbo/GoTeleMD/internal"
	"github.com/sshturbo/GoTeleMD/pkg/formatter"
)

func TestTgMarkdown(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expected    string
		safetyLevel formatter.SafetyLevel
	}{
		// Blocos de Código
		{"Bloco de código simples", "```\nfmt.Println(\"hello\")\n```", "```\nfmt.Println(\"hello\")\n```", formatter.SafetyLevelNone},
		{"Bloco de código com linguagem", "```go\nfmt.Println(\"hello\")\n```", "```go\nfmt.Println(\"hello\")\n```", formatter.SafetyLevelNone},

		// Código Inline
		{"Código inline", "This is `inline code` here", "This is `inline code` here", formatter.SafetyLevelNone},

		// Negrito
		{"Negrito com asteriscos", "**bold text**", "*bold text*", formatter.SafetyLevelNone},
		{"Negrito com underscores", "__bold text__", "*bold text*", formatter.SafetyLevelNone},

		// Itálico
		{"Itálico com asteriscos", "*italic text*", "_italic text_", formatter.SafetyLevelMedium},
		{"Itálico com underscores", "_italic text_", "_italic text_", formatter.SafetyLevelMedium},

		// Riscado
		{"Texto riscado", "~~strikethrough text~~", "~strikethrough text~", formatter.SafetyLevelMedium},

		// Links
		{"Link simples", "[link text](https://example.com)", "[link text](https://example.com)", formatter.SafetyLevelMedium},
		{"Link com formatação", "[**Bold** and _italic_](https://example.com)", "[*Bold* and _italic_](https://example.com)", formatter.SafetyLevelMedium},

		// Listas
		{"Lista não ordenada", "- Item 1\n- Item 2", "• Item 1\n• Item 2", formatter.SafetyLevelMedium},
		{"Lista ordenada", "1. First item\n2. Second item", "1. First item\n2. Second item", formatter.SafetyLevelMedium},
		{"Lista mista com formatação", "1. **Bold** item\n- _Italic_ item", "1. *Bold* item\n• _Italic_ item", formatter.SafetyLevelMedium},

		// Citações
		{"Citação simples", "> Quoted text", "> Quoted text", formatter.SafetyLevelMedium},
		{"Citação com formatação", "> **Bold** and _italic_", "> *Bold* and _italic_", formatter.SafetyLevelMedium},

		// Títulos
		{"Título H1", "# Heading 1", "*Heading 1*", formatter.SafetyLevelMedium},
		{"Título H3", "### Heading 3", "_Heading 3_", formatter.SafetyLevelMedium},

		// Tabelas
		{"Tabela simples", "| Col1 | Col2 |\n|------|------|\n| Val1 | Val2 |", "\n• Col1 | Col2\n• Val1 | Val2", formatter.SafetyLevelMedium},
		{"Tabela alinhada", "| Col1 | Col2 |\n|:----:|:-----|\n| Val1 | Val2 |", "\n•  Col1  | Col2\n•  Val1  | Val2", formatter.SafetyLevelMedium},

		// Texto simples com caracteres especiais
		{"Texto simples com caracteres especiais", "Hello #world! (test)", "Hello \\#world\\! \\(test\\)", formatter.SafetyLevelMedium},

		// Novo teste para múltiplos caracteres especiais
		{
			name:        "Texto com múltiplos caracteres especiais",
			input:       "Test # + - = | ! * _ [ ] ( ) { } .",
			expected:    "Test \\# \\+ \\- \\= \\| \\! * _ [ ] \\( \\) \\{ \\} \\.",
			safetyLevel: formatter.SafetyLevelMedium,
		},

		// Modo Seguro
		{"Nível de segurança estrito", "**bold** and _italic_", "\\*\\*bold\\*\\* and \\_italic\\_", formatter.SafetyLevelHigh},

		{
			name: "Teste de HTML",
			input: "```html\n" +
				"<!DOCTYPE html>\n" +
				`<html lang="pt-BR">` + "\n" +
				"<head>\n" +
				`  <meta charset="UTF-8">` + "\n" +
				"  <title>Minha Página</title>\n" +
				"</head>\n" +
				"<body>\n" +
				"  <h1>Bem-vindo</h1>\n" +
				"</body>\n" +
				"</html>\n```",
			expected: "```html\n" +
				"<!DOCTYPE html>\n" +
				`<html lang="pt-BR">` + "\n" +
				"<head>\n" +
				`  <meta charset="UTF-8">` + "\n" +
				"  <title>Minha Página</title>\n" +
				"</head>\n" +
				"<body>\n" +
				"  <h1>Bem-vindo</h1>\n" +
				"</body>\n" +
				"</html>\n```",
			safetyLevel: formatter.SafetyLevelMedium,
		},
		{
			name:        "Código inline com caracteres especiais",
			input:       "Text with `code#with*special(chars)` inline",
			expected:    "Text with `code\\#with\\*special\\(chars\\)` inline",
			safetyLevel: formatter.SafetyLevelMedium,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resultado := Convert(tt.input, false, false, int(tt.safetyLevel))
			if resultado != tt.expected {
				t.Errorf("Convert() = %v, want %v", resultado, tt.expected)
			}
		})
	}
}

func TestLongMessages(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expected    string
		safetyLevel formatter.SafetyLevel
	}{
		{
			name: "Mensagem longa com código",
			input: "```php\n" +
				"<?php\n" +
				"// Código PHP muito longo\n" +
				"function example() {\n" +
				"    echo \"Hello, world!\";\n" +
				"}\n" +
				"```",
			expected: "```php\n" +
				"<?php\n" +
				"// Código PHP muito longo\n" +
				"function example() {\n" +
				"    echo \"Hello, world!\";\n" +
				"}\n" +
				"```",
			safetyLevel: formatter.SafetyLevelMedium,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resultado := Convert(tt.input, false, false, int(tt.safetyLevel))
			if resultado != tt.expected {
				t.Errorf("Convert() = %v, want %v", resultado, tt.expected)
			}
		})
	}
}
