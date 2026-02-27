package bot

import "context"

// STTService defines the port for speech-to-text transcription.
type STTService interface {
	// Transcribe converts raw audio (Opus/PCM) into text.
	Transcribe(ctx context.Context, audioData []byte) (string, error)
}
