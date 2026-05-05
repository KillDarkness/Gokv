package protocol

import (
	"bufio"
	"io"
	"testing"
)

func BenchmarkWriteReplyDirect(b *testing.B) {
	reply := BulkString{Value: "kill"}
	for b.Loop() {
		if err := WriteReply(io.Discard, reply); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkWriteReplyBuffered(b *testing.B) {
	reply := BulkString{Value: "kill"}
	writer := bufio.NewWriterSize(io.Discard, 32*1024)
	for b.Loop() {
		if err := WriteReply(writer, reply); err != nil {
			b.Fatal(err)
		}
	}
	if err := writer.Flush(); err != nil {
		b.Fatal(err)
	}
}
