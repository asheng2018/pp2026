package state

import (
	"fmt"

	"github.com/ab-payment-system/internal/order/model"
)

type StateMachine struct{}

func NewStateMachine() *StateMachine { return &StateMachine{} }

func (sm *StateMachine) Transition(order *model.Order, next model.OrderStatus, data map[string]interface{}) error {
	if !order.Status.CanTransitionTo(next) {
		return fmt.Errorf("invalid transition: %s -> %s", order.Status, next)
	}
	order.Status = next
	return nil
}

func (sm *StateMachine) Pay(order *model.Order, gatewayOrderID string) error {
	return sm.Transition(order, model.StatusPaid, map[string]interface{}{
		"gateway_order_id": gatewayOrderID,
	})
}

func (sm *StateMachine) Fail(order *model.Order) error {
	return sm.Transition(order, model.StatusFailed, nil)
}

func (sm *StateMachine) Cancel(order *model.Order) error {
	return sm.Transition(order, model.StatusCanceled, nil)
}

func (sm *StateMachine) Expire(order *model.Order) error {
	return sm.Transition(order, model.StatusExpired, nil)
}

func (sm *StateMachine) RequestRefund(order *model.Order) error {
	return sm.Transition(order, model.StatusRefunding, nil)
}

func (sm *StateMachine) Refund(order *model.Order, partial bool) error {
	if partial {
		return sm.Transition(order, model.StatusPartiallyRefunded, nil)
	}
	return sm.Transition(order, model.StatusRefunded, nil)
}

func (sm *StateMachine) ResolveDispute(order *model.Order, won bool) error {
	if won {
		return sm.Transition(order, model.StatusDisputeWon, nil)
	}
	return sm.Transition(order, model.StatusDisputeLost, nil)
}
