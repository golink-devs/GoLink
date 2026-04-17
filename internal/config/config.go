package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server  ServerConfig  `yaml:"server"`
	Sources SourcesConfig `yaml:"sources"`
	Cache   CacheConfig   `yaml:"cache"`
	Metrics MetricsConfig `yaml:"metrics"`
	Plugins PluginsConfig `yaml:"plugins"`
	Logging LoggingConfig `yaml:"logging"`
}

type ServerConfig struct {
	Port     int    `yaml:"port"`
	Host     string `yaml:"host"`
	Password string `yaml:"password"`
}

type SourcesConfig struct {
	YouTube             bool   `yaml:"youtube"`
	Spotify             bool   `yaml:"spotify"`
	SpotifyClientID     string `yaml:"spotifyClientID"`
	SpotifyClientSecret string `yaml:"spotifyClientSecret"`
	AppleMusic          bool   `yaml:"applemusic"`
	SoundCloud          bool   `yaml:"soundcloud"`
	Deezer              bool   `yaml:"deezer"`
	HTTP                bool   `yaml:"http"`
}

type CacheConfig struct {
	Enabled bool `yaml:"enabled"`
	TTL     int  `yaml:"ttl"`
}

type MetricsConfig struct {
	Enabled bool `yaml:"enabled"`
}

type PluginsConfig struct {
	Enabled bool   `yaml:"enabled"`
	Path    string `yaml:"path"`
}

type LoggingConfig struct {
	Level string `yaml:"level"`
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
