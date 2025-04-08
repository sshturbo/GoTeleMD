package main

import (
	"encoding/json"
	"log"
	"os"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	tgmarkdown "github.com/sshturbo/GoTeleMD"
	"github.com/sshturbo/GoTeleMD/internal"
)

// Service representa o serviço de mensagens
type Service struct{}

func init() {
	tgmarkdown.EnableLogs = true
	tgmarkdown.TruncateInsteadOfBreak = false
}

func (s *Service) processarMensagem(text string) (*tgbotapi.Message, error) {
	log.Printf("📝 Processando mensagem...")

	// Converte e já recebe no formato MessageResponse
	response := tgmarkdown.Convert(
		text,
		false,
		false,
		tgmarkdown.SAFETYLEVELBASIC,
	)

	log.Printf("✅ Mensagem processada. Total de partes: %d", response.TotalParts)
	return &tgbotapi.Message{Text: response.MessageID}, nil
}

// enviarMensagens envia as partes da mensagem sequencialmente
func enviarMensagens(bot *tgbotapi.BotAPI, chatID int64, response internal.MessageResponse) error {
	// Primeiro envia o JSON com informações sobre a mensagem
	jsonBytes, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		return err
	}

	msg := tgbotapi.NewMessage(chatID, "```json\n"+string(jsonBytes)+"\n```")
	msg.ParseMode = "MarkdownV2"

	if _, err := bot.Send(msg); err != nil {
		return err
	}

	log.Printf("Enviando mensagem ID %s em %d partes...",
		response.MessageID, response.TotalParts)

	// Envia cada parte da mensagem
	for _, parte := range response.Parts {
		msg := tgbotapi.NewMessage(chatID, parte.Content)
		msg.ParseMode = "MarkdownV2"

		_, err := bot.Send(msg)
		if err != nil {
			log.Printf("❌ Erro ao enviar parte %d da mensagem: %v", parte.Part, err)
			return err
		}

		log.Printf("✅ Parte %d/%d enviada com sucesso",
			parte.Part, response.TotalParts)

		// Aguarda um pouco entre as mensagens para evitar rate limiting
		time.Sleep(500 * time.Millisecond)
	}

	return nil
}

func main() {
	service := &Service{}

	// Lendo o arquivo message.txt
	conteudo, err := os.ReadFile("message.txt")
	if err != nil {
		log.Fatalf("Erro ao ler arquivo: %v", err)
	}

	// Processa a mensagem e envia
	response := tgmarkdown.Convert(
		string(conteudo),
		false,
		false,
		tgmarkdown.SAFETYLEVELBASIC,
	)

	// Configuração do bot com credenciais fixas
	const (
		token  = "6881016701:AAGXDGM-CILWRekjJg5C6ejSYlWL-9jY2II"
		chatID = 889168461
	)

	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Fatalf("Erro ao criar bot: %v", err)
	}

	// Enviando as partes da mensagem
	err = enviarMensagens(bot, chatID, response)
	if err != nil {
		log.Fatalf("Erro ao enviar mensagem: %v", err)
	}
}
