package formatter

import (
	"strconv"
	"strings"
	"time"

	"github.com/sshturbo/GoTeleMD/internal"
	"github.com/sshturbo/GoTeleMD/pkg/types"
	"github.com/sshturbo/GoTeleMD/pkg/utils"
)

func RenderBlock(b internal.Block, config *types.Config) string {
	renderStart := time.Now()
	defer func() {
		utils.LogPerformance("renderBlock "+strconv.Itoa(int(b.Type)), time.Since(renderStart))
	}()

	utils.LogDebug("Renderizando bloco tipo: %v", b.Type)

	switch b.Type {
	case internal.BlockCode:
		return b.Content
	case internal.BlockText:
		return ProcessText(strings.TrimSpace(b.Content), config.SafetyLevel)
	case internal.BlockTable:
		lines := strings.Split(b.Content, "\n")
		return ConvertTable(lines, config.AlignTableColumns, config.IgnoreTableSeparator)
	case internal.BlockTitle:
		return ProcessTitle(b.Content, config.SafetyLevel)
	case internal.BlockList:
		return ProcessList(b.Content, config.SafetyLevel)
	case internal.BlockQuote:
		return ProcessQuote(b.Content, config.SafetyLevel)
	default:
		return ProcessText(strings.TrimSpace(b.Content), config.SafetyLevel)
	}
}
