// Package tgmarkdown fornece funcionalidades para converter markdown para o formato MarkdownV2 do Telegram
package tgmarkdown

import (
	"crypto/rand"
	"encoding/hex"
	"strings"

	"github.com/sshturbo/GoTeleMD/internal"
	"github.com/sshturbo/GoTeleMD/pkg/formatter"
	"github.com/sshturbo/GoTeleMD/pkg/utils"
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

	// Inicializa o logger
	utils.InitLogger(&EnableLogs)
}

// generateMessageID gera um ID único para a mensagem
func generateMessageID() string {
	b := make([]byte, 8)
	rand.Read(b)
	return hex.EncodeToString(b)
}

// Convert converte texto markdown para o formato MarkdownV2 do Telegram
// com suporte a divisão de mensagens longas e preservação de blocos de código.
// alignTableCols: alinha colunas de tabelas
// ignoreTableSeparators: ignora separadores de tabela
// safetyLevel: nível de segurança para escape de caracteres
// Retorna uma MessageResponse com as partes da mensagem formatadas
func Convert(input string, alignTableCols, ignoreTableSeparators bool, safetyLevel ...int) internal.MessageResponse {
	level := internal.SAFETYLEVELBASIC
	if len(safetyLevel) > 0 {
		level = safetyLevel[0]
	}

	// Converte o texto usando o formatter
	resultado := formatter.ConvertMarkdown(input, alignTableCols, ignoreTableSeparators, level)

	// Divide o texto em partes usando \n\n como separador
	partes := strings.Split(resultado, "\n\n")

	// Remove partes vazias e aplica trim
	var partesLimpas []string
	for _, parte := range partes {
		if trimmed := strings.TrimSpace(parte); trimmed != "" {
			partesLimpas = append(partesLimpas, trimmed)
		}
	}

	// Cria a estrutura da resposta
	messageParts := make([]internal.MessagePart, len(partesLimpas))
	for i, content := range partesLimpas {
		messageParts[i] = internal.MessagePart{
			Part:    i + 1,
			Content: content,
		}
	}

	return internal.MessageResponse{
		MessageID:  generateMessageID(),
		TotalParts: len(messageParts),
		Parts:      messageParts,
	}
}
