package api

import (
	"errors"
	"net/http"

	"saltcrm/internal/metadata"

	"github.com/gin-gonic/gin"
)

type metadataHandler struct {
	reg *metadata.Registry
}

func (h *metadataHandler) get(c *gin.Context) {
	m, err := h.reg.Get(c.Param("module"))
	if err != nil {
		if errors.Is(err, metadata.ErrUnknownModule) {
			c.JSON(http.StatusNotFound, gin.H{"error": "unknown module"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load module metadata"})
		return
	}
	c.JSON(http.StatusOK, m)
}

// defaultRegistry registers the metadata for every module the generic views can
// render. For now this is hand-written per module (the typed-core approach);
// custom fields and saved layouts are merged on in later slices.
func defaultRegistry() *metadata.Registry {
	r := metadata.NewRegistry()
	r.Register(metadata.ModuleMeta{
		Module:        "products",
		Label:         "Products",
		LabelSingular: "Product",
		Fields: []metadata.Field{
			{Name: "name", Type: metadata.FieldString, Label: "Name"},
			{Name: "category", Type: metadata.FieldEnum, Label: "Category", Options: []string{"fiber", "router", "tv"}},
			{Name: "monthlyPrice", Type: metadata.FieldCurrency, Label: "Monthly price"},
			{Name: "available", Type: metadata.FieldBool, Label: "Status"},
		},
		ListView: metadata.ListView{Columns: []string{"name", "category", "monthlyPrice", "available"}},
	})
	r.Register(metadata.ModuleMeta{
		Module:        "accounts",
		Label:         "Accounts",
		LabelSingular: "Account",
		Fields: []metadata.Field{
			{Name: "name", Type: metadata.FieldString, Label: "Name"},
			{Name: "email", Type: metadata.FieldString, Label: "Email"},
			{Name: "phone", Type: metadata.FieldString, Label: "Phone"},
			{Name: "serviceAddress", Type: metadata.FieldString, Label: "Service address"},
			{Name: "accountNumber", Type: metadata.FieldString, Label: "Account number"},
			{Name: "status", Type: metadata.FieldEnum, Label: "Status", Options: []string{"active", "suspended"}},
		},
		ListView: metadata.ListView{Columns: []string{"name", "email", "accountNumber", "status"}},
	})
	return r
}
