package config

import "testing"

func TestParseFlagsNoTelemetry(t *testing.T) {
	action, ov, positional, err := ParseFlags([]string{"--no-telemetry", "track.mp3"})
	if err != nil {
		t.Fatalf("ParseFlags: %v", err)
	}
	if action != "" {
		t.Fatalf("action = %q, want empty", action)
	}
	if ov.Telemetry == nil {
		t.Fatalf("ov.Telemetry = nil, want false pointer")
	}
	if *ov.Telemetry {
		t.Fatalf("ov.Telemetry = true, want false")
	}
	if len(positional) != 1 || positional[0] != "track.mp3" {
		t.Fatalf("positional = %v, want [track.mp3]", positional)
	}
}

func TestOverridesApplyTelemetry(t *testing.T) {
	cfg := Default()
	ov := Overrides{Telemetry: ptrBool(false)}
	ov.Apply(&cfg)
	if cfg.Telemetry {
		t.Fatalf("cfg.Telemetry = true, want false")
	}
}
