package protocol

import (
	"encoding/json"
)

type SiteID uint64
type State json.RawMessage

type Site struct {
	ID SiteID `json:"id"`
}

type River struct {
	Source SiteID `json:"source"`
	Target SiteID `json:"target"`
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

type Setup struct {
	Punter  uint64 `json:"punter"`
	Punters uint64 `json:"punters"`
	Map     Map    `json:"map"`
}

type Ready struct {
	Ready uint64 `json:"ready"`
	State State  `json:"state"`
}

type Claim struct {
	Punter uint64 `json:"punter"`
	Source SiteID `json:"source"`
	Target SiteID `json:"target"`
}

type Pass struct {
	Punter uint64 `json:"punter"`
}

type Move struct {
	Setup
	Claim *Claim `json:"claim,omitempty"`
	Pass  *Pass  `json:"pass,omitempty"`
}

type GameplayInput struct {
	Setup
	Move *struct {
		Moves []Move `json:"moves"`
	} `json:"move"`
	Stop *struct {
		Moves  []Move  `json:"moves"`
		Scores []Score `json:"scores"`
	} `json:"stop"`
	State State `json:"state"`
}

type GameplayOutput struct {
	Move
	State State `json:"state"`
}

type Score struct {
	Punter uint64 `json:"punter"`
	Score  uint64 `json:"score"`
}

type Stop struct {
	Moves  []Move  `json:"moves"`
	Scores []Score `json:"scores"`
}

type StopInput struct {
	Stop  Stop  `json:"stop"`
	State State `json:"state"`
}

type Timeout struct {
	Timeout float64 `json:"timeout"`
}
