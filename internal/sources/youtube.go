package sources

import (
	"context"
	"encoding/json"
	"os/exec"
	"strings"
)

type YouTubeResolver struct{}

func NewYouTubeResolver() *YouTubeResolver {
	return &YouTubeResolver{}
}

func (r *YouTubeResolver) Name() string { return "youtube" }

func (r *YouTubeResolver) CanResolve(identifier string) bool {
	return strings.HasPrefix(identifier, "ytsearch:") ||
		strings.HasPrefix(identifier, "ytmsearch:") ||
		strings.Contains(identifier, "youtube.com/") ||
		strings.Contains(identifier, "youtu.be/")
}

type ytdlpOutput struct {
	ID         string  `json:"id"`
	Title      string  `json:"title"`
	Uploader   string  `json:"uploader"`
	Duration   float64 `json:"duration"`
	URL        string  `json:"url"` // direct stream URL
	WebpageURL string  `json:"webpage_url"`
	Thumbnail  string  `json:"thumbnail"`
	IsLive     bool    `json:"is_live"`
}

func (r *YouTubeResolver) LoadItem(ctx context.Context, identifier string) (*LoadResult, error) {
	args := []string{
		"--dump-json",
		"--no-playlist",
		"-f", "bestaudio[ext=webm]/bestaudio/best",
		"--no-warnings",
	}

	// Handle search prefixes
	if strings.HasPrefix(identifier, "ytsearch:") {
		query := strings.TrimPrefix(identifier, "ytsearch:")
		args = append(args, "ytsearch1:"+query) // ytsearch1 to get only one result
	} else if strings.HasPrefix(identifier, "ytmsearch:") {
		query := strings.TrimPrefix(identifier, "ytmsearch:")
		args = append(args, "https://music.youtube.com/search?q="+query)
	} else {
		args = append(args, identifier)
	}

	cmd := exec.CommandContext(ctx, "yt-dlp", args...)
	out, err := cmd.Output()
	if err != nil {
		// If it's a search, it might be empty
		if strings.Contains(identifier, "search:") {
			return &LoadResult{LoadType: LoadTypeEmpty, Data: map[string]interface{}{}}, nil
		}
		return &LoadResult{LoadType: LoadTypeError, Data: map[string]interface{}{"message": err.Error(), "severity": "common"}}, err
	}

	// yt-dlp might return multiple JSON objects if it's a search
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	if len(lines) == 0 || lines[0] == "" {
		return &LoadResult{LoadType: LoadTypeEmpty, Data: map[string]interface{}{}}, nil
	}

	if strings.Contains(identifier, "search:") {
		tracks := make([]Track, 0)
		for _, line := range lines {
			var info ytdlpOutput
			if err := json.Unmarshal([]byte(line), &info); err != nil {
				continue
			}
			track := r.mapToTrack(info)
			tracks = append(tracks, track)
		}
		return &LoadResult{
			LoadType: LoadTypeSearch,
			Data:     tracks,
		}, nil
	}

	var info ytdlpOutput
	if err := json.Unmarshal([]byte(lines[0]), &info); err != nil {
		return nil, err
	}

	track := r.mapToTrack(info)

	return &LoadResult{
		LoadType: LoadTypeTrack,
		Data:     track,
	}, nil
}

func (r *YouTubeResolver) mapToTrack(info ytdlpOutput) Track {
	trackInfo := TrackInfo{
		Identifier: info.ID,
		IsSeekable: !info.IsLive,
		Author:     info.Uploader,
		Length:     int64(info.Duration * 1000),
		IsStream:   info.IsLive,
		Position:   0,
		Title:      info.Title,
		URI:        info.WebpageURL,
		ArtworkURL: info.Thumbnail,
		SourceName: "youtube",
	}
	return Track{
		Encoded: EncodeTrack(trackInfo),
		Info:    trackInfo,
	}
}

// GetStreamURL returns the direct stream URL for a track identifier.
func (r *YouTubeResolver) GetStreamURL(ctx context.Context, identifier string) (string, error) {
	args := []string{
		"-g", // get URL
		"-f", "bestaudio[ext=webm]/bestaudio/best",
		identifier,
	}
	cmd := exec.CommandContext(ctx, "yt-dlp", args...)
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}
