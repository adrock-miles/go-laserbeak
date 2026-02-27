package conversation

// Repository defines the interface for conversation persistence.
type Repository interface {
	// FindByChannel retrieves a conversation by its Discord channel ID.
	FindByChannel(channelID string) (*Conversation, bool)

	// Save persists a conversation.
	Save(conv *Conversation)

	// Delete removes a conversation by channel ID.
	Delete(channelID string)
}
