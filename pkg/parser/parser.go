package parser

import (
	"strings"
	"unicode/utf8"

	"github.com/sshturbo/GoTeleMD/internal"
	"github.com/sshturbo/GoTeleMD/pkg/utils"
)

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

func BreakLongText(input string) []string {
	effectiveLimit := internal.TelegramMaxLength

	if utf8.RuneCountInString(input) <= effectiveLimit {
		return []string{input}
	}

	var result []string
	inputLen := utf8.RuneCountInString(input)

	// Special handling for code blocks
	if strings.HasPrefix(input, "```") {
		// Find the end of the code block
		endIndex := strings.Index(input, "```")
		if endIndex == -1 {
			return []string{input}
		}

		// Check if there's a title before the code block
		titleStart := strings.LastIndex(input[:endIndex], "**")
		if titleStart != -1 {
			titleEnd := strings.Index(input[titleStart:], ":**") + titleStart + 3
			if titleEnd > titleStart {
				// Include the title in the code block
				header := input[:titleEnd]
				content := input[titleEnd:]
				return splitLongCodeBlockWithTitle(header, content, effectiveLimit)
			}
		}

		// If no title found, proceed with normal code block splitting
		parts := strings.SplitN(input, "\n", 2)
		if len(parts) != 2 {
			return []string{input}
		}
		header := parts[0]
		content := parts[1]
		content = strings.TrimSuffix(content, "```")
		
		return splitLongCodeBlockWithTitle(header, content, effectiveLimit)
	}

	// Handle regular text with embedded code blocks
	paragraphs := strings.Split(input, "\n\n")
	currentPart := strings.Builder{}
	inCodeBlock := false
	codeBlockBuffer := strings.Builder{}
	var titleBuffer strings.Builder

	for _, paragraph := range paragraphs {
		if strings.HasPrefix(paragraph, "```") {
			if inCodeBlock {
				inCodeBlock = false
				codeBlockBuffer.WriteString(paragraph)
				codeBlockContent := codeBlockBuffer.String()
				if utf8.RuneCountInString(codeBlockContent) <= effectiveLimit {
					if currentPart.Len() > 0 {
						result = append(result, currentPart.String())
						currentPart.Reset()
					}
					result = append(result, codeBlockContent)
				} else {
					parts := splitLongCodeBlockWithTitle(titleBuffer.String(), codeBlockContent, effectiveLimit)
					if currentPart.Len() > 0 {
						result = append(result, currentPart.String())
						currentPart.Reset()
					}
					result = append(result, parts...)
				}
				codeBlockBuffer.Reset()
				titleBuffer.Reset()
			} else {
				if currentPart.Len() > 0 {
					result = append(result, currentPart.String())
					currentPart.Reset()
				}
				inCodeBlock = true
				codeBlockBuffer.WriteString(paragraph)
				codeBlockBuffer.WriteString("\n\n")
			}
			continue
		}

		if inCodeBlock {
			if strings.HasPrefix(paragraph, "**") && strings.HasSuffix(paragraph, ":**") {
				titleBuffer.WriteString(paragraph)
				titleBuffer.WriteString("\n\n")
			} else {
				codeBlockBuffer.WriteString(paragraph)
				codeBlockBuffer.WriteString("\n\n")
			}
			continue
		}

		paragraphLen := utf8.RuneCountInString(paragraph)

		if paragraphLen > effectiveLimit {
			if currentPart.Len() > 0 {
				result = append(result, currentPart.String())
				currentPart.Reset()
			}
			parts := splitLongLine(paragraph, effectiveLimit)
			result = append(result, parts...)
		} else if currentPart.Len()+paragraphLen+2 > effectiveLimit {
			result = append(result, currentPart.String())
			currentPart.Reset()
			currentPart.WriteString(paragraph)
		} else {
			if currentPart.Len() > 0 {
				currentPart.WriteString("\n\n")
			}
			currentPart.WriteString(paragraph)
		}
	}

	if inCodeBlock {
		codeBlockContent := codeBlockBuffer.String()
		parts := splitLongCodeBlockWithTitle(titleBuffer.String(), codeBlockContent, effectiveLimit)
		if currentPart.Len() > 0 {
			result = append(result, currentPart.String())
			currentPart.Reset()
		}
		result = append(result, parts...)
	} else if currentPart.Len() > 0 {
		result = append(result, currentPart.String())
	}

	return result
}

func splitLongCodeBlockWithTitle(title, content string, limit int) []string {
	var parts []string
	lines := strings.Split(content, "\n")
	currentPart := strings.Builder{}
	
	// Add title to first part
	if title != "" {
		currentPart.WriteString(title)
		currentPart.WriteString("\n\n")
	}

	// Adiciona a marcação de código no início
	if !strings.HasPrefix(content, "```") {
		currentPart.WriteString("```")
		currentPart.WriteString("\n")
	}

	for i, line := range lines {
		lineWithNewline := line + "\n"
		if i == len(lines)-1 {
			lineWithNewline = line
		}

		// Verifica se a linha atual é uma marcação de código
		if strings.TrimSpace(line) == "```" {
			// Se estamos no início, ignora
			if currentPart.Len() == 0 {
				continue
			}
			// Se estamos no final, adiciona a marcação de fechamento
			if i == len(lines)-1 {
				currentPart.WriteString("```")
				parts = append(parts, currentPart.String())
				return parts
			}
			// Se estamos no meio, fecha o bloco atual e começa um novo
			currentPart.WriteString("```")
			parts = append(parts, currentPart.String())
			currentPart.Reset()
			if title != "" && len(parts) == 1 {
				currentPart.WriteString(title)
				currentPart.WriteString("\n\n")
			}
			currentPart.WriteString("```")
			currentPart.WriteString("\n")
			continue
		}

		if utf8.RuneCountInString(currentPart.String()+lineWithNewline) > limit-3 {
			// Close current block and start new one
			currentPart.WriteString("```")
			parts = append(parts, currentPart.String())
			currentPart.Reset()
			// Start new block with title if it's the first part
			if title != "" && len(parts) == 1 {
				currentPart.WriteString(title)
				currentPart.WriteString("\n\n")
			}
			currentPart.WriteString("```")
			currentPart.WriteString("\n")
			currentPart.WriteString(lineWithNewline)
		} else {
			currentPart.WriteString(lineWithNewline)
		}
	}

	// Close the last block
	if currentPart.Len() > 0 {
		currentPart.WriteString("```")
		parts = append(parts, currentPart.String())
	}

	return parts
}

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
