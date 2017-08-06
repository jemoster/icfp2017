package protocol

import (
	"bufio"
	"fmt"
	"io"

	. "github.com/jemoster/icfp2017/src/protocol/io"
)

// Play communicates with rw to play the next stage of the game.
func Play(r io.Reader, w io.Writer, g Game) error {
	h := HandshakeClientServer{Me: g.Name()}
	if err := Send(w, &h); err != nil {
		return fmt.Errorf("failed sending handshake: %v", err)
	}

	br := bufio.NewReader(r)

	var hr HandshakeServerClient
	if err := Recv(br, &hr); err != nil {
		return fmt.Errorf("failed receiving handshake: %v", err)
	}

	var input CombinedInput
	if err := Recv(br, &input); err != nil {
		return fmt.Errorf("failed to receive gameplay input: %v", err)
	}

	switch {
	case input.Setup != nil:
		r, err := g.Setup(input.Setup)
		if err != nil {
			return fmt.Errorf("setup failed: %v", err)
		}

		if err := Send(w, r); err != nil {
			return fmt.Errorf("failed to send ready: %v", err)
		}
	case input.Move != nil:
		r, err := g.Play(input.Move.Moves, input.State)
		if err != nil {
			return fmt.Errorf("move failed: %v", err)
		}

		if err := Send(w, r); err != nil {
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
