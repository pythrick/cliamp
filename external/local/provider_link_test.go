package local

import (
	"testing"

	"cliamp/playlist"
)

func TestSaveLinkedPlaylistPersistsSourceURL(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	p := &Provider{dir: dir}
	tracks := []playlist.Track{
		{Path: "https://example.com/stream", Title: "Example", Stream: true},
	}

	if err := p.SaveLinkedPlaylist("linked", "https://example.com/list.m3u", tracks); err != nil {
		t.Fatalf("SaveLinkedPlaylist: %v", err)
	}

	gotURL, err := p.SourceURL("linked")
	if err != nil {
		t.Fatalf("SourceURL: %v", err)
	}
	if gotURL != "https://example.com/list.m3u" {
		t.Fatalf("SourceURL = %q, want %q", gotURL, "https://example.com/list.m3u")
	}

	gotTracks, err := p.Tracks("linked")
	if err != nil {
		t.Fatalf("Tracks: %v", err)
	}
	if len(gotTracks) != 1 || gotTracks[0].Title != "Example" {
		t.Fatalf("Tracks = %#v, want one Example track", gotTracks)
	}
}

func TestSavePlaylistPreservesExistingSourceURL(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	p := &Provider{dir: dir}

	if err := p.SaveLinkedPlaylist("linked", "https://example.com/feed.xml", []playlist.Track{
		{Path: "https://example.com/old", Title: "Old", Stream: true},
	}); err != nil {
		t.Fatalf("SaveLinkedPlaylist: %v", err)
	}

	if err := p.SavePlaylist("linked", []playlist.Track{
		{Path: "https://example.com/new", Title: "New", Stream: true},
	}); err != nil {
		t.Fatalf("SavePlaylist: %v", err)
	}

	gotURL, err := p.SourceURL("linked")
	if err != nil {
		t.Fatalf("SourceURL: %v", err)
	}
	if gotURL != "https://example.com/feed.xml" {
		t.Fatalf("SourceURL after SavePlaylist = %q, want %q", gotURL, "https://example.com/feed.xml")
	}
}
