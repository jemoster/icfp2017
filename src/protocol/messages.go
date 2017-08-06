package protocol

import (
	"encoding/json"
	"fmt"
)

type SiteID uint64

func (sid SiteID) ID() int64 {
	return int64(sid)
}

type Site struct {
	ID SiteID `json:"id"`
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

type River struct {
	Source SiteID `json:"source"`
	Target SiteID `json:"target"`

	// State additions not covered in the official protocol.
	IsOwned     bool   `json:"isowned,omitempty"`
	OwnerPunter uint64 `json:"owner,omitempty"`

	IsOptioned   bool   `json:"isoptioned,omitempty"`
	OptionPunter uint64 `json:"optionowner,omitempty"`
}

type Map struct {
	Sites  []Site   `json:"sites"`
	Rivers []River  `json:"rivers"`
	Mines  []SiteID `json:"mines"`
}

type HandshakeClientServer struct {
	Me string `json:"me"`
}

type HandshakeServerClient struct {
	You string `json:"me"`
}

type Settings struct {
	Futures  bool `json:futures"`
	Splurges bool `json:splurges"`
	Options  bool `json:options"`
}

type Setup struct {
	Punter   uint64   `json:"punter"`
	Punters  uint64   `json:"punters"`
	Map      Map      `json:"map"`
	Settings Settings `json:"settings"`
}

type Ready struct {
	Ready uint64 `json:"ready"`

	// State is the Game's internal state, which will be marshalled to
	// JSON.
	State interface{} `json:"state"`
}

type Claim struct {
	Punter uint64 `json:"punter"`
	Source SiteID `json:"source"`
	Target SiteID `json:"target"`
}

type Pass struct {
	Punter uint64 `json:"punter"`
}

type Splurge struct {
	Punter uint64   `json:"punter"`
	Route  []SiteID `json:"route"`
}

// Option is equivalent to a claim.
type Option Claim

type Move struct {
	Claim   *Claim   `json:"claim,omitempty"`
	Pass    *Pass    `json:"pass,omitempty"`
	Splurge *Splurge `json:"splurge,omitempty"`
	Option  *Option  `json:"option,omitempty"`
}

func (m Move) String() string {
	if m.Claim != nil {
		return fmt.Sprintf("{Claim: %+v}", m.Claim)
	}
	if m.Pass != nil {
		return fmt.Sprintf("{Pass: %+v}", m.Pass)
	}
	return "{<nil>}"
}

// CombinedInput contains all the fields from the setup, gameplay, and stop
// states.
type CombinedInput struct {
	*Setup
	Move *struct {
		Moves []Move `json:"moves"`
	} `json:"move"`
	Stop *Stop `json:"stop"`

	// State is the Game's internal state, which cannot be decoded by this
	// package.
	State json.RawMessage `json:"state"`
}

type GameplayOutput struct {
	Move

	// State is the Game's internal state, which will be marshalled to
	// JSON.
	State interface{} `json:"state"`
}

type Score struct {
	Punter uint64 `json:"punter"`
	Score  int64  `json:"score"`
}

type Stop struct {
	Moves  []Move  `json:"moves"`
	Scores []Score `json:"scores"`
}

type Timeout struct {
	Timeout float64 `json:"timeout"`
}
