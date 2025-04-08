# GoTeleMD

GoTeleMD é uma biblioteca Go para converter Markdown em formato compatível com Telegram MarkdownV2.

## Características

- Conversão completa de Markdown para Telegram MarkdownV2
- Suporte para todos os elementos Markdown comuns
- Divisão automática de mensagens longas
- Sistema de configuração flexível
- Logs detalhados para debug
- Formatação de tabelas com alinhamento
- Suporte para blocos de código com syntax highlighting
- Preservação inteligente de quebras de linha

## Instalação

```bash
go get github.com/sshturbo/GoTeleMD@latest
```

## Uso Rápido

```go
package main

import (
    "fmt"
    "github.com/sshturbo/GoTeleMD"
    "github.com/sshturbo/GoTeleMD/pkg/types"
)

func main() {
    // Criar um conversor com configurações personalizadas
    converter := GoTeleMD.NewConverter(
        types.WithSafetyLevel(GoTeleMD.SAFETYLEVELBASIC),
        types.WithMaxMessageLength(4096),
        types.WithDebugLogs(true), // Ativa logs detalhados
    )

    // Converter markdown
    response, err := converter.Convert("# Título\nTexto em **negrito** e _itálico_")
    if err != nil {
        panic(err)
    }

    // Usar as partes convertidas
    for _, part := range response.Parts {
        fmt.Printf("Parte %d: %s\n", part.Part, part.Content)
    }
}
```

## Configurações Disponíveis

- `WithSafetyLevel(level int)`: Define o nível de segurança da conversão
  - `SAFETYLEVELNONE`: Sem escape de caracteres especiais
  - `SAFETYLEVELBASIC`: Escape básico mantendo formatação
  - `SAFETYLEVELSTRICT`: Escape completo sem formatação

- `WithMaxMessageLength(length int)`: Define tamanho máximo de mensagem (padrão: 4096)
- `WithDebugLogs(enable bool)`: Ativa/desativa logs detalhados de debug
  - Mostra informações sobre o processo de conversão
  - Exibe estrutura JSON das mensagens
  - Fornece métricas de tempo e tamanho
  - Detalha cada bloco processado

## Sistema de Logs

Quando ativado com `WithDebugLogs(true)`, o sistema de logs mostra:

- 📊 Configurações utilizadas
- 📝 Texto original recebido
- 📦 Estrutura da divisão em partes (JSON)
- 🔍 Detalhes do processamento de cada parte
- 📤 Conteúdo formatado de cada parte
- ✅ Resumo final da conversão

## Elementos Suportados

- Títulos (H1-H6)
- Texto em negrito e itálico
- Links
- Listas ordenadas e não ordenadas
- Blocos de código (com e sem highlight)
- Tabelas (com alinhamento)
- Citações
- Texto riscado

## Tratamento de Erros

A biblioteca fornece tipos de erro específicos para melhor tratamento:

```go
switch err := err.(type) {
case *types.Error:
    switch err.Type {
    case types.ErrInvalidInput:
        // Tratar erro de entrada inválida
    case types.ErrInvalidFormat:
        // Tratar erro de formato
    case types.ErrMessageTooLong:
        // Tratar erro de mensagem muito longa
    case types.ErrProcessingFailed:
        // Tratar erro de processamento
    }
}
```

## Contribuindo

Contribuições são bem-vindas! Por favor, leia nossas diretrizes de contribuição antes de submeter pull requests.

## Licença

Este projeto está licenciado sob a MIT License - veja o arquivo LICENSE para detalhes.