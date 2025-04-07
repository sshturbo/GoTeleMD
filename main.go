package tgmarkdown

import (
	"strings"

	"github.com/sshturbo/GoTeleMD/internal"
	"github.com/sshturbo/GoTeleMD/pkg/formatter"
	"github.com/sshturbo/GoTeleMD/pkg/parser"
)

// Níveis de segurança para processamento de texto
const (
	SAFETYLEVELNONE   = internal.SAFETYLEVELNONE   // Sem segurança adicional
	SAFETYLEVELBASIC  = internal.SAFETYLEVELBASIC  // Escapa caracteres especiais mantendo formatação
	SAFETYLEVELSTRICT = internal.SAFETYLEVELSTRICT // Escapa todo o texto sem formatação
)

// Variáveis de configuração global
var (
	EnableLogs             = false // Ativa logs de debug
	TruncateInsteadOfBreak = false // Trunca texto ao invés de quebrar em pontos seguros
	MaxWordLength          = 200   // Tamanho máximo de palavra antes de forçar quebra
)

func init() {
	// Sincroniza as configurações com o pacote internal
	internal.EnableLogs = &EnableLogs
	internal.TruncateInsteadOfBreak = &TruncateInsteadOfBreak
	internal.MaxWordLength = &MaxWordLength
}

func Convert(input string, alignTableCols, ignoreTableSeparators bool, safetyLevel ...int) string {
	level := internal.SAFETYLEVELBASIC
	if len(safetyLevel) > 0 {
		level = safetyLevel[0]
	}

	parts := parser.BreakLongText(input)
	var outputParts []string

	for _, part := range parts {
		blocks := parser.Tokenize(part)
		var output strings.Builder

		for i, b := range blocks {
			rendered := renderBlock(b, alignTableCols, ignoreTableSeparators, level)
			if i > 0 {
				output.WriteString("\n\n")
			}
			output.WriteString(rendered)
		}
		outputParts = append(outputParts, output.String())
	}

	return strings.Join(outputParts, "\n\n")
}

func renderBlock(b internal.Block, alignTableCols, ignoreTableSeparators bool, safetyLevel int) string {
	switch b.Type {
	case internal.BlockCode:
		if safetyLevel == internal.SAFETYLEVELSTRICT {
			return formatter.ProcessText(b.Content, safetyLevel)
		}
		// Remove as marcações existentes se houver
		content := b.Content
		if strings.HasPrefix(content, "```") && strings.HasSuffix(content, "```") {
			content = content[3 : len(content)-3]
		}
		if strings.TrimSpace(content) == "" {
			return "```\n```"
		}
		return "```\n" + content + "\n```"
	case internal.BlockText:
		return formatter.ProcessText(b.Content, safetyLevel)
	case internal.BlockTable:
		lines := strings.Split(b.Content, "\n")
		return formatter.ConvertTable(lines, alignTableCols, ignoreTableSeparators)
	case internal.BlockTitle:
		return formatter.ProcessTitle(b.Content, safetyLevel)
	case internal.BlockList:
		return formatter.ProcessList(b.Content, safetyLevel)
	case internal.BlockQuote:
		return formatter.ProcessQuote(b.Content, safetyLevel)
	default:
		return formatter.ProcessText(b.Content, safetyLevel)
	}
}
