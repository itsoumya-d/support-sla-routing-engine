package router

import (
	"testing"
	"time"
)

func TestTypesScaffold(t *testing.T) {
	ticket := Ticket{
		ID:        "t-01",
		Priority:  PriorityP0,
		CreatedAt: time.Now(),
	}
	if ticket.ID != "t-01" || ticket.Priority != PriorityP0 {
		t.Fatalf("unexpected ticket initialization: %+v", ticket)
	}
}
