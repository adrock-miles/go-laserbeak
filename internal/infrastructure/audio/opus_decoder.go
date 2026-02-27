package audio

import (
	"bytes"
	"encoding/binary"
	"fmt"

	"gopkg.in/hraban/opus.v2"
)

const (
	// Discord sends Opus audio at 48kHz stereo.
	SampleRate = 48000
	Channels   = 2
	FrameSize  = 960 // 20ms at 48kHz
)

// OpusDecoder decodes Opus frames to PCM samples.
type OpusDecoder struct {
	decoder *opus.Decoder
}

// NewOpusDecoder creates a new Opus decoder for Discord audio.
func NewOpusDecoder() (*OpusDecoder, error) {
	dec, err := opus.NewDecoder(SampleRate, Channels)
	if err != nil {
		return nil, fmt.Errorf("create opus decoder: %w", err)
	}
	return &OpusDecoder{decoder: dec}, nil
}

// Decode decodes a single Opus frame into PCM int16 samples.
func (d *OpusDecoder) Decode(opusData []byte) ([]int16, error) {
	pcm := make([]int16, FrameSize*Channels)
	n, err := d.decoder.Decode(opusData, pcm)
	if err != nil {
		return nil, fmt.Errorf("decode opus frame: %w", err)
	}
	return pcm[:n*Channels], nil
}

// PCMToWAV converts raw PCM int16 samples to a WAV byte slice.
func PCMToWAV(samples []int16, sampleRate, channels int) ([]byte, error) {
	var buf bytes.Buffer

	dataSize := len(samples) * 2 // 2 bytes per int16

	// WAV header
	buf.WriteString("RIFF")
	if err := binary.Write(&buf, binary.LittleEndian, uint32(36+dataSize)); err != nil {
		return nil, fmt.Errorf("write RIFF size: %w", err)
	}
	buf.WriteString("WAVE")

	// fmt subchunk
	buf.WriteString("fmt ")
	if err := binary.Write(&buf, binary.LittleEndian, uint32(16)); err != nil {
		return nil, fmt.Errorf("write fmt size: %w", err)
	}
	if err := binary.Write(&buf, binary.LittleEndian, uint16(1)); err != nil { // PCM
		return nil, fmt.Errorf("write audio format: %w", err)
	}
	if err := binary.Write(&buf, binary.LittleEndian, uint16(channels)); err != nil {
		return nil, fmt.Errorf("write channels: %w", err)
	}
	if err := binary.Write(&buf, binary.LittleEndian, uint32(sampleRate)); err != nil {
		return nil, fmt.Errorf("write sample rate: %w", err)
	}
	byteRate := uint32(sampleRate * channels * 2)
	if err := binary.Write(&buf, binary.LittleEndian, byteRate); err != nil {
		return nil, fmt.Errorf("write byte rate: %w", err)
	}
	blockAlign := uint16(channels * 2)
	if err := binary.Write(&buf, binary.LittleEndian, blockAlign); err != nil {
		return nil, fmt.Errorf("write block align: %w", err)
	}
	if err := binary.Write(&buf, binary.LittleEndian, uint16(16)); err != nil { // bits per sample
		return nil, fmt.Errorf("write bits per sample: %w", err)
	}

	// data subchunk
	buf.WriteString("data")
	if err := binary.Write(&buf, binary.LittleEndian, uint32(dataSize)); err != nil {
		return nil, fmt.Errorf("write data size: %w", err)
	}
	if err := binary.Write(&buf, binary.LittleEndian, samples); err != nil {
		return nil, fmt.Errorf("write PCM data: %w", err)
	}

	return buf.Bytes(), nil
}
