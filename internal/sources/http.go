package sources

import (
	"context"
	"strings"
)

type HTTPResolver struct{}

func NewHTTPResolver() *HTTPResolver {
	return &HTTPResolver{}
}

func (r *HTTPResolver) Name() string { return "http" }

func (r *HTTPResolver) CanResolve(identifier string) bool {
	return strings.HasPrefix(identifier, "http://") || strings.HasPrefix(identifier, "https://")
}

func (r *HTTPResolver) LoadItem(ctx context.Context, identifier string) (*LoadResult, error) {
	// For HTTP, we just return the URL as the title and identifier
	trackInfo := TrackInfo{
		Identifier: identifier,
		IsSeekable: true,
		Author:     "Unknown",
		Length:     0,
		IsStream:   true,
		Position:   0,
		Title:      identifier,
		URI:        identifier,
		SourceName: "http",
	}
	track := Track{
		Encoded: EncodeTrack(trackInfo),
		Info:    trackInfo,
	}
	return &LoadResult{
		LoadType: LoadTypeTrack,
		Data:     track,
	}, nil
}
