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

func Convert(input string, alignTableCols, ignoreTableSeparators bool, safetyLevel ...int) string {
	level := SAFETYLEVELBASIC
	if len(safetyLevel) > 0 {
		level = safetyLevel[0]
	}

	parts := breakLongText(input)
	var outputParts []string

	for _, part := range parts {
		blocks := tokenize(part)
		var output strings.Builder

		for i, b := range blocks {
			rendered := renderBlock(b, alignTableCols, ignoreTableSeparators, level)
			if i > 0 {
				output.WriteString("\n\n")
			}
			output.WriteString(rendered)
		}
		outputParts = append(outputParts, output.String())
	}

	return strings.Join(outputParts, "\n\n")
}

func breakLongText(input string) []string {
	// Usar sempre o limite máximo do Telegram como padrão
	effectiveLimit := TelegramMaxLength

	if utf8.RuneCountInString(input) <= effectiveLimit {
		return []string{input}
	}

	var result []string
	inputLen := utf8.RuneCountInString(input)

	// Se o input for um bloco de código completo, dividir apenas se exceder o limite
	if strings.HasPrefix(input, "```") && strings.HasSuffix(input, "```") {
		if inputLen <= effectiveLimit {
			return []string{input}
		}
		// Dividir o bloco de código em partes menores que 4096 caracteres
		return splitLongCodeBlock(input, effectiveLimit)
	}

	paragraphs := strings.Split(input, "\n\n")
	currentPart := strings.Builder{}
	inCodeBlock := false
	codeBlockBuffer := strings.Builder{}

	for _, paragraph := range paragraphs {
		// Verificar se estamos entrando ou saindo de um bloco de código
		if strings.HasPrefix(paragraph, "```") {
			if inCodeBlock {
				// Fim do bloco de código
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
					// Dividir o bloco de código em partes menores
					parts := splitLongCodeBlock(codeBlockContent, effectiveLimit)
					if currentPart.Len() > 0 {
						result = append(result, currentPart.String())
						currentPart.Reset()
					}
					result = append(result, parts...)
				}
				codeBlockBuffer.Reset()
			} else {
				// Início do bloco de código
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
		// Se o bloco de código não foi fechado, tratá-lo como texto normal
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

// Função auxiliar para dividir blocos de código longos
func splitLongCodeBlock(input string, limit int) []string {
	var parts []string
	lines := strings.Split(input, "\n")
	currentPart := strings.Builder{}
	header := lines[0] // Primeira linha com ``` ou ```lang

	for i, line := range lines[1:] { // Começar após o header
		lineWithNewline := line + "\n"
		if i == len(lines)-2 { // Última linha, remover \n extra
			lineWithNewline = line
		}

		if currentPart.Len() == 0 {
			currentPart.WriteString(header)
			currentPart.WriteString("\n")
		}

		if utf8.RuneCountInString(currentPart.String()+lineWithNewline) > limit-3 { // Reservar espaço para ```
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

func tokenize(input string) []Block {
	var blocks []Block
	lines := strings.Split(input, "\n")
	var buffer []string
	inCodeBlock := false
	var currentBlockType BlockType = BlockText

	flushBuffer := func() {
		if len(buffer) > 0 {
			content := strings.Join(buffer, "\n")
			if currentBlockType == BlockCode {
				// Preserva o conteúdo do bloco de código exatamente como está
				blocks = append(blocks, Block{currentBlockType, content})
			} else {
				// Para outros blocos, aplica trim
				blocks = append(blocks, Block{currentBlockType, strings.TrimSpace(content)})
			}
			buffer = []string{}
			currentBlockType = BlockText
		}
	}

	for i := 0; i < len(lines); i++ {
		line := lines[i]

		if strings.HasPrefix(line, "```") {
			if inCodeBlock {
				// Fim do bloco de código
				inCodeBlock = false
				buffer = append(buffer, line) // Inclui a linha de fechamento ```
				flushBuffer()
			} else {
				// Início do bloco de código
				if len(buffer) > 0 {
					flushBuffer()
				}
				inCodeBlock = true
				currentBlockType = BlockCode
				buffer = append(buffer, line) // Inclui a linha de abertura ```
			}
			continue
		}

		if inCodeBlock {
			// Adiciona a linha ao buffer sem modificações
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

func renderBlock(b Block, alignTableCols, ignoreTableSeparators bool, safetyLevel int) string {
	switch b.Type {
	case BlockCode:
		if safetyLevel == SAFETYLEVELSTRICT {
			return escapeSpecialChars(b.Content) // Escapar tudo em modo estrito
		}
		// Retorna o conteúdo do bloco de código exatamente como está, sem modificação
		if strings.HasPrefix(b.Content, "```") && strings.HasSuffix(b.Content, "```") {
			return b.Content
		}
		if strings.TrimSpace(b.Content) == "" {
			return "```\n```"
		}
		return "```\n" + b.Content + "\n```"
	case BlockText:
		return processText(b.Content, safetyLevel)
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

func processText(input string, safetyLevel int) string {
	if safetyLevel == SAFETYLEVELSTRICT {
		// Escapar tudo, incluindo conteúdo dentro de blocos inline
		return escapeSpecialChars(input)
	}

	if safetyLevel == SAFETYLEVELBASIC {
		text := input
		text = processInlineFormatting(text)
		text = linkPattern.ReplaceAllStringFunc(text, func(m string) string {
			match := linkPattern.FindStringSubmatch(m)
			linkText := match[1]
			linkText = boldPattern.ReplaceAllStringFunc(linkText, func(b string) string {
				boldMatch := boldPattern.FindStringSubmatch(b)
				if boldMatch[2] != "" {
					return "*" + strings.TrimSpace(boldMatch[2]) + "*"
				} else if boldMatch[4] != "" {
					return "*" + strings.TrimSpace(boldMatch[4]) + "*"
				} else if boldMatch[6] != "" {
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
			return fmt.Sprintf("[%s](%s)", linkText, match[2])
		})

		// Escapar caracteres especiais fora dos blocos inline
		var result strings.Builder
		lastIndex := 0
		for _, match := range inlineCodePattern.FindAllStringSubmatchIndex(text, -1) {
			prefix := text[lastIndex:match[0]]
			result.WriteString(escapeNonFormatChars(prefix))
			codeContent := text[match[2]:match[3]] // Conteúdo dentro de `
			result.WriteString("`")
			result.WriteString(codeContent) // Preservar conteúdo sem escapamento
			result.WriteString("`")
			lastIndex = match[1]
		}
		if lastIndex < len(text) {
			result.WriteString(escapeNonFormatChars(text[lastIndex:]))
		}
		return result.String()
	}

	// SAFETYLEVELNONE: Não escapar nada, apenas aplicar formatação
	text := input
	text = processInlineFormatting(text)
	text = linkPattern.ReplaceAllStringFunc(text, func(m string) string {
		match := linkPattern.FindStringSubmatch(m)
		linkText := match[1]
		linkText = boldPattern.ReplaceAllStringFunc(linkText, func(b string) string {
			boldMatch := boldPattern.FindStringSubmatch(b)
			if boldMatch[2] != "" {
				return "*" + strings.TrimSpace(boldMatch[2]) + "*"
			} else if boldMatch[4] != "" {
				return "*" + strings.TrimSpace(boldMatch[4]) + "*"
			} else if boldMatch[6] != "" {
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
		return fmt.Sprintf("[%s](%s)", linkText, match[2])
	})
	return text
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
	specialChars := []string{"#", "+", "-", "=", "|", ".", "!", "(", ")", "{", "}"} // Adicionado | e {} à lista
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
