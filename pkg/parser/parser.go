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

// generateMessageID gera um ID único para a mensagem
func generateMessageID() string {
	b := make([]byte, 8)
	rand.Read(b)
	return hex.EncodeToString(b)
}

// Tokenize breaks down the input text into blocks of different types
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

// BreakLongText divide texto longo em partes menores de forma inteligente
func BreakLongText(input string) types.MessageResponse {
	blocks := Tokenize(input)
	effectiveLimit := internal.TelegramMaxLength

	// Se o texto completo é menor que o limite, retorna em uma única parte
	if totalLen := utf8.RuneCountInString(input); totalLen <= effectiveLimit {
		return types.MessageResponse{
			MessageID:  generateMessageID(),
			TotalParts: 1,
			Parts: []types.MessagePart{
				{
					Part:    1,
					Content: input,
				},
			},
		}
	}

	var parts []string
	var currentGroup []internal.Block
	var currentLength int

	// Verifica se um bloco deve iniciar uma nova parte
	shouldStartNewPart := func(block internal.Block, nextBlock *internal.Block) bool {
		// Se o bloco atual é um título e o próximo bloco existe,
		// mantenha-os juntos a menos que exceda muito o limite
		if block.Type == internal.BlockTitle && nextBlock != nil {
			return currentLength > int(float64(effectiveLimit)*1.5) // Corrigido a conversão float para int
		}

		// Se é um bloco de código grande, cria uma nova parte
		if block.Type == internal.BlockCode && utf8.RuneCountInString(block.Content) > effectiveLimit/2 {
			return true
		}

		// Para outros tipos de blocos, só divide se realmente necessário
		return currentLength > effectiveLimit
	}

	// Função auxiliar para processar um grupo de blocos
	processGroup := func() {
		if len(currentGroup) == 0 {
			return
		}

		var groupContent strings.Builder
		for i, block := range currentGroup {
			if i > 0 {
				// Garante quebra de linha dupla entre blocos diferentes
				if block.Type == internal.BlockTitle {
					groupContent.WriteString("\n\n")
				} else if block.Type == internal.BlockCode {
					// Garante que blocos de código tenham quebras de linha adequadas
					if currentGroup[i-1].Type != internal.BlockCode {
						groupContent.WriteString("\n\n")
					}
				} else if currentGroup[i-1].Type == internal.BlockCode {
					groupContent.WriteString("\n\n")
				} else {
					groupContent.WriteString("\n\n")
				}
			}

			// Adiciona o conteúdo do bloco
			if block.Type == internal.BlockCode {
				// Garante que blocos de código tenham suas próprias linhas
				content := strings.TrimSpace(block.Content)
				if !strings.HasPrefix(content, "```") {
					content = "```\n" + content + "\n```"
				}
				groupContent.WriteString(content)
			} else {
				groupContent.WriteString(block.Content)
			}
		}

		content := groupContent.String()
		// Só cria uma nova parte se o conteúdo não for muito pequeno
		if utf8.RuneCountInString(content) > 50 { // Threshold mínimo
			parts = append(parts, content)
		} else {
			// Se for muito pequeno, tenta combinar com a parte anterior
			if len(parts) > 0 {
				lastPart := parts[len(parts)-1]
				// Garante quebra de linha dupla entre partes diferentes
				if !strings.HasSuffix(lastPart, "\n") {
					lastPart += "\n"
				}
				if !strings.HasPrefix(content, "\n") {
					content = "\n" + content
				}
				combinedContent := lastPart + "\n" + content

				if utf8.RuneCountInString(combinedContent) <= effectiveLimit {
					parts[len(parts)-1] = combinedContent
					return
				}
			}
			parts = append(parts, content)
		}

		currentGroup = nil
		currentLength = 0
	}

	// Processa os blocos
	for i := 0; i < len(blocks); i++ {
		block := blocks[i]
		nextBlock := (*internal.Block)(nil)
		if i+1 < len(blocks) {
			nextBlock = &blocks[i+1]
		}

		blockLength := utf8.RuneCountInString(block.Content)

		// Se é um bloco de código que excede o limite
		if block.Type == internal.BlockCode && blockLength > effectiveLimit {
			// Processa o grupo atual primeiro
			processGroup()
			// Divide o bloco de código
			codeParts := divideCodeBlock(block.Content, effectiveLimit)
			parts = append(parts, codeParts...)
			continue
		}

		// Calcula o tamanho que será adicionado (incluindo separadores)
		additionalLength := blockLength
		if len(currentGroup) > 0 {
			additionalLength += 2 // \n\n entre blocos
		}

		// Verifica se deve começar uma nova parte
		if shouldStartNewPart(block, nextBlock) {
			if len(currentGroup) > 0 {
				processGroup()
			}
			if block.Type == internal.BlockCode {
				// Bloco de código vai em uma parte separada
				parts = append(parts, block.Content)
				continue
			}
		}

		// Verifica se adicionar este bloco excederia o limite
		if currentLength+additionalLength > effectiveLimit {
			processGroup()
		}

		// Adiciona o bloco ao grupo atual
		currentGroup = append(currentGroup, block)
		currentLength += additionalLength
	}

	// Processa o último grupo se houver
	processGroup()

	// Converte as partes para o formato de resposta
	messageParts := make([]types.MessagePart, len(parts))
	for i, content := range parts {
		messageParts[i] = types.MessagePart{
			Part:    i + 1,
			Content: strings.TrimSpace(content),
		}
	}

	return types.MessageResponse{
		MessageID:  generateMessageID(),
		TotalParts: len(messageParts),
		Parts:      messageParts,
	}
}

// divideCodeBlock divide um bloco de código grande mantendo a sintaxe correta
func divideCodeBlock(content string, limit int) []string {
	lines := strings.Split(content, "\n")
	if len(lines) < 2 {
		return []string{content}
	}

	// Extrai a linguagem do bloco de código
	language := ""
	if strings.HasPrefix(lines[0], "```") {
		language = strings.TrimPrefix(lines[0], "```")
	}

	// Remove as marcações de início e fim
	codeContent := strings.Join(lines[1:len(lines)-1], "\n")

	// Calcula o limite efetivo considerando as marcações
	overhead := 6 // ```\n no início e \n``` no fim
	if language != "" {
		overhead += len(language)
	}
	effectiveLimit := limit - overhead

	// Divide o conteúdo em partes
	var parts []string
	var currentPart strings.Builder
	currentLength := 0

	for _, line := range strings.Split(codeContent, "\n") {
		lineLength := utf8.RuneCountInString(line) + 1 // +1 para o \n

		// Se a linha é maior que o limite efetivo
		if lineLength > effectiveLimit {
			// Se tem conteúdo pendente, salva primeiro
			if currentPart.Len() > 0 {
				parts = append(parts, formatCodeBlock(currentPart.String(), language))
				currentPart.Reset()
				currentLength = 0
			}

			// Divide a linha grande em partes menores
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

		// Se adicionar a linha atual ultrapassa o limite
		if currentLength+lineLength > effectiveLimit {
			parts = append(parts, formatCodeBlock(currentPart.String(), language))
			currentPart.Reset()
			currentLength = 0
		}

		// Adiciona a linha à parte atual
		if currentPart.Len() > 0 {
			currentPart.WriteString("\n")
		}
		currentPart.WriteString(line)
		currentLength += lineLength
	}

	// Adiciona a última parte se houver conteúdo
	if currentPart.Len() > 0 {
		parts = append(parts, formatCodeBlock(currentPart.String(), language))
	}

	return parts
}

// formatCodeBlock formata uma parte de código com as marcações apropriadas
func formatCodeBlock(content string, language string) string {
	content = strings.TrimSpace(content)
	if language != "" {
		return "```" + language + "\n" + content + "\n```"
	}
	return "```\n" + content + "\n```"
}
