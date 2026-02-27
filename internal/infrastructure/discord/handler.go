package discord

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/bwmarrin/discordgo"
)

// ChatHandler defines the callback for processing a chat message and returning a response.
type ChatHandler func(ctx context.Context, channelID, userID, content string) (string, error)

// VoiceTranscribeHandler defines the callback for transcribing audio and generating a response.
type VoiceTranscribeHandler func(ctx context.Context, channelID, userID string, audioWAV []byte) (string, error)

// Bot wraps the Discord session and routes messages to application-layer handlers.
type Bot struct {
	session        *discordgo.Session
	commandPrefix  string
	chatHandler    ChatHandler
	voiceHandler   VoiceTranscribeHandler
	voiceListener  *VoiceListener
}

// NewBot creates a new Discord Bot.
func NewBot(token, commandPrefix string) (*Bot, error) {
	s, err := discordgo.New("Bot " + token)
	if err != nil {
		return nil, fmt.Errorf("create discord session: %w", err)
	}

	s.Identify.Intents = discordgo.IntentsGuildMessages |
		discordgo.IntentsGuildVoiceStates |
		discordgo.IntentsMessageContent

	b := &Bot{
		session:       s,
		commandPrefix: commandPrefix,
		voiceListener: NewVoiceListener(),
	}

	s.AddHandler(b.onMessageCreate)

	return b, nil
}

// SetChatHandler sets the handler for text chat messages.
func (b *Bot) SetChatHandler(h ChatHandler) {
	b.chatHandler = h
}

// SetVoiceHandler sets the handler for voice transcriptions.
func (b *Bot) SetVoiceHandler(h VoiceTranscribeHandler) {
	b.voiceHandler = h
}

// Start opens the Discord websocket connection and begins listening.
func (b *Bot) Start() error {
	if err := b.session.Open(); err != nil {
		return fmt.Errorf("open discord session: %w", err)
	}

	// Start processing voice transcriptions
	if b.voiceHandler != nil {
		go b.processVoiceResults()
	}

	log.Println("Bot is online and listening")
	return nil
}

// Stop cleanly shuts down the bot.
func (b *Bot) Stop() {
	b.voiceListener.LeaveAll()
	b.session.Close()
}

// onMessageCreate handles incoming Discord messages.
func (b *Bot) onMessageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID {
		return
	}

	if !strings.HasPrefix(m.Content, b.commandPrefix) {
		return
	}

	content := strings.TrimPrefix(m.Content, b.commandPrefix)
	content = strings.TrimSpace(content)

	if content == "" {
		return
	}

	// Handle built-in commands
	switch {
	case content == "join":
		b.handleJoinVoice(s, m)
		return
	case content == "leave":
		b.handleLeaveVoice(s, m)
		return
	case content == "clear":
		b.handleClear(s, m)
		return
	case content == "help":
		b.handleHelp(s, m)
		return
	}

	// Route to chat handler
	if b.chatHandler == nil {
		return
	}

	s.ChannelTyping(m.ChannelID)

	reply, err := b.chatHandler(context.Background(), m.ChannelID, m.Author.ID, content)
	if err != nil {
		log.Printf("chat handler error: %v", err)
		s.ChannelMessageSend(m.ChannelID, "Sorry, I encountered an error processing your message.")
		return
	}

	b.sendLongMessage(s, m.ChannelID, reply)
}

// handleJoinVoice joins the voice channel the user is currently in.
func (b *Bot) handleJoinVoice(s *discordgo.Session, m *discordgo.MessageCreate) {
	guild, err := s.State.Guild(m.GuildID)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Could not find guild information.")
		return
	}

	var voiceChannelID string
	for _, vs := range guild.VoiceStates {
		if vs.UserID == m.Author.ID {
			voiceChannelID = vs.ChannelID
			break
		}
	}

	if voiceChannelID == "" {
		s.ChannelMessageSend(m.ChannelID, "You need to be in a voice channel first.")
		return
	}

	if err := b.voiceListener.Join(s, m.GuildID, voiceChannelID, m.ChannelID); err != nil {
		log.Printf("error joining voice: %v", err)
		s.ChannelMessageSend(m.ChannelID, "Failed to join your voice channel.")
		return
	}

	s.ChannelMessageSend(m.ChannelID, "Joined voice channel. I'll listen and respond here in text.")
}

// handleLeaveVoice leaves the voice channel in the current guild.
func (b *Bot) handleLeaveVoice(s *discordgo.Session, m *discordgo.MessageCreate) {
	b.voiceListener.Leave(m.GuildID)
	s.ChannelMessageSend(m.ChannelID, "Left voice channel.")
}

// handleClear resets conversation history for this channel.
func (b *Bot) handleClear(s *discordgo.Session, m *discordgo.MessageCreate) {
	// This is handled by sending "clear" content to the chat handler
	// which the application layer interprets
	if b.chatHandler != nil {
		b.chatHandler(context.Background(), m.ChannelID, m.Author.ID, "/clear")
	}
	s.ChannelMessageSend(m.ChannelID, "Conversation history cleared.")
}

// handleHelp sends usage information.
func (b *Bot) handleHelp(s *discordgo.Session, m *discordgo.MessageCreate) {
	help := fmt.Sprintf(`**Laserbeak Bot Commands**
`+"`%s <message>`"+` — Chat with the LLM
`+"`%s join`"+` — Join your voice channel and listen
`+"`%s leave`"+` — Leave voice channel
`+"`%s clear`"+` — Clear conversation history
`+"`%s help`"+` — Show this help`,
		b.commandPrefix, b.commandPrefix, b.commandPrefix, b.commandPrefix, b.commandPrefix)
	s.ChannelMessageSend(m.ChannelID, help)
}

// processVoiceResults consumes voice transcription results and forwards them to the voice handler.
func (b *Bot) processVoiceResults() {
	for trans := range b.voiceListener.Results() {
		if b.voiceHandler == nil {
			continue
		}

		go func(t VoiceTranscription) {
			b.session.ChannelTyping(t.ChannelID)

			reply, err := b.voiceHandler(context.Background(), t.ChannelID, t.UserID, t.Audio)
			if err != nil {
				log.Printf("voice handler error: %v", err)
				return
			}

			if strings.TrimSpace(reply) == "" {
				return
			}

			b.sendLongMessage(b.session, t.ChannelID, reply)
		}(trans)
	}
}

// sendLongMessage splits messages that exceed Discord's 2000 character limit.
func (b *Bot) sendLongMessage(s *discordgo.Session, channelID, content string) {
	const maxLen = 2000
	for len(content) > 0 {
		end := maxLen
		if end > len(content) {
			end = len(content)
		}
		s.ChannelMessageSend(channelID, content[:end])
		content = content[end:]
	}
}
