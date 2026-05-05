package protocol

import (
	"io"
	"strconv"
)

type Reply interface {
	WriteRESP(w io.Writer) error
}

type SimpleString string

func (r SimpleString) WriteRESP(w io.Writer) error {
	if _, err := io.WriteString(w, "+"); err != nil {
		return err
	}
	if _, err := io.WriteString(w, string(r)); err != nil {
		return err
	}
	_, err := io.WriteString(w, "\r\n")
	return err
}

type Error string

func (r Error) WriteRESP(w io.Writer) error {
	if _, err := io.WriteString(w, "-ERR "); err != nil {
		return err
	}
	if _, err := io.WriteString(w, string(r)); err != nil {
		return err
	}
	_, err := io.WriteString(w, "\r\n")
	return err
}

type Integer int64

func (r Integer) WriteRESP(w io.Writer) error {
	buf := strconv.AppendInt([]byte{':'}, int64(r), 10)
	buf = append(buf, '\r', '\n')
	_, err := w.Write(buf)
	return err
}

type BulkString struct {
	Value string
}

func (r BulkString) WriteRESP(w io.Writer) error {
	buf := strconv.AppendInt([]byte{'$'}, int64(len(r.Value)), 10)
	buf = append(buf, '\r', '\n')
	if _, err := w.Write(buf); err != nil {
		return err
	}
	if _, err := io.WriteString(w, r.Value); err != nil {
		return err
	}
	_, err := io.WriteString(w, "\r\n")
	return err
}

type NullBulkString struct{}

func (r NullBulkString) WriteRESP(w io.Writer) error {
	_, err := io.WriteString(w, "$-1\r\n")
	return err
}

type Array []Reply

func (r Array) WriteRESP(w io.Writer) error {
	buf := strconv.AppendInt([]byte{'*'}, int64(len(r)), 10)
	buf = append(buf, '\r', '\n')
	if _, err := w.Write(buf); err != nil {
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
