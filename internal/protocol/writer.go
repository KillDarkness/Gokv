package protocol

import "io"

func WriteReply(w io.Writer, reply Reply) error {
	if reply == nil {
		reply = NullBulkString{}
	}
	return reply.WriteRESP(w)
}
