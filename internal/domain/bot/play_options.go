package bot

import "context"

// PlayOption represents a single playable item from the external bot.
type PlayOption struct {
	Name string
}

// PlayOptionsService defines the port for fetching available play options.
type PlayOptionsService interface {
	// GetOptions returns the current list of available play options.
	GetOptions(ctx context.Context) ([]PlayOption, error)
}
