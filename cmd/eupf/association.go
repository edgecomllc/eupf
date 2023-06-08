package main

import "log"

type SessionMap map[uint64]Session

type NodeAssociation struct {
	ID            string
	Addr          string
	NextSessionID uint64
	Sessions      SessionMap
}

func (a *NodeAssociation) NewLocalSEID() uint64 {
	a.NextSessionID += 1
	return a.NextSessionID
}

func (a *NodeAssociation) AddSession(localSEID uint64, s Session) {
	log.Printf("AddSession: localSEID=%d, s=%v\n", localSEID, s)
	a.Sessions[localSEID] = s
}

func (a *NodeAssociation) UpdateSession(localSEID uint64, s Session) {
	log.Printf("UpdateSession: localSEID=%d, s=%v\n", localSEID, s)
	a.Sessions[localSEID] = s
}

func (a *NodeAssociation) GetSession(localSEID uint64) (Session, bool) {
	log.Printf("GetSession: localSEID=%d\n", localSEID)
	s, ok := a.Sessions[localSEID]
	return s, ok
}

func (a *NodeAssociation) DeleteSession(localSEID uint64) {
	log.Printf("DeleteSession: localSEID=%d\n", localSEID)
	delete(a.Sessions, localSEID)
}
