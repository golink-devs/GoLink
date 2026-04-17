package voice

import (
	"context"

	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/voice"
	"github.com/disgoorg/godave/golibdave"
	"github.com/disgoorg/snowflake/v2"
)

type VoiceConn struct {
	conn    voice.Conn
	manager voice.Manager
}

func NewVoiceConn(userID snowflake.ID, guildID snowflake.ID) *VoiceConn {
	// voiceStateUpdateFunc is not needed as GoLink doesn't have a bot client,
	// it just receives state from the user's bot.
	manager := voice.NewManager(nil,
		voice.WithDaveSessionCreateFunc(golibdave.NewSession),
	)

	conn := manager.CreateConn(guildID)

	return &VoiceConn{
		conn:    conn,
		manager: manager,
	}
}

func (c *VoiceConn) Open(ctx context.Context, channelID snowflake.ID, userID snowflake.ID, sessionID string, token string, endpoint string) error {
	// We need to provide the voice server update to the connection.
	c.conn.HandleVoiceStateUpdate(bot.EventVoiceStateUpdate{
		VoiceState: discord.VoiceState{
			GuildID:   c.conn.GuildID(),
			ChannelID: &channelID,
			UserID:    userID,
			SessionID: sessionID,
		},
	})

	c.conn.HandleVoiceServerUpdate(bot.EventVoiceServerUpdate{
		VoiceServerUpdate: discord.VoiceServerUpdate{
			GuildID:  c.conn.GuildID(),
			Token:    token,
			Endpoint: endpoint,
		},
	})

	return c.conn.Open(ctx, channelID, false, false)
}

func (c *VoiceConn) IsConnected() bool {
	// Check if gateway is open
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
