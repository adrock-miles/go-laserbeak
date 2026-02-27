package application

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/adrock-miles/GoBot-Laserbeak/internal/domain/bot"
)

// VoiceService handles voice-to-text-to-LLM pipeline.
// It transcribes audio, then routes the text through the ChatService.
type VoiceService struct {
	stt         bot.STTService
	chatService *ChatService
}

// NewVoiceService creates a new VoiceService.
func NewVoiceService(stt bot.STTService, chatService *ChatService) *VoiceService {
	return &VoiceService{
		stt:         stt,
		chatService: chatService,
	}
}

// HandleVoice transcribes audio and generates an LLM response.
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

	reply, err := s.chatService.HandleMessage(ctx, channelID, userID, text)
	if err != nil {
		return "", fmt.Errorf("handle transcribed message: %w", err)
	}

	return reply, nil
}
