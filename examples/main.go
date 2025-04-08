package main

import (
	"log"
	"os"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	tgmarkdown "github.com/sshturbo/GoTeleMD"
	"github.com/sshturbo/GoTeleMD/pkg/types"
)

func init() {
	tgmarkdown.EnableLogs = true
}

// enviarMensagem envia uma única mensagem para o Telegram
func enviarMensagem(bot *tgbotapi.BotAPI, chatID int64, texto string) error {
	log.Printf("📤 Enviando mensagem única...")
	log.Printf("📝 Conteúdo da mensagem:\n%s", texto)
	msg := tgbotapi.NewMessage(chatID, texto)
	msg.ParseMode = "MarkdownV2"

	_, err := bot.Send(msg)
	if err != nil {
		log.Printf("❌ Erro ao enviar mensagem: %v", err)
		return err
	}

	log.Printf("✅ Mensagem enviada com sucesso")
	return nil
}

// enviarMensagemEmPartes envia uma mensagem dividida em partes
func enviarMensagemEmPartes(bot *tgbotapi.BotAPI, chatID int64, msgResponse types.MessageResponse) error {
	log.Printf("📨 Iniciando envio de mensagem em %d partes (ID: %s)...",
		msgResponse.TotalParts, msgResponse.MessageID)

	for _, parte := range msgResponse.Parts {
		log.Printf("📤 Enviando parte %d/%d...", parte.Part, msgResponse.TotalParts)
		log.Printf("📝 Conteúdo da parte %d:\n%s", parte.Part, parte.Content)

		msg := tgbotapi.NewMessage(chatID, parte.Content)
		msg.ParseMode = "MarkdownV2"

		_, err := bot.Send(msg)
		if err != nil {
			log.Printf("❌ Erro ao enviar parte %d: %v", parte.Part, err)
			return err
		}

		log.Printf("✅ Parte %d/%d enviada com sucesso", parte.Part, msgResponse.TotalParts)

		// Aguarda entre cada parte para evitar rate limiting
		if parte.Part < msgResponse.TotalParts {
			time.Sleep(500 * time.Millisecond)
		}
	}

	return nil
}

func main() {
	// Lendo o arquivo message.txt
	conteudo, err := os.ReadFile("message.txt")
	if err != nil {
		log.Fatalf("❌ Erro ao ler arquivo: %v", err)
	}

	// Processa o texto usando a lib tgmarkdown
	response := tgmarkdown.Convert(
		string(conteudo),
		false, // alignTableCols
		false, // ignoreTableSeparators
		tgmarkdown.SAFETYLEVELBASIC,
	)

	// Log do JSON para debug
	/*jsonBytes, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		log.Fatalf("❌ Erro ao gerar JSON: %v", err)
	}
	log.Printf("📋 Resposta da lib:\n%s", string(jsonBytes))*/

	// Configuração do bot
	const (
		token  = "6881016701:AAGXDGM-CILWRekjJg5C6ejSYlWL-9jY2II"
		chatID = 889168461
	)

	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Fatalf("❌ Erro ao criar bot: %v", err)
	}

	// Verifica se a mensagem precisa ser enviada em partes
	if response.TotalParts <= 1 {
		// Mensagem única - envia diretamente
		err = enviarMensagem(bot, chatID, response.Parts[0].Content)
	} else {
		// Múltiplas partes - envia sequencialmente
		err = enviarMensagemEmPartes(bot, chatID, response)
	}

	if err != nil {
		log.Fatalf("❌ Erro no processo de envio: %v", err)
	}

	log.Printf("✨ Processo concluído com sucesso!")
}


