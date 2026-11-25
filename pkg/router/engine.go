package router

import (
	"context"
	"errors"
	"sync"
	"time"
)

var (
	ErrNoAvailableAgent = errors.New("no available agent matching skill criteria")
	ErrTicketNotFound   = errors.New("ticket not found")
)

type Router struct {
	mu           sync.RWMutex
	agents       map[string]*Agent
	activeQueue  map[string]*Ticket
	escalations  []string
}

func NewRouter(agents []*Agent) *Router {
	agentMap := make(map[string]*Agent, len(agents))
	for _, a := range agents {
		agentMap[a.ID] = a
	}
	return &Router{
		agents:      agentMap,
		activeQueue: make(map[string]*Ticket),
	}
}

// CalculateSLA calculates the SLA deadline based on urgency and VIP tiers.
func CalculateSLA(priority Priority, customerTier string, createdAt time.Time) time.Time {
	var duration time.Duration
	switch priority {
	case PriorityP0:
		duration = 15 * time.Minute
	case PriorityP1:
		duration = 60 * time.Minute
	default:
		duration = 240 * time.Minute
	}

	// VIP tier customers receive a 50% accelerated SLA
	if customerTier == "VIP" || customerTier == "Enterprise" {
		duration = duration / 2
	}

	return createdAt.Add(duration)
}

// RouteTicket matches an incoming ticket with the least loaded agent carrying the skill.
func (r *Router) RouteTicket(ctx context.Context, ticket *Ticket) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if ticket.SLADeadline.IsZero() {
		ticket.SLADeadline = CalculateSLA(ticket.Priority, ticket.CustomerTier, ticket.CreatedAt)
	}

	var bestAgent *Agent
	minLoad := 999999

	for _, a := range r.agents {
		if !a.Available {
			continue
		}
		// Check skill match
		matchesSkill := false
		for _, s := range a.Skills {
			if s == ticket.Category || s == "all" {
				matchesSkill = true
				break
			}
		}
		if matchesSkill && a.ActiveLoad < minLoad {
			bestAgent = a
			minLoad = a.ActiveLoad
		}
	}

	if bestAgent == nil {
		r.activeQueue[ticket.ID] = ticket
		return ErrNoAvailableAgent
	}

	bestAgent.ActiveLoad++
	ticket.AssignedTo = bestAgent.ID
	r.activeQueue[ticket.ID] = ticket
	return nil
}

// ResolveTicket marks a ticket complete and frees agent capacity.
func (r *Router) ResolveTicket(ticketID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	ticket, exists := r.activeQueue[ticketID]
	if !exists {
		return ErrTicketNotFound
	}

	if ticket.AssignedTo != "" {
		if agent, ok := r.agents[ticket.AssignedTo]; ok && agent.ActiveLoad > 0 {
			agent.ActiveLoad--
		}
	}
	delete(r.activeQueue, ticketID)
	return nil
}

// CheckSLABreaches returns all tickets currently past deadline.
func (r *Router) CheckSLABreaches(currentTime time.Time) []*Ticket {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var breached []*Ticket
	for _, t := range r.activeQueue {
		if currentTime.After(t.SLADeadline) {
			breached = append(breached, t)
		}
	}
	return breached
}
