package agent

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
)

// readTrackerKey is the context key for ReadTracker.
type readTrackerKey struct{}

// ReadTrackerKey returns the context key used to store a ReadTracker.
// Exported for use in tests that need to inject a tracker into context.
func ReadTrackerKey() any { return readTrackerKey{} }

// ReadTracker tracks which files have been read during the current agent turn.
// Used to enforce read-before-edit: file_edit and file_write on existing files
// must be preceded by a file_read of that file.
type ReadTracker struct {
	read map[string]bool
}

// NewReadTracker creates a new ReadTracker.
func NewReadTracker() *ReadTracker {
	return &ReadTracker{read: make(map[string]bool)}
}

// MarkRead records that a file has been read.
func (rt *ReadTracker) MarkRead(path string) {
	norm := normalizePath(path)
	if norm != "" {
		rt.read[norm] = true
	}
}

// HasRead returns true if the file has been read in this turn.
func (rt *ReadTracker) HasRead(path string) bool {
	norm := normalizePath(path)
	if norm == "" {
		return false
	}
	return rt.read[norm]
}

// CheckReadBeforeWrite extracts the ReadTracker from context and returns an error
// if the given path has not been read. Returns nil if the tracker is absent (e.g.,
// tool called outside the agent loop) or the file has been read.
func CheckReadBeforeWrite(ctx context.Context, path string) error {
	rt, ok := ctx.Value(readTrackerKey{}).(*ReadTracker)
	if !ok || rt == nil {
		return nil
	}
	if !rt.HasRead(path) {
		return fmt.Errorf("You must read this file with file_read before editing it. Path: %s", path)
	}
	return nil
}

// normalizePath resolves a path to an absolute, clean, symlink-resolved form.
func normalizePath(path string) string {
	if path == "" {
		return ""
	}
	if !filepath.IsAbs(path) {
		cwd, err := os.Getwd()
		if err != nil {
			return filepath.Clean(path)
		}
		path = filepath.Join(cwd, path)
	}
	path = filepath.Clean(path)
	// Try to resolve symlinks; if it fails (file doesn't exist yet), use the clean path.
	if resolved, err := filepath.EvalSymlinks(path); err == nil {
		path = resolved
	}
	return path
}
