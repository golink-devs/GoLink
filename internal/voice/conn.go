package voice

import (
	"context"
	"log/slog"
	"strings"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/gateway"
	"github.com/disgoorg/disgo/voice"
	"github.com/disgoorg/godave"
	"github.com/disgoorg/godave/golibdave"
	"github.com/disgoorg/snowflake/v2"
)

type VoiceConn struct {
	conn    voice.Conn
	manager voice.Manager
}

func NewVoiceConn(userID snowflake.ID, guildID snowflake.ID) *VoiceConn {
	// Provide a no-op voiceStateUpdateFunc to avoid "shard is not ready" or other gateway errors.
	noOpStateUpdate := func(ctx context.Context, guildID snowflake.ID, channelID *snowflake.ID, selfMute bool, selfDeaf bool) error {
		return nil
	}

	manager := voice.NewManager(noOpStateUpdate,
		userID,
		voice.WithDaveSessionCreateFunc(func(logger *slog.Logger, userId godave.UserID, callbacks godave.Callbacks) godave.Session {
			return golibdave.NewSession(logger, userId, callbacks)
		}),
	)

	conn := manager.CreateConn(guildID)

	return &VoiceConn{
		conn:    conn,
		manager: manager,
	}
}

func (c *VoiceConn) Open(ctx context.Context, userID snowflake.ID, sessionID string, token string, endpoint string) error {
	// Trim port from endpoint if present
	endpoint = strings.TrimSuffix(endpoint, ":80")
	endpoint = strings.TrimSuffix(endpoint, ":443")

	// We use a dummy channel ID because Lavalink doesn't receive it from the bot in player update.
	channelID := snowflake.ID(1)

	c.conn.HandleVoiceStateUpdate(gateway.EventVoiceStateUpdate{
		VoiceState: discord.VoiceState{
			GuildID:   c.conn.GuildID(),
			ChannelID: &channelID,
			UserID:    userID,
			SessionID: sessionID,
		},
	})

	c.conn.HandleVoiceServerUpdate(gateway.EventVoiceServerUpdate{
		GuildID:  c.conn.GuildID(),
		Token:    token,
		Endpoint: &endpoint,
	})

	// Use correct signature for voice.Gateway.Open
	return c.conn.Gateway().Open(ctx, voice.State{
		GuildID:   c.conn.GuildID(),
		UserID:    userID,
		SessionID: sessionID,
		Token:     token,
		Endpoint:  endpoint,
	})
}

func (c *VoiceConn) IsConnected() bool {
	return c.conn.Gateway().Status() == voice.StatusReady
}

func (c *VoiceConn) Close(ctx context.Context) {
	c.conn.Close(ctx)
}

func (c *VoiceConn) SetSpeaking(ctx context.Context, flags voice.SpeakingFlags) error {
	return c.conn.SetSpeaking(ctx, flags)
}

func (c *VoiceConn) SetOpusFrameProvider(provider voice.OpusFrameProvider) {
	c.conn.SetOpusFrameProvider(provider)
}
