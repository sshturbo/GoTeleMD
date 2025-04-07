package formatter

import (
	"fmt"
	"strings"

	"github.com/sshturbo/GoTeleMD/internal"
	"github.com/sshturbo/GoTeleMD/pkg/utils"
)

func processCodeBlocks(text string) string {
	var result strings.Builder
	var inCodeBlock bool
	var codeBlocks []string
	var currentBlock strings.Builder

	// Remover quebras de linha extras
	text = strings.TrimSpace(text)
	lines := strings.Split(text, "\n")

	// Primeiro passo: identificar e armazenar blocos de código únicos
	for _, line := range lines {
		if strings.HasPrefix(line, "```") {
			if inCodeBlock {
				// Finaliza o bloco atual
				codeBlocks = append(codeBlocks, currentBlock.String())
				currentBlock.Reset()
			}
			inCodeBlock = !inCodeBlock
			continue
		}

		if inCodeBlock {
			currentBlock.WriteString(line + "\n")
		}
	}

	// Remove blocos duplicados mantendo a ordem
	seen := make(map[string]bool)
	uniqueBlocks := make([]string, 0, len(codeBlocks))
	for _, block := range codeBlocks {
		if !seen[block] {
			seen[block] = true
			uniqueBlocks = append(uniqueBlocks, block)
		}
	}

	// Segundo passo: reconstruir o texto com blocos únicos
	inCodeBlock = false
	blockIndex := 0
	lastLineWasEmpty := false

	for i, line := range lines {
		isCodeBlockMarker := strings.HasPrefix(line, "```")

		if isCodeBlockMarker {
			if inCodeBlock {
				// Finaliza o bloco atual
				result.WriteString("```")
				blockIndex++
				if i < len(lines)-1 {
					result.WriteString("\n")
				}
			} else if blockIndex < len(uniqueBlocks) {
				// Evita quebras de linha extras antes do bloco de código
				if i > 0 && !lastLineWasEmpty {
					result.WriteString("\n")
				}
				result.WriteString("```")
				if line != "```" {
					result.WriteString(strings.TrimPrefix(line, "```"))
				}
				result.WriteString("\n")
			}
			inCodeBlock = !inCodeBlock
			lastLineWasEmpty = false
			continue
		}

		if inCodeBlock && blockIndex < len(uniqueBlocks) {
			// Escapar caracteres especiais dentro do bloco de código
			escaped := strings.NewReplacer(
				"_", "\\_",
				"*", "\\*",
				"[", "\\[",
				"]", "\\]",
				"(", "\\(",
				")", "\\)",
				"~", "\\~",
				"`", "\\`",
				">", "\\>",
				"#", "\\#",
				"+", "\\+",
				"-", "\\-",
				"=", "\\=",
				"|", "\\|",
				"{", "\\{",
				"}", "\\}",
				".", "\\.",
				"!", "\\!",
			).Replace(line)
			result.WriteString(escaped + "\n")
			lastLineWasEmpty = line == ""
		} else if !inCodeBlock {
			result.WriteString(line)
			if i < len(lines)-1 {
				result.WriteString("\n")
			}
			lastLineWasEmpty = line == ""
		}
	}

	return strings.TrimSpace(result.String())
}

func ProcessText(input string, safetyLevel int) string {
	if safetyLevel == internal.SAFETYLEVELSTRICT {
		// No modo strict, escapa tudo, incluindo as marcações ``` e conteúdo
		return escapeSpecialChars(input)
	}

	if safetyLevel == internal.SAFETYLEVELBASIC {
		text := processCodeBlocks(input)
		text = ProcessInlineFormatting(text)
		text = processLinks(text)
		return text
	}

	// Para SAFETYLEVEL_NONE
	text := input
	text = ProcessInlineFormatting(text)
	text = processLinks(text)
	return text
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
	escaped := text
	specialChars := []string{"#", "+", "-", "=", "|", ".", "!", "(", ")", "{", "}"}
	for _, char := range specialChars {
		escaped = strings.ReplaceAll(escaped, char, "\\"+char)
	}
	return escaped
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
			item := ProcessInlineFormatting(match[1])
			builder.WriteString(fmt.Sprintf("• %s\n", item))
		case utils.OrderedListPattern.MatchString(line):
			match := utils.OrderedListPattern.FindStringSubmatch(line)
			item := ProcessInlineFormatting(match[1])
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
