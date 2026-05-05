package protocol

import (
	"fmt"
	"io"
)

func WriteReply(w io.Writer, reply Reply) error {
	if reply == nil {
		reply = NullBulkString{}
	}
	return reply.WriteRESP(w)
}

func WriteCommand(w io.Writer, args []string) error {
	if _, err := fmt.Fprintf(w, "*%d\r\n", len(args)); err != nil {
		return err
	}
	for _, arg := range args {
		if _, err := fmt.Fprintf(w, "$%d\r\n%s\r\n", len([]byte(arg)), arg); err != nil {
			return err
		}
	}
	return nil
}
