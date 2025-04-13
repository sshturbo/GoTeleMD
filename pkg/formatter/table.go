package formatter

import (
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/sshturbo/GoTeleMD/pkg/utils"
)

func ConvertTable(lines []string, align, ignoreSeparators bool) string {
	var rows [][]string
	maxCols := 0
	var alignments []string

	for i, line := range lines {
		if utils.SeparatorLine.MatchString(line) {
			if !ignoreSeparators && i == 1 && i < len(lines)-1 {
				alignments = parseTableAlignment(line)
			}
			continue
		}

		// Remove apenas os pipes das extremidades
		line = strings.TrimSpace(line)
		line = strings.TrimPrefix(line, "|")
		line = strings.TrimSuffix(line, "|")

		// Divide a linha considerando escapes
		var cols []string
		var currentCol strings.Builder
		escaped := false

		for _, char := range line {
			if char == '\\' && !escaped {
				escaped = true
				continue
			}

			if char == '|' && !escaped {
				cols = append(cols, currentCol.String())
				currentCol.Reset()
			} else {
				if escaped && char != '|' {
					currentCol.WriteRune('\\')
				}
				currentCol.WriteRune(char)
				escaped = false
			}
		}
		cols = append(cols, currentCol.String())

		var clean []string
		for _, col := range cols {
			clean = append(clean, ProcessInlineFormatting(strings.TrimSpace(col)))
		}
		if len(clean) > 0 {
			rows = append(rows, clean)
			if len(clean) > maxCols {
				maxCols = len(clean)
			}
		}
	}

	alignments = normalizeAlignments(alignments, maxCols)
	colWidths := calculateColumnWidths(rows, maxCols)
	return formatTable(rows, colWidths, alignments, align)
}

func normalizeAlignments(alignments []string, maxCols int) []string {
	if len(alignments) < maxCols {
		newAlignments := make([]string, maxCols)
		for i := range newAlignments {
			if i < len(alignments) {
				newAlignments[i] = alignments[i]
			} else {
				newAlignments[i] = "l"
			}
		}
		return newAlignments
	}
	return alignments
}

func calculateColumnWidths(rows [][]string, maxCols int) []int {
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
	return colWidths
}

func formatTable(rows [][]string, colWidths []int, alignments []string, align bool) string {
	var builder strings.Builder
	builder.WriteString("\n")

	for _, row := range rows {
		var formattedColumns []string
		for i := 0; i < len(colWidths); i++ {
			var col string
			if i < len(row) {
				col = row[i]
			} else {
				col = ""
			}

			if align {
				col = alignColumn(col, colWidths[i], getAlignType(alignments, i))
			}
			formattedColumns = append(formattedColumns, col)
		}

		line := formatTableRow(formattedColumns, align)
		builder.WriteString(line + "\n")
	}

	return strings.TrimRight(builder.String(), "\n")
}

func alignColumn(col string, width int, alignType string) string {
	if width < 5 {
		width = 5
	}

	switch alignType {
	case "c":
		pad := width - utf8.RuneCountInString(col)
		leftPad := pad / 2
		rightPad := pad - leftPad
		return strings.Repeat(" ", leftPad) + col + strings.Repeat(" ", rightPad)
	case "r":
		return strings.Repeat(" ", width-utf8.RuneCountInString(col)) + col
	default: // "l"
		return col + strings.Repeat(" ", width-utf8.RuneCountInString(col))
	}
}

func getAlignType(alignments []string, index int) string {
	if index < len(alignments) {
		return alignments[index]
	}
	return "l"
}

func formatTableRow(columns []string, align bool) string {
	// Escapa os caracteres | em cada coluna
	escapedColumns := make([]string, len(columns))
	specialChars := []string{"[", "]", "(", ")", ">", "#", "+", "-", "=", "|", "{", "}", ".", "!", "\\"}

	for i, col := range columns {
		// Primeiro escapa os pontos no conteúdo
		escapedCol := col
		runes := []rune(escapedCol)
		var result strings.Builder
		lastWasDigit := false

		for i, r := range runes {
			c := string(r)

			if unicode.IsDigit(r) {
				lastWasDigit = true
				result.WriteRune(r)
				continue
			}

			isSpecial := false
			for _, special := range specialChars {
				if c == special {
					if c == "." && lastWasDigit && i+1 < len(runes) && unicode.IsDigit(runes[i+1]) {
						// Não escapa ponto entre números
						isSpecial = false
						break
					}
					result.WriteString("\\" + c)
					isSpecial = true
					break
				}
			}

			if !isSpecial {
				result.WriteRune(r)
			}
			lastWasDigit = false
		}

		escapedCol = result.String()
		escapedColumns[i] = escapedCol
	}

	prefix := "\\•  "
	if !align {
		prefix = "\\• "
	}

	return strings.TrimSpace(prefix + strings.Join(escapedColumns, " \\| "))
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
