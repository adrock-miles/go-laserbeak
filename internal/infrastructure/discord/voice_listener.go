package discord

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/adrock-miles/go-laserbeak/internal/infrastructure/audio"
	"github.com/bwmarrin/discordgo"
)

const (
	// silenceTimeout is how long to wait after the last audio packet
	// before considering a user has stopped speaking.
	silenceTimeout = 1500 * time.Millisecond

	// minSpeechFrames filters out very short audio bursts (noise/pings).
	minSpeechFrames = 25 // ~500ms at 20ms per frame
)

// VoiceTranscription represents a completed voice utterance from a user.
type VoiceTranscription struct {
	UserID    string
	GuildID   string
	ChannelID string // text channel to respond in
	Audio     []byte // WAV-encoded audio
}

// VoiceListener manages voice connections and collects user audio.
type VoiceListener struct {
	mu          sync.RWMutex
	connections map[string]*voiceConn // guildID -> voiceConn
	resultChan  chan VoiceTranscription

	ssrcMu     sync.RWMutex
	ssrcToUser map[uint32]string // SSRC -> userID (populated by VoiceSpeakingUpdate)
}

type voiceConn struct {
	vc            *discordgo.VoiceConnection
	textChannelID string
	cancel        context.CancelFunc
}

// NewVoiceListener creates a new VoiceListener.
func NewVoiceListener() *VoiceListener {
	return &VoiceListener{
		connections: make(map[string]*voiceConn),
		resultChan:  make(chan VoiceTranscription, 64),
		ssrcToUser:  make(map[uint32]string),
	}
}

// Results returns the channel that delivers completed transcriptions.
func (vl *VoiceListener) Results() <-chan VoiceTranscription {
	return vl.resultChan
}

// Join connects to a voice channel and begins listening.
func (vl *VoiceListener) Join(s *discordgo.Session, guildID, voiceChannelID, textChannelID string) error {
	vl.mu.Lock()
	defer vl.mu.Unlock()

	// Leave existing connection in this guild if any
	if existing, ok := vl.connections[guildID]; ok {
		existing.cancel()
		existing.vc.Disconnect()
		delete(vl.connections, guildID)
	}

	vc, err := s.ChannelVoiceJoin(guildID, voiceChannelID, true, false) // mute=true (listen only), deaf=false
	if err != nil {
		return err
	}

	// Register handler to map SSRC -> UserID via speaking events
	vc.AddHandler(vl.onSpeakingUpdate)

	ctx, cancel := context.WithCancel(context.Background())
	conn := &voiceConn{
		vc:            vc,
		textChannelID: textChannelID,
		cancel:        cancel,
	}
	vl.connections[guildID] = conn

	go vl.listenLoop(ctx, conn)

	return nil
}

// onSpeakingUpdate handles VoiceSpeakingUpdate events to map SSRC -> UserID.
func (vl *VoiceListener) onSpeakingUpdate(vc *discordgo.VoiceConnection, vs *discordgo.VoiceSpeakingUpdate) {
	if vs.UserID != "" {
		vl.ssrcMu.Lock()
		vl.ssrcToUser[uint32(vs.SSRC)] = vs.UserID
		vl.ssrcMu.Unlock()
	}
}

// getUserID resolves a SSRC to a user ID.
func (vl *VoiceListener) getUserID(ssrc uint32) string {
	vl.ssrcMu.RLock()
	defer vl.ssrcMu.RUnlock()
	return vl.ssrcToUser[ssrc]
}

// Leave disconnects from voice in a guild. Returns true if a connection was found and disconnected.
func (vl *VoiceListener) Leave(guildID string) bool {
	vl.mu.Lock()
	defer vl.mu.Unlock()

	conn, ok := vl.connections[guildID]
	if !ok {
		return false
	}

	conn.cancel()
	conn.vc.Disconnect()
	delete(vl.connections, guildID)
	log.Printf("Disconnected from voice in guild %s", guildID)
	return true
}

// LeaveAll disconnects from all voice channels.
func (vl *VoiceListener) LeaveAll() {
	vl.mu.Lock()
	defer vl.mu.Unlock()

	for guildID, conn := range vl.connections {
		conn.cancel()
		conn.vc.Disconnect()
		delete(vl.connections, guildID)
	}
}

// listenLoop receives Opus packets from Discord and assembles per-user utterances.
func (vl *VoiceListener) listenLoop(ctx context.Context, conn *voiceConn) {
	type userBuffer struct {
		pcm      []int16 // accumulated PCM samples (flat)
		lastSeen time.Time
		frames   int // number of frames collected (for minSpeechFrames check)
	}

	decoder, err := audio.NewOpusDecoder()
	if err != nil {
		log.Printf("error creating opus decoder: %v", err)
		return
	}

	// Reusable decode buffer — avoids allocation per packet.
	decodeBuf := make([]int16, audio.FrameSize*audio.Channels)

	buffers := make(map[uint32]*userBuffer) // SSRC -> buffer

	ticker := time.NewTicker(200 * time.Millisecond)
	defer ticker.Stop()

	opusChan := conn.vc.OpusRecv

	for {
		select {
		case <-ctx.Done():
			return

		case pkt, ok := <-opusChan:
			if !ok {
				return
			}

			n, err := decoder.DecodeInto(pkt.Opus, decodeBuf)
			if err != nil {
				continue
			}

			buf, exists := buffers[pkt.SSRC]
			if !exists {
				// Pre-allocate for ~2s of audio to reduce grow-copies.
				buf = &userBuffer{pcm: make([]int16, 0, audio.SampleRate*audio.Channels*2)}
				buffers[pkt.SSRC] = buf
			}
			buf.pcm = append(buf.pcm, decodeBuf[:n]...)
			buf.frames++
			buf.lastSeen = time.Now()

		case <-ticker.C:
			now := time.Now()
			for ssrc, buf := range buffers {
				if now.Sub(buf.lastSeen) < silenceTimeout {
					continue
				}

				if buf.frames >= minSpeechFrames {
					userID := vl.getUserID(ssrc)
					vl.emitPCM(conn, userID, buf.pcm)
				}

				delete(buffers, ssrc)
			}
		}
	}
}

// emitPCM converts accumulated PCM samples to WAV and sends to results channel.
func (vl *VoiceListener) emitPCM(conn *voiceConn, userID string, pcm []int16) {
	wav, err := audio.PCMToWAV(pcm, audio.SampleRate, audio.Channels)
	if err != nil {
		log.Printf("error encoding WAV: %v", err)
		return
	}

	vl.resultChan <- VoiceTranscription{
		UserID:    userID,
		GuildID:   conn.vc.GuildID,
		ChannelID: conn.textChannelID,
		Audio:     wav,
	}
}
