package formatter

import (
	"strconv"
	"strings"
	"time"

	"github.com/sshturbo/GoTeleMD/internal"
	"github.com/sshturbo/GoTeleMD/pkg/utils"
)

// RenderBlock converte um bloco de texto para o formato MarkdownV2 do Telegram
func RenderBlock(b internal.Block, alignTableCols, ignoreTableSeparators bool, safetyLevel int) string {
	renderStart := time.Now()
	defer func() {
		utils.LogPerformance("renderBlock "+strconv.Itoa(int(b.Type)), time.Since(renderStart))
	}()

	utils.LogDebug("Renderizando bloco tipo: %v", b.Type)

	switch b.Type {
	case internal.BlockCode:
		return b.Content // Retorna o conteúdo do bloco de código sem alterações
	case internal.BlockText:
		return ProcessText(strings.TrimSpace(b.Content), safetyLevel)
	case internal.BlockTable:
		lines := strings.Split(b.Content, "\n")
		return ConvertTable(lines, alignTableCols, ignoreTableSeparators)
	case internal.BlockTitle:
		return ProcessTitle(b.Content, safetyLevel)
	case internal.BlockList:
		return ProcessList(b.Content, safetyLevel)
	case internal.BlockQuote:
		return ProcessQuote(b.Content, safetyLevel)
	default:
		return ProcessText(strings.TrimSpace(b.Content), safetyLevel)
	}
}
