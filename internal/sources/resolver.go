package sources

import (
	"context"
	"encoding/base64"
	"encoding/json"
)

// LoadType values — exact strings Lavalink v4 uses
const (
	LoadTypeTrack    = "track"
	LoadTypePlaylist = "playlist"
	LoadTypeSearch   = "search"
	LoadTypeEmpty    = "empty"
	LoadTypeError    = "error"
)

type Resolver interface {
	Name() string
	CanResolve(identifier string) bool
	LoadItem(ctx context.Context, identifier string) (*LoadResult, error)
}

type LoadResult struct {
	LoadType string      `json:"loadType"`
	Data     interface{} `json:"data"`
}

type Track struct {
	Encoded string    `json:"encoded"` // base64 encoded track info
	Info    TrackInfo `json:"info"`
}

type TrackInfo struct {
	Identifier string `json:"identifier"`
	IsSeekable bool   `json:"isSeekable"`
	Author     string `json:"author"`
	Length     int64  `json:"length"` // milliseconds
	IsStream   bool   `json:"isStream"`
	Position   int64  `json:"position"`
	Title      string `json:"title"`
	URI        string `json:"uri"`
	ArtworkURL string `json:"artworkUrl"`
	ISRC       string `json:"isrc"`
	SourceName string `json:"sourceName"`
}

type Playlist struct {
	Info   PlaylistInfo `json:"info"`
	Tracks []Track      `json:"tracks"`
}

type PlaylistInfo struct {
	Name          string `json:"name"`
	SelectedTrack int    `json:"selectedTrack"`
}

// Registry holds all registered resolvers
type Registry struct {
	resolvers []Resolver
	cache     *TrackCache
}

func NewRegistry(cache *TrackCache) *Registry {
	return &Registry{
		resolvers: make([]Resolver, 0),
		cache:     cache,
	}
}

func (r *Registry) Register(resolver Resolver) {
	r.resolvers = append(r.resolvers, resolver)
}

func (r *Registry) Resolve(ctx context.Context, identifier string) (*LoadResult, error) {
	if r.cache != nil {
		if result, ok := r.cache.Get(identifier); ok {
			return result, nil
		}
	}

	for _, resolver := range r.resolvers {
		if resolver.CanResolve(identifier) {
			result, err := resolver.LoadItem(ctx, identifier)
			if err == nil && r.cache != nil && result.LoadType != LoadTypeEmpty && result.LoadType != LoadTypeError {
				r.cache.Set(identifier, result)
			}
			return result, err
		}
	}
	return &LoadResult{LoadType: LoadTypeEmpty, Data: map[string]interface{}{}}, nil
}

// EncodeTrack encodes track info to base64.
// Lavalink uses a custom binary format, but for GoLink we can use JSON base64-encoded for simplicity,
// though a real drop-in replacement should match Lavaplayer's binary format if possible.
// For now, let's use JSON base64.
func EncodeTrack(info TrackInfo) string {
	data, _ := json.Marshal(info)
	return base64.StdEncoding.EncodeToString(data)
}

// DecodeTrack decodes base64 track info.
func DecodeTrack(encoded string) (*TrackInfo, error) {
	data, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return nil, err
	}
	var info TrackInfo
	if err := json.Unmarshal(data, &info); err != nil {
		return nil, err
	}
	return &info, nil
}
