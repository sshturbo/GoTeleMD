package formatter

import (
	"fmt"
	"regexp"
	"strings"
	"unicode"

	"github.com/sshturbo/GoTeleMD/internal"
	"github.com/sshturbo/GoTeleMD/pkg/utils"
)

// isBalancedFormatting verifica se os caracteres de formatação estão balanceados
func isBalancedFormatting(text string, char string) bool {
	count := 0
	escaped := false
	for _, r := range text {
		c := string(r)
		if c == "\\" {
			escaped = !escaped
			continue
		}
		if !escaped && c == char {
			count++
		}
		escaped = false
	}
	return count%2 == 0
}

// findUnbalancedFormatting encontra posições de caracteres de formatação não balanceados
func findUnbalancedFormatting(text string) map[string][]int {
	unbalanced := make(map[string][]int)
	formatChars := []string{"*", "_", "~", "`"}

	for _, char := range formatChars {
		var stack []int
		escaped := false

		for i, r := range text {
			c := string(r)
			if c == "\\" {
				escaped = !escaped
				continue
			}
			if !escaped && c == char {
				if len(stack) > 0 {
					stack = stack[:len(stack)-1] // Remove posição pareada
				} else {
					stack = append(stack, i) // Adiciona posição não pareada
				}
			}
			escaped = false
		}

		if len(stack) > 0 {
			unbalanced[char] = stack
		}
	}

	return unbalanced
}

func escapeSpecialChars(text string) string {
	escaped := text
	specialChars := []string{"[", "]", "(", ")", ">", "#", "+", "-", "=", "|", "{", "}", ".", "!", "\\."}

	// Primeiro escapa as barras invertidas
	escaped = strings.ReplaceAll(escaped, "\\", "\\\\")

	// Depois escapa os caracteres especiais não relacionados à formatação
	for _, char := range specialChars {
		if char == "\\." {
			// Trata especialmente o ponto para não escapar dentro de números
			escaped = regexp.MustCompile(`(\d*?)\.(\d+)`).ReplaceAllString(escaped, "${1}\\.${2}")
		} else {
			escaped = strings.ReplaceAll(escaped, char, "\\"+char)
		}
	}

	// Por fim, trata os caracteres de formatação
	formatChars := []string{"*", "_", "~", "`"}
	for _, char := range formatChars {
		if !isBalancedFormatting(text, char) {
			// Se não estiver balanceado, escapa todas as ocorrências
			escaped = strings.ReplaceAll(escaped, char, "\\"+char)
		}
	}

	return escaped
}

func escapeNonFormatChars(text string) string {
	var result strings.Builder
	result.Grow(len(text) * 2)

	// Encontra caracteres de formatação não balanceados
	unbalanced := findUnbalancedFormatting(text)

	specialChars := []string{"#", "+", "-", "=", "|", ".", "!", "(", ")", "{", "}"}
	escaped := false
	lastWasDigit := false

	for i, r := range text {
		c := string(r)

		if c == "\\" {
			escaped = !escaped
			result.WriteString(c)
			lastWasDigit = false
			continue
		}

		if escaped {
			result.WriteString(c)
			escaped = false
			lastWasDigit = unicode.IsDigit(r)
			continue
		}

		// Verifica se é um caractere de formatação
		if c == "*" || c == "_" || c == "~" || c == "`" {
			// Se a posição atual está na lista de não balanceados, escapa
			if positions, exists := unbalanced[c]; exists {
				isUnbalanced := false
				for _, pos := range positions {
					if pos == i {
						isUnbalanced = true
						break
					}
				}
				if isUnbalanced {
					result.WriteString("\\")
				}
			}
			result.WriteString(c)
			lastWasDigit = false
			continue
		}

		// Escapa caracteres especiais não relacionados à formatação
		for _, special := range specialChars {
			if c == special {
				// Não escapa o ponto se estiver entre dígitos
				if c == "." && lastWasDigit {
					nextIsDigit := false
					if i+1 < len(text) {
						nextRune := []rune(text)[i+1]
						nextIsDigit = unicode.IsDigit(nextRune)
					}
					if nextIsDigit {
						break
					}
				}
				result.WriteString("\\")
				break
			}
		}

		result.WriteString(c)
		escaped = false
		lastWasDigit = unicode.IsDigit(r)
	}

	return result.String()
}

func ProcessText(input string, safetyLevel int) string {
	if safetyLevel == internal.SAFETYLEVELSTRICT {
		return escapeSpecialChars(input)
	}

	if safetyLevel == internal.SAFETYLEVELBASIC {
		parts := strings.Split(input, "```")
		for i := range parts {
			if i%2 == 0 {
				text := parts[i]

				// Primeiro faz o processamento de formatação inline
				text = ProcessInlineFormatting(text)
				text = processLinks(text)

				// Depois trata os pontos e outros caracteres especiais
				text = processSpecialChars(text)

				// Por fim, processa blocos de código inline
				var result strings.Builder
				lastIndex := 0
				for _, match := range utils.InlineCodePattern.FindAllStringSubmatchIndex(text, -1) {
					prefix := text[lastIndex:match[0]]
					result.WriteString(escapeNonFormatChars(prefix))
					codeContent := text[match[2]:match[3]]
					result.WriteString("`")
					result.WriteString(escapeSpecialChars(codeContent))
					result.WriteString("`")
					lastIndex = match[1]
				}
				if lastIndex < len(text) {
					result.WriteString(escapeNonFormatChars(text[lastIndex:]))
				}
				parts[i] = result.String()
			} else {
				parts[i] = escapeSpecialChars(parts[i])
			}
		}
		return strings.Join(parts, "```")
	}

	text := input
	text = ProcessInlineFormatting(text)
	text = processLinks(text)
	return text
}

func processSpecialChars(text string) string {
	var result strings.Builder
	runes := []rune(text)
	lastWasDigit := false

	for i := 0; i < len(runes); i++ {
		r := runes[i]
		c := string(r)

		if unicode.IsDigit(r) {
			lastWasDigit = true
			result.WriteRune(r)
			continue
		}

		// Escapa bullet point no início da linha
		if c == "•" {
			prevIsNewline := i == 0 || (i > 0 && string(runes[i-1]) == "\n")
			if prevIsNewline {
				result.WriteString("\\•")
				continue
			}
		}

		if c == "." {
			// Verifica se é um ponto entre números (número decimal)
			if lastWasDigit && i+1 < len(runes) && unicode.IsDigit(runes[i+1]) {
				result.WriteRune(r)
				continue
			}

			// Escapa o ponto em todas as outras situações
			result.WriteString("\\.")
		} else {
			result.WriteRune(r)
		}

		lastWasDigit = false
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
	text = utils.BoldPattern.ReplaceAllStringFunc(text, func(m string) string {
		match := utils.BoldPattern.FindStringSubmatch(m)
		if match[2] != "" {
			return "*" + strings.TrimSpace(match[2]) + "*"
		} else if match[4] != "" {
			return "*" + strings.TrimSpace(match[4]) + "*"
		} else if match[6] != "" {
			return "_" + strings.TrimSpace(match[6]) + "_"
		}
		return m
	})

	text = utils.ItalicPattern.ReplaceAllStringFunc(text, func(m string) string {
		match := utils.ItalicPattern.FindStringSubmatch(m)
		if match[2] != "" {
			return "_" + strings.TrimSpace(match[2]) + "_"
		}
		return m
	})

	text = utils.InlineCodePattern.ReplaceAllString(text, "`$1`")

	text = utils.RiscadoPattern.ReplaceAllStringFunc(text, func(m string) string {
		match := utils.RiscadoPattern.FindStringSubmatch(m)
		if len(match) < 2 {
			return m
		}
		return "~" + match[1] + "~"
	})

	return text
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
	isFirstItem := true
	lastLineWasList := false

	for _, line := range lines {
		trimmedLine := strings.TrimSpace(line)
		if trimmedLine == "" {
			if lastLineWasList {
				builder.WriteString("\n\n")
				lastLineWasList = false
			} else {
				builder.WriteString("\n")
			}
			isFirstItem = true
			continue
		}

		switch {
		case utils.ListItemPattern.MatchString(line):
			if !isFirstItem && lastLineWasList {
				builder.WriteString("\n")
			}
			match := utils.ListItemPattern.FindStringSubmatch(line)
			item := match[1]
			item = ProcessInlineFormatting(item)
			item = escapeNonFormatChars(item)
			builder.WriteString(fmt.Sprintf("• %s", item))
			isFirstItem = false
			lastLineWasList = true
		case utils.OrderedListPattern.MatchString(line):
			if !isFirstItem && lastLineWasList {
				builder.WriteString("\n")
			}
			match := utils.OrderedListPattern.FindStringSubmatch(line)
			item := match[1]
			item = ProcessInlineFormatting(item)
			item = escapeNonFormatChars(item)
			builder.WriteString(fmt.Sprintf("%d\\. %s", listCounter, item))
			listCounter++
			isFirstItem = false
			lastLineWasList = true
		default:
			if lastLineWasList {
				builder.WriteString("\n\n")
			} else if !isFirstItem {
				builder.WriteString("\n")
			}
			if safetyLevel == internal.SAFETYLEVELBASIC {
				line = escapeNonFormatChars(line)
			}
			builder.WriteString(line)
			listCounter = 1
			isFirstItem = false
			lastLineWasList = false
		}
	}

	result := strings.TrimSpace(builder.String())
	if lastLineWasList {
		result += "\n"
	}
	return result
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
