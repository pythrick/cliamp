// Package resume persists session playlist + playback metadata so playback
// can be restored on the next launch.
package resume

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"cliamp/internal/appdir"
	"cliamp/internal/tomlutil"
	"cliamp/playlist"
)

const (
	resumeStateFileName  = "resume.json"
	sessionQueueFileName = "session_queue.toml"
)

// State stores lightweight playback metadata.
type State struct {
	CurrentIndex int `json:"current_index"`
	PositionSec  int `json:"position_sec"`

	// Legacy fields kept for migration from older builds.
	Path   string           `json:"path,omitempty"`
	Tracks []playlist.Track `json:"tracks,omitempty"`
}

func stateFile() (string, error) {
	dir, err := appdir.Dir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, resumeStateFileName), nil
}

func queueFilePath() (string, error) {
	dir, err := appdir.Dir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, sessionQueueFileName), nil
}

// Save writes lightweight playback metadata.
func Save(s State) {
	f, err := stateFile()
	if err != nil {
		return
	}
	if s.CurrentIndex < 0 {
		s.CurrentIndex = 0
	}
	if s.PositionSec < 0 {
		s.PositionSec = 0
	}
	data, err := json.Marshal(s)
	if err != nil {
		return
	}
	_ = os.MkdirAll(filepath.Dir(f), 0o755)
	_ = os.WriteFile(f, data, 0o600)
}

// SaveQueue writes the current in-memory session queue to disk.
func SaveQueue(tracks []playlist.Track) {
	f, err := queueFilePath()
	if err != nil {
		return
	}
	if len(tracks) == 0 {
		_ = os.Remove(f)
		return
	}
	_ = os.MkdirAll(filepath.Dir(f), 0o755)
	out, err := os.Create(f)
	if err != nil {
		return
	}
	defer out.Close()
	for i, t := range tracks {
		if i > 0 {
			_, _ = fmt.Fprintln(out)
		}
		_, _ = fmt.Fprintln(out, "[[track]]")
		_, _ = fmt.Fprintf(out, "path = %q\n", t.Path)
		_, _ = fmt.Fprintf(out, "title = %q\n", t.Title)
		if t.Artist != "" {
			_, _ = fmt.Fprintf(out, "artist = %q\n", t.Artist)
		}
		if t.Album != "" {
			_, _ = fmt.Fprintf(out, "album = %q\n", t.Album)
		}
		if t.Genre != "" {
			_, _ = fmt.Fprintf(out, "genre = %q\n", t.Genre)
		}
		if t.Year != 0 {
			_, _ = fmt.Fprintf(out, "year = %d\n", t.Year)
		}
		if t.TrackNumber != 0 {
			_, _ = fmt.Fprintf(out, "track_number = %d\n", t.TrackNumber)
		}
	}
}

// SaveSession writes both queue playlist + metadata in one call.
func SaveSession(tracks []playlist.Track, s State) {
	SaveQueue(tracks)
	Save(s)
}

// Load reads lightweight playback metadata.
func Load() State {
	f, err := stateFile()
	if err != nil {
		return State{}
	}
	data, err := os.ReadFile(f)
	if err != nil {
		return State{}
	}
	var s State
	if err := json.Unmarshal(data, &s); err != nil {
		return State{}
	}
	if s.CurrentIndex < 0 {
		s.CurrentIndex = 0
	}
	if s.PositionSec < 0 {
		s.PositionSec = 0
	}

	// Migrate legacy embedded track list into the session queue file once.
	if len(s.Tracks) > 0 {
		if tracks := LoadQueue(); len(tracks) == 0 {
			SaveQueue(s.Tracks)
		}
		// If old file had path+position but no index, derive best effort index.
		if s.Path != "" && s.CurrentIndex == 0 {
			for i, t := range s.Tracks {
				if t.Path == s.Path {
					s.CurrentIndex = i
					break
				}
			}
		}
		s.Tracks = nil
		s.Path = ""
		Save(s)
	}
	return s
}

// LoadQueue reads the current session queue from disk.
func LoadQueue() []playlist.Track {
	f, err := queueFilePath()
	if err != nil {
		return nil
	}
	data, err := os.ReadFile(f)
	if err != nil {
		return nil
	}
	return parseQueueTOML(data)
}

func parseQueueTOML(data []byte) []playlist.Track {
	var tracks []playlist.Track
	var current *playlist.Track
	for _, rawLine := range strings.Split(string(data), "\n") {
		line := strings.TrimSpace(rawLine)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if line == "[[track]]" {
			if current != nil {
				tracks = append(tracks, *current)
			}
			current = &playlist.Track{}
			continue
		}
		if current == nil {
			continue
		}
		key, val, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		key = strings.TrimSpace(key)
		val = tomlutil.Unquote(strings.TrimSpace(val))
		switch key {
		case "path":
			current.Path = val
			current.Stream = playlist.IsURL(val)
		case "title":
			current.Title = val
		case "artist":
			current.Artist = val
		case "album":
			current.Album = val
		case "genre":
			current.Genre = val
		case "year":
			if n, err := strconv.Atoi(val); err == nil {
				current.Year = n
			}
		case "track_number":
			if n, err := strconv.Atoi(val); err == nil {
				current.TrackNumber = n
			}
		}
	}
	if current != nil {
		tracks = append(tracks, *current)
	}
	return tracks
}
