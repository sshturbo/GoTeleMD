package formatter

import (
	"fmt"
	"strings"

	"github.com/sshturbo/GoTeleMD/internal"
	"github.com/sshturbo/GoTeleMD/pkg/utils"
)

func ProcessText(input string, safetyLevel int) string {
	if safetyLevel == internal.SAFETYLEVELSTRICT {
		return escapeSpecialChars(input)
	}

	if safetyLevel == internal.SAFETYLEVELBASIC {
		parts := strings.Split(input, "```")
		for i := range parts {
			if i%2 == 0 {
				text := parts[i]
				text = ProcessInlineFormatting(text)
				text = processLinks(text)

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
	specialChars := []string{"#", "+", "-", "_", "=", "|", ".", "!", "(", ")", "{", "}"}
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
