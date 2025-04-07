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

## Uso

```go
import "github.com/sshturbo/GoTeleMD"

// Habilitar logs para debug (opcional)
tgmarkdown.EnableLogs = true

// Configurar comportamento de quebra (opcional)
tgmarkdown.TruncateInsteadOfBreak = false

// Converter markdown com opções básicas
texto := `# Título
**Negrito** e _itálico_`
converted := tgmarkdown.Convert(texto, false, false, tgmarkdown.SAFETYLEVELBASIC)

// Converter markdown com todas as opções explicadas
converted = tgmarkdown.Convert(
    texto,           // texto em markdown para converter
    true,           // alignTableCols: alinhar colunas das tabelas
    false,          // ignoreTableSeparators: manter linhas separadoras das tabelas
    tgmarkdown.SAFETYLEVELBASIC, // nível de segurança
)
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
texto := "**Negrito** _itálico_ ~~riscado~~ [link](https://exemplo.com)"
resultado := tgmarkdown.Convert(texto, false, false, true)
```

### Mensagens Longas
A biblioteca quebra automaticamente mensagens longas respeitando o limite do Telegram:

```go
textoLongo := strings.Repeat("Texto muito longo... ", 100)
partes := tgmarkdown.Convert(textoLongo, false, false, true)
// Resultado será quebrado em partes menores que 4096 caracteres
```