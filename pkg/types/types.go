package types

type MessagePart struct {
	Part    int    `json:"part"`
	Content string `json:"content"`
}

type MessageResponse struct {
	MessageID  string        `json:"message_id"`
	TotalParts int           `json:"total_parts"`
	Parts      []MessagePart `json:"parts"`
}
