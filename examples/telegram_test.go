package examples

import (
	"log"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers"
	"github.com/sshturbo/GoTeleMD"
)

const (
	// Substitua estes valores pelos seus
	botToken  = "6881016701:AAGXDGM-CILWRekjJg5C6ejSYlWL-9jY2II"   // Token do seu bot
	chatIDStr = "889168461" // ID do chat para testes (será convertido para int64)
)

// Service representa o serviço de mensagens
type Service struct{}

func (s *Service) escapeCodeTags(text string) (string, error) {
	log.Printf("📝 Texto antes do escape: %s", text)

	// Converter o texto com nível de segurança específico
	result := tgmarkdown.Convert(
		text,                        // texto de entrada
		false,                       // alignTableCols
		false,                       // ignoreTableSeparators
		tgmarkdown.SAFETYLEVELBASIC, // nível de segurança
	)

	log.Printf("✅ Texto após escape: %s", result)
	return result, nil
}

func init() {
	// Configurações globais do tgmarkdown
	tgmarkdown.EnableLogs = true
	tgmarkdown.TruncateInsteadOfBreak = false
}

func TestTelegramMarkdown(t *testing.T) {
	// Converte chatID para int64
	chatID, err := strconv.ParseInt(chatIDStr, 10, 64)
	if err != nil {
		t.Fatalf("Erro ao converter chatID: %v", err)
	}

	// Crie uma nova instância do bot e do serviço
	bot, err := gotgbot.NewBot(botToken, &gotgbot.BotOpts{})
	if err != nil {
		t.Fatalf("Erro ao criar bot: %v", err)
	}

	service := &Service{}

	// Lê o conteúdo do arquivo message.txt
	content, err := os.ReadFile(filepath.Join("message.txt"))
	if err != nil {
		t.Fatalf("Erro ao ler arquivo message.txt: %v", err)
	}

	// Primeiro passa pelo escapeCodeTags
	textoFormatado, err := service.escapeCodeTags(string(content))
	if err != nil {
		t.Fatalf("Erro ao escapar texto: %v", err)
	}

	// Envie a mensagem
	_, err = bot.SendMessage(chatID, textoFormatado, &gotgbot.SendMessageOpts{
		ParseMode: "MarkdownV2",
	})

	if err != nil {
		t.Errorf("Erro ao enviar mensagem: %v", err)
	}
}

// TestTelegramMarkdownInterativo é uma função auxiliar para testes manuais
func TestTelegramMarkdownInterativo(t *testing.T) {
	bot, err := gotgbot.NewBot(botToken, &gotgbot.BotOpts{})
	if err != nil {
		t.Fatalf("Erro ao criar bot: %v", err)
	}

	service := &Service{}

	// Crie um updater e dispatcher para o bot
	dispatcher := ext.NewDispatcher(&ext.DispatcherOpts{
		// Opções do dispatcher, se necessário
	})

	// Handler para comandos /markdown
	dispatcher.AddHandler(handlers.NewCommand("markdown", func(b *gotgbot.Bot, ctx *ext.Context) error {
		texto := ctx.EffectiveMessage.Text
		if len(texto) > 9 { // Remove "/markdown "
			texto = texto[9:]
		} else {
			// Tenta ler do arquivo message.txt se nenhum texto foi fornecido
			content, err := os.ReadFile(filepath.Join("message.txt"))
			if err != nil {
				texto = "Envie um texto após o comando /markdown ou crie um arquivo message.txt"
			} else {
				texto = string(content)
			}
		}

		// Primeiro passa pelo escapeCodeTags
		textoFormatado, err := service.escapeCodeTags(texto)
		if err != nil {
			log.Printf("Erro ao escapar texto: %v", err)
			return err
		}

		_, err = ctx.EffectiveMessage.Reply(bot, textoFormatado, &gotgbot.SendMessageOpts{
			ParseMode: "MarkdownV2",
		})

		if err != nil {
			log.Printf("Erro ao enviar mensagem: %v", err)
		}
		return nil
	}))

	updater := ext.NewUpdater(dispatcher, &ext.UpdaterOpts{
		ErrorLog: nil,
	})

	// Inicie o bot
	err = updater.StartPolling(bot, &ext.PollingOpts{
		DropPendingUpdates: true,
	})
	if err != nil {
		t.Fatalf("Erro ao iniciar polling: %v", err)
	}

	// Mantenha o bot rodando por um tempo para testes manuais
	log.Println("Bot iniciado! Use /markdown <texto> para testar ou use /markdown sem texto para usar o conteúdo de message.txt")

	// Idle sem contexto
	updater.Idle()
}
