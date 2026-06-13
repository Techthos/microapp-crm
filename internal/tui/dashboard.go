package tui

import "github.com/rivo/tview"

// newDashboard builds the read-only pipeline-summary screen (UC-18).
func (t *tui) newDashboard() *tview.TextView {
	tv := tview.NewTextView().SetDynamicColors(true)
	tv.SetBorder(true).SetTitle(" Dashboard — pipeline summary ")
	return tv
}
