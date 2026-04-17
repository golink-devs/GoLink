package player

import (
	"context"
	"sync"
	"time"

	"github.com/golink-devs/golink/internal/audio"
	"github.com/golink-devs/golink/internal/hub"
	"github.com/golink-devs/golink/internal/sources"
	"github.com/golink-devs/golink/internal/voice"
)

type LoopMode int

const (
	LoopNone  LoopMode = 0
	LoopTrack LoopMode = 1
	LoopQueue LoopMode = 2
)

type Player struct {
	GuildID   string
	SessionID string
	UserID    string

	Track *sources.Track
	Paused   bool
	Volume   int // 0-1000, default 100
	Position int64
	Ping     int

	Filters  Filters
	Queue    *Queue
	LoopMode LoopMode

	Pipeline  *audio.Pipeline
	VoiceConn *voice.VoiceConn
	Hub       *hub.Hub

	mu sync.RWMutex
}

func NewPlayer(sessionID, userID, guildID string, h *hub.Hub) *Player {
	return &Player{
		GuildID:   guildID,
		SessionID: sessionID,
		UserID:    userID,
		Volume:    100,
		Queue:     NewQueue(),
		Hub:       h,
		mu:        sync.RWMutex{},
	}
}

type PlayerState struct {
	Time      int64 `json:"time"`
	Position  int64 `json:"position"`
	Connected bool  `json:"connected"`
	Ping      int   `json:"ping"`
}

func (p *Player) State() PlayerState {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return PlayerState{
		Time:      time.Now().UnixMilli(),
		Position:  p.Position,
		Connected: p.VoiceConn != nil && p.VoiceConn.IsConnected(),
		Ping:      p.Ping,
	}
}

func (p *Player) Play(ctx context.Context, track sources.Track, streamURL string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.Pipeline != nil {
		p.Pipeline.Close()
	}

	ffmpeg, err := audio.NewFFmpegProcess(ctx, streamURL, p.Filters.BuildFilterChain())
	if err != nil {
		return err
	}

	pipeline, err := audio.NewPipeline(ffmpeg)
	if err != nil {
		ffmpeg.Stop()
		return err
	}

	p.Track = &track
	p.Paused = false
	p.Position = 0
	p.Pipeline = pipeline

	if p.VoiceConn != nil {
		p.VoiceConn.SetOpusFrameProvider(pipeline)
	}

	p.Hub.Send(p.SessionID, map[string]interface{}{
		"op":      "event",
		"type":    "TrackStartEvent",
		"guildId": p.GuildID,
		"track":   track,
	})

	return nil
}

func (p *Player) Stop(ctx context.Context) error {
	p.mu.Lock()
	track := p.Track
	p.Track = nil
	if p.Pipeline != nil {
		p.Pipeline.Close()
		p.Pipeline = nil
	}
	p.mu.Unlock()

	if track != nil {
		p.Hub.Send(p.SessionID, map[string]interface{}{
			"op":      "event",
			"type":    "TrackEndEvent",
			"guildId": p.GuildID,
			"track":   track,
			"reason":  "stopped",
		})
	}
	return nil
}

func (p *Player) Pause(paused bool) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.Paused = paused
	return nil
}

func (p *Player) Seek(positionMs int64) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.Position = positionMs
	// TODO: Restart pipeline with seek
	return nil
}

func (p *Player) SetVolume(volume int) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.Volume = volume
	return nil
}

func (p *Player) SetFilters(filters Filters) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.Filters = filters
	return nil
}

func (p *Player) Destroy(ctx context.Context) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.Pipeline != nil {
		p.Pipeline.Close()
	}
	if p.VoiceConn != nil {
		p.VoiceConn.Close(ctx)
	}
	return nil
}
