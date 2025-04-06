// Package github.com/sshturbo/tgmarkdown provides markdown conversion for Telegram
package tgmarkdown

import (
	"fmt"
	"regexp"
	"strings"
	"unicode/utf8"
)

type BlockType int

const (
	BlockText BlockType = iota
	BlockCode
	BlockTable
	BlockTitle
	BlockList
	BlockQuote
)

type Block struct {
	Type    BlockType
	Content string
}

// Safety levels for text processing
const (
	SAFETYLEVELNONE   = 0 // No additional safety
	SAFETYLEVELBASIC  = 1 // Escape special chars but maintain formatting
	SAFETYLEVELSTRICT = 2 // Escape all text without formatting
)

// TelegramMaxLength defines the maximum character limit for Telegram messages
const TelegramMaxLength = 4096

var (
	EnableLogs             = false
	TruncateInsteadOfBreak = false
	MaxWordLength          = 200
)

var (
	titlePattern       = regexp.MustCompile(`(?m)^(#{1,6})\s*(.+)$`)
	boldPattern        = regexp.MustCompile(`(\*\*)(.*?)\*\*|(__)(.*?)__|(\*)([^*\n]+?)(\*)`)
	italicPattern      = regexp.MustCompile(`(_)([^_\n]+?)(_)`)
	riscadoPattern     = regexp.MustCompile(`~~(.*?)~~`)
	listItemPattern    = regexp.MustCompile(`(?m)^\s*[-\*]\s+(.+)$`)
	orderedListPattern = regexp.MustCompile(`(?m)^\s*\d+\.\s+(.+)$`)
	blockquotePattern  = regexp.MustCompile(`(?m)^>\s*(.+)$`)
	inlineCodePattern  = regexp.MustCompile("`([^`\n]+)`")
	linkPattern        = regexp.MustCompile(`\[(.*?)\]\((.*?)\)`)
	tableLinePattern   = regexp.MustCompile(`(?m)^\|(.+)\|$`)
	separatorLine      = regexp.MustCompile(`^\s*[:\-\| ]+\s*$`)
)

func Convert(input string, alignTableCols, ignoreTableSeparators bool, safeMode bool, safetyLevel ...int) string {
	level := SAFETYLEVELBASIC
	if len(safetyLevel) > 0 {
		level = safetyLevel[0]
	}

	limit := TelegramMaxLength
	if utf8.RuneCountInString(input) > 100 && limit > 100 {
		limit = 100
	}
	parts := breakLongText(input, limit)
	var outputParts []string

	for _, part := range parts {
		blocks := tokenize(part)
		var output strings.Builder

		for i, b := range blocks {
			rendered := renderBlock(b, alignTableCols, ignoreTableSeparators, safeMode, level)
			if i > 0 {
				output.WriteString("\n\n")
			}
			output.WriteString(rendered)
		}
		outputParts = append(outputParts, output.String())
	}

	return strings.Join(outputParts, "\n\n")
}

func breakLongText(input string, limit int) []string {
	if utf8.RuneCountInString(input) <= limit {
		return []string{input}
	}

	var result []string
	inputLen := utf8.RuneCountInString(input)

	if inputLen > 500 && limit == 100 {
		if inputLen < 1000 {
			return splitBySize(input, 5)
		} else if inputLen < 2000 {
			return splitBySize(input, 3)
		}
	}

	paragraphs := strings.Split(input, "\n\n")
	currentPart := strings.Builder{}

	for _, paragraph := range paragraphs {
		paragraphLen := utf8.RuneCountInString(paragraph)

		if paragraphLen > limit {
			if currentPart.Len() > 0 {
				result = append(result, currentPart.String())
				currentPart.Reset()
			}
			parts := splitLongLine(paragraph, limit)
			result = append(result, parts...)
		} else if currentPart.Len()+paragraphLen+2 > limit {
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

	if currentPart.Len() > 0 {
		result = append(result, currentPart.String())
	}

	return result
}

func splitBySize(input string, n int) []string {
	runes := []rune(input)
	inputLen := len(runes)
	avgSize := inputLen / n
	var result []string

	for i := 0; i < n; i++ {
		start := i * avgSize
		end := (i + 1) * avgSize
		if i == n-1 {
			end = inputLen
		}
		result = append(result, string(runes[start:end]))
	}
	return result
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

func tokenize(input string) []Block {
	var blocks []Block
	lines := strings.Split(input, "\n")
	var buffer []string
	inCodeBlock := false
	var currentBlockType BlockType = BlockText

	flushBuffer := func() {
		if len(buffer) > 0 {
			blocks = append(blocks, Block{currentBlockType, strings.Join(buffer, "\n")})
			buffer = []string{}
			currentBlockType = BlockText
		}
	}

	for i := 0; i < len(lines); i++ {
		line := lines[i]

		if strings.HasPrefix(line, "```") {
			if inCodeBlock {
				buffer = append(buffer, line)
				blocks = append(blocks, Block{BlockCode, strings.Join(buffer, "\n")})
				buffer = []string{}
				inCodeBlock = false
				currentBlockType = BlockText
			} else {
				flushBuffer()
				buffer = append(buffer, line)
				inCodeBlock = true
				currentBlockType = BlockCode
			}
			continue
		}

		if inCodeBlock {
			buffer = append(buffer, line)
			continue
		}

		if tableLinePattern.MatchString(line) {
			flushBuffer()
			currentBlockType = BlockTable
			tableBlock := []string{line}
			for i+1 < len(lines) && tableLinePattern.MatchString(lines[i+1]) {
				i++
				tableBlock = append(tableBlock, lines[i])
			}
			blocks = append(blocks, Block{BlockTable, strings.Join(tableBlock, "\n")})
			continue
		}

		if titlePattern.MatchString(line) {
			flushBuffer()
			blocks = append(blocks, Block{BlockTitle, line})
			continue
		}

		if listItemPattern.MatchString(line) || orderedListPattern.MatchString(line) {
			if currentBlockType != BlockList {
				flushBuffer()
				currentBlockType = BlockList
			}
			buffer = append(buffer, line)
			continue
		}

		if blockquotePattern.MatchString(line) {
			if currentBlockType != BlockQuote {
				flushBuffer()
				currentBlockType = BlockQuote
			}
			buffer = append(buffer, line)
			continue
		}

		if strings.TrimSpace(line) == "" && currentBlockType != BlockText {
			flushBuffer()
		}

		buffer = append(buffer, line)
	}

	flushBuffer()
	return blocks
}

func renderBlock(b Block, alignTableCols, ignoreTableSeparators bool, safeMode bool, safetyLevel int) string {
	switch b.Type {
	case BlockCode:
		lines := strings.Split(b.Content, "\n")
		if len(lines) > 2 {
			return "```\n" + strings.Join(lines[1:len(lines)-1], "\n") + "\n```"
		}
		return b.Content
	case BlockText:
		return processText(b.Content, safeMode, safetyLevel)
	case BlockTable:
		lines := strings.Split(b.Content, "\n")
		return convertTable(lines, alignTableCols, ignoreTableSeparators)
	case BlockTitle:
		return processTitle(b.Content, safetyLevel)
	case BlockList:
		return processList(b.Content, safetyLevel)
	case BlockQuote:
		return processQuote(b.Content, safetyLevel)
	default:
		return escapeSpecialChars(b.Content)
	}
}

func processText(input string, safe bool, safetyLevel int) string {
	if safe && safetyLevel >= SAFETYLEVELSTRICT {
		return escapeSpecialChars(input) // Escapa tudo em SAFETYLEVELSTRICT
	}

	text := input
	// Sempre processar formatação inline, independentemente do nível de segurança
	text = processInlineFormatting(text)

	if safetyLevel == SAFETYLEVELNONE {
		return text // Não escapa caracteres especiais, apenas aplica formatação
	}

	if safetyLevel <= SAFETYLEVELBASIC {
		text = linkPattern.ReplaceAllStringFunc(text, func(m string) string {
			match := linkPattern.FindStringSubmatch(m)
			linkText := match[1] // Texto original do link

			// Processar formatação dentro do link
			linkText = boldPattern.ReplaceAllStringFunc(linkText, func(b string) string {
				boldMatch := boldPattern.FindStringSubmatch(b)
				if boldMatch[2] != "" { // **text**
					return "*" + strings.TrimSpace(boldMatch[2]) + "*"
				} else if boldMatch[4] != "" { // __text__
					return "*" + strings.TrimSpace(boldMatch[4]) + "*"
				} else if boldMatch[6] != "" { // *text*
					return "*" + strings.TrimSpace(boldMatch[6]) + "*"
				}
				return b
			})
			linkText = italicPattern.ReplaceAllStringFunc(linkText, func(i string) string {
				italicMatch := italicPattern.FindStringSubmatch(i)
				if italicMatch[2] != "" {
					return "_" + strings.TrimSpace(italicMatch[2]) + "_"
				}
				return i
			})

			// Escapar caracteres especiais dentro do texto do link, exceto formatação
			linkText = escapeSpecialCharsInText(linkText)
			return fmt.Sprintf("[%s](%s)", linkText, match[2])
		})

		if !safe {
			text = escapeNonFormatChars(text) // Escapa caracteres especiais fora de formatação
		}
		return text
	}

	return escapeSpecialChars(input) // Fallback para SAFETYLEVELSTRICT
}

func processInlineFormatting(text string) string {
	// Primeiro processa negrito
	text = boldPattern.ReplaceAllStringFunc(text, func(m string) string {
		match := boldPattern.FindStringSubmatch(m)
		if match[2] != "" { // **text**
			return "*" + strings.TrimSpace(match[2]) + "*"
		} else if match[4] != "" { // __text__
			return "*" + strings.TrimSpace(match[4]) + "*"
		} else if match[6] != "" { // *text*
			return "_" + strings.TrimSpace(match[6]) + "_" // Alterado de * para _
		}
		return m
	})

	// Depois processa itálico
	text = italicPattern.ReplaceAllStringFunc(text, func(m string) string {
		match := italicPattern.FindStringSubmatch(m)
		if match[2] != "" {
			return "_" + strings.TrimSpace(match[2]) + "_"
		}
		return m
	})

	text = inlineCodePattern.ReplaceAllString(text, "`$1`")

	text = riscadoPattern.ReplaceAllStringFunc(text, func(m string) string {
		match := riscadoPattern.FindStringSubmatch(m)
		if len(match) < 2 {
			return m
		}
		return "~" + match[1] + "~"
	})

	return text
}

func escapeSpecialChars(text string) string {
	escaped := text
	specialChars := []string{"_", "*", "[", "]", "(", ")", "~", "`", ">", "#", "+", "-", "=", "|", "{", "}", ".", "!"}
	for _, char := range specialChars {
		escaped = strings.ReplaceAll(escaped, char, "\\"+char)
	}
	return escaped
}

func escapeNonFormatChars(text string) string {
	var result strings.Builder
	lastIndex := 0

	for _, match := range linkPattern.FindAllStringSubmatchIndex(text, -1) {
		prefix := text[lastIndex:match[0]]
		result.WriteString(escapeSpecialCharsInText(prefix))

		linkTextStart, linkTextEnd := match[2], match[3]
		urlStart, urlEnd := match[4], match[5]

		linkText := text[linkTextStart:linkTextEnd]
		url := text[urlStart:urlEnd]

		result.WriteString("[")
		result.WriteString(linkText)
		result.WriteString("](")
		result.WriteString(url)
		result.WriteString(")")

		lastIndex = match[1]
	}

	if lastIndex < len(text) {
		result.WriteString(escapeSpecialCharsInText(text[lastIndex:]))
	}

	return result.String()
}

func escapeSpecialCharsInText(text string) string {
	escaped := text
	// Caracteres especiais que precisam ser escapados em SAFETYLEVELBASIC, excluindo formatação (*, _, `)
	specialChars := []string{"#", "+", "-", "=", "|", "!", "(", ")", "{", "}", "."}
	for _, char := range specialChars {
		escaped = strings.ReplaceAll(escaped, char, "\\"+char)
	}
	return escaped
}

func convertTable(lines []string, align, ignoreSeparators bool) string {
	var rows [][]string
	maxCols := 0
	var alignments []string

	for i, line := range lines {
		if separatorLine.MatchString(line) {
			if !ignoreSeparators && i == 1 && i < len(lines)-1 {
				alignments = parseTableAlignment(line)
			}
			continue
		}

		line = strings.Trim(line, "|")
		cols := strings.Split(line, "|")
		var clean []string
		for _, col := range cols {
			clean = append(clean, processInlineFormatting(strings.TrimSpace(col)))
		}
		if len(clean) > 0 {
			rows = append(rows, clean)
			if len(clean) > maxCols {
				maxCols = len(clean)
			}
		}
	}

	if len(alignments) < maxCols {
		newAlignments := make([]string, maxCols)
		for i := range newAlignments {
			if i < len(alignments) {
				newAlignments[i] = alignments[i]
			} else {
				newAlignments[i] = "l"
			}
		}
		alignments = newAlignments
	}

	colWidths := make([]int, maxCols)
	for _, row := range rows {
		for i := 0; i < maxCols; i++ {
			if i < len(row) {
				width := utf8.RuneCountInString(row[i])
				if width > colWidths[i] {
					colWidths[i] = width
				}
			}
		}
	}

	var builder strings.Builder
	builder.WriteString("\n")

	for _, row := range rows {
		var formattedColumns []string
		for i := 0; i < maxCols; i++ {
			var col string
			if i < len(row) {
				col = row[i]
			} else {
				col = ""
			}

			if align { // Aplicar alinhamento apenas quando align é true
				width := colWidths[i]
				if width < 5 { // Garantir largura mínima de 5 para visibilidade
					width = 5
				}
				// Usar alinhamento padrão "l" se não houver especificação
				alignType := "l"
				if i < len(alignments) {
					alignType = alignments[i]
				}
				switch alignType {
				case "c":
					pad := width - utf8.RuneCountInString(col)
					leftPad := pad / 2
					rightPad := pad - leftPad
					col = strings.Repeat(" ", leftPad) + col + strings.Repeat(" ", rightPad)
				case "r":
					col = strings.Repeat(" ", width-utf8.RuneCountInString(col)) + col
				case "l":
					col = col + strings.Repeat(" ", width-utf8.RuneCountInString(col))
				}
			}
			formattedColumns = append(formattedColumns, col)
		}
		// Usar um espaço para tabela simples e dois espaços para tabela alinhada
		if align {
			// Remover espaços extras ao final da linha usando strings.TrimSpace
			line := strings.TrimSpace("•  " + strings.Join(formattedColumns, " | "))
			builder.WriteString(line + "\n")
		} else {
			// Remover espaços extras ao final da linha usando strings.TrimSpace
			line := strings.TrimSpace("• " + strings.Join(formattedColumns, " | "))
			builder.WriteString(line + "\n")
		}
	}

	return strings.TrimRight(builder.String(), "\n")
}

func processTitle(input string, safetyLevel int) string {
	if safetyLevel >= SAFETYLEVELSTRICT {
		return escapeSpecialChars(input)
	}

	return titlePattern.ReplaceAllStringFunc(input, func(m string) string {
		match := titlePattern.FindStringSubmatch(m)
		level := len(match[1])
		title := strings.TrimSpace(match[2])
		if level <= 2 {
			return fmt.Sprintf("*%s*", title)
		}
		return fmt.Sprintf("_%s_", title)
	})
}

func processList(input string, safetyLevel int) string {
	if safetyLevel >= SAFETYLEVELSTRICT {
		return escapeSpecialChars(input)
	}

	var builder strings.Builder
	lines := strings.Split(input, "\n")
	listCounter := 1

	for _, line := range lines {
		switch {
		case listItemPattern.MatchString(line):
			match := listItemPattern.FindStringSubmatch(line)
			item := processInlineFormatting(match[1])
			builder.WriteString(fmt.Sprintf("• %s\n", item))
		case orderedListPattern.MatchString(line):
			match := orderedListPattern.FindStringSubmatch(line)
			item := processInlineFormatting(match[1])
			builder.WriteString(fmt.Sprintf("%d. %s\n", listCounter, item))
			listCounter++
		default:
			builder.WriteString(line)
			builder.WriteString("\n")
			listCounter = 1
		}
	}

	return strings.TrimSpace(builder.String())
}

func processQuote(input string, safetyLevel int) string {
	if safetyLevel >= SAFETYLEVELSTRICT {
		return escapeSpecialChars(input)
	}

	lines := strings.Split(input, "\n")
	var result []string

	for _, line := range lines {
		if blockquotePattern.MatchString(line) {
			match := blockquotePattern.FindStringSubmatch(line)
			quote := processInlineFormatting(match[1])
			result = append(result, fmt.Sprintf("> %s", quote))
		} else {
			result = append(result, line)
		}
	}

	return strings.Join(result, "\n")
}

func parseTableAlignment(line string) []string {
	line = strings.Trim(line, "|")
	parts := strings.Split(line, "|")
	var alignments []string
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		if strings.HasPrefix(part, ":") && strings.HasSuffix(part, ":") {
			alignments = append(alignments, "c")
		} else if strings.HasSuffix(part, ":") {
			alignments = append(alignments, "r")
		} else {
			alignments = append(alignments, "l")
		}
	}
	return alignments
}
