package audio

import (
	"encoding/binary"
	"fmt"
	"unsafe"

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

// DecodeInto decodes a single Opus frame into the provided buffer, avoiding allocation.
// The buffer must be at least FrameSize*Channels in length.
// Returns the number of decoded samples (per channel).
func (d *OpusDecoder) DecodeInto(opusData []byte, pcm []int16) (int, error) {
	n, err := d.decoder.Decode(opusData, pcm)
	if err != nil {
		return 0, fmt.Errorf("decode opus frame: %w", err)
	}
	return n * Channels, nil
}

// PCMToWAV converts raw PCM int16 samples to a WAV byte slice.
// This uses a single pre-allocated buffer and direct byte copy for the sample
// data, avoiding the reflection overhead of binary.Write on large slices.
func PCMToWAV(samples []int16, sampleRate, channels int) ([]byte, error) {
	dataSize := len(samples) * 2 // 2 bytes per int16
	headerSize := 44             // standard WAV header
	buf := make([]byte, headerSize+dataSize)

	// RIFF header
	copy(buf[0:4], "RIFF")
	binary.LittleEndian.PutUint32(buf[4:8], uint32(36+dataSize))
	copy(buf[8:12], "WAVE")

	// fmt subchunk
	copy(buf[12:16], "fmt ")
	binary.LittleEndian.PutUint32(buf[16:20], 16)       // subchunk size
	binary.LittleEndian.PutUint16(buf[20:22], 1)        // PCM format
	binary.LittleEndian.PutUint16(buf[22:24], uint16(channels))
	binary.LittleEndian.PutUint32(buf[24:28], uint32(sampleRate))
	binary.LittleEndian.PutUint32(buf[28:32], uint32(sampleRate*channels*2)) // byte rate
	binary.LittleEndian.PutUint16(buf[32:34], uint16(channels*2))           // block align
	binary.LittleEndian.PutUint16(buf[34:36], 16)                           // bits per sample

	// data subchunk
	copy(buf[36:40], "data")
	binary.LittleEndian.PutUint32(buf[40:44], uint32(dataSize))

	// Copy PCM data directly — reinterpret []int16 as []byte to avoid
	// per-sample binary.Write overhead.
	pcmBytes := unsafe.Slice((*byte)(unsafe.Pointer(unsafe.SliceData(samples))), dataSize)
	copy(buf[headerSize:], pcmBytes)

	return buf, nil
}
