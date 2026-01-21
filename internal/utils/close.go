package utils

import (
	"io"
	"log/slog"
)

type CancelOnClose struct {
	io.ReadCloser
	Cancel func()
}

// Close closes c and ignores any error.
// Use for best-effort cleanup in defer where error handling is not critical.
func Close(c io.Closer) {
	_ = c.Close()
}

// MustClose closes c and logs any error.
// Use for defer statements where we want to track close errors.
func MustClose(c io.Closer) {
	if err := c.Close(); err != nil {
		slog.Warn("failed to close", "error", err)
	}
}

func (c *CancelOnClose) CancelOnCloseFunc() error {
	if c.Cancel != nil {
		c.Cancel()
	}
	return c.Close()
}
