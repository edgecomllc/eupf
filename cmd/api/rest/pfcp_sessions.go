package rest

import (
	"github.com/edgecomllc/eupf/cmd/core"
	"github.com/gin-gonic/gin"
	"net"
	"net/http"
	"strconv"
)

// ListPfcpSessionsFiltered godoc
//	@Summary	If no parameters are given, list all PFCP sessions. If ip or teid is given, single session will be returned. If both ip and teid are given, it is possible to return two sessions.
//	@Tags		PFCP
//	@Produce	json
//	@Param		ip		query		string	false	"ip"
//	@Param		teid	query		int		false	"teid"
//	@Success	200		{object}	[]core.Session
//	@Router		/pfcp_sessions [get]
func (h *ApiHandler) listPfcpSessionsFiltered(c *gin.Context) {
	var sessions []core.Session
	sIp := c.Query("ip")
	sTeid := c.Query("teid")
	if sIp == "" && sTeid == "" {
		sessions = GetAllSessions(&h.PfcpSrv.NodeAssociations)
		c.IndentedJSON(http.StatusOK, sessions)
		return // early return if no parameters are given
	}

	if sIp != "" {
		if ip := net.ParseIP(sIp); ip != nil {
			if session := FilterSessionsByIP(&h.PfcpSrv.NodeAssociations, ip); session != nil {
				sessions = append(sessions, *session) // Append session by IP match
			}
		} else {
			c.IndentedJSON(http.StatusBadRequest, "Failed to parse IP")
		}
	}

	if sTeid != "" {
		if teid, err := strconv.Atoi(sTeid); err == nil {
			if session := FilterSessionsByTeid(&h.PfcpSrv.NodeAssociations, uint32(teid)); session != nil {
				sessions = append(sessions, *session) // Append session by TEID match
			}
		} else {
			c.IndentedJSON(http.StatusBadRequest, "Failed to parse TEID")
		}
	}
	c.IndentedJSON(http.StatusOK, sessions)
}

func GetAllSessions(nodeMap *map[string]*core.NodeAssociation) (sessions []core.Session) {
	for _, nodeAssoc := range *nodeMap {
		for _, session := range nodeAssoc.Sessions {
			sessions = append(sessions, *session)
		}
	}
	return
}

func FilterSessionsByIP(nodeMap *map[string]*core.NodeAssociation, filterByIP net.IP) *core.Session {
	for _, nodeAssoc := range *nodeMap {
		for _, session := range nodeAssoc.Sessions {
			for _, PDR := range session.PDRs {
				if PDR.Ipv4.Equal(filterByIP) {
					return session
				}
			}
		}
	}
	return nil
}

func FilterSessionsByTeid(nodeMap *map[string]*core.NodeAssociation, filterByTeid uint32) *core.Session {
	for _, nodeAssoc := range *nodeMap {
		for _, session := range nodeAssoc.Sessions {
			for _, PDR := range session.PDRs {
				if PDR.Teid == filterByTeid {
					return session
				}
			}
		}
	}
	return nil
}
