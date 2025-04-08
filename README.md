# GoTeleMD

GoTeleMD é uma biblioteca Go para converter MarkdownV2 em formato compatível com Telegram MarkdownV2.

## Características

- Conversão completa de Markdown para Telegram MarkdownV2
- Processamento assíncrono e paralelo (automático)
- Divisão automática de mensagens longas
- Sistema de configuração flexível
- Logs detalhados para debug
- Formatação de tabelas com alinhamento
- Suporte para blocos de código com syntax highlighting
- Preservação inteligente de quebras de linha
- Alta performance com processamento paralelo

## Instalação

```bash
go get github.com/sshturbo/GoTeleMD@latest
```

## Uso Rápido

### Uso Básico
```go
package main

import (
    "fmt"
    "github.com/sshturbo/GoTeleMD"
    "github.com/sshturbo/GoTeleMD/pkg/types"
)

func main() {
    // Criar um conversor com configurações básicas
    converter := GoTeleMD.NewConverter(
        types.WithSafetyLevel(GoTeleMD.SAFETYLEVELBASIC),
        types.WithMaxMessageLength(4096),
        types.WithDebugLogs(true), // Opcional: ativa logs detalhados
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

### Uso Avançado (com configurações de performance)
```go
// Criar um conversor com configurações avançadas de performance
converter := GoTeleMD.NewConverter(
    // Configurações básicas
    types.WithSafetyLevel(GoTeleMD.SAFETYLEVELBASIC),
    types.WithMaxMessageLength(4096),
    
    // Configurações opcionais de performance
    types.WithNumWorkers(8),              // Opcional: padrão é 4
    types.WithWorkerQueueSize(64),        // Opcional: padrão é 32
    types.WithMaxConcurrentParts(4),      // Opcional: padrão é 8
    
    // Configurações de debug
    types.WithDebugLogs(true),            // Opcional
)
```

## Configurações Disponíveis

### Configurações Básicas (Obrigatórias)
- `WithSafetyLevel(level int)`: Define o nível de segurança da conversão
  - `SAFETYLEVELNONE`: Sem escape de caracteres especiais
  - `SAFETYLEVELBASIC`: Escape básico mantendo formatação
  - `SAFETYLEVELSTRICT`: Escape completo sem formatação
- `WithMaxMessageLength(length int)`: Define tamanho máximo de mensagem (padrão: 4096)

### Configurações de Performance (Opcionais)
Todas as configurações de performance são opcionais e já possuem valores padrão otimizados:

- `WithNumWorkers(num int)`: Define número de workers para processamento paralelo
  - **Padrão**: 4 workers
  - **Quando ajustar**: Aumente para textos muito grandes ou muitas conversões simultâneas
  - Se definido como 0, usa número de CPUs disponíveis

- `WithWorkerQueueSize(size int)`: Define tamanho da fila de trabalho por worker
  - **Padrão**: 32 tarefas
  - **Quando ajustar**: Aumente para melhor throughput com muitos blocos pequenos

- `WithMaxConcurrentParts(max int)`: Define máximo de partes processadas simultaneamente
  - **Padrão**: 8 partes
  - **Quando ajustar**: Diminua para controlar uso de memória em textos muito grandes

### Configurações de Debug (Opcionais)
- `WithDebugLogs(enable bool)`: Ativa/desativa logs detalhados de debug
  - Mostra informações sobre o processo de conversão
  - Exibe estrutura JSON das mensagens
  - Fornece métricas de tempo e tamanho
  - Detalha cada bloco processado
  - Mostra informações de workers e performance

## Sistema de Logs

Quando ativado com `WithDebugLogs(true)`, o sistema de logs mostra:

- 📊 Configurações utilizadas e métricas de performance
- 📝 Texto original recebido
- 📦 Estrutura da divisão em partes (JSON)
- 🔧 Informações de cada worker e seu processamento
- 🔍 Detalhes do processamento de cada parte
- 📤 Conteúdo formatado de cada parte
- ✅ Resumo final da conversão

## Performance e Concorrência

A biblioteca utiliza processamento paralelo automaticamente:

- **Configuração Automática**: Valores padrão otimizados para a maioria dos casos
- **Workers Paralelos**: 4 workers por padrão (ajustável se necessário)
- **Controle de Recursos**: Limites automáticos de memória e CPU
- **Escalabilidade**: Adapta-se ao número de cores disponíveis
- **Recuperação de Erros**: Tratamento robusto de falhas

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