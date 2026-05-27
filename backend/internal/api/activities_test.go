package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"saltcrm/internal/activity"
	"saltcrm/internal/agent"
)

func TestGetActivitiesRequiresAuth(t *testing.T) {
	_, router := newTestRouter(t)

	req := httptest.NewRequest(http.MethodGet, "/activities", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestLogActivityThenListForParent(t *testing.T) {
	db, router := newTestRouter(t)
	cookie := loginAs(t, db, router, "agent@isp.example", agent.RoleAgent, nil)

	body := `{"type":"call","subject":"Intro call","parentType":"account","parentId":42}`
	req := httptest.NewRequest(http.MethodPost, "/activities", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Cookie", cookie)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("log status = %d, want %d; body=%s", rec.Code, http.StatusCreated, rec.Body.String())
	}

	// The record's timeline lists it.
	listRec := authGet(t, router, "/activities?parentType=account&parentId=42", cookie)
	var items []map[string]any
	json.Unmarshal(listRec.Body.Bytes(), &items)
	if len(items) != 1 || items[0]["subject"] != "Intro call" {
		t.Fatalf("timeline = %+v, want the logged call", items)
	}
}

func TestLogActivityInvalidTypeReturns400(t *testing.T) {
	db, router := newTestRouter(t)
	cookie := loginAs(t, db, router, "agent@isp.example", agent.RoleAgent, nil)

	req := httptest.NewRequest(http.MethodPost, "/activities", strings.NewReader(`{"type":"party","subject":"x","parentType":"account","parentId":1}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Cookie", cookie)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestCompleteActivityMarksDone(t *testing.T) {
	db, router := newTestRouter(t)
	cookie := loginAs(t, db, router, "agent@isp.example", agent.RoleAgent, nil)
	a := activity.Activity{Type: activity.TypeTask, Subject: "Follow up", Status: activity.StatusOpen, ParentType: "account", ParentID: 1}
	db.Create(&a)

	req := httptest.NewRequest(http.MethodPost, "/activities/"+strconv.FormatUint(uint64(a.ID), 10)+"/complete", nil)
	req.Header.Set("Cookie", cookie)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}
	var got map[string]any
	json.Unmarshal(rec.Body.Bytes(), &got)
	if got["status"] != "done" {
		t.Errorf("status = %v, want done", got["status"])
	}
}
