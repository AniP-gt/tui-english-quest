package ui

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

const hpTickDuration = 20 * time.Millisecond

// hpTickMsg signals that the HP animation should advance.
type hpTickMsg struct{}

// hpTickCmd returns a command that fires a tick message after a short delay.
func hpTickCmd() tea.Cmd {
	return tea.Tick(hpTickDuration, func(time.Time) tea.Msg {
		return hpTickMsg{}
	})
}

// HPAnimator keeps track of the display HP and whether an animation is active.
type HPAnimator struct {
	display   int
	animating bool
}

// NewHPAnimator initializes an HPAnimator with a starting value.
func NewHPAnimator(initialHP int) HPAnimator {
	return HPAnimator{display: initialHP}
}

// Display returns the current value that should be rendered.
func (a *HPAnimator) Display() int {
	return a.display
}

// Sync sets the display HP to the target when no animation is running.
func (a *HPAnimator) Sync(target int) {
	if !a.animating {
		a.display = target
	}
}

// StartAnimation begins animating from the previous value down to the target HP.
func (a *HPAnimator) StartAnimation(prev, target int) tea.Cmd {
	start := prev
	if a.display > start {
		start = a.display
	}
	if target >= start {
		a.display = target
		a.animating = false
		return nil
	}
	a.display = start
	a.animating = true
	return hpTickCmd()
}

// Tick advances the animation toward the target HP.
func (a *HPAnimator) Tick(target int) tea.Cmd {
	if !a.animating {
		a.display = target
		return nil
	}
	newDisplay, still := stepDownHP(a.display, target)
	a.display = newDisplay
	if still {
		return hpTickCmd()
	}
	a.animating = false
	return nil
}

func stepDownHP(current, target int) (int, bool) {
	if current <= target {
		return target, false
	}
	diff := current - target
	step := diff/2 + 1
	current -= step
	if current < target {
		current = target
	}
	return current, current > target
}
