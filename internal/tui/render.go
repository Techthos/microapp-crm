// Package tui is the terminal UI for microapp-crm, built on rivo/tview. It owns
// a single tview.Application and pulls all data through the db.Store — no bbolt
// access or business logic lives here. See docs/SPECIFICATIONS.md (TUI Surface)
// and .claude/rules/tui-rules.md.
package tui

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/techthos/microapp-crm/internal/models"
)

// Column headers for each list screen.
var (
	leadHeader    = []string{"ID", "Name", "Company", "Email", "Status", "Source"}
	contactHeader = []string{"ID", "Name", "Company", "Email", "Phone"}
	dealHeader    = []string{"ID", "Title", "Contact", "Value", "Cur", "Stage"}
)

// leadRows renders leads as table cells (header row first). Pure: no tview.
func leadRows(leads []models.Lead) [][]string {
	rows := [][]string{leadHeader}
	for _, l := range leads {
		rows = append(rows, []string{
			strconv.FormatUint(l.ID, 10), l.Name, l.Company, l.Email,
			string(l.Status), string(l.Source),
		})
	}
	return rows
}

// contactRows renders contacts as table cells (header row first).
func contactRows(contacts []models.Contact) [][]string {
	rows := [][]string{contactHeader}
	for _, c := range contacts {
		rows = append(rows, []string{
			strconv.FormatUint(c.ID, 10), c.Name, c.Company, c.Email, c.Phone,
		})
	}
	return rows
}

// dealRows renders deals as table cells (header row first).
func dealRows(deals []models.Deal) [][]string {
	rows := [][]string{dealHeader}
	for _, d := range deals {
		rows = append(rows, []string{
			strconv.FormatUint(d.ID, 10), d.Title, strconv.FormatUint(d.ContactID, 10),
			formatMoney(d.Value), d.Currency, string(d.Stage),
		})
	}
	return rows
}

// formatMoney renders a monetary value with two decimals.
func formatMoney(v float64) string {
	return strconv.FormatFloat(v, 'f', 2, 64)
}

// sourceOptions lists lead-source dropdown choices (blank = unset, allowed).
func sourceOptions() []string {
	opts := []string{""}
	for _, s := range models.Sources() {
		opts = append(opts, string(s))
	}
	return opts
}

// statusOptions lists lead-status dropdown choices.
func statusOptions() []string {
	opts := make([]string, 0, len(models.LeadStatuses()))
	for _, s := range models.LeadStatuses() {
		opts = append(opts, string(s))
	}
	return opts
}

// stageOptions lists deal-stage dropdown choices.
func stageOptions() []string {
	opts := make([]string, 0, len(models.DealStages()))
	for _, s := range models.DealStages() {
		opts = append(opts, string(s))
	}
	return opts
}

// indexOf returns the position of val in opts, or 0 if absent.
func indexOf(opts []string, val string) int {
	for i, o := range opts {
		if o == val {
			return i
		}
	}
	return 0
}

// summaryLines renders the pipeline summary as display lines (UC-18). Deal
// values are shown grouped by currency, never summed across currencies.
func summaryLines(s models.PipelineSummary) []string {
	var lines []string

	var funnel strings.Builder
	funnel.WriteString("LEADS  ")
	for i, sc := range s.LeadsByStatus {
		if i > 0 {
			funnel.WriteString("  ")
		}
		fmt.Fprintf(&funnel, "%s:%d", sc.Status, sc.Count)
	}
	lines = append(lines, funnel.String(), "")

	lines = append(lines, "DEALS")
	for _, ss := range s.DealsByStage {
		totals := "—"
		if len(ss.Totals) > 0 {
			parts := make([]string, 0, len(ss.Totals))
			for _, ct := range ss.Totals {
				parts = append(parts, fmt.Sprintf("%s %s", ct.Currency, formatMoney(ct.Total)))
			}
			totals = strings.Join(parts, " / ")
		}
		lines = append(lines, fmt.Sprintf("  %-14s %3d deals   %s", ss.Stage, ss.Count, totals))
	}
	return lines
}
