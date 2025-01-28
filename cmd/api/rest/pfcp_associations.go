package rest

import (
	"maps"
	"net/http"

	"github.com/edgecomllc/eupf/cmd/core"
	"github.com/gin-gonic/gin"
)

type NodeAssociationDescription struct {
	ID   string
	Addr string
}

// ListPfcpAssociations godoc
//
//	@Summary		List PFCP associations
//	@Description	List PFCP associations
//	@Tags			PFCP
//	@Produce		json
//	@Success		200	{object}	NodeAssociationDescription
//	@Router			/pfcp_associations [get]
func (h *ApiHandler) listPfcpAssociations(c *gin.Context) {

	nodeAssociationsList := []NodeAssociationDescription{}

	for _, c := range h.PfcpSrv {
		for _, v := range c.NodeAssociations {
			nodeAssociationsList = append(nodeAssociationsList, NodeAssociationDescription{
				ID:   v.ID,
				Addr: v.Addr})
		}
	}
	c.IndentedJSON(http.StatusOK, nodeAssociationsList)
}

// ListPfcpAssociationsFull godoc
//
//	@Summary		List PFCP associations
//	@Description	List PFCP associations
//	@Tags			PFCP
//	@Produce		json
//	@Success		200	{object}	map[string]core.NodeAssociation
//	@Router			/pfcp_associations/full [get]
func (h *ApiHandler) listPfcpAssociationsFull(c *gin.Context) {

	associations := map[string]*core.NodeAssociation{}
	for _, c := range h.PfcpSrv {
		maps.Copy(associations, c.NodeAssociations)
	}

	c.IndentedJSON(http.StatusOK, associations)
}
