package glance

import (
	"context"
	"fmt"
	"html/template"
	"log/slog"
	"os"
	"strconv"
	"time"

	"github.com/stripe/stripe-go/v81"
	"github.com/stripe/stripe-go/v81/customer"
	"github.com/stripe/stripe-go/v81/subscription"
)

var customersWidgetTemplate = mustParseTemplate("customers.html", "widget-base.html")

type customersWidget struct {
	widgetBase       `yaml:",inline"`
	StripeAPIKey     string `yaml:"stripe-api-key"`
	StripeMode       string `yaml:"stripe-mode"` // 'live' or 'test'

	// Customer metrics
	TotalCustomers   int     `yaml:"-"`
	NewCustomers     int     `yaml:"-"`
	ChurnedCustomers int     `yaml:"-"`
	ChurnRate        float64 `yaml:"-"`
	ActiveCustomers  int     `yaml:"-"`

	// Financial metrics (if available)
	CAC              float64 `yaml:"-"` // Customer Acquisition Cost
	LTV              float64 `yaml:"-"` // Lifetime Value
	LTVtoCAC         float64 `yaml:"-"` // LTV/CAC ratio

	// Trend data
	TrendLabels      []string  `yaml:"-"`
	TrendValues      []int     `yaml:"-"`
}

func (w *customersWidget) initialize() error {
	w.widgetBase.withTitle("Customer Metrics").withCacheDuration(time.Hour)

	if w.StripeAPIKey == "" {
		return fmt.Errorf("stripe-api-key is required for customers widget")
	}

	if w.StripeMode == "" {
		w.StripeMode = "live"
	}

	if w.StripeMode != "live" && w.StripeMode != "test" {
		return fmt.Errorf("stripe-mode must be 'live' or 'test', got: %s", w.StripeMode)
	}

	return nil
}

func (w *customersWidget) update(ctx context.Context) {
	// Get decrypted API key
	encService, err := GetEncryptionService()
	if err != nil {
		w.withError(fmt.Errorf("encryption service unavailable: %w", err))
		return
	}

	apiKey, err := encService.DecryptIfNeeded(w.StripeAPIKey)
	if err != nil {
		w.withError(fmt.Errorf("failed to decrypt API key: %w", err))
		return
	}

	// Get Stripe client with resilience
	pool := GetStripeClientPool()
	client, err := pool.GetClient(apiKey, w.StripeMode)
	if err != nil {
		w.withError(fmt.Errorf("failed to get Stripe client: %w", err))
		return
	}

	// Set Stripe API key for direct API calls
	stripe.Key = apiKey

	// Try to load from database first for trend data
	db, dbErr := GetMetricsDatabase("")
	if dbErr == nil {
		// Get historical data from database
		endTime := time.Now()
		startTime := endTime.AddDate(0, -6, 0) // Last 6 months
		history, err := db.GetCustomerHistory(ctx, w.StripeMode, startTime, endTime)
		if err == nil && len(history) > 0 {
			w.loadHistoricalData(history)
		}
	}

	// Get total customers with retry
	totalCustomers, err := w.getTotalCustomersWithRetry(ctx, client)
	if !w.canContinueUpdateAfterHandlingErr(err) {
		return
	}
	w.TotalCustomers = totalCustomers

	// Get active customers (with active subscriptions)
	activeCustomers, err := w.getActiveCustomersWithRetry(ctx, client)
	if err != nil {
		slog.Error("Failed to get active customers", "error", err)
	} else {
		w.ActiveCustomers = activeCustomers
	}

	// Get new customers this month
	newCustomers, err := w.getNewCustomersWithRetry(ctx, client)
	if err != nil {
		slog.Error("Failed to get new customers", "error", err)
	} else {
		w.NewCustomers = newCustomers
	}

	// Get churned customers this month
	churnedCustomers, err := w.getChurnedCustomersWithRetry(ctx, client)
	if err != nil {
		slog.Error("Failed to get churned customers", "error", err)
	} else {
		w.ChurnedCustomers = churnedCustomers
	}

	// Calculate churn rate
	if w.TotalCustomers > 0 {
		w.ChurnRate = (float64(w.ChurnedCustomers) / float64(w.TotalCustomers)) * 100
	}

	// Calculate LTV using actual MRR data
	// LTV = Average MRR per customer / Monthly churn rate
	if w.ActiveCustomers > 0 && w.ChurnRate > 0 {
		// Try to get current MRR from database first (most efficient)
		var avgRevenuePerCustomer float64
		if dbErr == nil {
			revenueSnapshot, err := db.GetLatestRevenue(ctx, w.StripeMode)
			if err == nil && revenueSnapshot != nil && revenueSnapshot.MRR > 0 {
				// Use actual MRR data
				avgRevenuePerCustomer = revenueSnapshot.MRR / float64(w.ActiveCustomers)
				slog.Debug("Calculated LTV from database MRR",
					"mrr", revenueSnapshot.MRR,
					"active_customers", w.ActiveCustomers,
					"avg_revenue", avgRevenuePerCustomer)
			} else {
				// Fallback: Calculate MRR directly from Stripe
				currentMRR, err := w.calculateCurrentMRRWithRetry(ctx, client)
				if err == nil && currentMRR > 0 {
					avgRevenuePerCustomer = currentMRR / float64(w.ActiveCustomers)
					slog.Debug("Calculated LTV from fresh MRR calculation",
						"mrr", currentMRR,
						"active_customers", w.ActiveCustomers,
						"avg_revenue", avgRevenuePerCustomer)
				} else {
					// Ultimate fallback: use a conservative estimate
					avgRevenuePerCustomer = 29.0 // Conservative default for SaaS
					slog.Warn("Using default average revenue for LTV calculation - could not fetch MRR",
						"default", avgRevenuePerCustomer,
						"error", err)
				}
			}
		} else {
			// No database, calculate fresh
			currentMRR, err := w.calculateCurrentMRRWithRetry(ctx, client)
			if err == nil && currentMRR > 0 {
				avgRevenuePerCustomer = currentMRR / float64(w.ActiveCustomers)
			} else {
				avgRevenuePerCustomer = 29.0 // Conservative default
				slog.Warn("Using default average revenue for LTV calculation",
					"default", avgRevenuePerCustomer)
			}
		}

		monthlyChurnRate := w.ChurnRate / 100.0
		if monthlyChurnRate > 0 {
			w.LTV = avgRevenuePerCustomer / monthlyChurnRate
		}
	}

	// CAC: Allow manual override via environment variable
	// In production, integrate with Google Ads, Facebook Ads, etc.
	cacEnv := os.Getenv("BUSINESS_CAC")
	if cacEnv != "" {
		// Parse CAC from environment variable
		if cacValue, err := strconv.ParseFloat(cacEnv, 64); err == nil {
			w.CAC = cacValue
			slog.Debug("Using CAC from environment variable", "cac", cacValue)
		}
	}
	// If no CAC set, leave it as 0 (will be displayed as N/A in UI)

	// Calculate LTV/CAC ratio
	if w.CAC > 0 {
		w.LTVtoCAC = w.LTV / w.CAC
	}

	// Generate trend data
	w.generateTrendData()

	// Save to database for historical tracking
	if dbErr == nil {
		snapshot := &CustomerSnapshot{
			Timestamp:        time.Now(),
			TotalCustomers:   w.TotalCustomers,
			NewCustomers:     w.NewCustomers,
			ChurnedCustomers: w.ChurnedCustomers,
			ChurnRate:        w.ChurnRate,
			ActiveCustomers:  w.ActiveCustomers,
			Mode:             w.StripeMode,
		}

		if err := db.SaveCustomerSnapshot(ctx, snapshot); err != nil {
			slog.Error("Failed to save customer snapshot", "error", err)
		}
	}
}

func (w *customersWidget) getTotalCustomers(ctx context.Context) (int, error) {
	params := &stripe.CustomerListParams{}
	params.Context = ctx

	count := 0
	iter := customer.List(params)

	for iter.Next() {
		count++
	}

	if err := iter.Err(); err != nil {
		return 0, fmt.Errorf("failed to list customers: %w", err)
	}

	return count, nil
}

func (w *customersWidget) getActiveCustomers(ctx context.Context) (int, error) {
	// Get customers with active subscriptions
	params := &stripe.SubscriptionListParams{}
	params.Status = stripe.String("active")
	params.Context = ctx

	// Use a map to track unique customers
	uniqueCustomers := make(map[string]bool)
	iter := subscription.List(params)

	for iter.Next() {
		sub := iter.Subscription()
		if sub.Customer != nil {
			uniqueCustomers[sub.Customer.ID] = true
		}
	}

	if err := iter.Err(); err != nil {
		return 0, fmt.Errorf("failed to list active subscriptions: %w", err)
	}

	return len(uniqueCustomers), nil
}

func (w *customersWidget) getNewCustomers(ctx context.Context) (int, error) {
	// Get customers created this month
	now := time.Now()
	startOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)

	params := &stripe.CustomerListParams{}
	params.Filters.AddFilter("created", "gte", fmt.Sprintf("%d", startOfMonth.Unix()))
	params.Context = ctx

	count := 0
	iter := customer.List(params)

	for iter.Next() {
		count++
	}

	if err := iter.Err(); err != nil {
		return 0, fmt.Errorf("failed to list new customers: %w", err)
	}

	return count, nil
}

func (w *customersWidget) getChurnedCustomers(ctx context.Context) (int, error) {
	// Get subscriptions canceled this month
	now := time.Now()
	startOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)

	params := &stripe.SubscriptionListParams{}
	params.Status = stripe.String("canceled")
	params.Filters.AddFilter("canceled_at", "gte", fmt.Sprintf("%d", startOfMonth.Unix()))
	params.Context = ctx

	// Use a map to track unique customers who churned
	uniqueCustomers := make(map[string]bool)
	iter := subscription.List(params)

	for iter.Next() {
		sub := iter.Subscription()
		if sub.Customer != nil {
			uniqueCustomers[sub.Customer.ID] = true
		}
	}

	if err := iter.Err(); err != nil {
		return 0, fmt.Errorf("failed to list churned subscriptions: %w", err)
	}

	return len(uniqueCustomers), nil
}

func (w *customersWidget) generateTrendData() {
	// For MVP, generate simple trend based on current data
	// In production, query historical data

	now := time.Now()
	months := 6

	w.TrendLabels = make([]string, months)
	w.TrendValues = make([]int, months)

	// Generate last 6 months
	for i := months - 1; i >= 0; i-- {
		monthDate := now.AddDate(0, -i, 0)
		w.TrendLabels[months-1-i] = monthDate.Format("Jan")

		// For MVP, simulate growth trend
		// In production, fetch actual historical data
		if i == 0 {
			w.TrendValues[months-1-i] = w.TotalCustomers
		} else {
			// Simulate historical customer count with growth
			growthPerMonth := w.NewCustomers - w.ChurnedCustomers
			w.TrendValues[months-1-i] = w.TotalCustomers - (growthPerMonth * i)

			// Ensure non-negative
			if w.TrendValues[months-1-i] < 0 {
				w.TrendValues[months-1-i] = 0
			}
		}
	}
}

func (w *customersWidget) Render() template.HTML {
	return w.renderTemplate(w, customersWidgetTemplate)
}

// getTotalCustomersWithRetry wraps getTotalCustomers with circuit breaker and retry logic
func (w *customersWidget) getTotalCustomersWithRetry(ctx context.Context, client *StripeClientWrapper) (int, error) {
	var result int
	err := client.ExecuteWithRetry(ctx, "getTotalCustomers", func() error {
		count, err := w.getTotalCustomers(ctx)
		result = count
		return err
	})
	return result, err
}

// getActiveCustomersWithRetry wraps getActiveCustomers with circuit breaker and retry logic
func (w *customersWidget) getActiveCustomersWithRetry(ctx context.Context, client *StripeClientWrapper) (int, error) {
	var result int
	err := client.ExecuteWithRetry(ctx, "getActiveCustomers", func() error {
		count, err := w.getActiveCustomers(ctx)
		result = count
		return err
	})
	return result, err
}

// getNewCustomersWithRetry wraps getNewCustomers with circuit breaker and retry logic
func (w *customersWidget) getNewCustomersWithRetry(ctx context.Context, client *StripeClientWrapper) (int, error) {
	var result int
	err := client.ExecuteWithRetry(ctx, "getNewCustomers", func() error {
		count, err := w.getNewCustomers(ctx)
		result = count
		return err
	})
	return result, err
}

// getChurnedCustomersWithRetry wraps getChurnedCustomers with circuit breaker and retry logic
func (w *customersWidget) getChurnedCustomersWithRetry(ctx context.Context, client *StripeClientWrapper) (int, error) {
	var result int
	err := client.ExecuteWithRetry(ctx, "getChurnedCustomers", func() error {
		count, err := w.getChurnedCustomers(ctx)
		result = count
		return err
	})
	return result, err
}

// calculateCurrentMRR calculates the current MRR from active subscriptions
// This is used for LTV calculation when database snapshot is not available
func (w *customersWidget) calculateCurrentMRR(ctx context.Context) (float64, error) {
	// Fetch all active subscriptions
	params := &stripe.SubscriptionListParams{}
	params.Status = stripe.String("active")
	params.Context = ctx

	totalMRR := 0.0
	iter := subscription.List(params)

	for iter.Next() {
		sub := iter.Subscription()

		// Calculate MRR for this subscription
		for _, item := range sub.Items.Data {
			if item.Price == nil {
				continue
			}

			// Get the amount in dollars (Stripe uses cents)
			amount := float64(item.Price.UnitAmount) / 100.0

			// Normalize to monthly based on interval
			interval := item.Price.Recurring.Interval
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
			default:
				slog.Warn("Unknown subscription interval", "interval", interval)
				continue
			}

			// Multiply by quantity
			monthlyAmount *= float64(item.Quantity)
			totalMRR += monthlyAmount
		}
	}

	if err := iter.Err(); err != nil {
		return 0, fmt.Errorf("failed to list subscriptions for MRR: %w", err)
	}

	return totalMRR, nil
}

// calculateCurrentMRRWithRetry wraps calculateCurrentMRR with circuit breaker and retry logic
func (w *customersWidget) calculateCurrentMRRWithRetry(ctx context.Context, client *StripeClientWrapper) (float64, error) {
	var result float64
	err := client.ExecuteWithRetry(ctx, "calculateCurrentMRR", func() error {
		mrr, err := w.calculateCurrentMRR(ctx)
		result = mrr
		return err
	})
	return result, err
}

// loadHistoricalData loads historical data from database snapshots
func (w *customersWidget) loadHistoricalData(history []*CustomerSnapshot) {
	if len(history) == 0 {
		return
	}

	// Use database data to populate trend chart
	maxPoints := 6
	if len(history) > maxPoints {
		history = history[:maxPoints]
	}

	w.TrendLabels = make([]string, len(history))
	w.TrendValues = make([]int, len(history))

	// Reverse chronological order (oldest first for chart)
	for i := range history {
		idx := len(history) - 1 - i
		w.TrendLabels[i] = history[idx].Timestamp.Format("Jan")
		w.TrendValues[i] = history[idx].TotalCustomers
	}
}
