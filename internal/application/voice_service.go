package application

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/adrock-miles/GoBot-Laserbeak/internal/domain/bot"
)

// VoiceCommand represents a parsed voice command result.
type VoiceCommand struct {
	// Text is the message to send to the output text channel.
	Text string
}

// VoiceService handles voice-to-text-to-command pipeline.
// It transcribes audio, checks for the wake phrase, and parses voice commands.
type VoiceService struct {
	stt        bot.STTService
	wakePhrase string
}

// NewVoiceService creates a new VoiceService.
func NewVoiceService(stt bot.STTService, wakePhrase string) *VoiceService {
	return &VoiceService{
		stt:        stt,
		wakePhrase: strings.ToLower(wakePhrase),
	}
}

// HandleVoice transcribes audio and parses voice commands.
// Returns the command text to send to chat, or empty string if no valid command.
func (s *VoiceService) HandleVoice(ctx context.Context, channelID, userID string, audioWAV []byte) (string, error) {
	text, err := s.stt.Transcribe(ctx, audioWAV)
	if err != nil {
		return "", fmt.Errorf("transcribe audio: %w", err)
	}

	text = strings.TrimSpace(text)
	if text == "" {
		return "", nil
	}

	log.Printf("voice transcription from user %s: %s", userID, text)

	cmd, ok := s.parseCommand(text)
	if !ok {
		return "", nil
	}

	log.Printf("voice command from user %s: %s", userID, cmd.Text)
	return cmd.Text, nil
}

// parseCommand checks if the transcription starts with the wake phrase
// and parses the subsequent command.
func (s *VoiceService) parseCommand(transcription string) (VoiceCommand, bool) {
	lower := strings.ToLower(transcription)

	// Check for wake phrase
	if !strings.HasPrefix(lower, s.wakePhrase) {
		return VoiceCommand{}, false
	}

	// Extract everything after the wake phrase
	rest := strings.TrimSpace(transcription[len(s.wakePhrase):])
	restLower := strings.ToLower(rest)

	switch {
	case restLower == "stop":
		return VoiceCommand{Text: "!stop"}, true

	case strings.HasPrefix(restLower, "play"):
		query := strings.TrimSpace(rest[len("play"):])
		if query == "" {
			return VoiceCommand{}, false
		}
		return VoiceCommand{Text: "!play " + query}, true
	}

	return VoiceCommand{}, false
}
