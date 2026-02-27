package conversation

// Conversation is the aggregate root for a chat conversation.
// Each conversation is scoped to a Discord channel.
type Conversation struct {
	ChannelID    string
	SystemPrompt string
	Messages     []Message
	MaxHistory   int
}

// NewConversation creates a new Conversation for the given channel.
func NewConversation(channelID string, systemPrompt string, maxHistory int) *Conversation {
	c := &Conversation{
		ChannelID:    channelID,
		SystemPrompt: systemPrompt,
		Messages:     make([]Message, 0),
		MaxHistory:   maxHistory,
	}
	return c
}

// AddMessage appends a message and trims history if it exceeds MaxHistory.
func (c *Conversation) AddMessage(msg Message) {
	c.Messages = append(c.Messages, msg)
	if c.MaxHistory > 0 && len(c.Messages) > c.MaxHistory {
		c.Messages = c.Messages[len(c.Messages)-c.MaxHistory:]
	}
}

// AllMessages returns the system prompt plus conversation messages
// in the format expected by LLM APIs.
func (c *Conversation) AllMessages() []Message {
	msgs := make([]Message, 0, len(c.Messages)+1)
	if c.SystemPrompt != "" {
		msgs = append(msgs, NewMessage(RoleSystem, c.SystemPrompt))
	}
	msgs = append(msgs, c.Messages...)
	return msgs
}

// Clear resets the conversation history.
func (c *Conversation) Clear() {
	c.Messages = c.Messages[:0]
}
