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

	if strings.HasPrefix(input, "```") && strings.HasSuffix(input, "```") {
		if inputLen <= effectiveLimit {
			return []string{input}
		}
		return splitLongCodeBlock(input, effectiveLimit)
	}

	paragraphs := strings.Split(input, "\n\n")
	currentPart := strings.Builder{}
	inCodeBlock := false
	codeBlockBuffer := strings.Builder{}

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
					parts := splitLongCodeBlock(codeBlockContent, effectiveLimit)
					if currentPart.Len() > 0 {
						result = append(result, currentPart.String())
						currentPart.Reset()
					}
					result = append(result, parts...)
				}
				codeBlockBuffer.Reset()
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
			codeBlockBuffer.WriteString(paragraph)
			codeBlockBuffer.WriteString("\n\n")
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
		parts := splitLongLine(codeBlockContent, effectiveLimit)
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

func splitLongCodeBlock(input string, limit int) []string {
	var parts []string
	lines := strings.Split(input, "\n")
	currentPart := strings.Builder{}
	header := lines[0]

	for i, line := range lines[1:] {
		lineWithNewline := line + "\n"
		if i == len(lines)-2 {
			lineWithNewline = line
		}

		if currentPart.Len() == 0 {
			currentPart.WriteString(header)
			currentPart.WriteString("\n")
		}

		if utf8.RuneCountInString(currentPart.String()+lineWithNewline) > limit-3 {
			currentPart.WriteString("```")
			parts = append(parts, currentPart.String())
			currentPart.Reset()
			currentPart.WriteString(header)
			currentPart.WriteString("\n")
			currentPart.WriteString(lineWithNewline)
		} else {
			currentPart.WriteString(lineWithNewline)
		}
	}

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
