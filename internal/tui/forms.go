package tui

import (
	"strconv"
	"strings"

	"github.com/rivo/tview"
)

// formText reads an InputField's current text by label.
func formText(form *tview.Form, label string) string {
	return form.GetFormItemByLabel(label).(*tview.InputField).GetText()
}

// formDropdown reads a DropDown's selected option text by label.
func formDropdown(form *tview.Form, label string) string {
	_, opt := form.GetFormItemByLabel(label).(*tview.DropDown).GetCurrentOption()
	return opt
}

// parseFloat parses a money input; a blank or invalid value is 0.
func parseFloat(s string) float64 {
	v, err := strconv.ParseFloat(strings.TrimSpace(s), 64)
	if err != nil {
		return 0
	}
	return v
}
