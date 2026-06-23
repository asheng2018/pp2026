package alert

import (
	"context"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
)

type AlertLevel string

const (
	AlertInfo     AlertLevel = "info"
	AlertWarning  AlertLevel = "warning"
	AlertCritical AlertLevel = "critical"
)

type Alert struct {
	ID           string     `json:"id"`
	Level        AlertLevel `json:"level"`
	Title        string     `json:"title"`
	Message      string     `json:"message"`
	ResourceType string     `json:"resource_type"`
	ResourceID   string     `json:"resource_id"`
	Timestamp    time.Time  `json:"timestamp"`
	Acknowledged bool       `json:"acknowledged"`
}

type AlertChannel interface {
	Send(ctx context.Context, alert *Alert) error
	Name() string
}

type AlertManager struct {
	mu           sync.RWMutex
	channels     []AlertChannel
	alerts       []*Alert
	cooldowns    map[string]time.Time
	silenceUntil map[string]time.Time
}

func NewAlertManager() *AlertManager {
	return &AlertManager{
		channels:     make([]AlertChannel, 0),
		alerts:       make([]*Alert, 0),
		cooldowns:    make(map[string]time.Time),
		silenceUntil: make(map[string]time.Time),
	}
}

func (am *AlertManager) RegisterChannel(ch AlertChannel) {
	am.mu.Lock()
	defer am.mu.Unlock()
	am.channels = append(am.channels, ch)
	log.Info().Str("channel", ch.Name()).Msg("alert channel registered")
}

func (am *AlertManager) Fire(ctx context.Context, level AlertLevel, title, message, resourceType, resourceID string) error {
	am.mu.Lock()

	// Check cooldown (prevent alert storms)
	key := title + ":" + resourceType + ":" + resourceID
	if cooldownUntil, ok := am.cooldowns[key]; ok && time.Now().Before(cooldownUntil) {
		am.mu.Unlock()
		return nil
	}
	am.cooldowns[key] = time.Now().Add(5 * time.Minute)

	// Check silence
	if silenceUntil, ok := am.silenceUntil[resourceType]; ok && time.Now().Before(silenceUntil) {
		am.mu.Unlock()
		return nil
	}

	alert := &Alert{
		ID:           generateAlertID(),
		Level:        level,
		Title:        title,
		Message:      message,
		ResourceType: resourceType,
		ResourceID:   resourceID,
		Timestamp:    time.Now(),
	}
	am.alerts = append(am.alerts, alert)
	channels := am.channels
	am.mu.Unlock()

	log.Warn().
		Str("level", string(level)).
		Str("title", title).
		Str("resource", resourceID).
		Msg("alert fired")

	for _, ch := range channels {
		go func(ch AlertChannel) {
			if err := ch.Send(ctx, alert); err != nil {
				log.Error().Err(err).Str("channel", ch.Name()).Msg("failed to send alert")
			}
		}(ch)
	}

	return nil
}

func (am *AlertManager) Silence(resourceType string, duration time.Duration) {
	am.mu.Lock()
	defer am.mu.Unlock()
	am.silenceUntil[resourceType] = time.Now().Add(duration)
}

func (am *AlertManager) Acknowledge(alertID string) {
	am.mu.Lock()
	defer am.mu.Unlock()
	for _, a := range am.alerts {
		if a.ID == alertID {
			a.Acknowledged = true
			return
		}
	}
}

func (am *AlertManager) GetActiveAlerts() []*Alert {
	am.mu.RLock()
	defer am.mu.RUnlock()
	var active []*Alert
	for _, a := range am.alerts {
		if !a.Acknowledged && time.Since(a.Timestamp) < 24*time.Hour {
			active = append(active, a)
		}
	}
	return active
}

func generateAlertID() string {
	return "alert_" + time.Now().Format("20060102150405")
}
