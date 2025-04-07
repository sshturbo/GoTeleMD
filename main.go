package tgmarkdown

import (
	"strings"

	"github.com/sshturbo/GoTeleMD/internal"
	"github.com/sshturbo/GoTeleMD/pkg/formatter"
	"github.com/sshturbo/GoTeleMD/pkg/parser"
)

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
		if strings.HasPrefix(b.Content, "```") && strings.HasSuffix(b.Content, "```") {
			return b.Content
		}
		if strings.TrimSpace(b.Content) == "" {
			return "```\n```"
		}
		return "```\n" + b.Content + "\n```"
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
