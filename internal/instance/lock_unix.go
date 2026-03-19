//go:build unix

package instance

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	"cliamp/internal/appdir"
)

const (
	takeoverPollInterval = 100 * time.Millisecond
	takeoverTimeout      = 8 * time.Second
	takeoverKillAfter    = 3 * time.Second
)

type Lock struct {
	f *os.File
}

// LockedError reports that another cliamp instance currently holds the lock.
type LockedError struct {
	PID int
}

func (e LockedError) Error() string {
	if e.PID > 0 {
		return fmt.Sprintf("another cliamp instance is running (pid %d). Re-run with --takeover to stop it.", e.PID)
	}
	return "another cliamp instance is running. Re-run with --takeover to stop it."
}

func lockFile() (string, error) {
	dir, err := appdir.Dir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "session.lock"), nil
}

func readPID(f *os.File) int {
	if _, err := f.Seek(0, 0); err != nil {
		return 0
	}
	data, err := os.ReadFile(f.Name())
	if err != nil {
		return 0
	}
	pid, _ := strconv.Atoi(strings.TrimSpace(string(data)))
	return pid
}

func writePID(f *os.File, pid int) {
	if err := f.Truncate(0); err != nil {
		return
	}
	if _, err := f.Seek(0, 0); err != nil {
		return
	}
	_, _ = fmt.Fprintf(f, "%d\n", pid)
}

func tryLock(f *os.File) error {
	return syscall.Flock(int(f.Fd()), syscall.LOCK_EX|syscall.LOCK_NB)
}

// Acquire obtains a process lock for cliamp.
// If takeover is true, it sends SIGTERM to the lock holder and retries.
func Acquire(takeover bool) (*Lock, error) {
	path, err := lockFile()
	if err != nil {
		return nil, err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, err
	}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0o644)
	if err != nil {
		return nil, err
	}
	lockNow := func() error {
		err := tryLock(f)
		if err == nil {
			writePID(f, os.Getpid())
			return nil
		}
		if errors.Is(err, syscall.EWOULDBLOCK) || errors.Is(err, syscall.EAGAIN) {
			return err
		}
		return err
	}

	if err := lockNow(); err == nil {
		return &Lock{f: f}, nil
	} else if !errors.Is(err, syscall.EWOULDBLOCK) && !errors.Is(err, syscall.EAGAIN) {
		_ = f.Close()
		return nil, err
	}

	pid := readPID(f)
	if !takeover {
		_ = f.Close()
		return nil, LockedError{PID: pid}
	}

	if pid > 0 && pid != os.Getpid() {
		if p, err := os.FindProcess(pid); err == nil {
			_ = p.Signal(syscall.SIGTERM)
		}
	}
	start := time.Now()
	deadline := start.Add(takeoverTimeout)
	sentKill := false
	for time.Now().Before(deadline) {
		time.Sleep(takeoverPollInterval)
		if err := lockNow(); err == nil {
			return &Lock{f: f}, nil
		}
		if !sentKill && pid > 0 && time.Since(start) >= takeoverKillAfter {
			if p, err := os.FindProcess(pid); err == nil {
				_ = p.Signal(syscall.SIGKILL)
			}
			sentKill = true
		}
	}
	_ = f.Close()
	return nil, fmt.Errorf("failed to take over running instance (pid %d)", pid)
}

func (l *Lock) Close() {
	if l == nil || l.f == nil {
		return
	}
	_ = syscall.Flock(int(l.f.Fd()), syscall.LOCK_UN)
	_ = l.f.Close()
}
