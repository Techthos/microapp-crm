package tui

import (
	"strings"
	"testing"

	"github.com/gdamore/tcell/v2"
)

// runTUI starts a seeded TUI on a simulation screen and returns it plus a
// teardown that stops the app cleanly.
func runTUI(t *testing.T) (*tui, tcell.SimulationScreen) {
	t.Helper()
	ti := newTUI(seededStore(t))
	ti.loadSync()
	sim := tcell.NewSimulationScreen("UTF-8")
	ti.app.SetScreen(sim)
	sim.SetSize(120, 40)
	runErr := make(chan error, 1)
	go func() { runErr <- ti.app.Run() }()
	t.Cleanup(func() {
		ti.app.Stop()
		<-runErr
	})
	return ti, sim
}

// waitForGone polls until the screen no longer contains text.
func waitForGone(t *testing.T, ti *tui, sim tcell.SimulationScreen, text string) {
	t.Helper()
	var last string
	for i := 0; i < 200; i++ {
		last = snapshot(ti, sim)
		if !strings.Contains(last, text) {
			return
		}
	}
	t.Fatalf("timed out waiting for %q to disappear; screen:\n%s", text, last)
}

func typeRunes(sim tcell.SimulationScreen, s string) {
	for _, r := range s {
		sim.InjectKey(tcell.KeyRune, r, tcell.ModNone)
	}
}

func TestCreateLeadThroughForm(t *testing.T) {
	t.Parallel()
	ti, sim := runTUI(t)

	waitFor(t, ti, sim, "LEADS")
	sim.InjectKey(tcell.KeyF2, 0, tcell.ModNone)
	waitFor(t, ti, sim, "Zara")

	// Open the new-lead form.
	sim.InjectKey(tcell.KeyRune, 'n', tcell.ModNone)
	waitFor(t, ti, sim, "New Lead")

	// The Name field is focused first; type a name, then Tab to the Save button.
	typeRunes(sim, "Newbie")
	// Name -> Company -> Email -> Phone -> Source -> Status -> Notes -> Save
	for i := 0; i < 7; i++ {
		sim.InjectKey(tcell.KeyTab, 0, tcell.ModNone)
	}
	sim.InjectKey(tcell.KeyEnter, 0, tcell.ModNone)

	// The new lead appears in the table and the form is gone.
	waitFor(t, ti, sim, "Newbie")
	waitForGone(t, ti, sim, "New Lead")
}

func TestLeadFormCancel(t *testing.T) {
	t.Parallel()
	ti, sim := runTUI(t)
	waitFor(t, ti, sim, "LEADS")
	sim.InjectKey(tcell.KeyF2, 0, tcell.ModNone)
	waitFor(t, ti, sim, "Zara")

	sim.InjectKey(tcell.KeyRune, 'n', tcell.ModNone)
	waitFor(t, ti, sim, "New Lead")
	sim.InjectKey(tcell.KeyEscape, 0, tcell.ModNone)
	waitForGone(t, ti, sim, "New Lead")
}

func TestChangeDealStageThroughPicker(t *testing.T) {
	t.Parallel()
	ti, sim := runTUI(t)

	waitFor(t, ti, sim, "LEADS")
	sim.InjectKey(tcell.KeyF4, 0, tcell.ModNone)
	waitFor(t, ti, sim, "Megadeal")
	// Seeded deal starts at "proposal".
	waitFor(t, ti, sim, "proposal")

	// Open the stage picker and choose the first button (qualification).
	sim.InjectKey(tcell.KeyRune, 's', tcell.ModNone)
	waitFor(t, ti, sim, "which stage")
	sim.InjectKey(tcell.KeyEnter, 0, tcell.ModNone)

	waitFor(t, ti, sim, "qualification")
}

func TestDeleteContactThroughModal(t *testing.T) {
	t.Parallel()
	ti, sim := runTUI(t)

	waitFor(t, ti, sim, "LEADS")
	sim.InjectKey(tcell.KeyF3, 0, tcell.ModNone)
	waitFor(t, ti, sim, "Quentin")

	// Open the cascade-delete confirmation; the first button is "Delete".
	sim.InjectKey(tcell.KeyRune, 'd', tcell.ModNone)
	waitFor(t, ti, sim, "Delete contact")
	sim.InjectKey(tcell.KeyEnter, 0, tcell.ModNone)

	waitForGone(t, ti, sim, "Quentin")
}
