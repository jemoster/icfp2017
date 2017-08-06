package io

import (
	"io"
	"fmt"
	"log"
	"bufio"
	"strconv"
	"encoding/json"
)

func WriteMessage(w io.Writer, b []byte) error {
	if _, err := fmt.Fprintf(w, "%d:", len(b)); err != nil {
		return fmt.Errorf("failed to write prefix: %v", err)
	}

	if _, err := w.Write(b); err != nil {
		return fmt.Errorf("error writing buffer: %v", err)
	}

	return nil
}

func ReadMessage(r *bufio.Reader) ([]byte, error) {
	b, err := r.ReadSlice(':')
	if err != nil {
		return nil, fmt.Errorf("failed to read size: %v", err)
	}

	if len(b) == 1 {
		return nil, fmt.Errorf("length missing: %q", string(b))
	}
	b = b[:len(b)-1]

	l, err := strconv.ParseUint(string(b), 10, 64)
	if err != nil {
		return nil, fmt.Errorf("received bad size %q: %v", string(b), err)
	}

	b = make([]byte, l)
	if _, err = io.ReadFull(r, b); err != nil {
		return nil, fmt.Errorf("error reading message: %v", err)
	}
	return b, err
}

func Send(w io.Writer, v interface{}) error {
	b, err := json.Marshal(v)
	if err != nil {
		return fmt.Errorf("failed marshaling %+v: %v", v, err)
	}
	log.Printf("Send(%T): %s\n", v, string(b))
	if err := WriteMessage(w, b); err != nil {
		return fmt.Errorf("failed to write message: %v", err)
	}

	return nil
}

func Recv(r *bufio.Reader, v interface{}) error {
	b, err := ReadMessage(r)
	if err != nil {
		return fmt.Errorf("failed to read message: %v", err)
	}
	log.Printf("Recv(%T): %s\n", v, string(b))
	if err := json.Unmarshal(b, v); err != nil {
		return fmt.Errorf("failed to unmarshal: %v, %s", err, string(b))
	}

	return nil
}