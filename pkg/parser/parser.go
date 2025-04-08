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

func BreakLongText(input string) types.MessageResponse {
	blocks := Tokenize(input)
	effectiveLimit := internal.TelegramMaxLength

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

	shouldStartNewPart := func(block internal.Block, nextBlock *internal.Block) bool {
		if block.Type == internal.BlockTitle && nextBlock != nil {
			return currentLength > int(float64(effectiveLimit)*1.5) 
		}

		if block.Type == internal.BlockCode && utf8.RuneCountInString(block.Content) > effectiveLimit/2 {
			return true
		}

		return currentLength > effectiveLimit
	}

	processGroup := func() {
		if len(currentGroup) == 0 {
			return
		}

		var groupContent strings.Builder
		for i, block := range currentGroup {
			if i > 0 {
				if block.Type == internal.BlockTitle {
					groupContent.WriteString("\n\n")
				} else if block.Type == internal.BlockCode {
					if currentGroup[i-1].Type != internal.BlockCode {
						groupContent.WriteString("\n\n")
					}
				} else if currentGroup[i-1].Type == internal.BlockCode {
					groupContent.WriteString("\n\n")
				} else {
					groupContent.WriteString("\n\n")
				}
			}

			if block.Type == internal.BlockCode {
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
		if utf8.RuneCountInString(content) > 50 { 
			parts = append(parts, content)
		} else {
			if len(parts) > 0 {
				lastPart := parts[len(parts)-1]
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

	for i := 0; i < len(blocks); i++ {
		block := blocks[i]
		nextBlock := (*internal.Block)(nil)
		if i+1 < len(blocks) {
			nextBlock = &blocks[i+1]
		}

		blockLength := utf8.RuneCountInString(block.Content)

		if block.Type == internal.BlockCode && blockLength > effectiveLimit {
			processGroup()
			codeParts := divideCodeBlock(block.Content, effectiveLimit)
			parts = append(parts, codeParts...)
			continue
		}

		additionalLength := blockLength
		if len(currentGroup) > 0 {
			additionalLength += 2 
		}

		if shouldStartNewPart(block, nextBlock) {
			if len(currentGroup) > 0 {
				processGroup()
			}
			if block.Type == internal.BlockCode {
				parts = append(parts, block.Content)
				continue
			}
		}

		if currentLength+additionalLength > effectiveLimit {
			processGroup()
		}

		currentGroup = append(currentGroup, block)
		currentLength += additionalLength
	}

	processGroup()

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

func divideCodeBlock(content string, limit int) []string {
	lines := strings.Split(content, "\n")
	if len(lines) < 2 {
		return []string{content}
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
	effectiveLimit := limit - overhead

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

	return parts
}

func formatCodeBlock(content string, language string) string {
	content = strings.TrimSpace(content)
	if language != "" {
		return "```" + language + "\n" + content + "\n```"
	}
	return "```\n" + content + "\n```"
}
