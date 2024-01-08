package rest

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type NodeAssociationNoSession struct {
	ID            string
	Addr          string
	NextSessionID uint64
}

type NodeAssociationMapNoSession map[string]NodeAssociationNoSession

// ListPfcpAssociations godoc
//	@Summary		List PFCP associations
//	@Description	List PFCP associations
//	@Tags			PFCP
//	@Produce		json
//	@Success		200	{object}	NodeAssociationMapNoSession
//	@Router			/pfcp_associations [get]
func (h *ApiHandler) listPfcpAssociations(c *gin.Context) {

	nodeAssociationsNoSession := make(NodeAssociationMapNoSession)
	for k, v := range h.PfcpSrv.NodeAssociations {
		nodeAssociationsNoSession[k] = NodeAssociationNoSession{
			ID:            v.ID,
			Addr:          v.Addr,
			NextSessionID: v.NextSessionID,
		}
	}
	c.IndentedJSON(http.StatusOK, nodeAssociationsNoSession)
}

// ListPfcpAssociationsFull godoc
//	@Summary		List PFCP associations
//	@Description	List PFCP associations
//	@Tags			PFCP
//	@Produce		json
//	@Success		200	{object}	map[string]core.NodeAssociation
//	@Router			/pfcp_associations/full [get]
func (h *ApiHandler) listPfcpAssociationsFull(c *gin.Context) {
	c.IndentedJSON(http.StatusOK, h.PfcpSrv.NodeAssociations)
}
