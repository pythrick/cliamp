//go:build unix

package instance

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"testing"
	"time"
)

func TestLockHelperProcess(t *testing.T) {
	if os.Getenv("CLIAMP_LOCK_HELPER") != "1" {
		return
	}
	l, err := Acquire(false)
	if err != nil {
		fmt.Printf("acquire error: %v\n", err)
		os.Exit(2)
	}
	defer l.Close()
	fmt.Println("ready")
	for {
		time.Sleep(5 * time.Second)
	}
}

func TestAcquireAndTakeover(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	cmd := exec.Command(os.Args[0], "-test.run=TestLockHelperProcess")
	cmd.Env = append(os.Environ(),
		"CLIAMP_LOCK_HELPER=1",
		"HOME="+home,
	)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		t.Fatalf("StdoutPipe: %v", err)
	}
	if err := cmd.Start(); err != nil {
		t.Fatalf("Start helper: %v", err)
	}
	defer func() {
		_ = cmd.Process.Kill()
		_, _ = cmd.Process.Wait()
	}()

	sc := bufio.NewScanner(stdout)
	deadline := time.Now().Add(5 * time.Second)
	ready := false
	for time.Now().Before(deadline) {
		if sc.Scan() && sc.Text() == "ready" {
			ready = true
			break
		}
		time.Sleep(20 * time.Millisecond)
	}
	if !ready {
		t.Fatalf("helper did not become ready")
	}

	_, err = Acquire(false)
	var le LockedError
	if err == nil || !errors.As(err, &le) || le.PID <= 0 {
		t.Fatalf("Acquire(false) err = %v, want LockedError with pid", err)
	}

	l, err := Acquire(true)
	if err != nil {
		t.Fatalf("Acquire(true): %v", err)
	}
	defer l.Close()

	waitCh := make(chan error, 1)
	go func() { waitCh <- cmd.Wait() }()
	select {
	case <-time.After(5 * time.Second):
		t.Fatalf("helper process still alive after takeover")
	case <-waitCh:
	}

	// Lock file should contain our own pid now.
	data, err := os.ReadFile(filepath.Join(home, ".config", "cliamp", "session.lock"))
	if err != nil {
		t.Fatalf("Read lock file: %v", err)
	}
	if string(data) == "" {
		t.Fatalf("lock file is empty after takeover")
	}
}

func TestAcquireLockedErrorIncludesPIDFromFile(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	lockPath := filepath.Join(home, ".config", "cliamp", "session.lock")
	if err := os.MkdirAll(filepath.Dir(lockPath), 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	f, err := os.OpenFile(lockPath, os.O_CREATE|os.O_RDWR, 0o644)
	if err != nil {
		t.Fatalf("OpenFile: %v", err)
	}
	defer f.Close()
	if err := syscall.Flock(int(f.Fd()), syscall.LOCK_EX|syscall.LOCK_NB); err != nil {
		t.Fatalf("Flock: %v", err)
	}
	_, _ = f.WriteString("12345\n")

	_, err = Acquire(false)
	var le LockedError
	if !errors.As(err, &le) || le.PID != 12345 {
		t.Fatalf("Acquire(false) err = %v, want LockedError{PID:12345}", err)
	}
}
