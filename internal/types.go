package internal

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
