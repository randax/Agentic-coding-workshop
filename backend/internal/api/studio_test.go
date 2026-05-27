package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"saltcrm/internal/agent"
	"saltcrm/internal/customer"
	"saltcrm/internal/metadata"
)

func addField(t *testing.T, router http.Handler, cookie, body string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, "/studio/fields", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Cookie", cookie)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	return rec
}

func TestAddCustomFieldRequiresAdmin(t *testing.T) {
	db, router := newTestRouter(t)
	cookie := loginAs(t, db, router, "manager@isp.example", agent.RoleManager, nil)

	rec := addField(t, router, cookie, `{"module":"accounts","name":"churnRisk","type":"string","label":"Churn risk"}`)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d (only admins add fields)", rec.Code, http.StatusForbidden)
	}
}

func TestCustomFieldAppearsInMetadataAndPersistsThroughEdit(t *testing.T) {
	db, router := newTestRouter(t)
	admin := loginAs(t, db, router, "admin@isp.example", agent.RoleAdmin, nil)
	c := customer.Customer{Name: "Globex", AccountNumber: "G-1", Status: customer.StatusActive}
	db.Create(&c)

	// Admin defines a custom field on accounts.
	if rec := addField(t, router, admin, `{"module":"accounts","name":"churnRisk","type":"enum","label":"Churn risk","options":["low","high"]}`); rec.Code != http.StatusCreated {
		t.Fatalf("addField status = %d; body=%s", rec.Code, rec.Body.String())
	}

	// It now appears in the accounts metadata, flagged custom.
	metaRec := authGet(t, router, "/metadata/accounts", admin)
	var m metadata.ModuleMeta
	json.Unmarshal(metaRec.Body.Bytes(), &m)
	var found *metadata.Field
	for i := range m.Fields {
		if m.Fields[i].Name == "churnRisk" {
			found = &m.Fields[i]
		}
	}
	if found == nil || !found.Custom || found.Type != metadata.FieldEnum {
		t.Fatalf("churnRisk not merged as a custom enum field: %+v", m.Fields)
	}

	// Editing the account with the custom value persists it...
	idStr := strconv.FormatUint(uint64(c.ID), 10)
	putReq := httptest.NewRequest(http.MethodPut, "/accounts/"+idStr,
		strings.NewReader(`{"name":"Globex","email":"g@x.example","accountNumber":"G-1","status":"active","churnRisk":"high"}`))
	putReq.Header.Set("Content-Type", "application/json")
	putReq.Header.Set("Cookie", admin)
	putRec := httptest.NewRecorder()
	router.ServeHTTP(putRec, putReq)
	if putRec.Code != http.StatusOK {
		t.Fatalf("PUT status = %d; body=%s", putRec.Code, putRec.Body.String())
	}

	// ...and is flattened back onto the record on read.
	getRec := authGet(t, router, "/accounts/"+idStr, admin)
	var got map[string]any
	json.Unmarshal(getRec.Body.Bytes(), &got)
	if got["churnRisk"] != "high" {
		t.Errorf("account = %+v, want churnRisk=high flattened onto the record", got)
	}
}
