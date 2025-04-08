package formatter

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/sshturbo/GoTeleMD/internal"
	"github.com/sshturbo/GoTeleMD/pkg/parser"
	"github.com/sshturbo/GoTeleMD/pkg/types"
	"github.com/sshturbo/GoTeleMD/pkg/utils"
)

func ConvertMarkdown(input string, config *types.Config) (string, error) {
	if input == "" {
		return "", types.NewError(types.ErrInvalidInput, "input cannot be empty", nil)
	}

	startTime := time.Now()
	defer func() {
		utils.LogPerformance("Convert total", time.Since(startTime))
	}()

	utils.LogDebug("🔄 Iniciando conversão do Markdown")
	utils.LogDebug("📊 Configurações:")
	utils.LogDebug("   - Nível de segurança: %d", config.SafetyLevel)
	utils.LogDebug("   - Tamanho máximo: %d", config.MaxMessageLength)
	utils.LogDebug("   - Alinhamento de tabelas: %v", config.AlignTableColumns)

	utils.LogDebug("📝 Texto original:\n%s", input)

	response, err := parser.BreakLongText(strings.TrimSpace(input), config.MaxMessageLength)
	if err != nil {
		return "", types.NewError(types.ErrProcessingFailed, "failed to break text", err)
	}

	if config.EnableDebugLogs {
		jsonResponse, _ := json.MarshalIndent(response, "", "  ")
		utils.LogDebug("📦 Divisão em partes:")
		utils.LogDebug("   - Total de partes: %d", response.TotalParts)
		utils.LogDebug("   - ID da mensagem: %s", response.MessageID)
		utils.LogDebug("   - Estrutura JSON:\n%s", string(jsonResponse))
	}

	var outputParts []string
	for _, part := range response.Parts {
		blocks := parser.Tokenize(strings.TrimSpace(part.Content))
		var output strings.Builder
		output.Grow(len(part.Content))

		utils.LogDebug("🔍 Processando parte %d/%d", part.Part, response.TotalParts)
		utils.LogDebug("   - Blocos encontrados: %d", len(blocks))

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

			blockStartTime := time.Now()
			rendered := RenderBlock(b, config)
			utils.LogDebug("   - Bloco %d: Tipo=%d, Tamanho=%d, Tempo=%v",
				i+1, b.Type, len(rendered), time.Since(blockStartTime))
			output.WriteString(rendered)
		}

		formattedContent := strings.TrimSpace(output.String())
		outputParts = append(outputParts, formattedContent)

		if config.EnableDebugLogs {
			utils.LogDebug("📤 Parte %d formatada:", part.Part)
			utils.LogDebug("   - Tamanho: %d caracteres", len(formattedContent))
			utils.LogDebug("   - Conteúdo:\n%s", formattedContent)
		}
	}

	result := strings.TrimSpace(strings.Join(outputParts, "\n\n"))
	utils.LogDebug("✅ Conversão finalizada")
	utils.LogDebug("   - Tamanho final: %d caracteres", len(result))
	utils.LogDebug("   - Partes geradas: %d", len(outputParts))

	return result, nil
}
