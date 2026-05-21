package main

import (
	"context"
	"fmt"
	"io"
	"log/slog"
)

type simpleHandler struct {
	w     io.Writer
	level slog.Level
}

func (h *simpleHandler) Enabled(_ context.Context, level slog.Level) bool {
	return level >= h.level
}

func (h *simpleHandler) Handle(_ context.Context, r slog.Record) error {
	time := r.Time.Format("2006-01-02 15:04:05")
	level := r.Level.String()
	msg := r.Message

	line := fmt.Sprintf("[%s] [%s] %s", time, level, msg)
	r.Attrs(func(a slog.Attr) bool {
		line += fmt.Sprintf(" %s=%v", a.Key, a.Value.Any())
		return true
	})
	line += "\n"
	_, err := h.w.Write([]byte(line))
	return err
}

func (h *simpleHandler) WithAttrs([]slog.Attr) slog.Handler {
	return h
}

func (h *simpleHandler) WithGroup(string) slog.Handler {
	return h
}
