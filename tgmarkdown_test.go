package tgmarkdown

import (
	"strings"
	"testing"
)

func TestTgMarkdown(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expected    string
		safetyLevel int
	}{
		// Blocos de Código
		{"Bloco de código simples", "```\nfmt.Println(\"hello\")\n```", "```\nfmt.Println(\"hello\")\n```", SAFETYLEVELNONE},
		{"Bloco de código com linguagem", "```go\nfmt.Println(\"hello\")\n```", "```go\nfmt.Println(\"hello\")\n```", SAFETYLEVELNONE},

		// Código Inline
		{"Código inline", "This is `inline code` here", "This is `inline code` here", SAFETYLEVELNONE},

		// Negrito
		{"Negrito com asteriscos", "**bold text**", "*bold text*", SAFETYLEVELNONE},
		{"Negrito com underscores", "__bold text__", "*bold text*", SAFETYLEVELNONE},

		// Itálico
		{"Itálico com asteriscos", "*italic text*", "_italic text_", SAFETYLEVELBASIC},
		{"Itálico com underscores", "_italic text_", "_italic text_", SAFETYLEVELBASIC},

		// Riscado
		{"Texto riscado", "~~strikethrough text~~", "~strikethrough text~", SAFETYLEVELBASIC},

		// Links
		{"Link simples", "[link text](https://example.com)", "[link text](https://example.com)", SAFETYLEVELBASIC},
		{"Link com formatação", "[**Bold** and _italic_](https://example.com)", "[*Bold* and _italic_](https://example.com)", SAFETYLEVELBASIC},

		// Listas
		{"Lista não ordenada", "- Item 1\n- Item 2", "• Item 1\n• Item 2", SAFETYLEVELBASIC},
		{"Lista ordenada", "1. First item\n2. Second item", "1. First item\n2. Second item", SAFETYLEVELBASIC},
		{"Lista mista com formatação", "1. **Bold** item\n- _Italic_ item", "1. *Bold* item\n• _Italic_ item", SAFETYLEVELBASIC},

		// Citações
		{"Citação simples", "> Quoted text", "> Quoted text", SAFETYLEVELBASIC},
		{"Citação com formatação", "> **Bold** and _italic_", "> *Bold* and _italic_", SAFETYLEVELBASIC},

		// Títulos
		{"Título H1", "# Heading 1", "*Heading 1*", SAFETYLEVELBASIC},
		{"Título H3", "### Heading 3", "_Heading 3_", SAFETYLEVELBASIC},

		// Tabelas
		{"Tabela simples", "| Col1 | Col2 |\n|------|------|\n| Val1 | Val2 |", "\n• Col1 | Col2\n• Val1 | Val2", SAFETYLEVELBASIC},
		{"Tabela alinhada", "| Col1 | Col2 |\n|:----:|:-----|\n| Val1 | Val2 |", "\n•  Col1  | Col2\n•  Val1  | Val2", SAFETYLEVELBASIC},

		// Texto simples com caracteres especiais
		{"Texto simples com caracteres especiais", "Hello #world! (test)", "Hello \\#world\\! \\(test\\)", SAFETYLEVELBASIC},

		// Novo teste para múltiplos caracteres especiais
		{
			name:        "Texto com múltiplos caracteres especiais",
			input:       "Test # + - = | ! * _ [ ] ( ) { } .",
			expected:    "Test \\# \\+ \\- \\= \\| \\! * _ [ ] \\( \\) \\{ \\} \\.", // Não escapa *, _, [, ]
			safetyLevel: SAFETYLEVELBASIC,
		},

		// Modo Seguro
		{"Nível de segurança estrito", "**bold** and _italic_", "\\*\\*bold\\*\\* and \\_italic\\_", SAFETYLEVELSTRICT},

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
			safetyLevel: SAFETYLEVELBASIC,
		},
		{
			name:        "Código inline com caracteres especiais",
			input:       "Text with `code#with*special(chars)` inline",
			expected:    "Text with `code#with*special(chars)` inline",
			safetyLevel: SAFETYLEVELBASIC,
		},
		{
			name:        "Código inline em modo estrito",
			input:       "Text with `code#with*special(chars)` inline",
			expected:    "Text with \\`code#with\\*special\\(chars\\)\\` inline",
			safetyLevel: SAFETYLEVELSTRICT,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			alignTableCols := tt.name == "Tabela alinhada"
			resultado := Convert(tt.input, alignTableCols, false, tt.safetyLevel)
			if resultado != tt.expected {
				t.Errorf("\nEsperado:\n%v\nObtido:\n%v", tt.expected, resultado)
			}
		})
	}
}

func TestLongMessages(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		limit  int
		partes int
	}{
		{
			name:   "Mensagem curta",
			input:  "Uma mensagem curta que não deve ser quebrada",
			limit:  TelegramMaxLength,
			partes: 1,
		},
		{
			name:   "Mensagem_longa",
			input:  strings.Repeat("Texto longo que deve ser quebrado em várias partes. ", 320),
			limit:  TelegramMaxLength, // Ajustar para refletir o limite real usado
			partes: 5,
		},
		{
			name:   "Código_longo",
			input:  "```\n" + strings.Repeat("Bloco de código muito longo que precisa ser quebrado\n", 205) + "```",
			limit:  TelegramMaxLength, // Refletir o limite real usado no código
			partes: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resultado := Convert(tt.input, false, false, SAFETYLEVELBASIC)
			partes := strings.Split(resultado, "\n\n")
			if len(partes) != tt.partes {
				t.Errorf("Esperado %d partes, obtido %d", tt.partes, len(partes))
			}
		})
	}
}
