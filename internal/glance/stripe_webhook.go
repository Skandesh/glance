package glance

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/stripe/stripe-go/v81"
	"github.com/stripe/stripe-go/v81/webhook"
)

// WebhookHandler handles Stripe webhook events for real-time updates
type WebhookHandler struct {
	secret         string
	eventHandlers  map[string][]EventHandlerFunc
	mu             sync.RWMutex
	eventLog       []WebhookEvent
	maxEventLog    int
	cacheInvalidator CacheInvalidator
}

// EventHandlerFunc is a function that handles a Stripe webhook event
type EventHandlerFunc func(ctx context.Context, event stripe.Event) error

// WebhookEvent represents a processed webhook event
type WebhookEvent struct {
	ID        string    `json:"id"`
	Type      string    `json:"type"`
	Processed time.Time `json:"processed"`
	Success   bool      `json:"success"`
	Error     string    `json:"error,omitempty"`
}

// CacheInvalidator is an interface for invalidating widget caches
type CacheInvalidator interface {
	InvalidateCache(widgetType string) error
}

var (
	globalWebhookHandler *WebhookHandler
	webhookHandlerOnce   sync.Once
)

// GetWebhookHandler returns the global webhook handler (singleton)
func GetWebhookHandler(secret string, invalidator CacheInvalidator) *WebhookHandler {
	webhookHandlerOnce.Do(func() {
		globalWebhookHandler = &WebhookHandler{
			secret:           secret,
			eventHandlers:    make(map[string][]EventHandlerFunc),
			eventLog:         make([]WebhookEvent, 0, 100),
			maxEventLog:      100,
			cacheInvalidator: invalidator,
		}

		// Register default event handlers
		globalWebhookHandler.RegisterHandler("customer.subscription.created", handleSubscriptionCreated)
		globalWebhookHandler.RegisterHandler("customer.subscription.updated", handleSubscriptionUpdated)
		globalWebhookHandler.RegisterHandler("customer.subscription.deleted", handleSubscriptionDeleted)
		globalWebhookHandler.RegisterHandler("customer.created", handleCustomerCreated)
		globalWebhookHandler.RegisterHandler("customer.deleted", handleCustomerDeleted)
		globalWebhookHandler.RegisterHandler("invoice.payment_succeeded", handleInvoicePaymentSucceeded)
		globalWebhookHandler.RegisterHandler("invoice.payment_failed", handleInvoicePaymentFailed)
	})

	return globalWebhookHandler
}

// RegisterHandler registers a handler for a specific event type
func (wh *WebhookHandler) RegisterHandler(eventType string, handler EventHandlerFunc) {
	wh.mu.Lock()
	defer wh.mu.Unlock()

	if wh.eventHandlers[eventType] == nil {
		wh.eventHandlers[eventType] = make([]EventHandlerFunc, 0)
	}

	wh.eventHandlers[eventType] = append(wh.eventHandlers[eventType], handler)
}

// HandleWebhook handles an incoming webhook request
func (wh *WebhookHandler) HandleWebhook(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	payload, err := io.ReadAll(r.Body)
	if err != nil {
		slog.Error("Failed to read webhook body", "error", err)
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}

	// Verify signature
	signature := r.Header.Get("Stripe-Signature")
	event, err := webhook.ConstructEvent(payload, signature, wh.secret)
	if err != nil {
		slog.Error("Failed to verify webhook signature", "error", err)
		http.Error(w, "Invalid signature", http.StatusUnauthorized)
		return
	}

	slog.Info("Received Stripe webhook",
		"event_id", event.ID,
		"event_type", event.Type,
		"livemode", event.Livemode)

	// Process event asynchronously
	go wh.processEvent(event)

	// Respond immediately to Stripe
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"received": true,
		"event_id": event.ID,
	})
}

// processEvent processes a webhook event
func (wh *WebhookHandler) processEvent(event stripe.Event) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	eventTypeStr := string(event.Type)

	webhookEvent := WebhookEvent{
		ID:        event.ID,
		Type:      eventTypeStr,
		Processed: time.Now(),
		Success:   true,
	}

	wh.mu.RLock()
	handlers, exists := wh.eventHandlers[eventTypeStr]
	wh.mu.RUnlock()

	if !exists || len(handlers) == 0 {
		slog.Debug("No handlers registered for event type", "type", eventTypeStr)
		return
	}

	// Execute all handlers for this event type
	for _, handler := range handlers {
		if err := handler(ctx, event); err != nil {
			webhookEvent.Success = false
			webhookEvent.Error = err.Error()
			slog.Error("Webhook handler failed",
				"event_id", event.ID,
				"event_type", eventTypeStr,
				"error", err)
		}
	}

	// Invalidate relevant caches
	if wh.cacheInvalidator != nil {
		if err := wh.invalidateCachesForEvent(eventTypeStr); err != nil {
			slog.Error("Failed to invalidate cache", "event_type", eventTypeStr, "error", err)
		}
	}

	// Log the event
	wh.logEvent(webhookEvent)
}

// invalidateCachesForEvent invalidates caches based on event type
func (wh *WebhookHandler) invalidateCachesForEvent(eventType string) error {
	switch {
	case eventType == "customer.subscription.created" ||
		eventType == "customer.subscription.updated" ||
		eventType == "customer.subscription.deleted" ||
		eventType == "invoice.payment_succeeded" ||
		eventType == "invoice.payment_failed":
		// Invalidate revenue cache
		return wh.cacheInvalidator.InvalidateCache("revenue")

	case eventType == "customer.created" ||
		eventType == "customer.deleted" ||
		eventType == "customer.updated":
		// Invalidate customer cache
		return wh.cacheInvalidator.InvalidateCache("customers")
	}

	return nil
}

// logEvent adds an event to the event log
func (wh *WebhookHandler) logEvent(event WebhookEvent) {
	wh.mu.Lock()
	defer wh.mu.Unlock()

	wh.eventLog = append(wh.eventLog, event)

	// Keep only the last N events
	if len(wh.eventLog) > wh.maxEventLog {
		wh.eventLog = wh.eventLog[len(wh.eventLog)-wh.maxEventLog:]
	}
}

// GetEventLog returns recent webhook events
func (wh *WebhookHandler) GetEventLog() []WebhookEvent {
	wh.mu.RLock()
	defer wh.mu.RUnlock()

	// Return a copy
	log := make([]WebhookEvent, len(wh.eventLog))
	copy(log, wh.eventLog)
	return log
}

// Default event handlers

func handleSubscriptionCreated(ctx context.Context, event stripe.Event) error {
	var subscription stripe.Subscription
	if err := json.Unmarshal(event.Data.Raw, &subscription); err != nil {
		return fmt.Errorf("failed to unmarshal subscription: %w", err)
	}

	slog.Info("Subscription created",
		"subscription_id", subscription.ID,
		"customer_id", subscription.Customer.ID,
		"status", subscription.Status)

	// Store in database if available
	db, err := GetMetricsDatabase("")
	if err == nil {
		// Calculate MRR for this subscription
		mrr := calculateSubscriptionMRR(&subscription)

		mode := "live"
		if !event.Livemode {
			mode = "test"
		}

		snapshot := &RevenueSnapshot{
			Timestamp: time.Now(),
			NewMRR:    mrr,
			Mode:      mode,
		}

		if err := db.SaveRevenueSnapshot(ctx, snapshot); err != nil {
			slog.Error("Failed to save revenue snapshot", "error", err)
		}
	}

	return nil
}

func handleSubscriptionUpdated(ctx context.Context, event stripe.Event) error {
	var subscription stripe.Subscription
	if err := json.Unmarshal(event.Data.Raw, &subscription); err != nil {
		return fmt.Errorf("failed to unmarshal subscription: %w", err)
	}

	slog.Info("Subscription updated",
		"subscription_id", subscription.ID,
		"customer_id", subscription.Customer.ID,
		"status", subscription.Status)

	return nil
}

func handleSubscriptionDeleted(ctx context.Context, event stripe.Event) error {
	var subscription stripe.Subscription
	if err := json.Unmarshal(event.Data.Raw, &subscription); err != nil {
		return fmt.Errorf("failed to unmarshal subscription: %w", err)
	}

	slog.Info("Subscription deleted",
		"subscription_id", subscription.ID,
		"customer_id", subscription.Customer.ID)

	// Store in database if available
	db, err := GetMetricsDatabase("")
	if err == nil {
		mrr := calculateSubscriptionMRR(&subscription)

		mode := "live"
		if !event.Livemode {
			mode = "test"
		}

		snapshot := &RevenueSnapshot{
			Timestamp:  time.Now(),
			ChurnedMRR: mrr,
			Mode:       mode,
		}

		if err := db.SaveRevenueSnapshot(ctx, snapshot); err != nil {
			slog.Error("Failed to save revenue snapshot", "error", err)
		}
	}

	return nil
}

func handleCustomerCreated(ctx context.Context, event stripe.Event) error {
	var customer stripe.Customer
	if err := json.Unmarshal(event.Data.Raw, &customer); err != nil {
		return fmt.Errorf("failed to unmarshal customer: %w", err)
	}

	slog.Info("Customer created", "customer_id", customer.ID)

	// Store in database if available
	db, err := GetMetricsDatabase("")
	if err == nil {
		mode := "live"
		if !event.Livemode {
			mode = "test"
		}

		snapshot := &CustomerSnapshot{
			Timestamp:    time.Now(),
			NewCustomers: 1,
			Mode:         mode,
		}

		if err := db.SaveCustomerSnapshot(ctx, snapshot); err != nil {
			slog.Error("Failed to save customer snapshot", "error", err)
		}
	}

	return nil
}

func handleCustomerDeleted(ctx context.Context, event stripe.Event) error {
	var customer stripe.Customer
	if err := json.Unmarshal(event.Data.Raw, &customer); err != nil {
		return fmt.Errorf("failed to unmarshal customer: %w", err)
	}

	slog.Info("Customer deleted", "customer_id", customer.ID)

	// Store in database if available
	db, err := GetMetricsDatabase("")
	if err == nil {
		mode := "live"
		if !event.Livemode {
			mode = "test"
		}

		snapshot := &CustomerSnapshot{
			Timestamp:        time.Now(),
			ChurnedCustomers: 1,
			Mode:             mode,
		}

		if err := db.SaveCustomerSnapshot(ctx, snapshot); err != nil {
			slog.Error("Failed to save customer snapshot", "error", err)
		}
	}

	return nil
}

func handleInvoicePaymentSucceeded(ctx context.Context, event stripe.Event) error {
	var invoice stripe.Invoice
	if err := json.Unmarshal(event.Data.Raw, &invoice); err != nil {
		return fmt.Errorf("failed to unmarshal invoice: %w", err)
	}

	slog.Info("Invoice payment succeeded",
		"invoice_id", invoice.ID,
		"customer_id", invoice.Customer.ID,
		"amount", invoice.AmountPaid)

	return nil
}

func handleInvoicePaymentFailed(ctx context.Context, event stripe.Event) error {
	var invoice stripe.Invoice
	if err := json.Unmarshal(event.Data.Raw, &invoice); err != nil {
		return fmt.Errorf("failed to unmarshal invoice: %w", err)
	}

	slog.Warn("Invoice payment failed",
		"invoice_id", invoice.ID,
		"customer_id", invoice.Customer.ID,
		"amount", invoice.AmountDue)

	return nil
}

// calculateSubscriptionMRR calculates MRR for a single subscription
func calculateSubscriptionMRR(sub *stripe.Subscription) float64 {
	totalMRR := 0.0

	for _, item := range sub.Items.Data {
		if item.Price == nil {
			continue
		}

		amount := float64(item.Price.UnitAmount) / 100.0
		interval := string(item.Price.Recurring.Interval)
		intervalCount := item.Price.Recurring.IntervalCount

		var monthlyAmount float64
		switch interval {
		case "month":
			monthlyAmount = amount / float64(intervalCount)
		case "year":
			monthlyAmount = amount / (12.0 * float64(intervalCount))
		case "week":
			monthlyAmount = amount * 4.33 / float64(intervalCount)
		case "day":
			monthlyAmount = amount * 30 / float64(intervalCount)
		}

		monthlyAmount *= float64(item.Quantity)
		totalMRR += monthlyAmount
	}

	return totalMRR
}

// WebhookStatusHandler returns an HTTP handler for webhook status
func WebhookStatusHandler(handler *WebhookHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		eventLog := handler.GetEventLog()

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"total_events": len(eventLog),
			"recent_events": eventLog,
		})
	}
}
