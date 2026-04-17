package player

import (
	"errors"
	"math/rand"
	"sync"

	"github.com/golink-devs/golink/internal/sources"
)

type Queue struct {
	tracks []sources.Track
	mu     sync.RWMutex
}

func NewQueue() *Queue {
	return &Queue{
		tracks: make([]sources.Track, 0),
	}
}

func (q *Queue) Add(track sources.Track) {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.tracks = append(q.tracks, track)
}

func (q *Queue) Remove(index int) error {
	q.mu.Lock()
	defer q.mu.Unlock()
	if index < 0 || index >= len(q.tracks) {
		return errors.New("index out of bounds")
	}
	q.tracks = append(q.tracks[:index], q.tracks[index+1:]...)
	return nil
}

func (q *Queue) Move(from, to int) error {
	q.mu.Lock()
	defer q.mu.Unlock()
	if from < 0 || from >= len(q.tracks) || to < 0 || to >= len(q.tracks) {
		return errors.New("index out of bounds")
	}
	track := q.tracks[from]
	q.tracks = append(q.tracks[:from], q.tracks[from+1:]...)
	// Insert at 'to'
	q.tracks = append(q.tracks[:to], append([]sources.Track{track}, q.tracks[to:]...)...)
	return nil
}

func (q *Queue) Shuffle() {
	q.mu.Lock()
	defer q.mu.Unlock()
	rand.Shuffle(len(q.tracks), func(i, j int) {
		q.tracks[i], q.tracks[j] = q.tracks[j], q.tracks[i]
	})
}

func (q *Queue) Peek() *sources.Track {
	q.mu.RLock()
	defer q.mu.RUnlock()
	if len(q.tracks) == 0 {
		return nil
	}
	return &q.tracks[0]
}

func (q *Queue) Poll() *sources.Track {
	q.mu.Lock()
	defer q.mu.Unlock()
	if len(q.tracks) == 0 {
		return nil
	}
	track := q.tracks[0]
	q.tracks = q.tracks[1:]
	return &track
}

func (q *Queue) Len() int {
	q.mu.RLock()
	defer q.mu.RUnlock()
	return len(q.tracks)
}

func (q *Queue) Clear() {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.tracks = make([]sources.Track, 0)
}

func (q *Queue) Tracks() []sources.Track {
	q.mu.RLock()
	defer q.mu.RUnlock()
	tracks := make([]sources.Track, len(q.tracks))
	copy(tracks, q.tracks)
	return tracks
}
