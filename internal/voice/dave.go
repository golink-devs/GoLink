package voice

import (
	"log/slog"

	"github.com/disgoorg/godave"
	"github.com/disgoorg/godave/golibdave"
	"github.com/disgoorg/godave/libdave"
)

func init() {
	// Set libdave log level globally (optional)
	libdave.SetDefaultLogLoggerLevel(slog.LevelError)
}

// NewDAVESession creates a new DAVE session for a guild voice connection.
func NewDAVESession(logger *slog.Logger) (godave.Session, error) {
	return golibdave.NewSession(logger)
}
