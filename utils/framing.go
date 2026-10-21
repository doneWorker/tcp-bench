package utils

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
)

const MaxFrameSize = 10 * 1024 * 1024

var ErrFrameTooLarge = errors.New("frame too large")

func PackFrame(data []byte) ([]byte, error) {
	if len(data) > MaxFrameSize {
		return nil, fmt.Errorf("payload too large: %d bytes (max: %d)", len(data), MaxFrameSize)
	}

	frame := make([]byte, 4+len(data))

	binary.BigEndian.PutUint32(frame[:4], uint32(len(data)))
	copy(frame[4:], data)

	return frame, nil
}

func ReadFrame(r io.Reader) ([]byte, error) {
	var hdr [4]byte
	if _, err := io.ReadFull(r, hdr[:]); err != nil {
		return nil, err
	}
	n := binary.BigEndian.Uint32(hdr[:])
	if n > MaxFrameSize {
		return nil, ErrFrameTooLarge
	}
	msg := make([]byte, n)
	if _, err := io.ReadFull(r, msg); err != nil {
		if err == io.EOF {
			err = io.ErrUnexpectedEOF
		}
		return nil, err
	}
	return msg, nil
}
