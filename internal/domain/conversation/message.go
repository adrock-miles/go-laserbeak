package conversation

import "time"

// Role represents the sender role in a conversation message.
type Role string

const (
	RoleUser      Role = "user"
	RoleAssistant Role = "assistant"
	RoleSystem    Role = "system"
)

// Message is a value object representing a single message in a conversation.
type Message struct {
	Role      Role
	Content   string
	Timestamp time.Time
}

// NewMessage creates a new Message value object.
func NewMessage(role Role, content string) Message {
	return Message{
		Role:      role,
		Content:   content,
		Timestamp: time.Now(),
	}
}
