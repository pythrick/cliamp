package ui

import "testing"

func TestParseURLPlaylistImportInput(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantName string
		wantURL  string
		wantOK   bool
	}{
		{
			name:     "valid input",
			input:    "My Playlist | https://example.com/list.m3u",
			wantName: "My Playlist",
			wantURL:  "https://example.com/list.m3u",
			wantOK:   true,
		},
		{
			name:   "missing separator",
			input:  "My Playlist https://example.com/list.m3u",
			wantOK: false,
		},
		{
			name:   "empty name",
			input:  "   | https://example.com/list.m3u",
			wantOK: false,
		},
		{
			name:   "empty url",
			input:  "My Playlist |   ",
			wantOK: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotName, gotURL, gotOK := parseURLPlaylistImportInput(tt.input)
			if gotOK != tt.wantOK {
				t.Fatalf("ok = %v, want %v", gotOK, tt.wantOK)
			}
			if gotName != tt.wantName {
				t.Fatalf("name = %q, want %q", gotName, tt.wantName)
			}
			if gotURL != tt.wantURL {
				t.Fatalf("url = %q, want %q", gotURL, tt.wantURL)
			}
		})
	}
}
