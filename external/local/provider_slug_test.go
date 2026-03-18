package local

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSlugifyPlaylistName(t *testing.T) {
	t.Parallel()

	tests := []struct {
		in   string
		want string
	}{
		{"Ayla - Cover Songs", "ayla-cover-songs"},
		{"  But It hits different  ", "but-it-hits-different"},
		{"Nightcore!!!", "nightcore"},
		{"___", "playlist"},
	}

	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			if got := slugifyPlaylistName(tt.in); got != tt.want {
				t.Fatalf("slugifyPlaylistName(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestSafePathCreatesSlugFileForNewPlaylist(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	p := &Provider{dir: dir}

	path, err := p.safePath("Ayla - Cover Songs")
	if err != nil {
		t.Fatalf("safePath: %v", err)
	}
	want := filepath.Join(dir, "ayla-cover-songs.toml")
	if path != want {
		t.Fatalf("safePath = %q, want %q", path, want)
	}
}

func TestSafePathPrefersExistingLegacyFile(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	p := &Provider{dir: dir}
	legacy := filepath.Join(dir, "Ayla - Cover Songs.toml")
	if err := os.WriteFile(legacy, []byte("[[track]]\npath=\"x\"\ntitle=\"x\"\n"), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	path, err := p.safePath("Ayla - Cover Songs")
	if err != nil {
		t.Fatalf("safePath: %v", err)
	}
	if path != legacy {
		t.Fatalf("safePath = %q, want legacy %q", path, legacy)
	}
}
