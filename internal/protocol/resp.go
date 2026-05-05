package protocol

import (
	"fmt"
	"io"
)

type Reply interface {
	WriteRESP(w io.Writer) error
}

type SimpleString string

func (r SimpleString) WriteRESP(w io.Writer) error {
	_, err := fmt.Fprintf(w, "+%s\r\n", string(r))
	return err
}

type Error string

func (r Error) WriteRESP(w io.Writer) error {
	_, err := fmt.Fprintf(w, "-ERR %s\r\n", string(r))
	return err
}

type Integer int64

func (r Integer) WriteRESP(w io.Writer) error {
	_, err := fmt.Fprintf(w, ":%d\r\n", int64(r))
	return err
}

type BulkString struct {
	Value string
}

func (r BulkString) WriteRESP(w io.Writer) error {
	_, err := fmt.Fprintf(w, "$%d\r\n%s\r\n", len([]byte(r.Value)), r.Value)
	return err
}

type NullBulkString struct{}

func (r NullBulkString) WriteRESP(w io.Writer) error {
	_, err := io.WriteString(w, "$-1\r\n")
	return err
}

type Array []Reply

func (r Array) WriteRESP(w io.Writer) error {
	if _, err := fmt.Fprintf(w, "*%d\r\n", len(r)); err != nil {
		return err
	}
	for _, item := range r {
		if item == nil {
			item = NullBulkString{}
		}
		if err := item.WriteRESP(w); err != nil {
			return err
		}
	}
	return nil
}
