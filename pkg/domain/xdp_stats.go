package domain

type XdpStats struct {
	Aborted  uint64 `json:"aborted"`
	Drop     uint64 `json:"drop"`
	Pass     uint64 `json:"pass"`
	Tx       uint64 `json:"tx"`
	Redirect uint64 `json:"redirect"`
}
