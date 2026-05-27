package api

import (
	"errors"
	"net/http"

	"saltcrm/internal/metadata"
	"saltcrm/internal/studio"

	"github.com/gin-gonic/gin"
)

type metadataHandler struct {
	reg    *metadata.Registry
	studio *studio.Service
}

func (h *metadataHandler) get(c *gin.Context) {
	module := c.Param("module")
	base, err := h.reg.Get(module)
	if err != nil {
		if errors.Is(err, metadata.ErrUnknownModule) {
			c.JSON(http.StatusNotFound, gin.H{"error": "unknown module"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load module metadata"})
		return
	}
	// ?raw=1 serves the code+custom defaults without saved layouts applied — the
	// design-time palette the Studio layout editor needs to show hidden fields.
	var layouts map[string][]string
	if c.Query("raw") == "" {
		layouts = h.savedLayouts(c, module)
	}
	c.JSON(http.StatusOK, metadata.Resolve(base, h.customFields(c, module), layouts))
}

// customFields converts a module's runtime custom-field definitions (from
// Studio) into metadata fields for merging into its code-defined metadata.
func (h *metadataHandler) customFields(c *gin.Context, module string) []metadata.Field {
	if h.studio == nil {
		return nil
	}
	defs, err := h.studio.ListByModule(c.Request.Context(), module)
	if err != nil {
		return nil
	}
	out := make([]metadata.Field, 0, len(defs))
	for _, d := range defs {
		out = append(out, metadata.Field{
			Name: d.Name, Type: metadata.FieldType(d.Type), Label: d.Label, Options: d.Options, Custom: true,
		})
	}
	return out
}

// savedLayouts returns the module's saved view layouts, or nil if none/unavailable.
func (h *metadataHandler) savedLayouts(c *gin.Context, module string) map[string][]string {
	if h.studio == nil {
		return nil
	}
	layouts, err := h.studio.GetLayouts(c.Request.Context(), module)
	if err != nil {
		return nil
	}
	return layouts
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
		DetailView: metadata.DetailView{Panels: []metadata.Panel{
			{Label: "Product", Fields: []string{"name", "category", "monthlyPrice", "available"}},
		}},
		EditView: metadata.EditView{Fields: []string{"name", "category", "monthlyPrice"}},
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
			{Name: "customerSince", Type: metadata.FieldDate, Label: "Customer since"},
			{Name: "status", Type: metadata.FieldEnum, Label: "Status", Options: []string{"active", "suspended"}},
		},
		ListView: metadata.ListView{Columns: []string{"name", "email", "accountNumber", "status"}},
		DetailView: metadata.DetailView{Panels: []metadata.Panel{
			{Label: "Profile", Fields: []string{"name", "email", "phone", "serviceAddress"}},
			{Label: "Account", Fields: []string{"accountNumber", "customerSince", "status"}},
		}},
		EditView: metadata.EditView{Fields: []string{"name", "email", "phone", "serviceAddress", "accountNumber", "status"}},
		Subpanels: []metadata.Subpanel{
			{
				Label: "Contacts",
				Path:  "/accounts/{id}/contacts",
				Columns: []metadata.Field{
					{Name: "name", Type: metadata.FieldString, Label: "Name"},
					{Name: "title", Type: metadata.FieldString, Label: "Title"},
					{Name: "email", Type: metadata.FieldString, Label: "Email"},
					{Name: "phone", Type: metadata.FieldString, Label: "Phone"},
				},
			},
			{
				Label: "Cases",
				Path:  "/customers/{id}/cases",
				Columns: []metadata.Field{
					{Name: "subject", Type: metadata.FieldString, Label: "Subject"},
					{Name: "category", Type: metadata.FieldString, Label: "Category"},
					{Name: "priority", Type: metadata.FieldString, Label: "Priority"},
					{Name: "status", Type: metadata.FieldString, Label: "Status"},
				},
			},
			{
				Label: "Subscriptions",
				Path:  "/customers/{id}/subscriptions",
				Columns: []metadata.Field{
					{Name: "status", Type: metadata.FieldString, Label: "Status"},
					{Name: "quantity", Type: metadata.FieldString, Label: "Qty"},
					{Name: "monthlyPriceSnapshot", Type: metadata.FieldCurrency, Label: "Monthly price"},
				},
			},
			{
				Label: "Activities",
				Path:  "/activities?parentType=account&parentId={id}",
				Columns: []metadata.Field{
					{Name: "type", Type: metadata.FieldString, Label: "Type"},
					{Name: "subject", Type: metadata.FieldString, Label: "Subject"},
					{Name: "status", Type: metadata.FieldString, Label: "Status"},
					{Name: "occurredAt", Type: metadata.FieldDate, Label: "When"},
				},
			},
		},
	})
	r.Register(metadata.ModuleMeta{
		Module:        "contacts",
		Label:         "Contacts",
		LabelSingular: "Contact",
		Fields: []metadata.Field{
			{Name: "name", Type: metadata.FieldString, Label: "Name"},
			{Name: "title", Type: metadata.FieldString, Label: "Title"},
			{Name: "email", Type: metadata.FieldString, Label: "Email"},
			{Name: "phone", Type: metadata.FieldString, Label: "Phone"},
		},
		ListView: metadata.ListView{Columns: []string{"name", "title", "email", "phone"}},
		DetailView: metadata.DetailView{Panels: []metadata.Panel{
			{Label: "Contact", Fields: []string{"name", "title", "email", "phone"}},
		}},
		EditView: metadata.EditView{Fields: []string{"name", "title", "email", "phone"}},
	})
	r.Register(metadata.ModuleMeta{
		Module:        "leads",
		Label:         "Leads",
		LabelSingular: "Lead",
		Fields: []metadata.Field{
			{Name: "name", Type: metadata.FieldString, Label: "Name"},
			{Name: "company", Type: metadata.FieldString, Label: "Company"},
			{Name: "email", Type: metadata.FieldString, Label: "Email"},
			{Name: "phone", Type: metadata.FieldString, Label: "Phone"},
			{Name: "status", Type: metadata.FieldEnum, Label: "Status", Options: []string{"new", "working", "qualified", "unqualified", "converted"}},
		},
		ListView: metadata.ListView{Columns: []string{"name", "company", "status", "email"}},
		DetailView: metadata.DetailView{Panels: []metadata.Panel{
			{Label: "Lead", Fields: []string{"name", "company", "email", "phone", "status"}},
		}},
		EditView: metadata.EditView{Fields: []string{"name", "company", "email", "phone", "status"}},
	})
	r.Register(metadata.ModuleMeta{
		Module:        "opportunities",
		Label:         "Opportunities",
		LabelSingular: "Opportunity",
		Fields: []metadata.Field{
			{Name: "name", Type: metadata.FieldString, Label: "Name"},
			{Name: "amount", Type: metadata.FieldCurrency, Label: "Amount"},
			{Name: "stage", Type: metadata.FieldEnum, Label: "Stage", Options: []string{"prospecting", "qualification", "proposal", "negotiation", "closed_won", "closed_lost"}},
			{Name: "probability", Type: metadata.FieldString, Label: "Win %"},
			{Name: "expectedCloseDate", Type: metadata.FieldDate, Label: "Expected close"},
		},
		ListView: metadata.ListView{Columns: []string{"name", "stage", "amount", "expectedCloseDate"}},
		DetailView: metadata.DetailView{Panels: []metadata.Panel{
			{Label: "Opportunity", Fields: []string{"name", "amount", "stage", "probability", "expectedCloseDate"}},
		}},
		EditView: metadata.EditView{Fields: []string{"name", "amount", "stage"}},
		Subpanels: []metadata.Subpanel{
			{
				Label: "Line items",
				Path:  "/opportunities/{id}/line-items",
				Columns: []metadata.Field{
					{Name: "productName", Type: metadata.FieldString, Label: "Product"},
					{Name: "quantity", Type: metadata.FieldString, Label: "Qty"},
					{Name: "unitPrice", Type: metadata.FieldCurrency, Label: "Unit price"},
					{Name: "lineTotal", Type: metadata.FieldCurrency, Label: "Total"},
				},
			},
		},
	})
	r.Register(metadata.ModuleMeta{
		Module:        "subscriptions",
		Label:         "Subscriptions",
		LabelSingular: "Subscription",
		Fields: []metadata.Field{
			{Name: "status", Type: metadata.FieldString, Label: "Status"},
			{Name: "quantity", Type: metadata.FieldString, Label: "Qty"},
			{Name: "monthlyPriceSnapshot", Type: metadata.FieldCurrency, Label: "Monthly price"},
			{Name: "startDate", Type: metadata.FieldDate, Label: "Start"},
		},
		ListView: metadata.ListView{Columns: []string{"status", "quantity", "monthlyPriceSnapshot", "startDate"}},
		DetailView: metadata.DetailView{Panels: []metadata.Panel{
			{Label: "Subscription", Fields: []string{"status", "quantity", "monthlyPriceSnapshot", "startDate"}},
		}},
	})
	r.Register(metadata.ModuleMeta{
		Module:        "activities",
		Label:         "Activities",
		LabelSingular: "Activity",
		Fields: []metadata.Field{
			{Name: "type", Type: metadata.FieldEnum, Label: "Type", Options: []string{"call", "meeting", "task"}},
			{Name: "subject", Type: metadata.FieldString, Label: "Subject"},
			{Name: "status", Type: metadata.FieldEnum, Label: "Status", Options: []string{"open", "done"}},
			{Name: "occurredAt", Type: metadata.FieldDate, Label: "When"},
		},
		ListView: metadata.ListView{Columns: []string{"type", "subject", "status", "occurredAt"}},
		DetailView: metadata.DetailView{Panels: []metadata.Panel{
			{Label: "Activity", Fields: []string{"type", "subject", "status", "occurredAt"}},
		}},
	})
	return r
}
