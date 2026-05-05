package server

import (
	"context"
	"reflect"
	"testing"
)

type recordingAppender struct {
	commands [][]string
}

func (a *recordingAppender) Append(ctx context.Context, args []string) error {
	copyArgs := append([]string(nil), args...)
	a.commands = append(a.commands, copyArgs)
	return nil
}

func TestDatabaseAppenderSkipsDuplicateSelect(t *testing.T) {
	recorder := &recordingAppender{}
	appender := newDatabaseAppender(recorder)
	ctx := context.Background()

	appender.Select(0)
	if err := appender.Append(ctx, []string{"SET", "a", "1"}); err != nil {
		t.Fatalf("Append() error = %v", err)
	}
	if err := appender.Append(ctx, []string{"SET", "b", "2"}); err != nil {
		t.Fatalf("Append() error = %v", err)
	}
	appender.Select(1)
	if err := appender.Append(ctx, []string{"SET", "c", "3"}); err != nil {
		t.Fatalf("Append() error = %v", err)
	}

	want := [][]string{
		{"SELECT", "0"},
		{"SET", "a", "1"},
		{"SET", "b", "2"},
		{"SELECT", "1"},
		{"SET", "c", "3"},
	}
	if !reflect.DeepEqual(recorder.commands, want) {
		t.Fatalf("commands = %#v; want %#v", recorder.commands, want)
	}
}
