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

const (
	SAFETYLEVELNONE   = 0
	SAFETYLEVELBASIC  = 1
	SAFETYLEVELSTRICT = 2
)

const TelegramMaxLength = 4096

var (
	EnableLogs *bool
)
