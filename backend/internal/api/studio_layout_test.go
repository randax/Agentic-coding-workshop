package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"saltcrm/internal/agent"
	"saltcrm/internal/metadata"
)

func putLayouts(t *testing.T, router http.Handler, cookie, body string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodPut, "/studio/layouts", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Cookie", cookie)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	return rec
}

func TestSaveLayoutRequiresAdmin(t *testing.T) {
	db, router := newTestRouter(t)
	cookie := loginAs(t, db, router, "manager@isp.example", agent.RoleManager, nil)

	rec := putLayouts(t, router, cookie, `{"module":"accounts","views":{"list":["name"]}}`)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d (only admins edit layouts)", rec.Code, http.StatusForbidden)
	}
}

func TestSavedLayoutAppliesToMetadata(t *testing.T) {
	db, router := newTestRouter(t)
	admin := loginAs(t, db, router, "admin@isp.example", agent.RoleAdmin, nil)

	if rec := putLayouts(t, router, admin, `{"module":"accounts","views":{"list":["status","name"]}}`); rec.Code != http.StatusNoContent {
		t.Fatalf("save layout status = %d, want %d; body=%s", rec.Code, http.StatusNoContent, rec.Body.String())
	}

	metaRec := authGet(t, router, "/metadata/accounts", admin)
	var m metadata.ModuleMeta
	if err := json.Unmarshal(metaRec.Body.Bytes(), &m); err != nil {
		t.Fatalf("decode metadata: %v; body=%s", err, metaRec.Body.String())
	}
	if !reflect.DeepEqual(m.ListView.Columns, []string{"status", "name"}) {
		t.Errorf("list columns = %v, want [status name] from the saved layout", m.ListView.Columns)
	}
}

func TestRawMetadataIgnoresSavedLayout(t *testing.T) {
	db, router := newTestRouter(t)
	admin := loginAs(t, db, router, "admin@isp.example", agent.RoleAdmin, nil)
	putLayouts(t, router, admin, `{"module":"accounts","views":{"list":["status"]}}`)

	metaRec := authGet(t, router, "/metadata/accounts?raw=1", admin)
	var m metadata.ModuleMeta
	if err := json.Unmarshal(metaRec.Body.Bytes(), &m); err != nil {
		t.Fatalf("decode metadata: %v; body=%s", err, metaRec.Body.String())
	}
	// raw must serve the code-defined defaults, not the saved layout.
	want := []string{"name", "email", "accountNumber", "status"}
	if !reflect.DeepEqual(m.ListView.Columns, want) {
		t.Errorf("raw list columns = %v, want defaults %v (layout ignored)", m.ListView.Columns, want)
	}
}

func TestGetLayoutsReturnsSaved(t *testing.T) {
	db, router := newTestRouter(t)
	admin := loginAs(t, db, router, "admin@isp.example", agent.RoleAdmin, nil)
	putLayouts(t, router, admin, `{"module":"accounts","views":{"list":["status","name"],"edit":["name"]}}`)

	rec := authGet(t, router, "/studio/layouts?module=accounts", admin)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}
	var got map[string][]string
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode layouts: %v; body=%s", err, rec.Body.String())
	}
	want := map[string][]string{"list": {"status", "name"}, "edit": {"name"}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("layouts = %+v, want %+v", got, want)
	}
}
