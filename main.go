package GoTeleMD

import (
	"github.com/sshturbo/GoTeleMD/internal"
	"github.com/sshturbo/GoTeleMD/pkg/formatter"
	"github.com/sshturbo/GoTeleMD/pkg/parser"
	"github.com/sshturbo/GoTeleMD/pkg/types"
	"github.com/sshturbo/GoTeleMD/pkg/utils"
)

const (
	SAFETYLEVELNONE   = internal.SAFETYLEVELNONE
	SAFETYLEVELBASIC  = internal.SAFETYLEVELBASIC
	SAFETYLEVELSTRICT = internal.SAFETYLEVELSTRICT
)

var (
	EnableLogs = false
)

func init() {
	internal.EnableLogs = &EnableLogs
	utils.InitLogger(&EnableLogs)
}

func Convert(input string, alignTableCols, ignoreTableSeparators bool, safetyLevel ...int) types.MessageResponse {
	level := internal.SAFETYLEVELBASIC
	if len(safetyLevel) > 0 {
		level = safetyLevel[0]
	}

	resultado := formatter.ConvertMarkdown(input, alignTableCols, ignoreTableSeparators, level)
	return parser.BreakLongText(resultado)
}
