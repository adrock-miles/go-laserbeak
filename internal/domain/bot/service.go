package bot

import "context"

// LLMMessage represents a message sent to or received from an LLM.
type LLMMessage struct {
	Role    string
	Content string
}

// LLMService defines the port for interacting with a language model.
type LLMService interface {
	// ChatCompletion sends a list of messages and returns the assistant's reply.
	ChatCompletion(ctx context.Context, messages []LLMMessage) (string, error)
}
