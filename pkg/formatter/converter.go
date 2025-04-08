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
				if (b.Type == internal.BlockCode || blocks[i-1].Type == internal.BlockCode) && b.Type != blocks[i-1].Type {
					output.WriteString("\n\n")
				}
			}

			output.WriteString(RenderBlock(b, alignTableCols, ignoreTableSeparators, safetyLevel))
		}

		outputParts = append(outputParts, strings.TrimSpace(output.String()))
	}

	return strings.TrimSpace(strings.Join(outputParts, "\n\n"))
}
