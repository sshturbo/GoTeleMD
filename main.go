package tgmarkdown

import (
	"strings"

	"github.com/sshturbo/GoTeleMD/internal"
	"github.com/sshturbo/GoTeleMD/pkg/formatter"
	"github.com/sshturbo/GoTeleMD/pkg/parser"
)

// Níveis de segurança para processamento de texto
const (
	SAFETYLEVELNONE   = formatter.SafetyLevelNone   // Sem segurança adicional
	SAFETYLEVELBASIC  = formatter.SafetyLevelMedium // Escapa caracteres especiais mantendo formatação
	SAFETYLEVELSTRICT = formatter.SafetyLevelHigh   // Escapa todo o texto sem formatação
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
	level := formatter.SafetyLevelMedium
	if len(safetyLevel) > 0 {
		switch safetyLevel[0] {
		case internal.SAFETYLEVELNONE:
			level = formatter.SafetyLevelNone
		case internal.SAFETYLEVELBASIC:
			level = formatter.SafetyLevelMedium
		case internal.SAFETYLEVELSTRICT:
			level = formatter.SafetyLevelHigh
		}
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

func renderBlock(b internal.Block, alignTableCols, ignoreTableSeparators bool, safetyLevel formatter.SafetyLevel) string {
	switch b.Type {
	case internal.BlockCode:
		if safetyLevel == formatter.SafetyLevelHigh {
			return formatter.ProcessText(b.Content, safetyLevel)
		}
		content := strings.TrimSpace(b.Content)
		// Se já tem as marcações de código, retorna o conteúdo como está
		if strings.HasPrefix(content, "```") && strings.HasSuffix(content, "```") {
			return content
		}
		// Se está vazio, retorna marcadores vazios
		if content == "" {
			return "```\n```"
		}
		// Remove marcações existentes se houver e adiciona novas
		content = strings.TrimPrefix(content, "```")
		content = strings.TrimSuffix(content, "```")
		return "```\n" + strings.TrimSpace(content) + "\n```"
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
