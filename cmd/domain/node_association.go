package domain

type NodeAssociationNoSession struct {
	ID            string
	Addr          string
	NextSessionID uint64
}
type NodeAssociationMapNoSession map[string]NodeAssociationNoSession
