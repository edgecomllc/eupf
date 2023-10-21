package domain

type FarMapElement struct {
	Id                    uint32 `json:"id"`
	Action                uint8  `json:"action"`
	OuterHeaderCreation   uint8  `json:"outer_header_creation"`
	Teid                  uint32 `json:"teid"`
	RemoteIP              uint32 `json:"remote_ip"`
	LocalIP               uint32 `json:"local_ip"`
	TransportLevelMarking uint16 `json:"transport_level_marking"`
}
