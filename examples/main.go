package main

import (
	"log"
	"os"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	tgmarkdown "github.com/sshturbo/GoTeleMD"
)

// Service representa o serviço de mensagens
type Service struct{}

func init() {
	tgmarkdown.EnableLogs = true
	tgmarkdown.TruncateInsteadOfBreak = false
}

func (s *Service) escapeCodeTags(text string) (string, error) {
	log.Printf("📝 Texto antes do escape: %s", text)

	// Primeiro processar os blocos de código

	// Depois aplicar o conversor geral do tgmarkdown
	result := tgmarkdown.Convert(
		text,
		false,
		false,
		tgmarkdown.SAFETYLEVELBASIC,
	)

	log.Printf("✅ Texto após escape: %s", result)
	return result, nil
}

func main() {
	service := &Service{}

	// Lendo o arquivo message.txt
	conteudo, err := os.ReadFile("message.txt")
	if err != nil {
		log.Fatalf("Erro ao ler arquivo: %v", err)
	}

	resultado, err := service.escapeCodeTags(string(conteudo))
	if err != nil {
		log.Fatalf("Erro ao escapar texto: %v", err)
	}

	// Configuração do bot com credenciais fixas
	const (
		token  = "6881016701:AAGXDGM-CILWRekjJg5C6ejSYlWL-9jY2II"
		chatID = 889168461
	)

	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Fatalf("Erro ao criar bot: %v", err)
	}

	// Enviando mensagem
	msg := tgbotapi.NewMessage(chatID, resultado)
	msg.ParseMode = "MarkdownV2"

	_, err = bot.Send(msg)
	if err != nil {
		log.Fatalf("Erro ao enviar mensagem: %v", err)
	}
}
