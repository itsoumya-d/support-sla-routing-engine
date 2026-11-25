package router

import "time"

type Priority string

const (
	PriorityP0 Priority = "P0" // Critical: 15 min SLA
	PriorityP1 Priority = "P1" // High: 1 hour SLA
	PriorityP2 Priority = "P2" // Standard: 4 hours SLA
)

type Ticket struct {
	ID          string    `json:"id"`
	CustomerTier string   `json:"customer_tier"`
	Priority    Priority  `json:"priority"`
	Category    string    `json:"category"`
	CreatedAt   time.Time `json:"created_at"`
	SLADeadline time.Time `json:"sla_deadline"`
	AssignedTo  string    `json:"assigned_to,omitempty"`
}

type Agent struct {
	ID        string   `json:"id"`
	Skills    []string `json:"skills"`
	Available bool     `json:"available"`
	ActiveLoad int     `json:"active_load"`
}
