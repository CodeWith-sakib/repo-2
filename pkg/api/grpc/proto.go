package grpc

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
)

type WireType uint8

const (
	WireVarint  WireType = 0
	WireFixed64 WireType = 1
	WireLength  WireType = 2
	WireFixed32 WireType = 5
)

type WireMessage struct {
	fields map[int][]byte
}

func NewWireMessage() *WireMessage {
	return &WireMessage{fields: make(map[int][]byte)}
}

func (m *WireMessage) SetString(tag int, val string) {
	m.fields[tag] = []byte(val)
}

func (m *WireMessage) GetString(tag int) string {
	return string(m.fields[tag])
}

func (m *WireMessage) SetBytes(tag int, val []byte) {
	m.fields[tag] = val
}

func (m *WireMessage) GetBytes(tag int) []byte {
	return m.fields[tag]
}

func (m *WireMessage) SetVarint(tag int, val uint64) {
	buf := make([]byte, binary.MaxVarintLen64)
	n := binary.PutUvarint(buf, val)
	m.fields[tag] = buf[:n]
}

func (m *WireMessage) GetVarint(tag int) (uint64, error) {
	data, ok := m.fields[tag]
	if !ok || len(data) == 0 {
		return 0, errors.New("field not found")
	}
	val, n := binary.Uvarint(data)
	if n <= 0 {
		return 0, errors.New("invalid varint encoding")
	}
	return val, nil
}

func (m *WireMessage) Encode() []byte {
	var out []byte
	for tag, data := range m.fields {
		key := uint64(tag<<3) | uint64(WireLength)
		buf := make([]byte, binary.MaxVarintLen64)
		n := binary.PutUvarint(buf, key)
		out = append(out, buf[:n]...)

		lengthBuf := make([]byte, binary.MaxVarintLen64)
		ln := binary.PutUvarint(lengthBuf, uint64(len(data)))
		out = append(out, lengthBuf[:ln]...)
		out = append(out, data...)
	}
	return out
}

func DecodeWireMessage(data []byte) (*WireMessage, error) {
	msg := NewWireMessage()
	offset := 0
	for offset < len(data) {
		key, n := binary.Uvarint(data[offset:])
		if n <= 0 {
			return nil, io.ErrUnexpectedEOF
		}
		offset += n

		tag := int(key >> 3)
		wire := WireType(key & 0x7)

		switch wire {
		case WireLength:
			length, ln := binary.Uvarint(data[offset:])
			if ln <= 0 {
				return nil, io.ErrUnexpectedEOF
			}
			offset += ln
			if offset+int(length) > len(data) {
				return nil, fmt.Errorf("length overflow reading field tag %d", tag)
			}
			val := make([]byte, length)
			copy(val, data[offset:offset+int(length)])
			msg.fields[tag] = val
			offset += int(length)
		default:
			return nil, fmt.Errorf("unsupported wire type: %d", wire)
		}
	}
	return msg, nil
}
