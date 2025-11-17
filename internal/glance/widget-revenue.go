package glance

import (
	"context"
	"fmt"
	"html/template"
	"log/slog"
	"time"

	"github.com/stripe/stripe-go/v81"
	"github.com/stripe/stripe-go/v81/subscription"
)

var revenueWidgetTemplate = mustParseTemplate("revenue.html", "widget-base.html")

type revenueWidget struct {
	widgetBase      `yaml:",inline"`
	StripeAPIKey    string `yaml:"stripe-api-key"`
	StripeMode      string `yaml:"stripe-mode"` // 'live' or 'test'

	// Revenue metrics
	CurrentMRR   float64 `yaml:"-"`
	PreviousMRR  float64 `yaml:"-"`
	GrowthRate   float64 `yaml:"-"`
	ARR          float64 `yaml:"-"`
	NewMRR       float64 `yaml:"-"`
	ChurnedMRR   float64 `yaml:"-"`
	NetNewMRR    float64 `yaml:"-"`

	// Trend data for charts
	TrendLabels  []string  `yaml:"-"`
	TrendValues  []float64 `yaml:"-"`
}

type chartPoint struct {
	Month string
	Value float64
}

func (w *revenueWidget) initialize() error {
	w.widgetBase.withTitle("Revenue").withCacheDuration(time.Hour)

	if w.StripeAPIKey == "" {
		return fmt.Errorf("stripe-api-key is required for revenue widget")
	}

	if w.StripeMode == "" {
		w.StripeMode = "live"
	}

	if w.StripeMode != "live" && w.StripeMode != "test" {
		return fmt.Errorf("stripe-mode must be 'live' or 'test', got: %s", w.StripeMode)
	}

	return nil
}

func (w *revenueWidget) update(ctx context.Context) {
	// Set Stripe API key
	stripe.Key = w.StripeAPIKey

	// Calculate current MRR
	currentMRR, err := w.calculateMRR(ctx)
	if !w.canContinueUpdateAfterHandlingErr(err) {
		return
	}

	w.CurrentMRR = currentMRR
	w.ARR = currentMRR * 12

	// For MVP, we'll calculate growth by comparing to stored previous value
	// In production, you'd query historical data from Stripe or a database
	if w.PreviousMRR > 0 {
		w.GrowthRate = ((w.CurrentMRR - w.PreviousMRR) / w.PreviousMRR) * 100
	}

	// Calculate new MRR (subscriptions created this month)
	newMRR, err := w.calculateNewMRR(ctx)
	if err != nil {
		slog.Error("Failed to calculate new MRR", "error", err)
	} else {
		w.NewMRR = newMRR
	}

	// Calculate churned MRR (subscriptions canceled this month)
	churnedMRR, err := w.calculateChurnedMRR(ctx)
	if err != nil {
		slog.Error("Failed to calculate churned MRR", "error", err)
	} else {
		w.ChurnedMRR = churnedMRR
	}

	w.NetNewMRR = w.NewMRR - w.ChurnedMRR

	// Generate trend data (last 6 months for MVP)
	// In production, you'd store historical data
	w.generateTrendData()

	// Store current MRR for next iteration
	w.PreviousMRR = w.CurrentMRR
}

func (w *revenueWidget) calculateMRR(ctx context.Context) (float64, error) {
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
				monthlyAmount = amount * 4.33 / float64(intervalCount) // ~4.33 weeks per month
			case "day":
				monthlyAmount = amount * 30 / float64(intervalCount)
			default:
				slog.Warn("Unknown interval", "interval", interval)
				continue
			}

			// Multiply by quantity
			monthlyAmount *= float64(item.Quantity)

			totalMRR += monthlyAmount
		}
	}

	if err := iter.Err(); err != nil {
		return 0, fmt.Errorf("failed to list subscriptions: %w", err)
	}

	return totalMRR, nil
}

func (w *revenueWidget) calculateNewMRR(ctx context.Context) (float64, error) {
	// Get start of current month
	now := time.Now()
	startOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)

	// Fetch subscriptions created this month
	params := &stripe.SubscriptionListParams{}
	params.Status = stripe.String("active")
	params.Filters.AddFilter("created", "gte", fmt.Sprintf("%d", startOfMonth.Unix()))
	params.Context = ctx

	newMRR := 0.0
	iter := subscription.List(params)

	for iter.Next() {
		sub := iter.Subscription()

		// Calculate MRR for this subscription
		for _, item := range sub.Items.Data {
			if item.Price == nil {
				continue
			}

			amount := float64(item.Price.UnitAmount) / 100.0
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
				continue
			}

			monthlyAmount *= float64(item.Quantity)
			newMRR += monthlyAmount
		}
	}

	if err := iter.Err(); err != nil {
		return 0, fmt.Errorf("failed to list new subscriptions: %w", err)
	}

	return newMRR, nil
}

func (w *revenueWidget) calculateChurnedMRR(ctx context.Context) (float64, error) {
	// Get start of current month
	now := time.Now()
	startOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)

	// Fetch subscriptions canceled this month
	params := &stripe.SubscriptionListParams{}
	params.Status = stripe.String("canceled")
	params.Filters.AddFilter("canceled_at", "gte", fmt.Sprintf("%d", startOfMonth.Unix()))
	params.Context = ctx

	churnedMRR := 0.0
	iter := subscription.List(params)

	for iter.Next() {
		sub := iter.Subscription()

		// Calculate MRR that was lost
		for _, item := range sub.Items.Data {
			if item.Price == nil {
				continue
			}

			amount := float64(item.Price.UnitAmount) / 100.0
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
				continue
			}

			monthlyAmount *= float64(item.Quantity)
			churnedMRR += monthlyAmount
		}
	}

	if err := iter.Err(); err != nil {
		return 0, fmt.Errorf("failed to list churned subscriptions: %w", err)
	}

	return churnedMRR, nil
}

func (w *revenueWidget) generateTrendData() {
	// For MVP, generate simple trend based on current data
	// In production, you'd query historical data from database or Stripe

	now := time.Now()
	months := 6

	w.TrendLabels = make([]string, months)
	w.TrendValues = make([]float64, months)

	// Generate last 6 months
	for i := months - 1; i >= 0; i-- {
		monthDate := now.AddDate(0, -i, 0)
		w.TrendLabels[months-1-i] = monthDate.Format("Jan")

		// For MVP, simulate growth trend
		// In production, fetch actual historical data
		if i == 0 {
			w.TrendValues[months-1-i] = w.CurrentMRR
		} else {
			// Simulate historical data with some growth
			growthFactor := 1.0 + (w.GrowthRate/100.0)*float64(i)
			w.TrendValues[months-1-i] = w.CurrentMRR / growthFactor
		}
	}
}

func (w *revenueWidget) Render() template.HTML {
	return w.renderTemplate(w, revenueWidgetTemplate)
}
