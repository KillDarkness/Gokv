package protocol

import (
	"io"
	"strconv"
)

func WriteReply(w io.Writer, reply Reply) error {
	if reply == nil {
		reply = NullBulkString{}
	}
	return reply.WriteRESP(w)
}

func WriteCommand(w io.Writer, args []string) error {
	buf := strconv.AppendInt([]byte{'*'}, int64(len(args)), 10)
	buf = append(buf, '\r', '\n')
	if _, err := w.Write(buf); err != nil {
		return err
	}
	for _, arg := range args {
		buf = strconv.AppendInt([]byte{'$'}, int64(len(arg)), 10)
		buf = append(buf, '\r', '\n')
		if _, err := w.Write(buf); err != nil {
			return err
		}
		if _, err := io.WriteString(w, arg); err != nil {
			return err
		}
		if _, err := io.WriteString(w, "\r\n"); err != nil {
			return err
		}
	}
	return nil
}
