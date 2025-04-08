// Package tgmarkdown fornece funcionalidades para converter markdown para o formato MarkdownV2 do Telegram
package tgmarkdown

import (
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

// Convert converte texto markdown para o formato MarkdownV2 do Telegram
// com suporte a divisão de mensagens longas e preservação de blocos de código.
// alignTableCols: alinha colunas de tabelas
// ignoreTableSeparators: ignora separadores de tabela
// safetyLevel: nível de segurança para escape de caracteres
func Convert(input string, alignTableCols, ignoreTableSeparators bool, safetyLevel ...int) string {
	level := internal.SAFETYLEVELBASIC
	if len(safetyLevel) > 0 {
		level = safetyLevel[0]
	}

	return formatter.ConvertMarkdown(input, alignTableCols, ignoreTableSeparators, level)
}
