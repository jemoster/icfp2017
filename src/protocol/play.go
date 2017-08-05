package protocol

import (
	"bufio"
	//	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"strconv"
)

func writeMessage(w io.Writer, b []byte) error {
	if _, err := fmt.Fprintf(w, "%d:", len(b)); err != nil {
		return fmt.Errorf("failed to write prefix: %v", err)
	}

	did, err := w.Write(b)

	if err != nil {
		return fmt.Errorf("error writing to buffer: %v", err)
	}

	if did != len(b) {
		return fmt.Errorf("did not write all bytes (have %d, wrote %d)", len(b), did)
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
	_, err = io.ReadFull(r, b)
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
		return fmt.Errorf("failed marshaling: %v, %s", err, string(b))
	}

	if err := writeMessage(w, b); err != nil {
		return fmt.Errorf("failed to write message: %v", err)
	}

	return nil
}

func recv(r io.Reader, v interface{}) error {
	br := bufio.NewReader(r)
	b, err := readMessage(br)

	if err != nil {
		return fmt.Errorf("failed to read message: %v", err)
	}

	if err := json.Unmarshal(b, v); err != nil {
		return fmt.Errorf("failed to unmarshal: %v, %s", err, string(b))
	}

	return nil
}

func EncodeState(v interface{}) State {
	p, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return p
}

func DecodeState(i *GameplayInput, v interface{}) error {
	return json.Unmarshal(i.State, v)
}

// Play communicates with rw to play the next stage of the game.
func Play(r io.Reader, w io.Writer, g Game) error {
	log.Printf("\n\n\n")

	var err error

	h := HandshakeClientServer{Me: g.Name()}
	err = send(w, &h)
	if err != nil {
		return fmt.Errorf("failed sending handshake: %v", err)
	}

	var hr HandshakeServerClient
	err = recv(r, &hr)
	if err != nil {
		return fmt.Errorf("failed receiving handshake: %v", err)
	}

	var input GameplayInput
	err = recv(r, &input)
	if err != nil {
		return fmt.Errorf("Failed to receive gameplay input: %v", err)
	}

	if input.State != nil {
		if input.Stop != nil {
			g.Stop(&input)
		} else {
			res := g.Play(&input)
			err := send(w, &res)
			if err != nil {
				return fmt.Errorf("Failed to send move: %v", err)
			}
		}
	} else {
		res := g.Setup(&input)
		err := send(w, &res)
		if err != nil {
			return fmt.Errorf("Failed to send ready: %v", err)
		}
	}

	return nil
}
