package voice

import (
	"log/slog"

	"github.com/disgoorg/godave"
	"github.com/disgoorg/godave/golibdave"
	"github.com/disgoorg/godave/libdave"
	"github.com/disgoorg/snowflake/v2"
)

func init() {
	// Set libdave log level globally (optional)
	libdave.SetDefaultLogLoggerLevel(slog.LevelError)
}

// NewDAVESession creates a new DAVE session for a guild voice connection.
func NewDAVESession(logger *slog.Logger, userID snowflake.ID) (godave.Session, error) {
	// golibdave.NewSession matches the signature expected by godave
	session := golibdave.NewSession(logger, godave.UserID(userID.String()), nil)
	return session, nil
}
