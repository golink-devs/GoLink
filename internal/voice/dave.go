package voice

import (
	"log/slog"

	"github.com/disgoorg/godave"
	"github.com/thomas-vilte/dave-go/session"
	"github.com/disgoorg/snowflake/v2"
)

// NewDAVESession creates a new DAVE session for a guild voice connection.
func NewDAVESession(logger *slog.Logger, userID snowflake.ID) (godave.Session, error) {
	s := session.New(logger, godave.UserID(userID.String()), nil)
	return s, nil
}

