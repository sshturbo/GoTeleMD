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

type Converter struct {
	config *types.Config
}

func NewConverter(options ...types.Option) *Converter {
	config := types.DefaultConfig()
	for _, opt := range options {
		opt(config)
	}

	utils.InitLogger(&config.EnableDebugLogs)
	return &Converter{config: config}
}

func (c *Converter) Convert(input string) (types.MessageResponse, error) {
	if input == "" {
		return types.MessageResponse{}, types.NewError(types.ErrInvalidInput, "input cannot be empty", nil)
	}

	resultado, err := formatter.ConvertMarkdown(input, c.config)
	if err != nil {
		return types.MessageResponse{}, err
	}
	return parser.BreakLongText(resultado, c.config.MaxMessageLength)
}

// Deprecated: Use NewConverter and Convert instead
func Convert(input string, alignTableCols, ignoreTableSeparators bool, safetyLevel ...int) types.MessageResponse {
	level := internal.SAFETYLEVELBASIC
	if len(safetyLevel) > 0 {
		level = safetyLevel[0]
	}

	config := types.DefaultConfig()
	config.SafetyLevel = level
	config.AlignTableColumns = alignTableCols
	config.IgnoreTableSeparator = ignoreTableSeparators

	conv := &Converter{config: config}
	response, _ := conv.Convert(input)
	return response
}
