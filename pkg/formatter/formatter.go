package formatter

import (
	"fmt"
	"strings"

	"github.com/sshturbo/GoTeleMD/internal"
	"github.com/sshturbo/GoTeleMD/pkg/utils"
)

func ProcessText(text string, safetyLevel SafetyLevel) string {
	if safetyLevel == SafetyLevelNone {
		return text
	}

	var result strings.Builder
	parts := strings.Split(text, "```")
	
	for i, part := range parts {
		if i%2 == 0 {
			// Texto fora do bloco de código
			lines := strings.Split(part, "\n")
			for j, line := range lines {
				processed := line
				if safetyLevel == SafetyLevelHigh {
					processed = escapeNonFormatChars(line)
				} else if safetyLevel == SafetyLevelMedium {
					processed = escapeSpecialCharsInText(line)
				}
				result.WriteString(processed)
				if j < len(lines)-1 {
					result.WriteString("\n")
				}
			}
		} else {
			// Conteúdo do bloco de código
			result.WriteString("```")
			result.WriteString(escapeCodeBlockContent(part))
			result.WriteString("```")
		}
	}
	
	return result.String()
}

func processLinks(text string) string {
	return utils.LinkPattern.ReplaceAllStringFunc(text, func(m string) string {
		match := utils.LinkPattern.FindStringSubmatch(m)
		linkText := match[1]
		linkText = utils.BoldPattern.ReplaceAllStringFunc(linkText, func(b string) string {
			boldMatch := utils.BoldPattern.FindStringSubmatch(b)
			if boldMatch[2] != "" {
				return "*" + strings.TrimSpace(boldMatch[2]) + "*"
			} else if boldMatch[4] != "" {
				return "*" + strings.TrimSpace(boldMatch[4]) + "*"
			} else if boldMatch[6] != "" {
				return "*" + strings.TrimSpace(boldMatch[6]) + "*"
			}
			return b
		})
		linkText = utils.ItalicPattern.ReplaceAllStringFunc(linkText, func(i string) string {
			italicMatch := utils.ItalicPattern.FindStringSubmatch(i)
			if italicMatch[2] != "" {
				return "_" + strings.TrimSpace(italicMatch[2]) + "_"
			}
			return i
		})
		return fmt.Sprintf("[%s](%s)", linkText, match[2])
	})
}

func ProcessInlineFormatting(text string) string {
	// Processa negrito
	text = utils.BoldPattern.ReplaceAllStringFunc(text, func(m string) string {
		match := utils.BoldPattern.FindStringSubmatch(m)
		if match[2] != "" {
			return "*" + strings.TrimSpace(match[2]) + "*"
		} else if match[4] != "" {
			return "*" + strings.TrimSpace(match[4]) + "*"
		} else if match[6] != "" {
			return "_" + strings.TrimSpace(match[6]) + "_" // Caso de sublinhado como fallback
		}
		return m
	})

	// Processa itálico
	text = utils.ItalicPattern.ReplaceAllStringFunc(text, func(m string) string {
		match := utils.ItalicPattern.FindStringSubmatch(m)
		if match[2] != "" {
			return "_" + strings.TrimSpace(match[2]) + "_"
		}
		return m
	})

	// Processa código inline
	text = utils.InlineCodePattern.ReplaceAllString(text, "`$1`")

	// Processa riscado
	text = utils.RiscadoPattern.ReplaceAllStringFunc(text, func(m string) string {
		match := utils.RiscadoPattern.FindStringSubmatch(m)
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

	for _, match := range utils.LinkPattern.FindAllStringSubmatchIndex(text, -1) {
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
	// Preserva a formatação original, apenas escapa caracteres especiais
	lines := strings.Split(text, "\n")
	var result strings.Builder
	
	for i, line := range lines {
		// Escapa apenas os caracteres que precisam ser escapados no Telegram
		escaped := line
		specialChars := []string{"#", "+", "-", "=", "|", ".", "!", "(", ")", "{", "}"}
		for _, char := range specialChars {
			escaped = strings.ReplaceAll(escaped, char, "\\"+char)
		}
		
		result.WriteString(escaped)
		if i < len(lines)-1 {
			result.WriteString("\n")
		}
	}
	
	return result.String()
}

// Nova função para escapar apenas os caracteres necessários dentro de blocos de código
func escapeCodeBlockContent(content string) string {
	// Preserva a formatação original do bloco de código
	lines := strings.Split(content, "\n")
	var result strings.Builder
	
	for i, line := range lines {
		// Escapa apenas os caracteres que precisam ser escapados no Telegram
		escaped := line
		specialChars := []string{"#", "+", "-", "=", "|", ".", "!", "(", ")", "{", "}"}
		for _, char := range specialChars {
			escaped = strings.ReplaceAll(escaped, char, "\\"+char)
		}
		
		result.WriteString(escaped)
		if i < len(lines)-1 {
			result.WriteString("\n")
		}
	}
	
	return result.String()
}

func ProcessTitle(input string, safetyLevel int) string {
	if safetyLevel >= internal.SAFETYLEVELSTRICT {
		return escapeSpecialChars(input)
	}

	return utils.TitlePattern.ReplaceAllStringFunc(input, func(m string) string {
		match := utils.TitlePattern.FindStringSubmatch(m)
		level := len(match[1])
		title := strings.TrimSpace(match[2])
		if level <= 2 {
			return fmt.Sprintf("*%s*", title)
		}
		return fmt.Sprintf("_%s_", title)
	})
}

func ProcessList(input string, safetyLevel int) string {
	if safetyLevel >= internal.SAFETYLEVELSTRICT {
		return escapeSpecialChars(input)
	}

	var builder strings.Builder
	lines := strings.Split(input, "\n")
	listCounter := 1

	for _, line := range lines {
		switch {
		case utils.ListItemPattern.MatchString(line):
			match := utils.ListItemPattern.FindStringSubmatch(line)
			item := match[1] // Texto do item da lista
			// Processa a formatação inline primeiro
			item = ProcessInlineFormatting(item)
			// Escapa caracteres especiais que não fazem parte da formatação
			item = escapeNonFormatChars(item)
			builder.WriteString(fmt.Sprintf("• %s\n", item))
		case utils.OrderedListPattern.MatchString(line):
			match := utils.OrderedListPattern.FindStringSubmatch(line)
			item := match[1] // Texto do item da lista
			// Processa a formatação inline primeiro
			item = ProcessInlineFormatting(item)
			// Escapa caracteres especiais que não fazem parte da formatação
			item = escapeNonFormatChars(item)
			builder.WriteString(fmt.Sprintf("%d. %s\n", listCounter, item))
			listCounter++
		default:
			// Para linhas que não são itens de lista, escapa se necessário
			if safetyLevel == internal.SAFETYLEVELBASIC {
				line = escapeNonFormatChars(line)
			}
			builder.WriteString(line)
			builder.WriteString("\n")
			listCounter = 1
		}
	}

	return strings.TrimSpace(builder.String())
}

func ProcessQuote(input string, safetyLevel int) string {
	if safetyLevel >= internal.SAFETYLEVELSTRICT {
		return escapeSpecialChars(input)
	}

	lines := strings.Split(input, "\n")
	var result []string

	for _, line := range lines {
		if utils.BlockquotePattern.MatchString(line) {
			match := utils.BlockquotePattern.FindStringSubmatch(line)
			quote := ProcessInlineFormatting(match[1])
			result = append(result, fmt.Sprintf("> %s", quote))
		} else {
			result = append(result, line)
		}
	}

	return strings.Join(result, "\n")
}
