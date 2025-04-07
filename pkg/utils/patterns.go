package utils

import "regexp"

var (
	TitlePattern       = regexp.MustCompile(`(?m)^(#{1,6})\s*(.+)$`)
	BoldPattern        = regexp.MustCompile(`(\*\*)(.*?)\*\*|(__)(.*?)__|(\*)([^*\n]+?)(\*)`)
	ItalicPattern      = regexp.MustCompile(`(_)([^_\n]+?)(_)`)
	RiscadoPattern     = regexp.MustCompile(`~~(.*?)~~`)
	ListItemPattern    = regexp.MustCompile(`(?m)^\s*[-\*]\s+(.+)$`)
	OrderedListPattern = regexp.MustCompile(`(?m)^\s*\d+\.\s+(.+)$`)
	BlockquotePattern  = regexp.MustCompile(`(?m)^>\s*(.+)$`)
	InlineCodePattern  = regexp.MustCompile("`([^`\n]+)`")
	LinkPattern        = regexp.MustCompile(`\[(.*?)\]\((.*?)\)`)
	TableLinePattern   = regexp.MustCompile(`(?m)^\|(.+)\|$`)
	SeparatorLine      = regexp.MustCompile(`^\s*[:\-\| ]+\s*$`)
)
