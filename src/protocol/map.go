package protocol

type SiteID uint64

type Site struct {
	id SiteID
}

type River struct {
	source SiteID
	target SiteID
}

type Map struct {
	sites  []Site
	rivers []River
	mines  []SiteID
}
