package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"saltcrm/internal/metadata"
)

func TestGetMetadataReturnsModuleContract(t *testing.T) {
	_, router := newTestRouter(t)

	req := httptest.NewRequest(http.MethodGet, "/metadata/products", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}
	var got metadata.ModuleMeta
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode: %v; body=%s", err, rec.Body.String())
	}
	if got.Module != "products" || got.LabelSingular != "Product" {
		t.Errorf("meta = %+v, want module=products labelSingular=Product", got)
	}

	byName := map[string]metadata.Field{}
	for _, f := range got.Fields {
		byName[f.Name] = f
	}
	if byName["category"].Type != metadata.FieldEnum || len(byName["category"].Options) != 3 {
		t.Errorf("category field = %+v, want enum with 3 options", byName["category"])
	}
	if byName["monthlyPrice"].Type != metadata.FieldCurrency {
		t.Errorf("monthlyPrice field = %+v, want currency", byName["monthlyPrice"])
	}
	cols := got.ListView.Columns
	if len(cols) == 0 {
		t.Fatalf("listView has no columns")
	}
}

func TestGetMetadataUnknownModuleReturns404(t *testing.T) {
	_, router := newTestRouter(t)

	req := httptest.NewRequest(http.MethodGet, "/metadata/nope", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusNotFound, rec.Body.String())
	}
}
