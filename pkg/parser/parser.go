package parser

import (
	"strings"
	"unicode/utf8"

	"github.com/sshturbo/GoTeleMD/internal"
	"github.com/sshturbo/GoTeleMD/pkg/utils"
)

// Tokenize breaks down the input text into blocks of different types (code, text, table, etc.).
// It handles special markdown syntax like code blocks, tables, titles, and lists.
func Tokenize(input string) []internal.Block {
	var blocks []internal.Block
	lines := strings.Split(input, "\n")
	var buffer []string
	inCodeBlock := false
	var currentBlockType internal.BlockType = internal.BlockText

	flushBuffer := func() {
		if len(buffer) > 0 {
			content := strings.Join(buffer, "\n")
			if currentBlockType == internal.BlockCode {
				blocks = append(blocks, internal.Block{Type: currentBlockType, Content: content})
			} else {
				blocks = append(blocks, internal.Block{Type: currentBlockType, Content: strings.TrimSpace(content)})
			}
			buffer = []string{}
			currentBlockType = internal.BlockText
		}
	}

	for i := 0; i < len(lines); i++ {
		line := lines[i]

		if strings.HasPrefix(line, "```") {
			if inCodeBlock {
				inCodeBlock = false
				buffer = append(buffer, line)
				flushBuffer()
			} else {
				if len(buffer) > 0 {
					flushBuffer()
				}
				inCodeBlock = true
				currentBlockType = internal.BlockCode
				buffer = append(buffer, line)
			}
			continue
		}

		if inCodeBlock {
			buffer = append(buffer, line)
			continue
		}

		if utils.TableLinePattern.MatchString(line) {
			flushBuffer()
			currentBlockType = internal.BlockTable
			tableBlock := []string{line}
			for i+1 < len(lines) && utils.TableLinePattern.MatchString(lines[i+1]) {
				i++
				tableBlock = append(tableBlock, lines[i])
			}
			blocks = append(blocks, internal.Block{Type: internal.BlockTable, Content: strings.Join(tableBlock, "\n")})
			continue
		}

		if utils.TitlePattern.MatchString(line) {
			flushBuffer()
			blocks = append(blocks, internal.Block{Type: internal.BlockTitle, Content: line})
			continue
		}

		if utils.ListItemPattern.MatchString(line) || utils.OrderedListPattern.MatchString(line) {
			if currentBlockType != internal.BlockList {
				flushBuffer()
				currentBlockType = internal.BlockList
			}
			buffer = append(buffer, line)
			continue
		}

		if utils.BlockquotePattern.MatchString(line) {
			if currentBlockType != internal.BlockQuote {
				flushBuffer()
				currentBlockType = internal.BlockQuote
			}
			buffer = append(buffer, line)
			continue
		}

		if strings.TrimSpace(line) == "" && currentBlockType != internal.BlockText {
			flushBuffer()
		}

		buffer = append(buffer, line)
	}

	flushBuffer()
	return blocks
}

// BreakLongText splits long text into smaller parts while preserving code blocks intact.
// It ensures that:
// 1. Code blocks are never split - they are kept as complete units
// 2. Regular text is split at appropriate boundaries when needed
// 3. The resulting parts do not exceed Telegram's message length limit
func BreakLongText(input string) []string {
	effectiveLimit := internal.TelegramMaxLength

	if utf8.RuneCountInString(input) <= effectiveLimit {
		return []string{input}
	}

	var result []string
	var currentPart strings.Builder
	var codeBuffer strings.Builder
	var inCodeBlock bool

	lines := strings.Split(strings.TrimSpace(input), "\n")

	for i := 0; i < len(lines); i++ {
		line := lines[i]
		isCodeDelimiter := strings.HasPrefix(line, "```")

		if isCodeDelimiter {
			if inCodeBlock {
				// Fechando bloco de código
				if err := handleCodeBlockEnd(&codeBuffer, &currentPart, line, &result); err != nil {
					utils.LogError("Erro ao processar fim do bloco de código: %v", err)
				}
				inCodeBlock = false
			} else {
				// Iniciando bloco de código
				if err := handleCodeBlockStart(&currentPart, &codeBuffer, line, &result); err != nil {
					utils.LogError("Erro ao iniciar bloco de código: %v", err)
				}
				inCodeBlock = true
			}
			continue
		}

		if inCodeBlock {
			codeBuffer.WriteString(line)
			if i < len(lines)-1 {
				codeBuffer.WriteString("\n")
			}
			continue
		}

		// Processamento de texto normal
		if err := handleRegularText(&currentPart, line, effectiveLimit, &result); err != nil {
			utils.LogError("Erro ao processar texto regular: %v", err)
		}
	}

	if currentPart.Len() > 0 {
		result = append(result, strings.TrimSpace(currentPart.String()))
	}

	return result
}

// handleCodeBlockEnd processa o final de um bloco de código
func handleCodeBlockEnd(codeBuffer, currentPart *strings.Builder, line string, result *[]string) error {
	codeBuffer.WriteString(line)
	codeContent := codeBuffer.String()

	if currentPart.Len() > 0 {
		*result = append(*result, currentPart.String())
		currentPart.Reset()
	}

	// Se o conteúdo do bloco de código for maior que o limite, divide em partes
	if utf8.RuneCountInString(codeContent) > internal.TelegramMaxLength {
		lines := strings.Split(codeContent, "\n")
		language := ""

		// Extrai a linguagem se especificada
		if strings.HasPrefix(lines[0], "```") && len(lines[0]) > 3 {
			language = strings.TrimPrefix(lines[0], "```")
		}

		var currentBlock strings.Builder
		currentBlock.WriteString("```")
		if language != "" {
			currentBlock.WriteString(language)
		}
		currentBlock.WriteString("\n")

		for i := 1; i < len(lines)-1; i++ { // Ignora primeira e última linha (```)
			line := lines[i] + "\n"
			if currentBlock.Len()+len(line) > internal.TelegramMaxLength-4 { // -4 para o fechamento do bloco
				currentBlock.WriteString("```")
				*result = append(*result, currentBlock.String())
				currentBlock.Reset()
				currentBlock.WriteString("```")
				if language != "" {
					currentBlock.WriteString(language)
				}
				currentBlock.WriteString("\n")
			}
			currentBlock.WriteString(line)
		}

		currentBlock.WriteString("```")
		*result = append(*result, currentBlock.String())
	} else {
		*result = append(*result, codeContent)
	}

	codeBuffer.Reset()
	return nil
}

// handleCodeBlockStart processa o início de um bloco de código
func handleCodeBlockStart(currentPart, codeBuffer *strings.Builder, line string, result *[]string) error {
	if currentPart.Len() > 0 {
		*result = append(*result, currentPart.String())
		currentPart.Reset()
	}

	codeBuffer.WriteString(line)
	codeBuffer.WriteString("\n")
	return nil
}

// handleRegularText processa texto normal (fora de blocos de código)
func handleRegularText(currentPart *strings.Builder, line string, limit int, result *[]string) error {
	lineLen := utf8.RuneCountInString(line)
	if currentPart.Len()+lineLen+1 > limit {
		if lineLen > limit {
			parts := splitLongLine(line, limit)
			for _, part := range parts {
				if currentPart.Len() > 0 {
					*result = append(*result, currentPart.String())
					currentPart.Reset()
				}
				*result = append(*result, part)
			}
		} else {
			*result = append(*result, currentPart.String())
			currentPart.Reset()
			currentPart.WriteString(line)
		}
	} else {
		if currentPart.Len() > 0 {
			currentPart.WriteString("\n")
		}
		currentPart.WriteString(line)
	}
	return nil
}

// splitLongLine divide uma linha longa em partes menores respeitando
// o limite de caracteres e tentando quebrar em espaços entre palavras
func splitLongLine(input string, limit int) []string {
	if utf8.RuneCountInString(input) <= limit {
		return []string{input}
	}

	var parts []string
	words := strings.Fields(input)
	currentPart := strings.Builder{}

	for _, word := range words {
		wordLen := utf8.RuneCountInString(word)

		if wordLen > limit {
			if currentPart.Len() > 0 {
				parts = append(parts, currentPart.String())
				currentPart.Reset()
			}
			wordRunes := []rune(word)
			for i := 0; i < len(wordRunes); i += limit {
				end := i + limit
				if end > len(wordRunes) {
					end = len(wordRunes)
				}
				parts = append(parts, string(wordRunes[i:end]))
			}
		} else if currentPart.Len() > 0 && currentPart.Len()+wordLen+1 > limit {
			parts = append(parts, currentPart.String())
			currentPart.Reset()
			currentPart.WriteString(word)
		} else {
			if currentPart.Len() > 0 {
				currentPart.WriteString(" ")
			}
			currentPart.WriteString(word)
		}
	}

	if currentPart.Len() > 0 {
		parts = append(parts, currentPart.String())
	}

	return parts
}
