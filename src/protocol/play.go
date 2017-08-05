package protocol

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"io"
	"strconv"
)

func writeMessage(w io.Writer, b []byte) error {
	if _, err := fmt.Fprintf(w, "%d:", len(b)); err != nil {
		return fmt.Errorf("failed to write prefix: %v", err)
	}

	_, err := w.Write(b)
	return err
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

// Play communicates with rw to play the next stage of the game.
func Play(r io.Reader, w io.Writer, g Game) error {
	log.Printf("Play: sending handshake")
	h := HandshakeClientServer{Me: g.Name()}
	b, err := json.Marshal(&h)
	if err != nil {
		return fmt.Errorf("failed to marshal handshake: %v, %s", err, string(b))
	}

	if err := writeMessage(w, b); err != nil {
		return fmt.Errorf("failed to write handshake: %v", err)
	}

	log.Printf("Play: getting handshake response")

	br := bufio.NewReader(r)

	// Handshake completion.
	b, err = readMessage(br)
	if err != nil {
		return fmt.Errorf("failed to read handshake completion: %v", err)
	}

	var hr HandshakeServerClient
	if err := json.Unmarshal(b, &hr); err != nil {
		return fmt.Errorf("failed to unmarshal handshake: %v, %s", err, string(b))
	}

	log.Printf("Play: getting command")

	// The next message is either Setup, GameplayInput, or StopInput.
	b, err = readMessage(br)
	if err != nil {
		return fmt.Errorf("failed to read command: %v", err)
	}

	log.Printf("Play: received command: %s", string(b))

	// Unmarshal will allow missing and unknown fields, so we can't just
	// attempt to unmarshal each type. Instead, perform weak type detection
	// based on unique keys from each message.

	// Setup?
	if bytes.Contains(b, []byte(`"map"`)) {
		var setup Setup
		if err := json.Unmarshal(b, &setup); err != nil {
			return fmt.Errorf("failed to unmarshal setup: %v, %s", err, string(b))
		}

		log.Printf("Play: Setup")
		return sendResponse(w, g.Setup(&setup))
	}

	// Stop?
	if bytes.Contains(b, []byte(`"stop"`)) {
		var stop StopInput
		if err := json.Unmarshal(b, &stop); err != nil {
			return fmt.Errorf("failed to unmarshal stop input: %v, %s", err, string(b))
		}

		log.Printf("Play: Stop")
		g.Stop(&stop)
		return nil
	}

	// GameplayInput?
	//
	// N.B. StopInput also contains "move", so this must come afterwards.
	if bytes.Contains(b, []byte(`"move"`)) {
		var gi GameplayInput
		if err := json.Unmarshal(b, &gi); err != nil {
			return fmt.Errorf("failed to unmarshal game input: %v, %s", err, string(b))
		}

		log.Printf("Play: Game")
		return sendResponse(w, g.Play(&gi))
	}

	return fmt.Errorf("unknown command: %s", string(b))
}
