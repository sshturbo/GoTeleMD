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
		{"Tabela simples", "| Col1 | Col2 |\n|------|------|\n| Val1 | Val2 |", "• Col1 | Col2\n• Val1 | Val2", internal.SAFETYLEVELBASIC},
		{"Tabela alinhada", "| Col1 | Col2 |\n|:----:|:-----|\n| Val1 | Val2 |", "•  Col1  | Col2\n•  Val1  | Val2", internal.SAFETYLEVELBASIC},

		// Texto simples com caracteres especiais
		{"Texto simples com caracteres especiais", "Hello #world! (test).", "Hello \\#world\\! \\(test\\)\\.", internal.SAFETYLEVELBASIC},

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
			expected:    "Text with `code\\#with\\*special\\(chars\\)` inline",
			safetyLevel: internal.SAFETYLEVELBASIC,
		},
		{
			name:        "Código inline em modo estrito",
			input:       "Text with `code#with*special(chars)` inline",
			expected:    "Text with \\`code\\#with\\*special\\(chars\\)\\` inline",
			safetyLevel: internal.SAFETYLEVELSTRICT,
		},
		{
			name:        "Bloco de código longo não deve ser dividido",
			input:       "```python\n" + strings.Repeat("print('teste')\n", 200) + "```",
			expected:    "```python\n" + strings.Repeat("print('teste')\n", 200) + "```",
			safetyLevel: internal.SAFETYLEVELBASIC,
		},
		{
			name:        "Texto e bloco de código devem ser separados",
			input:       strings.TrimSpace(strings.Repeat("Texto antes do código. ", 50)) + "\n```\nprint('teste')\n```\n" + strings.TrimSpace(strings.Repeat("Texto depois do código. ", 50)),
			expected:    strings.TrimSpace(strings.Repeat("Texto antes do código\\. ", 50)) + "\n\n```\nprint('teste')\n```\n\n" + strings.TrimSpace(strings.Repeat("Texto depois do código\\. ", 50)),
			safetyLevel: internal.SAFETYLEVELBASIC,
		},
		{
			name:        "Múltiplos blocos de código devem ser preservados",
			input:       "```js\nconsole.log('primeiro');\n```\nTexto entre blocos.\n```python\nprint('segundo')\n```",
			expected:    "```js\nconsole.log('primeiro');\n```\n\nTexto entre blocos\\.\n\n```python\nprint('segundo')\n```",
			safetyLevel: internal.SAFETYLEVELBASIC,
		},
		{
			name:        "Bloco de código com caracteres especiais",
			input:       "```\n# Comentário com #hashtag e @menção.\nprint('teste!@#$%')\n```",
			expected:    "```\n# Comentário com #hashtag e @menção.\nprint('teste!@#$%')\n```",
			safetyLevel: internal.SAFETYLEVELBASIC,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			alignTableCols := tt.name == "Tabela alinhada"
			response := Convert(tt.input, alignTableCols, false, tt.safetyLevel)

			// Para testes, vamos verificar apenas o conteúdo da primeira parte
			if len(response.Parts) == 0 {
				t.Error("Resposta não contém nenhuma parte")
				return
			}

			resultado := response.Parts[0].Content
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
			response := Convert(tt.input, false, false, internal.SAFETYLEVELBASIC)
			if response.TotalParts != tt.partes {
				t.Errorf("Esperado %d partes, obtido %d", tt.partes, response.TotalParts)
			}
		})
	}
}

func TestLongMessagesWithCodeBlocks(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name: "Mensagem longa com blocos de código",
			input: strings.Repeat("Texto antes. ", 200) + "\n" +
				"```python\n" + strings.Repeat("print('teste')\n", 50) + "```\n" +
				strings.Repeat("Texto depois. ", 200),
			expected: []string{
				strings.Repeat("Texto antes\\. ", 200),
				"```python\n" + strings.Repeat("print('teste')\n", 50) + "```",
				strings.Repeat("Texto depois\\. ", 200),
			},
		},
		{
			name: "Múltiplos blocos de código em mensagem longa",
			input: "```js\nconsole.log('primeiro');\n```\n" +
				strings.Repeat("Texto no meio. ", 200) + "\n" +
				"```python\nprint('segundo')\n```",
			expected: []string{
				"```js\nconsole.log('primeiro');\n```",
				strings.Repeat("Texto no meio\\. ", 200),
				"```python\nprint('segundo')\n```",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response := Convert(tt.input, false, false, internal.SAFETYLEVELBASIC)

			if len(response.Parts) != len(tt.expected) {
				t.Errorf("Número de partes incorreto. Esperado %d, obtido %d",
					len(tt.expected), len(response.Parts))
				return
			}

			for i, expectedContent := range tt.expected {
				if response.Parts[i].Content != expectedContent {
					t.Errorf("Parte %d incorreta.\nEsperado:\n%s\nObtido:\n%s",
						i, expectedContent, response.Parts[i].Content)
				}
			}
		})
	}
}
