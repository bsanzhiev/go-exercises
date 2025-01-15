package events

import (
	"testing"
)

func TestEventProcessor(t *testing.T) {
	ep := NewEventProcessor(2)

	ep.Submit(Event{ID: 1, Priority: HighPriority})
	ep.Submit(Event{ID: 2, Priority: NormalPriority})
	ep.Stop()

	// ...verify results as needed...
}
