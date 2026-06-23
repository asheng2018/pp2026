package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/ab-payment-system/internal/order/model"
	"github.com/ab-payment-system/internal/order/state"
)

type OrderRepository interface {
	Create(ctx context.Context, o *model.Order) error
	FindByID(ctx context.Context, id string) (*model.Order, error)
	FindByOrderNo(ctx context.Context, orderNo string) (*model.Order, error)
	UpdateStatus(ctx context.Context, id string, status model.OrderStatus, extra map[string]interface{}) error
	ListByMerchant(ctx context.Context, merchantID string, status string, offset, limit int) ([]*model.Order, error)
	CountToday(ctx context.Context, merchantID string) (int, error)
	SumToday(ctx context.Context, accountID string) (decimal.Decimal, error)
}

type OrderService struct {
	repo     OrderRepository
	sm       *state.StateMachine
	orderTTL time.Duration
}

func NewOrderService(repo OrderRepository, orderTTL time.Duration) *OrderService {
	if orderTTL <= 0 {
		orderTTL = 10 * time.Minute
	}
	return &OrderService{
		repo:     repo,
		sm:       state.NewStateMachine(),
		orderTTL: orderTTL,
	}
}

func (s *OrderService) Create(ctx context.Context, merchantID, amount, currency, customerEmail, customerIP, customerCountry, gateway, aSiteReferer string, metadata map[string]interface{}) (*model.Order, error) {
	amt, err := decimal.NewFromString(amount)
	if err != nil {
		return nil, fmt.Errorf("invalid amount: %w", err)
	}

	order := &model.Order{
		OrderNo:         generateOrderNo(merchantID),
		MerchantID:      merchantID,
		Amount:          amt,
		Currency:        currency,
		Status:          model.StatusPending,
		CustomerEmail:   customerEmail,
		CustomerIP:      customerIP,
		CustomerCountry: customerCountry,
		Gateway:         gateway,
		ASiteReferer:    aSiteReferer,
		RiskLevel:       "low",
		RiskScore:       0,
		Metadata:        metadata,
		CallbackData:    make(map[string]interface{}),
		ExpiredAt:       timePtr(time.Now().Add(s.orderTTL)),
	}

	if err := s.repo.Create(ctx, order); err != nil {
		return nil, fmt.Errorf("create order: %w", err)
	}

	return order, nil
}

func (s *OrderService) Get(ctx context.Context, id string) (*model.Order, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *OrderService) GetByOrderNo(ctx context.Context, orderNo string) (*model.Order, error) {
	return s.repo.FindByOrderNo(ctx, orderNo)
}

func (s *OrderService) Pay(ctx context.Context, id string, gatewayOrderID string) error {
	order, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if err := s.sm.Pay(order, gatewayOrderID); err != nil {
		return err
	}
	return s.repo.UpdateStatus(ctx, id, order.Status, map[string]interface{}{
		"gateway_order_id": gatewayOrderID,
	})
}

func (s *OrderService) Fail(ctx context.Context, id string) error {
	order, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if err := s.sm.Fail(order); err != nil {
		return err
	}
	return s.repo.UpdateStatus(ctx, id, order.Status, nil)
}

func (s *OrderService) Cancel(ctx context.Context, id string) error {
	order, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if err := s.sm.Cancel(order); err != nil {
		return err
	}
	return s.repo.UpdateStatus(ctx, id, order.Status, nil)
}

func (s *OrderService) Expire(ctx context.Context, id string) error {
	order, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if err := s.sm.Expire(order); err != nil {
		return err
	}
	return s.repo.UpdateStatus(ctx, id, order.Status, nil)
}

func (s *OrderService) Refund(ctx context.Context, id string, partial bool) error {
	order, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if err := s.sm.Refund(order, partial); err != nil {
		return err
	}
	return s.repo.UpdateStatus(ctx, id, order.Status, nil)
}

func (s *OrderService) ListByMerchant(ctx context.Context, merchantID, status string, offset, limit int) ([]*model.Order, error) {
	return s.repo.ListByMerchant(ctx, merchantID, status, offset, limit)
}

func (s *OrderService) SetAccount(ctx context.Context, orderID string, accountID string, payToken string) error {
	hash := sha256.Sum256([]byte(payToken))
	// Update order with account assignment
	order, err := s.repo.FindByID(ctx, orderID)
	if err != nil {
		return err
	}
	order.AccountID = accountID
	order.PayTokenHash = hex.EncodeToString(hash[:])
	order.Status = model.StatusProcessing
	return s.repo.UpdateStatus(ctx, orderID, order.Status, map[string]interface{}{})
}

func generateOrderNo(merchantID string) string {
	prefix := merchantID
	if len(prefix) > 6 {
		prefix = prefix[:6]
	}
	return fmt.Sprintf("%s_%s_%s", prefix, time.Now().Format("20060102"), uuid.New().String()[:8])
}

func timePtr(t time.Time) *time.Time { return &t }
