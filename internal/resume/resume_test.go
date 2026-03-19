package resume

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"cliamp/playlist"
)

func TestSaveLoadQueueRoundTrip(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	in := []playlist.Track{
		{Path: "/tmp/a.mp3", Title: "A", Artist: "AA", Album: "Alb", Genre: "Pop", Year: 2025, TrackNumber: 1},
		{Path: "https://example.com/stream", Title: "Live"},
	}
	SaveQueue(in)
	out := LoadQueue()
	if len(out) != len(in) {
		t.Fatalf("LoadQueue len = %d, want %d", len(out), len(in))
	}
	for i := range in {
		if out[i].Path != in[i].Path || out[i].Title != in[i].Title || out[i].Artist != in[i].Artist {
			t.Fatalf("LoadQueue[%d] = %+v, want %+v", i, out[i], in[i])
		}
	}
}

func TestSaveLoadStateClamp(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	Save(State{CurrentIndex: -5, PositionSec: -9})
	got := Load()
	if got.CurrentIndex != 0 || got.PositionSec != 0 {
		t.Fatalf("Load() = %+v, want index=0 position=0", got)
	}
}

func TestLoadMigratesLegacyEmbeddedTracks(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	dir := filepath.Join(home, ".config", "cliamp")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	legacy := State{
		Path:        "/tmp/second.mp3",
		PositionSec: 42,
		Tracks: []playlist.Track{
			{Path: "/tmp/first.mp3", Title: "First"},
			{Path: "/tmp/second.mp3", Title: "Second"},
		},
	}
	data, err := json.Marshal(legacy)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, resumeStateFileName), data, 0o600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	got := Load()
	if got.CurrentIndex != 1 || got.PositionSec != 42 {
		t.Fatalf("Load() = %+v, want index=1 position=42", got)
	}
	q := LoadQueue()
	if len(q) != 2 || q[1].Path != "/tmp/second.mp3" {
		t.Fatalf("LoadQueue() = %+v, want migrated legacy tracks", q)
	}

	// Legacy fields should be removed from resume.json after migration.
	raw, err := os.ReadFile(filepath.Join(dir, resumeStateFileName))
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	s := string(raw)
	if strings.Contains(s, `"tracks"`) || strings.Contains(s, `"path"`) {
		t.Fatalf("resume.json still contains legacy fields: %s", s)
	}
}

func TestLoadMigrationDoesNotOverwriteExistingQueue(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	dir := filepath.Join(home, ".config", "cliamp")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}

	// Existing queue should win.
	SaveQueue([]playlist.Track{{Path: "/tmp/existing.mp3", Title: "Existing"}})

	legacy := State{
		Path:        "/tmp/legacy-second.mp3",
		PositionSec: 50,
		Tracks: []playlist.Track{
			{Path: "/tmp/legacy-first.mp3", Title: "Legacy First"},
			{Path: "/tmp/legacy-second.mp3", Title: "Legacy Second"},
		},
	}
	data, err := json.Marshal(legacy)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, resumeStateFileName), data, 0o600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	got := Load()
	if got.CurrentIndex != 1 || got.PositionSec != 50 {
		t.Fatalf("Load() = %+v, want index=1 position=50", got)
	}

	q := LoadQueue()
	if len(q) != 1 || q[0].Path != "/tmp/existing.mp3" {
		t.Fatalf("LoadQueue() = %+v, want existing queue preserved", q)
	}
}
