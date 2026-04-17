package voice

import (
	"context"
	"log/slog"

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
	manager := voice.NewManager(nil,
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

func (c *VoiceConn) Open(ctx context.Context, channelID snowflake.ID, userID snowflake.ID, sessionID string, token string, endpoint string) error {
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

	return c.conn.Open(ctx, channelID, false, false)
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
