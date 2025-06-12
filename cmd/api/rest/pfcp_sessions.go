package rest

import (
	"net"
	"net/http"
	"strconv"

	"github.com/edgecomllc/eupf/cmd/core"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

// DeletePfcpSessions godoc
//
//	@Summary	Delete PFCP session by IMSI, MSISDN or ID
//	@Description	Deletes a PFCP session based on provided parameters. At least one parameter must be provided.
//	@Tags		PFCP
//	@Accept		json
//	@Produce	json
//	@Param		imsi	query		string	false	"IMSI of the session to delete"
//	@Param		msisdn	query		string	false	"MSISDN of the session to delete"
//	@Param		id		query		string	false	"ID of the session to delete"
//	@Success	200		{string}	string	"Session successfully deleted"
//	@Failure	400		{string}	string	"Bad Request - No parameters provided"
//	@Failure	404		{string}	string	"Session not found"
//	@Router		/pfcp_sessions [delete]
func (h *ApiHandler) deletePfcpSessions(c *gin.Context) {
	query := c.Request.URL.Query()
	if len(query) != 1 {
		log.Warn().Msgf("Can't delete session by several query parameters: %v", query)
		c.JSON(http.StatusBadRequest, gin.H{"error": "ambiguous query parameter"})
		return
	}

	if id, err := strconv.ParseUint(c.Query("id"), 10, 32); err == nil {
		for _, con := range h.GetPFCPSrv() {
			if con.ReleaseSessionByID(id) {
				c.IndentedJSON(http.StatusOK, gin.H{"message": "OK"})
				return
			}
		}

		log.Warn().Err(err).Msgf("can't release session by id: %s", c.Query("id"))
		c.JSON(http.StatusNotFound, gin.H{"error": "can't release session by id"})
		return
	}

	imsi := c.Query("imsi")
	msisdn := c.Query("msisdn")
	if imsi != "" || msisdn != "" {
		for _, con := range h.GetPFCPSrv() {
			if con.ReleaseSessionByUserID(imsi, msisdn) {
				c.IndentedJSON(http.StatusOK, gin.H{"message": "OK"})
				return
			}
		}
	}

	log.Warn().Msgf("can't release session")
	c.JSON(http.StatusNotFound, gin.H{"error": "can't release session"})
}

// ListPfcpSessionsFiltered godoc
//
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
		var sessions []core.Session
		for _, c := range h.GetPFCPSrv() {
			newSessions := GetAllSessions(&c.NodeAssociations)
			sessions = append(sessions, newSessions...)
		}
		c.IndentedJSON(http.StatusOK, sessions)
		return // early return if no parameters are given
	}

	if sIp != "" {
		if ip := net.ParseIP(sIp); ip != nil {
			for _, c := range h.GetPFCPSrv() {
				if session := FilterSessionsByIP(&c.NodeAssociations, ip); session != nil {
					sessions = append(sessions, *session) // Append session by IP match
				}
			}
		} else {
			c.IndentedJSON(http.StatusBadRequest, "Failed to parse IP")
		}
	}

	if sTeid != "" {
		if teid, err := strconv.Atoi(sTeid); err == nil {
			for _, c := range h.GetPFCPSrv() {
				if session := FilterSessionsByTeid(&c.NodeAssociations, uint32(teid)); session != nil {
					sessions = append(sessions, *session) // Append session by TEID match
				}
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
