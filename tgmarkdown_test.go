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
		safeMode    bool
		safetyLevel int
	}{
		// Blocos de Código
		{"Bloco de código simples", "```\nfmt.Println(\"hello\")\n```", "```\nfmt.Println(\"hello\")\n```", false, SAFETYLEVELNONE},
		{"Bloco de código com linguagem", "```go\nfmt.Println(\"hello\")\n```", "```\nfmt.Println(\"hello\")\n```", false, SAFETYLEVELNONE},

		// Código Inline
		{"Código inline", "This is `inline code` here", "This is `inline code` here", false, SAFETYLEVELNONE},

		// Negrito
		{"Negrito com asteriscos", "**bold text**", "*bold text*", false, SAFETYLEVELNONE},
		{"Negrito com underscores", "__bold text__", "*bold text*", false, SAFETYLEVELNONE},

		// Itálico
		{"Itálico com asteriscos", "*italic text*", "_italic text_", false, SAFETYLEVELBASIC},
		{"Itálico com underscores", "_italic text_", "_italic text_", false, SAFETYLEVELBASIC},

		// Riscado
		{"Texto riscado", "~~strikethrough text~~", "~strikethrough text~", false, SAFETYLEVELBASIC},

		// Links
		{"Link simples", "[link text](https://example.com)", "[link text](https://example.com)", false, SAFETYLEVELBASIC},
		{"Link com formatação", "[**Bold** and _italic_](https://example.com)", "[*Bold* and _italic_](https://example.com)", false, SAFETYLEVELBASIC},

		// Listas
		{"Lista não ordenada", "- Item 1\n- Item 2", "• Item 1\n• Item 2", false, SAFETYLEVELBASIC},
		{"Lista ordenada", "1. First item\n2. Second item", "1. First item\n2. Second item", false, SAFETYLEVELBASIC},
		{"Lista mista com formatação", "1. **Bold** item\n- _Italic_ item", "1. *Bold* item\n• _Italic_ item", false, SAFETYLEVELBASIC},

		// Citações
		{"Citação simples", "> Quoted text", "> Quoted text", false, SAFETYLEVELBASIC},
		{"Citação com formatação", "> **Bold** and _italic_", "> *Bold* and _italic_", false, SAFETYLEVELBASIC},

		// Títulos
		{"Título H1", "# Heading 1", "*Heading 1*", false, SAFETYLEVELBASIC},
		{"Título H3", "### Heading 3", "_Heading 3_", false, SAFETYLEVELBASIC},

		// Tabelas
		{"Tabela simples", "| Col1 | Col2 |\n|------|------|\n| Val1 | Val2 |", "\n• Col1 | Col2\n• Val1 | Val2", false, SAFETYLEVELBASIC},
		{"Tabela alinhada", "| Col1 | Col2 |\n|:----:|:-----|\n| Val1 | Val2 |", "\n•  Col1  | Col2\n•  Val1  | Val2", true, SAFETYLEVELBASIC},

		// Texto simples com caracteres especiais
		{"Texto simples com caracteres especiais", "Hello #world! (test)", "Hello \\#world\\! \\(test\\)", false, SAFETYLEVELBASIC},

		// Novo teste para múltiplos caracteres especiais
		{
			name:        "Texto com múltiplos caracteres especiais",
			input:       "Test # + - = | ! * _ [ ] ( ) { } .",
			expected:    "Test \\# \\+ \\- \\= \\| \\! * _ [ ] \\( \\) \\{ \\} \\.", // Não escapa *, _, [, ]
			safeMode:    false,
			safetyLevel: SAFETYLEVELBASIC,
		},

		// Modo Seguro
		{"Nível de segurança estrito", "**bold** and _italic_", "\\*\\*bold\\*\\* and \\_italic\\_", true, SAFETYLEVELSTRICT},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			alignTableCols := tt.name == "Tabela alinhada"
			resultado := Convert(tt.input, alignTableCols, false, tt.safeMode, tt.safetyLevel)
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
			limit:  100,
			partes: 1,
		},
		{
			name:   "Mensagem longa",
			input:  strings.Repeat("Texto longo que deve ser quebrado em várias partes. ", 10),
			limit:  100,
			partes: 5,
		},
		{
			name:   "Código longo",
			input:  "```\n" + strings.Repeat("Bloco de código muito longo que precisa ser quebrado\n", 5) + "```",
			limit:  100,
			partes: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resultado := Convert(tt.input, false, false, false, SAFETYLEVELBASIC)
			partes := strings.Split(resultado, "\n\n")
			if len(partes) != tt.partes {
				t.Errorf("Esperado %d partes, obtido %d", tt.partes, len(partes))
			}
		})
	}
}


