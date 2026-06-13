package server_test

import (
	"encoding/json"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/techthos/microapp-crm/internal/models"
)

// TestEveryToolHappyPath exercises each remaining CRUD tool once so the full
// registered surface is covered end-to-end through the in-process client.
func TestEveryToolHappyPath(t *testing.T) {
	t.Parallel()
	c, ctx := setup(t)

	// --- Leads: create, list, update, delete ---
	res := callTool(t, c, ctx, "create_lead", map[string]any{"name": "L1", "source": "web"})
	var lead models.Lead
	mustJSON(t, res, &lead)

	if r := callTool(t, c, ctx, "list_leads", map[string]any{}); r.IsError {
		t.Fatalf("list_leads: %s", resultText(t, r))
	}
	if r := callTool(t, c, ctx, "list_leads", map[string]any{"status": "new"}); r.IsError {
		t.Fatalf("list_leads(new): %s", resultText(t, r))
	}

	upd := callTool(t, c, ctx, "update_lead", map[string]any{
		"id": lead.ID, "name": "L1b", "status": "contacted",
	})
	var updatedLead models.Lead
	mustJSON(t, upd, &updatedLead)
	if updatedLead.Status != models.StatusContacted || updatedLead.Name != "L1b" {
		t.Errorf("update_lead result: %+v", updatedLead)
	}

	if r := callTool(t, c, ctx, "delete_lead", map[string]any{"id": lead.ID}); r.IsError {
		t.Fatalf("delete_lead: %s", resultText(t, r))
	}

	// --- Contacts: create, list, get, update ---
	cres := callTool(t, c, ctx, "create_contact", map[string]any{"name": "C1", "email": "c1@x.io", "tags": []string{"vip"}})
	var contact models.Contact
	mustJSON(t, cres, &contact)

	for _, args := range []map[string]any{
		{}, {"query": "C1"}, {"email": "c1@x.io"}, {"tag": "vip"},
	} {
		if r := callTool(t, c, ctx, "list_contacts", args); r.IsError {
			t.Fatalf("list_contacts(%v): %s", args, resultText(t, r))
		}
	}
	if r := callTool(t, c, ctx, "get_contact", map[string]any{"id": contact.ID}); r.IsError {
		t.Fatalf("get_contact: %s", resultText(t, r))
	}
	if r := callTool(t, c, ctx, "update_contact", map[string]any{"id": contact.ID, "name": "C1b"}); r.IsError {
		t.Fatalf("update_contact: %s", resultText(t, r))
	}

	// --- Deals: create, list, get, update, delete ---
	dres := callTool(t, c, ctx, "create_deal", map[string]any{
		"title": "D1", "contact_id": contact.ID, "value": 1000.0, "currency": "EUR", "stage": "qualification",
	})
	var deal models.Deal
	mustJSON(t, dres, &deal)

	for _, args := range []map[string]any{
		{}, {"stage": "qualification"}, {"contact_id": contact.ID},
	} {
		if r := callTool(t, c, ctx, "list_deals", args); r.IsError {
			t.Fatalf("list_deals(%v): %s", args, resultText(t, r))
		}
	}
	if r := callTool(t, c, ctx, "get_deal", map[string]any{"id": deal.ID}); r.IsError {
		t.Fatalf("get_deal: %s", resultText(t, r))
	}
	updDeal := callTool(t, c, ctx, "update_deal", map[string]any{
		"id": deal.ID, "title": "D1", "contact_id": contact.ID, "value": 2000.0, "currency": "EUR", "stage": "won",
	})
	var updatedDeal models.Deal
	mustJSON(t, updDeal, &updatedDeal)
	if updatedDeal.Stage != models.StageWon {
		t.Errorf("update_deal stage = %q, want won", updatedDeal.Stage)
	}
	if r := callTool(t, c, ctx, "delete_deal", map[string]any{"id": deal.ID}); r.IsError {
		t.Fatalf("delete_deal: %s", resultText(t, r))
	}

	// Lead resource read path (the other resource handlers are covered elsewhere).
	lead2 := callTool(t, c, ctx, "create_lead", map[string]any{"name": "L2"})
	var l2 models.Lead
	mustJSON(t, lead2, &l2)
	if r := callTool(t, c, ctx, "get_lead", map[string]any{"id": l2.ID}); r.IsError {
		t.Fatalf("get_lead: %s", resultText(t, r))
	}
}

// mustJSON asserts a successful tool result and unmarshals its JSON text into v.
func mustJSON(t *testing.T, res *mcp.CallToolResult, v any) {
	t.Helper()
	if res.IsError {
		t.Fatalf("tool returned error: %s", resultText(t, res))
	}
	if err := json.Unmarshal([]byte(resultText(t, res)), v); err != nil {
		t.Fatalf("unmarshal tool result: %v", err)
	}
}
