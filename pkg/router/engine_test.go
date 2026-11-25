package router

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestRouterWorkflow(t *testing.T) {
	agents := []*Agent{
		{ID: "agent_1", Skills: []string{"refunds", "orders"}, Available: true, ActiveLoad: 0},
		{ID: "agent_2", Skills: []string{"technical", "orders"}, Available: true, ActiveLoad: 1},
	}
	r := NewRouter(agents)

	baseTime := time.Date(2026, 11, 25, 12, 0, 0, 0, time.UTC)

	t.Run("CalculatesSLACorrectly", func(t *testing.T) {
		p0SLA := CalculateSLA(PriorityP0, "Standard", baseTime)
		if p0SLA.Sub(baseTime) != 15*time.Minute {
			t.Fatalf("expected 15m, got %v", p0SLA.Sub(baseTime))
		}

		vipSLA := CalculateSLA(PriorityP0, "VIP", baseTime)
		if vipSLA.Sub(baseTime) != 7*time.Minute+30*time.Second {
			t.Fatalf("expected 7.5m, got %v", vipSLA.Sub(baseTime))
		}
	})

	t.Run("RoutesToLeastLoadedAgent", func(t *testing.T) {
		ticket := &Ticket{
			ID:           "t_101",
			Priority:     PriorityP1,
			Category:     "orders",
			CustomerTier: "Standard",
			CreatedAt:    baseTime,
		}

		err := r.RouteTicket(context.Background(), ticket)
		if err != nil {
			t.Fatalf("unexpected routing error: %v", err)
		}
		if ticket.AssignedTo != "agent_1" {
			t.Errorf("expected routing to agent_1 (least loaded), got %s", ticket.AssignedTo)
		}
	})

	t.Run("DetectsBreachedSLA", func(t *testing.T) {
		ticket := &Ticket{
			ID:           "t_breach",
			Priority:     PriorityP0,
			Category:     "refunds",
			CustomerTier: "Standard",
			CreatedAt:    baseTime,
		}
		_ = r.RouteTicket(context.Background(), ticket)

		// 20 minutes later (past 15m SLA)
		breaches := r.CheckSLABreaches(baseTime.Add(20 * time.Minute))
		if len(breaches) == 0 {
			t.Errorf("expected breach detection for overdue ticket")
		}
	})

	t.Run("ConcurrentDispatchSafety", func(t *testing.T) {
		var wg sync.WaitGroup
		for i := 0; i < 30; i++ {
			wg.Add(1)
			go func(idx int) {
				defer wg.Done()
				tk := &Ticket{
					ID:           fmt.Sprintf("t_conc_%d", idx),
					Priority:     PriorityP2,
					Category:     "orders",
					CustomerTier: "Standard",
					CreatedAt:    baseTime,
				}
				_ = r.RouteTicket(context.Background(), tk)
			}(i)
		}
		wg.Wait()
	})
}
