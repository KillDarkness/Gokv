package protocol

import (
	"reflect"
	"strings"
	"testing"
)

func TestParserReadRESPArray(t *testing.T) {
	parser := NewParser(strings.NewReader("*2\r\n$4\r\nPING\r\n$4\r\ntest\r\n"))

	got, err := parser.ReadCommand()
	if err != nil {
		t.Fatalf("ReadCommand() error = %v", err)
	}

	want := []string{"PING", "test"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("ReadCommand() = %#v; want %#v", got, want)
	}
}

func TestWriterBulkString(t *testing.T) {
	var b strings.Builder
	if err := WriteReply(&b, BulkString{Value: "kill"}); err != nil {
		t.Fatalf("WriteReply() error = %v", err)
	}
	if got, want := b.String(), "$4\r\nkill\r\n"; got != want {
		t.Fatalf("WriteReply() = %q; want %q", got, want)
	}
}
