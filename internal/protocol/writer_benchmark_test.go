package protocol

import (
	"bufio"
	"io"
	"strings"
	"testing"
)

func BenchmarkParserReadCommand(b *testing.B) {
	const command = "*3\r\n$3\r\nSET\r\n$4\r\nname\r\n$4\r\nkill\r\n"
	input := strings.Repeat(command, b.N)
	parser := NewParser(strings.NewReader(input))

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := parser.ReadCommand(); err != nil {
			b.Fatal(err)
		}
	}
}

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
