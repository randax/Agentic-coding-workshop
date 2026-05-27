package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"saltcrm/internal/agent"
	"saltcrm/internal/contact"
	"saltcrm/internal/customer"
	"saltcrm/internal/lead"

	"gorm.io/gorm"
)

// agentByEmail loads the seeded login user so a fixture lead can be owned by them
// (and therefore visible under own-or-team scoping).
func agentByEmail(t *testing.T, db *gorm.DB, email string) agent.Agent {
	t.Helper()
	var a agent.Agent
	if err := db.Where("email = ?", email).First(&a).Error; err != nil {
		t.Fatalf("load agent %s: %v", email, err)
	}
	return a
}

func convert(t *testing.T, router http.Handler, leadID uint, cookie string) *httptest.ResponseRecorder {
	t.Helper()
	return authPost(t, router, "/leads/"+strconv.FormatUint(uint64(leadID), 10)+"/convert", cookie, `{}`)
}

func TestConvertLeadCreatesAccountAndContact(t *testing.T) {
	db, router := newTestRouter(t)
	team := uint(1)
	cookie := loginAs(t, db, router, "manager@isp.example", agent.RoleManager, &team)
	me := agentByEmail(t, db, "manager@isp.example")
	l := lead.Lead{
		Name: "Sofia Berg", Email: "sofia@polarfoods.example", Phone: "+47 903 55 666",
		Company: "Polar Foods", Status: lead.StatusQualified, AssignedUserID: &me.ID, TeamID: &team,
	}
	db.Create(&l)

	rec := convert(t, router, l.ID, cookie)
	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusCreated, rec.Body.String())
	}
	var resp struct {
		AccountID uint `json:"accountId"`
		ContactID uint `json:"contactId"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v; body=%s", err, rec.Body.String())
	}
	if resp.AccountID == 0 || resp.ContactID == 0 {
		t.Fatalf("response = %+v, want non-zero account and contact ids", resp)
	}

	// Account: company → name; email/phone/owner/team carried; auto ISP-#### number.
	var acc customer.Customer
	db.First(&acc, resp.AccountID)
	if acc.Name != "Polar Foods" || acc.Email != "sofia@polarfoods.example" || acc.Phone != "+47 903 55 666" {
		t.Errorf("account = %+v, want company/email/phone carried from the lead", acc)
	}
	if !strings.HasPrefix(acc.AccountNumber, "ISP-") || acc.AccountNumber == "ISP-" {
		t.Errorf("account number = %q, want an auto-minted ISP-#### value", acc.AccountNumber)
	}
	if acc.AssignedUserID == nil || *acc.AssignedUserID != me.ID || acc.TeamID == nil || *acc.TeamID != team {
		t.Errorf("account owner/team not carried: %+v", acc)
	}

	// Contact: linked to the new account; lead identity carried.
	var con contact.Contact
	db.First(&con, resp.ContactID)
	if con.AccountID != resp.AccountID || con.Name != "Sofia Berg" || con.Email != "sofia@polarfoods.example" {
		t.Errorf("contact = %+v, want it linked to account %d with the lead's identity", con, resp.AccountID)
	}

	// Lead: terminal converted status + link to the account it became.
	var reloaded lead.Lead
	db.First(&reloaded, l.ID)
	if reloaded.Status != lead.StatusConverted {
		t.Errorf("lead status = %q, want %q", reloaded.Status, lead.StatusConverted)
	}
	if reloaded.ConvertedAccountID == nil || *reloaded.ConvertedAccountID != resp.AccountID {
		t.Errorf("lead.ConvertedAccountID = %v, want %d", reloaded.ConvertedAccountID, resp.AccountID)
	}
}

func TestConvertLeadAllowedForAgentRole(t *testing.T) {
	db, router := newTestRouter(t)
	team := uint(1)
	cookie := loginAs(t, db, router, "agent@isp.example", agent.RoleAgent, &team)
	me := agentByEmail(t, db, "agent@isp.example")
	l := lead.Lead{Name: "Lee Prospect", Company: "Acme", Status: lead.StatusQualified, AssignedUserID: &me.ID, TeamID: &team}
	db.Create(&l)

	rec := convert(t, router, l.ID, cookie)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d (agents may convert their own leads); body=%s", rec.Code, http.StatusCreated, rec.Body.String())
	}
}

func TestConvertNonQualifiedLeadReturns409(t *testing.T) {
	db, router := newTestRouter(t)
	team := uint(1)
	cookie := loginAs(t, db, router, "manager@isp.example", agent.RoleManager, &team)
	me := agentByEmail(t, db, "manager@isp.example")
	l := lead.Lead{Name: "Raw", Company: "Acme", Status: lead.StatusWorking, AssignedUserID: &me.ID, TeamID: &team}
	db.Create(&l)

	rec := convert(t, router, l.ID, cookie)

	if rec.Code != http.StatusConflict {
		t.Fatalf("status = %d, want %d (only qualified leads convert); body=%s", rec.Code, http.StatusConflict, rec.Body.String())
	}
}

func TestConvertAlreadyConvertedLeadReturns409(t *testing.T) {
	db, router := newTestRouter(t)
	team := uint(1)
	cookie := loginAs(t, db, router, "manager@isp.example", agent.RoleManager, &team)
	me := agentByEmail(t, db, "manager@isp.example")
	already := uint(123)
	l := lead.Lead{Name: "Done", Company: "Acme", Status: lead.StatusQualified, AssignedUserID: &me.ID, TeamID: &team, ConvertedAccountID: &already}
	db.Create(&l)

	rec := convert(t, router, l.ID, cookie)

	if rec.Code != http.StatusConflict {
		t.Fatalf("status = %d, want %d (cannot convert twice); body=%s", rec.Code, http.StatusConflict, rec.Body.String())
	}
}

func TestConvertLeadOutsideScopeReturns404(t *testing.T) {
	db, router := newTestRouter(t)
	teamA := uint(1)
	cookie := loginAs(t, db, router, "agent@isp.example", agent.RoleAgent, &teamA)
	// A qualified lead owned by another user on another team — invisible.
	otherUser, otherTeam := uint(999), uint(2)
	l := lead.Lead{Name: "Hidden", Company: "Acme", Status: lead.StatusQualified, AssignedUserID: &otherUser, TeamID: &otherTeam}
	db.Create(&l)

	rec := convert(t, router, l.ID, cookie)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d (a lead the rep can't see is not-found); body=%s", rec.Code, http.StatusNotFound, rec.Body.String())
	}
}

func TestConvertLeadRequiresAuth(t *testing.T) {
	db, router := newTestRouter(t)
	l := lead.Lead{Name: "Sofia", Company: "Polar Foods", Status: lead.StatusQualified}
	db.Create(&l)

	req := httptest.NewRequest(http.MethodPost, "/leads/"+strconv.FormatUint(uint64(l.ID), 10)+"/convert", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d (auth required); body=%s", rec.Code, http.StatusUnauthorized, rec.Body.String())
	}
}

// TestConvertIsAtomicOnFailure forces the contact insert to fail mid-conversion
// (by removing the contacts table) and asserts nothing is left behind: no orphan
// account, and the lead is untouched.
func TestConvertIsAtomicOnFailure(t *testing.T) {
	db, router := newTestRouter(t)
	team := uint(1)
	cookie := loginAs(t, db, router, "manager@isp.example", agent.RoleManager, &team)
	me := agentByEmail(t, db, "manager@isp.example")
	l := lead.Lead{Name: "Sofia Berg", Company: "Polar Foods", Status: lead.StatusQualified, AssignedUserID: &me.ID, TeamID: &team}
	db.Create(&l)

	if err := db.Migrator().DropTable(&contact.Contact{}); err != nil {
		t.Fatalf("drop contacts table: %v", err)
	}

	rec := convert(t, router, l.ID, cookie)
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d (conversion should fail); body=%s", rec.Code, http.StatusInternalServerError, rec.Body.String())
	}

	// No orphan account survived the rolled-back transaction.
	var accounts int64
	db.Model(&customer.Customer{}).Count(&accounts)
	if accounts != 0 {
		t.Errorf("found %d accounts after a failed conversion, want 0 (rollback)", accounts)
	}
	// The lead is unchanged: still qualified, still unconverted.
	var reloaded lead.Lead
	db.First(&reloaded, l.ID)
	if reloaded.Status != lead.StatusQualified || reloaded.ConvertedAccountID != nil {
		t.Errorf("lead = %+v, want it unchanged (qualified, unconverted) after rollback", reloaded)
	}
}
