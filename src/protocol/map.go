package protocol

type SiteID uint64

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
