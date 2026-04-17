package sources

import (
	"context"
	"fmt"
	"strings"

	"github.com/zmb3/spotify/v2"
	"golang.org/x/oauth2/clientcredentials"
)

type SpotifyResolver struct {
	client  *spotify.Client
	youtube *YouTubeResolver
}

func NewSpotifyResolver(clientID, clientSecret string, youtube *YouTubeResolver) *SpotifyResolver {
	if clientID == "" || clientSecret == "" {
		return &SpotifyResolver{youtube: youtube}
	}

	config := &clientcredentials.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		TokenURL:     "https://accounts.spotify.com/api/token",
	}

	httpClient := config.Client(context.Background())
	client := spotify.New(httpClient)

	return &SpotifyResolver{
		client:  client,
		youtube: youtube,
	}
}

func (r *SpotifyResolver) Name() string { return "spotify" }

func (r *SpotifyResolver) CanResolve(identifier string) bool {
	return strings.HasPrefix(identifier, "spsearch:") ||
		strings.Contains(identifier, "open.spotify.com/")
}

func (r *SpotifyResolver) LoadItem(ctx context.Context, identifier string) (*LoadResult, error) {
	if r.client == nil {
		return &LoadResult{LoadType: LoadTypeError, Data: map[string]interface{}{"message": "Spotify credentials not configured", "severity": "common"}}, nil
	}

	if strings.HasPrefix(identifier, "spsearch:") {
		query := strings.TrimPrefix(identifier, "spsearch:")
		results, err := r.client.Search(ctx, query, spotify.SearchTypeTrack)
		if err != nil {
			return nil, err
		}

		if results.Tracks == nil || len(results.Tracks.Tracks) == 0 {
			return &LoadResult{LoadType: LoadTypeEmpty, Data: map[string]interface{}{}}, nil
		}

		tracks := make([]Track, 0)
		for _, st := range results.Tracks.Tracks {
			tracks = append(tracks, r.mapSpotifyTrackToTrack(st))
		}
		return &LoadResult{
			LoadType: LoadTypeSearch,
			Data:     tracks,
		}, nil
	}

	// Handle Spotify URL
	if strings.Contains(identifier, "track/") {
		parts := strings.Split(identifier, "track/")
		id := strings.Split(parts[1], "?")[0]
		st, err := r.client.GetTrack(ctx, spotify.ID(id))
		if err != nil {
			return nil, err
		}

		return &LoadResult{
			LoadType: LoadTypeTrack,
			Data:     r.mapSpotifyTrackToTrack(*st),
		}, nil
	}

	return &LoadResult{LoadType: LoadTypeEmpty, Data: map[string]interface{}{}}, nil
}

func (r *SpotifyResolver) mapSpotifyTrackToTrack(st spotify.FullTrack) Track {
	artists := make([]string, 0)
	for _, a := range st.Artists {
		artists = append(artists, a.Name)
	}
	author := strings.Join(artists, ", ")

	trackInfo := TrackInfo{
		Identifier: string(st.ID),
		IsSeekable: true,
		Author:     author,
		Length:     int64(st.Duration),
		IsStream:   false,
		Position:   0,
		Title:      st.Name,
		URI:        st.ExternalURLs["spotify"],
		ArtworkURL: "", // Spotify tracks have albums with images
		SourceName: "spotify",
	}
	if len(st.Album.Images) > 0 {
		trackInfo.ArtworkURL = st.Album.Images[0].URL
	}

	return Track{
		Encoded: EncodeTrack(trackInfo),
		Info:    trackInfo,
	}
}

func (r *SpotifyResolver) ResolveToYouTube(ctx context.Context, title, artist string) (*LoadResult, error) {
	query := fmt.Sprintf("ytsearch:%s %s", artist, title)
	return r.youtube.LoadItem(ctx, query)
}
