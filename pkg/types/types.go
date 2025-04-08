package types

// MessagePart representa uma parte individual da mensagem
type MessagePart struct {
	Part    int    `json:"part"`
	Content string `json:"content"`
}

// MessageResponse representa a estrutura da resposta em JSON
type MessageResponse struct {
	MessageID  string        `json:"message_id"`
	TotalParts int           `json:"total_parts"`
	Parts      []MessagePart `json:"parts"`
}
