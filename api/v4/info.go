package v4

import (
	"time"

	"github.com/gofiber/fiber/v2"
)

type InfoResponse struct {
	Version        VersionInfo  `json:"version"`
	BuildTime      int64        `json:"buildTime"`
	Git            GitInfo      `json:"git"`
	JVM            string       `json:"jvm"`
	Lavalayer      string       `json:"lavaplayer"`
	SourceManagers []string     `json:"sourceManagers"`
	Filters        []string     `json:"filters"`
	Plugins        []PluginInfo `json:"plugins"`
}

type VersionInfo struct {
	Semver     string      `json:"semver"`
	Major      int         `json:"major"`
	Minor      int         `json:"minor"`
	Patch      int         `json:"patch"`
	PreRelease interface{} `json:"preRelease"`
	Build      interface{} `json:"build"`
}

type GitInfo struct {
	Branch     string `json:"branch"`
	Commit     string `json:"commit"`
	CommitTime int64  `json:"commitTime"`
}

type PluginInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

func GetInfo(c *fiber.Ctx) error {
	return c.JSON(InfoResponse{
		Version: VersionInfo{
			Semver: "1.0.0",
			Major:  1,
			Minor:  0,
			Patch:  0,
		},
		BuildTime: time.Now().UnixMilli(),
		Git: GitInfo{
			Branch: "main",
			Commit: "abc1234",
		},
		JVM:            "N/A",
		Lavalayer:      "N/A",
		SourceManagers: []string{"youtube", "spotify", "applemusic", "soundcloud", "deezer", "http"},
		Filters:        []string{"volume", "equalizer", "karaoke", "timescale", "tremolo", "vibrato", "rotation", "distortion", "channelMix", "lowPass"},
		Plugins:        []PluginInfo{},
	})
}

func GetVersion(c *fiber.Ctx) error {
	return c.SendString("1.0.0")
}

func GetStats(c *fiber.Ctx) error {
	// Dummy stats for now
	return c.JSON(fiber.Map{
		"players":        0,
		"playingPlayers": 0,
		"uptime":         0,
		"memory": fiber.Map{
			"free":       0,
			"used":       0,
			"allocated":  0,
			"reservable": 0,
		},
		"cpu": fiber.Map{
			"cores":        4,
			"systemLoad":   0.0,
			"lavalinkLoad": 0.0,
		},
		"frameStats": nil,
	})
}
