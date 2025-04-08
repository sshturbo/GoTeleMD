package types

type Config struct {
	SafetyLevel          int
	AlignTableColumns    bool
	IgnoreTableSeparator bool
	MaxMessageLength     int
	EnableDebugLogs      bool
	CustomEscapeChars    []string
	PreserveEmptyLines   bool
	StrictLineBreaks     bool
}

func DefaultConfig() *Config {
	return &Config{
		SafetyLevel:          1, // SAFETYLEVELBASIC
		AlignTableColumns:    true,
		IgnoreTableSeparator: false,
		MaxMessageLength:     4096, // TelegramMaxLength
		EnableDebugLogs:      false,
		PreserveEmptyLines:   true,
		StrictLineBreaks:     true,
	}
}

type Option func(*Config)

func WithSafetyLevel(level int) Option {
	return func(c *Config) {
		c.SafetyLevel = level
	}
}

func WithTableAlignment(align bool) Option {
	return func(c *Config) {
		c.AlignTableColumns = align
	}
}

func WithTableSeparators(ignore bool) Option {
	return func(c *Config) {
		c.IgnoreTableSeparator = ignore
	}
}

func WithMaxMessageLength(length int) Option {
	return func(c *Config) {
		c.MaxMessageLength = length
	}
}

func WithDebugLogs(enable bool) Option {
	return func(c *Config) {
		c.EnableDebugLogs = enable
	}
}

func WithCustomEscapeChars(chars []string) Option {
	return func(c *Config) {
		c.CustomEscapeChars = chars
	}
}

func WithPreserveEmptyLines(preserve bool) Option {
	return func(c *Config) {
		c.PreserveEmptyLines = preserve
	}
}

func WithStrictLineBreaks(strict bool) Option {
	return func(c *Config) {
		c.StrictLineBreaks = strict
	}
}
