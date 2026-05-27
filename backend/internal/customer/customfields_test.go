package customer

import (
	"encoding/json"
	"testing"
)

func TestCustomFieldsFlattenOnMarshal(t *testing.T) {
	c := Customer{Name: "Ada", Status: StatusActive, CustomFields: map[string]any{"churnRisk": "high"}}
	b, err := json.Marshal(c)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var m map[string]any
	json.Unmarshal(b, &m)
	if m["name"] != "Ada" || m["churnRisk"] != "high" {
		t.Errorf("marshaled = %s, want churnRisk flattened to top level", b)
	}
}

func TestUnknownKeysCapturedAsCustomFields(t *testing.T) {
	var c Customer
	if err := json.Unmarshal([]byte(`{"name":"Ada","status":"active","churnRisk":"low","nps":9}`), &c); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if c.Name != "Ada" {
		t.Errorf("name = %q, want Ada", c.Name)
	}
	if c.CustomFields["churnRisk"] != "low" || c.CustomFields["nps"].(float64) != 9 {
		t.Errorf("customFields = %+v, want churnRisk=low nps=9", c.CustomFields)
	}
}
