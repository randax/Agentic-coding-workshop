package metadata

import (
	"errors"
	"testing"
)

func TestRegistryGetReturnsRegisteredModule(t *testing.T) {
	r := NewRegistry()
	r.Register(ModuleMeta{
		Module: "products", Label: "Products", LabelSingular: "Product",
		Fields: []Field{
			{Name: "name", Type: FieldString, Label: "Name"},
			{Name: "category", Type: FieldEnum, Label: "Category", Options: []string{"fiber", "router", "tv"}},
		},
		ListView: ListView{Columns: []string{"name", "category"}},
	})

	got, err := r.Get("products")
	if err != nil {
		t.Fatalf("Get returned unexpected error: %v", err)
	}
	if got.Label != "Products" || got.LabelSingular != "Product" {
		t.Errorf("labels = %q/%q, want Products/Product", got.Label, got.LabelSingular)
	}
	if len(got.Fields) != 2 || got.Fields[1].Type != FieldEnum || len(got.Fields[1].Options) != 3 {
		t.Errorf("fields = %+v, want 2 fields with an enum carrying 3 options", got.Fields)
	}
	if len(got.ListView.Columns) != 2 {
		t.Errorf("listView columns = %v, want 2", got.ListView.Columns)
	}
}

func TestRegistryGetUnknownModuleReturnsError(t *testing.T) {
	r := NewRegistry()

	_, err := r.Get("nope")
	if !errors.Is(err, ErrUnknownModule) {
		t.Fatalf("Get error = %v, want ErrUnknownModule", err)
	}
}
