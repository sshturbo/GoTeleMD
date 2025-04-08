package formatter

import (
	"strings"
	"time"

	"github.com/sshturbo/GoTeleMD/internal"
	"github.com/sshturbo/GoTeleMD/pkg/parser"
	"github.com/sshturbo/GoTeleMD/pkg/utils"
)

func ConvertMarkdown(input string, alignTableCols, ignoreTableSeparators bool, safetyLevel int) string {
	startTime := time.Now()
	defer func() {
		utils.LogPerformance("Convert total", time.Since(startTime))
	}()

	utils.LogDebug("Iniciando conversão com nível de segurança: %d", safetyLevel)

	response := parser.BreakLongText(strings.TrimSpace(input))

	var outputParts []string
	for _, part := range response.Parts {
		blocks := parser.Tokenize(strings.TrimSpace(part.Content))
		var output strings.Builder
		output.Grow(len(part.Content))

		for i, b := range blocks {
			if i > 0 {
				if b.Type == internal.BlockTitle || blocks[i-1].Type == internal.BlockTitle {
					output.WriteString("\n\n")
				} else if b.Type == internal.BlockCode || blocks[i-1].Type == internal.BlockCode {
					output.WriteString("\n\n")
				} else if b.Type == internal.BlockList || blocks[i-1].Type == internal.BlockList {
					output.WriteString("\n\n")
				} else {
					output.WriteString("\n\n")
				}
			}

			rendered := RenderBlock(b, alignTableCols, ignoreTableSeparators, safetyLevel)
			output.WriteString(rendered)
		}

		outputParts = append(outputParts, strings.TrimSpace(output.String()))
	}

	return strings.TrimSpace(strings.Join(outputParts, "\n\n"))
}
