package parser

import (
	"crypto/rand"
	"encoding/hex"
	"strings"
	"unicode/utf8"

	"github.com/sshturbo/GoTeleMD/internal"
	"github.com/sshturbo/GoTeleMD/pkg/types"
	"github.com/sshturbo/GoTeleMD/pkg/utils"
)

// Constantes para gerenciamento de tamanho
const (
	// Margem de segurança para caracteres de escape e formatação
	safetyMargin = 256
	// Tamanho mínimo para tentar manter em cada parte
	minPartSize = 512
)

func generateMessageID() string {
	b := make([]byte, 8)
	rand.Read(b)
	return hex.EncodeToString(b)
}

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

func BreakLongText(input string, maxLength int) (types.MessageResponse, error) {
	if maxLength <= 0 {
		maxLength = internal.TelegramMaxLength
	}

	// Ajusta o tamanho máximo considerando a margem de segurança
	effectiveMaxLength := maxLength - safetyMargin

	// Se o texto completo é menor que o limite, retorna uma única parte
	if totalLen := utf8.RuneCountInString(input); totalLen <= effectiveMaxLength {
		return types.MessageResponse{
			MessageID:  generateMessageID(),
			TotalParts: 1,
			Parts: []types.MessagePart{
				{
					Part:    1,
					Content: input,
				},
			},
		}, nil
	}

	// Divide o texto em blocos mantendo a ordem
	blocks := Tokenize(input)
	var parts []types.MessagePart
	var currentPart strings.Builder
	currentPartSize := 0
	partNumber := 1

	// Função para finalizar a parte atual
	flushCurrentPart := func() {
		if currentPart.Len() > 0 {
			content := strings.TrimSpace(currentPart.String())
			if len(content) > 0 {
				parts = append(parts, types.MessagePart{
					Part:    partNumber,
					Content: content,
				})
				partNumber++
				currentPart.Reset()
				currentPartSize = 0
			}
		}
	}

	for i, block := range blocks {
		blockContent := block.Content
		blockSize := utf8.RuneCountInString(blockContent)

		// Se o bloco é maior que o limite efetivo, divide ele
		if blockSize > effectiveMaxLength {
			// Primeiro, finaliza a parte atual se houver conteúdo
			flushCurrentPart()

			// Divide o bloco grande em partes menores
			if block.Type == internal.BlockCode {
				codeParts, err := divideCodeBlock(blockContent, effectiveMaxLength)
				if err != nil {
					return types.MessageResponse{}, err
				}
				for _, codePart := range codeParts {
					parts = append(parts, types.MessagePart{
						Part:    partNumber,
						Content: codePart,
					})
					partNumber++
				}
			} else {
				textParts, err := divideContent(blockContent, effectiveMaxLength)
				if err != nil {
					return types.MessageResponse{}, err
				}
				for _, textPart := range textParts {
					parts = append(parts, types.MessagePart{
						Part:    partNumber,
						Content: textPart,
					})
					partNumber++
				}
			}
			continue
		}

		// Calcula o tamanho necessário incluindo separadores
		neededSize := blockSize
		if currentPartSize > 0 {
			neededSize += 2 // Para \n\n entre blocos
		}

		// Se adicionar este bloco excederia o limite, inicia uma nova parte
		if currentPartSize+neededSize > effectiveMaxLength {
			flushCurrentPart()
		}

		// Adiciona o bloco à parte atual
		if currentPart.Len() > 0 {
			if block.Type == internal.BlockCode || blocks[i-1].Type == internal.BlockCode {
				currentPart.WriteString("\n\n")
			} else if block.Type == internal.BlockTitle || (i > 0 && blocks[i-1].Type == internal.BlockTitle) {
				currentPart.WriteString("\n\n")
			} else if block.Type == internal.BlockList || (i > 0 && blocks[i-1].Type == internal.BlockList) {
				currentPart.WriteString("\n\n")
			} else {
				currentPart.WriteString("\n\n")
			}
			currentPartSize += 2
		}

		currentPart.WriteString(blockContent)
		currentPartSize += blockSize
	}

	// Finaliza a última parte
	flushCurrentPart()

	return types.MessageResponse{
		MessageID:  generateMessageID(),
		TotalParts: len(parts),
		Parts:      parts,
	}, nil
}

// divideContent divide um conteúdo grande em partes menores
func divideContent(content string, maxLength int) ([]string, error) {
	var parts []string
	lines := strings.Split(content, "\n")
	var currentPart strings.Builder
	currentLength := 0

	flushPart := func() {
		if currentPart.Len() > 0 {
			parts = append(parts, strings.TrimSpace(currentPart.String()))
			currentPart.Reset()
			currentLength = 0
		}
	}

	for _, line := range lines {
		lineLength := utf8.RuneCountInString(line)

		// Se a linha é maior que o limite, divide ela
		if lineLength > maxLength {
			flushPart()

			// Divide a linha em partes menores
			runes := []rune(line)
			for i := 0; i < len(runes); i += maxLength {
				end := i + maxLength
				if end > len(runes) {
					end = len(runes)
				}
				parts = append(parts, string(runes[i:end]))
			}
			continue
		}

		// Se adicionar esta linha excederia o limite
		if currentLength+lineLength+2 > maxLength {
			flushPart()
		}

		// Adiciona a linha à parte atual
		if currentPart.Len() > 0 {
			currentPart.WriteString("\n")
			currentLength += 1
		}
		currentPart.WriteString(line)
		currentLength += lineLength
	}

	flushPart()
	return parts, nil
}

func divideCodeBlock(content string, maxLength int) ([]string, error) {
	lines := strings.Split(content, "\n")
	if len(lines) < 2 {
		return []string{content}, nil
	}

	language := ""
	if strings.HasPrefix(lines[0], "```") {
		language = strings.TrimPrefix(lines[0], "```")
	}

	codeContent := strings.Join(lines[1:len(lines)-1], "\n")

	overhead := 6
	if language != "" {
		overhead += len(language)
	}
	effectiveLimit := maxLength - overhead

	var parts []string
	var currentPart strings.Builder
	currentLength := 0

	for _, line := range strings.Split(codeContent, "\n") {
		lineLength := utf8.RuneCountInString(line) + 1

		if lineLength > effectiveLimit {
			if currentPart.Len() > 0 {
				parts = append(parts, formatCodeBlock(currentPart.String(), language))
				currentPart.Reset()
				currentLength = 0
			}

			var linePart strings.Builder
			runes := []rune(line)
			for i := 0; i < len(runes); i += effectiveLimit {
				end := i + effectiveLimit
				if end > len(runes) {
					end = len(runes)
				}
				linePart.Reset()
				linePart.WriteString(string(runes[i:end]))
				parts = append(parts, formatCodeBlock(linePart.String(), language))
			}
			continue
		}

		if currentLength+lineLength > effectiveLimit {
			parts = append(parts, formatCodeBlock(currentPart.String(), language))
			currentPart.Reset()
			currentLength = 0
		}

		if currentPart.Len() > 0 {
			currentPart.WriteString("\n")
		}
		currentPart.WriteString(line)
		currentLength += lineLength
	}

	if currentPart.Len() > 0 {
		parts = append(parts, formatCodeBlock(currentPart.String(), language))
	}

	return parts, nil
}

func formatCodeBlock(content string, language string) string {
	content = strings.TrimSpace(content)
	if language != "" {
		return "```" + language + "\n" + content + "\n```"
	}
	return "```\n" + content + "\n```"
}
