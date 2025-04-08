# tgmarkdown

Conversor de Markdown estilo GitHub para MarkdownV2 do Telegram, com suporte a:

- Tabelas alinhadas (esquerda `:---`, centro `:---:`, direita `---:`)
- Códigos (inline e bloco)
- Negrito, itálico, riscado, links (com suporte a formatação interna)
- Listas ordenadas e não-ordenadas
- Citações (blockquotes)
- Títulos (H1-H2 em negrito, H3-H6 em itálico)
- Escape automático de caracteres
- Quebra segura de mensagens grandes
- Logs ativáveis para debug
- Níveis de segurança configuráveis

## Instalação

```bash
go get github.com/sshturbo/GoTeleMD@v0.1.0
```


## Configurações

### Variáveis Globais
- `EnableLogs`: ativa logs de debug (default: false)
- `TruncateInsteadOfBreak`: trunca texto ao invés de quebrar em pontos seguros (default: false)
- `MaxWordLength`: tamanho máximo de palavra antes de forçar quebra (default: 200)

### Níveis de Segurança
- `SAFETYLEVELNONE`: sem escape de caracteres especiais
- `SAFETYLEVELBASIC`: escape básico mantendo formatação (padrão)
- `SAFETYLEVELSTRICT`: escape completo sem formatação

## Exemplos

### Tabelas
```go
texto := `| Nome  | Idade |
|:------:|------:|
| João   | 25    |
| Maria  | 30    |`

// Tabela com alinhamento (centro para Nome, direita para Idade)
resultado := tgmarkdown.Convert(texto, true, false, tgmarkdown.SAFETYLEVELBASIC)
```

### Listas
```go
texto := `- Item não numerado
- Outro item
1. Item numerado
2. Outro numerado`
resultado := tgmarkdown.Convert(texto, false, false, tgmarkdown.SAFETYLEVELBASIC)
```

### Citações
```go
texto := `> Uma citação simples
> Com **formatação** em _markdown_`
resultado := tgmarkdown.Convert(texto, false, false, tgmarkdown.SAFETYLEVELBASIC)
```

### Links com Formatação
```go
texto := `[Link com **negrito** e _itálico_](https://exemplo.com)`
resultado := tgmarkdown.Convert(texto, false, false, tgmarkdown.SAFETYLEVELBASIC)
```

### Código
```go
texto := "Código `inline` e bloco:\n```go\nfmt.Println(\"olá\")\n```"
resultado := tgmarkdown.Convert(texto, false, false, tgmarkdown.SAFETYLEVELBASIC)
```

### Formatação
```go
texto := "**Negrito** _itálico_ ~~riscado~~ [link](https://exemplo.com)"
resultado := tgmarkdown.Convert(texto, false, false, tgmarkdown.SAFETYLEVELBASIC)
```

### Mensagens Longas
A biblioteca quebra automaticamente mensagens longas respeitando o limite do Telegram:

```go
textoLongo := strings.Repeat("Texto muito longo... ", 100)
response := tgmarkdown.Convert(textoLongo, false, false, tgmarkdown.SAFETYLEVELBASIC)
// Resultado será quebrado em partes menores que 4096 caracteres
```

### Exemplo uso
Exemplo completo de como enviar mensagens longas divididas em partes:

```go
import (
    "log"
    "time"
    tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
    tgmarkdown "github.com/sshturbo/GoTeleMD"
)

func init() {
    tgmarkdown.EnableLogs = true
}

// Função para enviar uma única mensagem
func enviarMensagem(bot *tgbotapi.BotAPI, chatID int64, texto string) error {
    msg := tgbotapi.NewMessage(chatID, texto)
    msg.ParseMode = "MarkdownV2"

    _, err := bot.Send(msg)
    if err != nil {
        log.Printf("❌ Erro ao enviar mensagem: %v", err)
        return err
    }
    return nil
}

// Função para enviar mensagem dividida em partes
func enviarMensagemEmPartes(bot *tgbotapi.BotAPI, chatID int64, msgResponse tgmarkdown.MessageResponse) error {
    log.Printf("📨 Iniciando envio de mensagem em %d partes (ID: %s)...",
        msgResponse.TotalParts, msgResponse.MessageID)

    for _, parte := range msgResponse.Parts {
        log.Printf("📤 Enviando parte %d/%d...", parte.Part, msgResponse.TotalParts)

        msg := tgbotapi.NewMessage(chatID, parte.Content)
        msg.ParseMode = "MarkdownV2"

        _, err := bot.Send(msg)
        if err != nil {
            log.Printf("❌ Erro ao enviar parte %d: %v", parte.Part, err)
            return err
        }

        // Aguarda entre cada parte para evitar rate limiting
        if parte.Part < msgResponse.TotalParts {
            time.Sleep(500 * time.Millisecond)
        }
    }
    return nil
}

func main() {
    // Configuração do bot
    bot, err := tgbotapi.NewBotAPI("SEU_TOKEN_AQUI")
    if err != nil {
        log.Fatal(err)
    }

    // Texto longo para enviar
    textoLongo := `# Título Grande
    
Um texto muito longo com **formatação** em _markdown_...
    
## Código de Exemplo
    
\`\`\`go
func exemplo() {
    fmt.Println("Olá, mundo!")
}
\`\`\`
`
    // Converte o texto usando a lib
    response := tgmarkdown.Convert(
        textoLongo,
        false,          // alignTableCols
        false,          // ignoreTableSeparators
        tgmarkdown.SAFETYLEVELBASIC,
    )

    // Verifica se precisa enviar em partes
    chatID := int64(123456789) // ID do chat/grupo/canal
    if response.TotalParts <= 1 {
        // Mensagem única
        err = enviarMensagem(bot, chatID, response.Parts[0].Content)
    } else {
        // Múltiplas partes
        err = enviarMensagemEmPartes(bot, chatID, response)
    }

    if err != nil {
        log.Fatalf("❌ Erro no processo de envio: %v", err)
    }
}
```

Este exemplo mostra:
- Como converter textos longos usando a biblioteca
- Como tratar o envio de mensagens únicas e múltiplas partes
- Como implementar delay entre as partes para evitar rate limiting do Telegram
- Como usar o modo MarkdownV2 corretamente
- Como tratar erros durante o envio

**Importante:**
- Use `time.Sleep(500 * time.Millisecond)` entre as partes para evitar rate limiting
- Sempre verifique `response.TotalParts` para decidir o método de envio
- Use `msg.ParseMode = "MarkdownV2"` para formatação correta
- Trate os erros de envio adequadamente