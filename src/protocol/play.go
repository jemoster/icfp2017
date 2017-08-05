package protocol

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
)

func writeMessage(w io.Writer, b []byte) error {
	if _, err := fmt.Fprintf(w, "%d:", len(b)); err != nil {
		return fmt.Errorf("failed to write prefix: %v", err)
	}

	if _, err := w.Write(b); err != nil {
		return fmt.Errorf("error writing buffer: %v", err)
	}

	return nil
}

func readMessage(r *bufio.Reader) ([]byte, error) {
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

func sendResponse(w io.Writer, v interface{}) error {
	b, err := json.Marshal(v)
	if err != nil {
		return fmt.Errorf("failed to marshal response %+v: %v", v, err)
	}

	return writeMessage(w, b)
}

func send(w io.Writer, v interface{}) error {
	b, err := json.Marshal(v)
	if err != nil {
		return fmt.Errorf("failed marshaling %+v: %v", v, err)
	}

	if err := writeMessage(w, b); err != nil {
		return fmt.Errorf("failed to write message: %v", err)
	}

	return nil
}

func recv(r *bufio.Reader, v interface{}) error {
	b, err := readMessage(r)
	if err != nil {
		return fmt.Errorf("failed to read message: %v", err)
	}

	if err := json.Unmarshal(b, v); err != nil {
		return fmt.Errorf("failed to unmarshal: %v, %s", err, string(b))
	}

	return nil
}

// Play communicates with rw to play the next stage of the game.
func Play(r io.Reader, w io.Writer, g Game) error {
	h := HandshakeClientServer{Me: g.Name()}
	if err := send(w, &h); err != nil {
		return fmt.Errorf("failed sending handshake: %v", err)
	}

	br := bufio.NewReader(r)

	var hr HandshakeServerClient
	if err := recv(br, &hr); err != nil {
		return fmt.Errorf("failed receiving handshake: %v", err)
	}

	var input CombinedInput
	if err := recv(br, &input); err != nil {
		return fmt.Errorf("failed to receive gameplay input: %v", err)
	}

	switch {
	case input.Setup != nil:
		r, err := g.Setup(input.Setup)
		if err != nil {
			return fmt.Errorf("setup failed: %v", err)
		}

		if err := send(w, r); err != nil {
			return fmt.Errorf("failed to send ready: %v", err)
		}
	case input.Move != nil:
		r, err := g.Play(input.Move.Moves, input.State)
		if err != nil {
			return fmt.Errorf("move failed: %v", err)
		}

		if err := send(w, r); err != nil {
			return fmt.Errorf("failed to send move: %v", err)
		}
	case input.Stop != nil:
		if err := g.Stop(input.Stop, input.State); err != nil {
			return fmt.Errorf("stop failed: %v", err)
		}
	default:
		return fmt.Errorf("unknown input %+v", input)
	}

	return nil
}
