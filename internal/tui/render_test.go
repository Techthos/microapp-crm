package tui

import (
	"strings"
	"testing"

	"github.com/techthos/microapp-crm/internal/models"
)

func TestLeadRows(t *testing.T) {
	t.Parallel()
	leads := []models.Lead{
		{ID: 1, Name: "Jane", Company: "Acme", Email: "j@x.io", Status: models.StatusNew, Source: models.SourceWeb},
	}
	rows := leadRows(leads)
	if len(rows) != 2 {
		t.Fatalf("rows = %d, want 2 (header + 1)", len(rows))
	}
	if rows[0][0] != "ID" || rows[0][4] != "Status" {
		t.Errorf("unexpected header: %v", rows[0])
	}
	if rows[1][0] != "1" || rows[1][1] != "Jane" || rows[1][4] != "new" {
		t.Errorf("unexpected row: %v", rows[1])
	}
}

func TestContactAndDealRows(t *testing.T) {
	t.Parallel()
	crows := contactRows([]models.Contact{{ID: 2, Name: "Ada", Phone: "555"}})
	if crows[1][0] != "2" || crows[1][1] != "Ada" || crows[1][4] != "555" {
		t.Errorf("contact row: %v", crows[1])
	}
	drows := dealRows([]models.Deal{{ID: 3, Title: "Big", ContactID: 2, Value: 1500, Currency: "EUR", Stage: models.StageWon}})
	if drows[1][0] != "3" || drows[1][3] != "1500.00" || drows[1][5] != "won" {
		t.Errorf("deal row: %v", drows[1])
	}
}

func TestSummaryLinesGroupsByCurrency(t *testing.T) {
	t.Parallel()
	s := models.PipelineSummary{
		DealsByStage: []models.StageSummary{
			{Stage: models.StageQualification, Count: 2, Totals: []models.CurrencyTotal{
				{Currency: "EUR", Total: 2500}, {Currency: "USD", Total: 3000},
			}},
			{Stage: models.StageWon, Count: 0},
		},
		LeadsByStatus: []models.StatusCount{
			{Status: models.StatusNew, Count: 4}, {Status: models.StatusLost, Count: 1},
		},
	}
	out := strings.Join(summaryLines(s), "\n")
	for _, want := range []string{"LEADS", "new:4", "lost:1", "qualification", "EUR 2500.00", "USD 3000.00"} {
		if !strings.Contains(out, want) {
			t.Errorf("summary missing %q in:\n%s", want, out)
		}
	}
	// EUR and USD must appear separately, not summed.
	if strings.Contains(out, "5500") {
		t.Errorf("currencies were summed across:\n%s", out)
	}
}

func TestOptionHelpers(t *testing.T) {
	t.Parallel()
	if got := sourceOptions(); got[0] != "" || len(got) != 6 {
		t.Errorf("sourceOptions = %v, want blank + 5", got)
	}
	if got := statusOptions(); len(got) != 5 {
		t.Errorf("statusOptions = %d, want 5", len(got))
	}
	if got := indexOf(stageOptions(), "won"); got != 3 {
		t.Errorf("indexOf(won) = %d, want 3", got)
	}
	if got := indexOf([]string{"a", "b"}, "missing"); got != 0 {
		t.Errorf("indexOf(missing) = %d, want 0", got)
	}
}

func TestFormatMoneyAndParse(t *testing.T) {
	t.Parallel()
	if got := formatMoney(1500); got != "1500.00" {
		t.Errorf("formatMoney = %q", got)
	}
	if got := parseFloat(" 12.5 "); got != 12.5 {
		t.Errorf("parseFloat = %v", got)
	}
	if got := parseFloat("nope"); got != 0 {
		t.Errorf("parseFloat(invalid) = %v, want 0", got)
	}
	if got := parseUint("42"); got != 42 {
		t.Errorf("parseUint = %v", got)
	}
}
