package protocol

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strings"
)

var ErrInvalidRESP = errors.New("invalid RESP command")

var inlinePingCommand = []string{"PING"}

type Parser struct {
	reader *bufio.Reader
}

func NewParser(r io.Reader) *Parser {
	return &Parser{reader: bufio.NewReader(r)}
}

func (p *Parser) Buffered() int {
	return p.reader.Buffered()
}

func (p *Parser) ReadCommand() ([]string, error) {
	prefix, err := p.reader.ReadByte()
	if err != nil {
		return nil, err
	}

	switch prefix {
	case '*':
		return p.readArray()
	default:
		if err := p.reader.UnreadByte(); err != nil {
			return nil, err
		}
		return p.readInline()
	}
}

func (p *Parser) readArray() ([]string, error) {
	count, err := p.readNumberLine()
	if err != nil {
		return nil, err
	}
	if count <= 0 {
		return nil, fmt.Errorf("%w: empty array", ErrInvalidRESP)
	}

	args := make([]string, count)
	for i := range count {
		prefix, err := p.reader.ReadByte()
		if err != nil {
			return nil, err
		}
		if prefix != '$' {
			return nil, fmt.Errorf("%w: expected bulk string", ErrInvalidRESP)
		}

		length, err := p.readNumberLine()
		if err != nil {
			return nil, err
		}
		if length < 0 {
			return nil, fmt.Errorf("%w: null bulk command argument", ErrInvalidRESP)
		}

		buf := make([]byte, length+2)
		if _, err := io.ReadFull(p.reader, buf); err != nil {
			return nil, err
		}
		if buf[length] != '\r' || buf[length+1] != '\n' {
			return nil, fmt.Errorf("%w: malformed bulk string", ErrInvalidRESP)
		}
		args[i] = string(buf[:length])
	}

	return args, nil
}

func (p *Parser) readInline() ([]string, error) {
	line, err := p.readLine()
	if err != nil {
		return nil, err
	}
	if line == "PING" {
		return inlinePingCommand, nil
	}
	args := strings.Fields(line)
	if len(args) == 0 {
		return nil, fmt.Errorf("%w: empty inline command", ErrInvalidRESP)
	}
	return args, nil
}

func (p *Parser) readNumberLine() (int, error) {
	line, err := p.reader.ReadSlice('\n')
	if err != nil {
		return 0, err
	}
	if len(line) < 2 || line[len(line)-2] != '\r' {
		return 0, fmt.Errorf("%w: missing CRLF", ErrInvalidRESP)
	}
	number, ok := parsePositiveInt(line[:len(line)-2])
	if !ok {
		return 0, fmt.Errorf("%w: invalid integer", ErrInvalidRESP)
	}
	return number, nil
}

func parsePositiveInt(data []byte) (int, bool) {
	if len(data) == 0 {
		return 0, false
	}
	value := 0
	for _, b := range data {
		if b < '0' || b > '9' {
			return 0, false
		}
		value = value*10 + int(b-'0')
	}
	return value, true
}

func (p *Parser) readLine() (string, error) {
	line, err := p.reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	if !strings.HasSuffix(line, "\r\n") {
		return "", fmt.Errorf("%w: missing CRLF", ErrInvalidRESP)
	}
	return strings.TrimSuffix(line, "\r\n"), nil
}
