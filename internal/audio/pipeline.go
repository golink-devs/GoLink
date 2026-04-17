package audio

import (
	"encoding/binary"
	"io"

	"github.com/disgoorg/disgo/voice"
	"layeh.com/gopus"
)

const (
	SampleRate     = 48000
	Channels       = 2
	FrameSize      = voice.OpusFrameSize      // 960 samples
	FrameSizeBytes = voice.OpusFrameSizeBytes // 3840 bytes of PCM
	MaxOpusSize    = voice.MaxOpusFrameSize    // 1400 bytes max
)

// Pipeline reads PCM from FFmpeg, encodes to Opus, and provides frames.
// It implements voice.OpusFrameProvider.
type Pipeline struct {
	ffmpeg  *FFmpegProcess
	encoder *gopus.Encoder
	buf     []byte
	stopped chan struct{}
}

func NewPipeline(ffmpeg *FFmpegProcess) (*Pipeline, error) {
	encoder, err := gopus.NewEncoder(SampleRate, Channels, gopus.Audio)
	if err != nil {
		return nil, err
	}
	encoder.SetBitrate(128000) // 128kbps

	return &Pipeline{
		ffmpeg:  ffmpeg,
		encoder: encoder,
		buf:     make([]byte, FrameSizeBytes),
		stopped: make(chan struct{}),
	}, nil
}

// ProvideOpusFrame implements voice.OpusFrameProvider
// Called every 20ms by disgoorg/disgo's AudioSender
func (p *Pipeline) ProvideOpusFrame() ([]byte, error) {
	select {
	case <-p.stopped:
		return voice.SilenceAudioFrame, io.EOF
	default:
	}

	// Read exactly one PCM frame (3840 bytes = 960 stereo samples)
	_, err := io.ReadFull(p.ffmpeg, p.buf)
	if err != nil {
		if err == io.EOF || err == io.ErrUnexpectedEOF {
			return voice.SilenceAudioFrame, io.EOF
		}
		return voice.SilenceAudioFrame, err
	}

	// Convert []byte to []int16
	pcm := make([]int16, FrameSize*Channels)
	for i := range pcm {
		pcm[i] = int16(binary.LittleEndian.Uint16(p.buf[i*2:]))
	}

	// Encode PCM to Opus
	opusData, err := p.encoder.Encode(pcm, FrameSize, MaxOpusSize)
	if err != nil {
		return voice.SilenceAudioFrame, err
	}

	return opusData, nil
}

func (p *Pipeline) Close() {
	select {
	case <-p.stopped:
		return
	default:
		close(p.stopped)
	}
	p.ffmpeg.Stop()
}
