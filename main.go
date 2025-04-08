// Package tgmarkdown fornece funcionalidades para converter markdown para o formato MarkdownV2 do Telegram
package tgmarkdown

import (
	"github.com/sshturbo/GoTeleMD/internal"
	"github.com/sshturbo/GoTeleMD/pkg/formatter"
	"github.com/sshturbo/GoTeleMD/pkg/parser"
	"github.com/sshturbo/GoTeleMD/pkg/types"
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
	EnableLogs = false // Ativa logs de debug
)

func init() {
	// Sincroniza as configurações com o pacote internal
	internal.EnableLogs = &EnableLogs

	// Inicializa o logger
	utils.InitLogger(&EnableLogs)
}

// Convert converte texto markdown para o formato MarkdownV2 do Telegram
// com suporte a divisão de mensagens longas e preservação de blocos de código.
// alignTableCols: alinha colunas de tabelas
// ignoreTableSeparators: ignora separadores de tabela
// safetyLevel: nível de segurança para escape de caracteres
// Retorna uma MessageResponse com as partes da mensagem formatadas
func Convert(input string, alignTableCols, ignoreTableSeparators bool, safetyLevel ...int) types.MessageResponse {
	level := internal.SAFETYLEVELBASIC
	if len(safetyLevel) > 0 {
		level = safetyLevel[0]
	}

	// Converte o texto usando o formatter
	resultado := formatter.ConvertMarkdown(input, alignTableCols, ignoreTableSeparators, level)

	// Usa o parser para dividir o texto de forma inteligente
	return parser.BreakLongText(resultado)
}
