package tgmarkdown

import (
	"strings"
	"testing"

	"github.com/sshturbo/GoTeleMD/internal"
)

func TestTgMarkdown(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expected    string
		safetyLevel int
	}{
		// Blocos de Código
		{"Bloco de código simples", "```\nfmt.Println(\"hello\")\n```", "```\nfmt.Println(\"hello\")\n```", internal.SAFETYLEVELNONE},
		{"Bloco de código com linguagem", "```go\nfmt.Println(\"hello\")\n```", "```go\nfmt.Println(\"hello\")\n```", internal.SAFETYLEVELNONE},

		// Código Inline
		{"Código inline", "This is `inline code` here", "This is `inline code` here", internal.SAFETYLEVELNONE},

		// Negrito
		{"Negrito com asteriscos", "**bold text**", "*bold text*", internal.SAFETYLEVELNONE},
		{"Negrito com underscores", "__bold text__", "*bold text*", internal.SAFETYLEVELNONE},

		// Itálico
		{"Itálico com asteriscos", "*italic text*", "_italic text_", internal.SAFETYLEVELBASIC},
		{"Itálico com underscores", "_italic text_", "_italic text_", internal.SAFETYLEVELBASIC},

		// Riscado
		{"Texto riscado", "~~strikethrough text~~", "~strikethrough text~", internal.SAFETYLEVELBASIC},

		// Links
		{"Link simples", "[link text](https://example.com)", "[link text](https://example.com)", internal.SAFETYLEVELBASIC},
		{"Link com formatação", "[**Bold** and _italic_](https://example.com)", "[*Bold* and _italic_](https://example.com)", internal.SAFETYLEVELBASIC},

		// Listas
		{"Lista não ordenada", "- Item 1\n- Item 2", "• Item 1\n• Item 2", internal.SAFETYLEVELBASIC},
		{"Lista ordenada", "1. First item\n2. Second item", "1. First item\n2. Second item", internal.SAFETYLEVELBASIC},
		{"Lista mista com formatação", "1. **Bold** item\n- _Italic_ item", "1. *Bold* item\n• _Italic_ item", internal.SAFETYLEVELBASIC},

		// Citações
		{"Citação simples", "> Quoted text", "> Quoted text", internal.SAFETYLEVELBASIC},
		{"Citação com formatação", "> **Bold** and _italic_", "> *Bold* and _italic_", internal.SAFETYLEVELBASIC},

		// Títulos
		{"Título H1", "# Heading 1", "*Heading 1*", internal.SAFETYLEVELBASIC},
		{"Título H3", "### Heading 3", "_Heading 3_", internal.SAFETYLEVELBASIC},

		// Tabelas
		{"Tabela simples", "| Col1 | Col2 |\n|------|------|\n| Val1 | Val2 |", "\n• Col1 | Col2\n• Val1 | Val2", internal.SAFETYLEVELBASIC},
		{"Tabela alinhada", "| Col1 | Col2 |\n|:----:|:-----|\n| Val1 | Val2 |", "\n•  Col1  | Col2\n•  Val1  | Val2", internal.SAFETYLEVELBASIC},

		// Texto simples com caracteres especiais
		{"Texto simples com caracteres especiais", "Hello #world! (test)", "Hello \\#world\\! \\(test\\)", internal.SAFETYLEVELBASIC},

		// Novo teste para múltiplos caracteres especiais
		{
			name:        "Texto com múltiplos caracteres especiais",
			input:       "Test # + - = | ! * _ [ ] ( ) { } .",
			expected:    "Test \\# \\+ \\- \\= \\| \\! * _ [ ] \\( \\) \\{ \\} \\.", // Não escapa *, _, [, ]
			safetyLevel: internal.SAFETYLEVELBASIC,
		},

		// Modo Seguro
		{"Nível de segurança estrito", "**bold** and _italic_", "\\*\\*bold\\*\\* and \\_italic\\_", internal.SAFETYLEVELSTRICT},

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
			safetyLevel: internal.SAFETYLEVELBASIC,
		},
		{
			name:        "Código inline com caracteres especiais",
			input:       "Text with `code#with*special(chars)` inline",
			expected:    "Text with `code#with*special(chars)` inline",
			safetyLevel: internal.SAFETYLEVELBASIC,
		},
		{
			name:        "Código inline em modo estrito",
			input:       "Text with `code#with*special(chars)` inline",
			expected:    "Text with \\`code\\#with\\*special\\(chars\\)\\` inline",
			safetyLevel: internal.SAFETYLEVELSTRICT,
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
			limit:  internal.TelegramMaxLength,
			partes: 1,
		},
		{
			name:   "Mensagem_longa",
			input:  strings.Repeat("Texto longo que deve ser quebrado em várias partes. ", 320),
			limit:  internal.TelegramMaxLength,
			partes: 5,
		},
		{
			name:   "Código_longo",
			input:  "```\n" + strings.Repeat("Bloco de código muito longo que precisa ser quebrado\n", 205) + "```",
			limit:  internal.TelegramMaxLength,
			partes: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resultado := Convert(tt.input, false, false, internal.SAFETYLEVELBASIC)
			partes := strings.Split(resultado, "\n\n")
			if len(partes) != tt.partes {
				t.Errorf("Esperado %d partes, obtido %d", tt.partes, len(partes))
			}
		})
	}
}
