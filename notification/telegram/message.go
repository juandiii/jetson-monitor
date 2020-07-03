package telegram

type SendMessage struct {
	ChatID int    `json:"chat_id"`
	Text   string `json:"text"`
}
