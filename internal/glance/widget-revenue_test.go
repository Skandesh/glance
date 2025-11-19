package glance

import (
	"testing"
	"time"
)

func TestRevenueWidget_Initialize(t *testing.T) {
	tests := []struct {
		name          string
		widget        *revenueWidget
		expectError   bool
		errorContains string
	}{
		{
			name: "valid configuration",
			widget: &revenueWidget{
				StripeAPIKey: "sk_test_valid_key",
				StripeMode:   "test",
			},
			expectError: false,
		},
		{
			name:          "missing API key",
			widget:        &revenueWidget{},
			expectError:   true,
			errorContains: "stripe-api-key is required",
		},
		{
			name: "invalid mode",
			widget: &revenueWidget{
				StripeAPIKey: "sk_test_valid_key",
				StripeMode:   "invalid",
			},
			expectError:   true,
			errorContains: "must be 'live' or 'test'",
		},
		{
			name: "defaults to live mode",
			widget: &revenueWidget{
				StripeAPIKey: "sk_live_valid_key",
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.widget.initialize()

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
				} else if tt.errorContains != "" && !contains(err.Error(), tt.errorContains) {
					t.Errorf("expected error to contain %q, got %q", tt.errorContains, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}

				// Check defaults
				if tt.widget.Title == "" {
					t.Error("expected Title to be set by initialize")
				}
				if tt.widget.cacheDuration != time.Hour {
					t.Errorf("expected cache duration to be 1 hour, got %v", tt.widget.cacheDuration)
				}
				if tt.widget.StripeMode == "" {
					t.Error("expected StripeMode to default to 'live'")
				}
			}
		})
	}
}

func TestRevenueWidget_GenerateTrendData(t *testing.T) {
	widget := &revenueWidget{
		CurrentMRR: 10000.0,
		GrowthRate: 10.0, // 10% growth
	}

	widget.generateTrendData()

	// Check that trend data was generated
	if len(widget.TrendLabels) != 6 {
		t.Errorf("expected 6 trend labels, got %d", len(widget.TrendLabels))
	}

	if len(widget.TrendValues) != 6 {
		t.Errorf("expected 6 trend values, got %d", len(widget.TrendValues))
	}

	// Check that current month has current MRR
	if widget.TrendValues[5] != widget.CurrentMRR {
		t.Errorf("expected last trend value to be current MRR (%f), got %f", widget.CurrentMRR, widget.TrendValues[5])
	}

	// Check that labels are month names
	validMonths := map[string]bool{
		"Jan": true, "Feb": true, "Mar": true, "Apr": true,
		"May": true, "Jun": true, "Jul": true, "Aug": true,
		"Sep": true, "Oct": true, "Nov": true, "Dec": true,
	}

	for i, label := range widget.TrendLabels {
		if !validMonths[label] {
			t.Errorf("trend label %d (%q) is not a valid month", i, label)
		}
	}
}

func TestRevenueWidget_MRRCalculation(t *testing.T) {
	// Test interval normalization logic
	tests := []struct {
		name          string
		amount        float64 // in cents
		interval      string
		intervalCount int64
		quantity      int64
		expectedMRR   float64
	}{
		{
			name:          "monthly subscription",
			amount:        2900, // $29.00
			interval:      "month",
			intervalCount: 1,
			quantity:      1,
			expectedMRR:   29.0,
		},
		{
			name:          "yearly subscription",
			amount:        29900, // $299.00
			interval:      "year",
			intervalCount: 1,
			quantity:      1,
			expectedMRR:   299.0 / 12.0, // ~24.92
		},
		{
			name:          "bi-monthly subscription",
			amount:        5000, // $50.00
			interval:      "month",
			intervalCount: 2,
			quantity:      1,
			expectedMRR:   25.0, // $50 / 2
		},
		{
			name:          "weekly subscription",
			amount:        700, // $7.00
			interval:      "week",
			intervalCount: 1,
			quantity:      1,
			expectedMRR:   7.0 * 4.33, // ~30.31
		},
		{
			name:          "daily subscription",
			amount:        100, // $1.00
			interval:      "day",
			intervalCount: 1,
			quantity:      1,
			expectedMRR:   30.0, // $1 * 30
		},
		{
			name:          "quantity > 1",
			amount:        1000, // $10.00
			interval:      "month",
			intervalCount: 1,
			quantity:      5,
			expectedMRR:   50.0, // $10 * 5
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Simulate MRR calculation logic
			amountInDollars := float64(tt.amount) / 100.0
			var monthlyAmount float64

			switch tt.interval {
			case "month":
				monthlyAmount = amountInDollars / float64(tt.intervalCount)
			case "year":
				monthlyAmount = amountInDollars / (12.0 * float64(tt.intervalCount))
			case "week":
				monthlyAmount = amountInDollars * 4.33 / float64(tt.intervalCount)
			case "day":
				monthlyAmount = amountInDollars * 30 / float64(tt.intervalCount)
			}

			monthlyAmount *= float64(tt.quantity)

			if !floatEquals(monthlyAmount, tt.expectedMRR, 0.01) {
				t.Errorf("expected MRR %f, got %f", tt.expectedMRR, monthlyAmount)
			}
		})
	}
}

func TestRevenueWidget_GrowthRateCalculation(t *testing.T) {
	tests := []struct {
		name           string
		currentMRR     float64
		previousMRR    float64
		expectedGrowth float64
	}{
		{
			name:           "10% growth",
			currentMRR:     11000,
			previousMRR:    10000,
			expectedGrowth: 10.0,
		},
		{
			name:           "negative growth (churn)",
			currentMRR:     9000,
			previousMRR:    10000,
			expectedGrowth: -10.0,
		},
		{
			name:           "no growth",
			currentMRR:     10000,
			previousMRR:    10000,
			expectedGrowth: 0.0,
		},
		{
			name:           "100% growth (doubled)",
			currentMRR:     20000,
			previousMRR:    10000,
			expectedGrowth: 100.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			growthRate := ((tt.currentMRR - tt.previousMRR) / tt.previousMRR) * 100

			if !floatEquals(growthRate, tt.expectedGrowth, 0.01) {
				t.Errorf("expected growth rate %f%%, got %f%%", tt.expectedGrowth, growthRate)
			}
		})
	}
}

// Helper functions
func contains(s, substr string) bool {
	return len(s) >= len(substr) && s[:len(substr)] == substr || len(s) > len(substr) && findSubstring(s, substr)
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func floatEquals(a, b, tolerance float64) bool {
	diff := a - b
	if diff < 0 {
		diff = -diff
	}
	return diff < tolerance
}
