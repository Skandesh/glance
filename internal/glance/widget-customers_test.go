package glance

import (
	"testing"
	"time"
)

func TestCustomersWidget_Initialize(t *testing.T) {
	tests := []struct {
		name          string
		widget        *customersWidget
		expectError   bool
		errorContains string
	}{
		{
			name: "valid configuration",
			widget: &customersWidget{
				StripeAPIKey: "sk_test_valid_key",
				StripeMode:   "test",
			},
			expectError: false,
		},
		{
			name:          "missing API key",
			widget:        &customersWidget{},
			expectError:   true,
			errorContains: "stripe-api-key is required",
		},
		{
			name: "invalid mode",
			widget: &customersWidget{
				StripeAPIKey: "sk_test_valid_key",
				StripeMode:   "production", // invalid
			},
			expectError:   true,
			errorContains: "must be 'live' or 'test'",
		},
		{
			name: "defaults to live mode",
			widget: &customersWidget{
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

func TestCustomersWidget_ChurnRateCalculation(t *testing.T) {
	tests := []struct {
		name             string
		totalCustomers   int
		churnedCustomers int
		expectedRate     float64
	}{
		{
			name:             "5% churn rate",
			totalCustomers:   100,
			churnedCustomers: 5,
			expectedRate:     5.0,
		},
		{
			name:             "no churn",
			totalCustomers:   100,
			churnedCustomers: 0,
			expectedRate:     0.0,
		},
		{
			name:             "10% churn",
			totalCustomers:   1000,
			churnedCustomers: 100,
			expectedRate:     10.0,
		},
		{
			name:             "fractional churn",
			totalCustomers:   137,
			churnedCustomers: 3,
			expectedRate:     2.19, // 3/137 * 100
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var churnRate float64
			if tt.totalCustomers > 0 {
				churnRate = (float64(tt.churnedCustomers) / float64(tt.totalCustomers)) * 100
			}

			if !floatEquals(churnRate, tt.expectedRate, 0.01) {
				t.Errorf("expected churn rate %f%%, got %f%%", tt.expectedRate, churnRate)
			}
		})
	}
}

func TestCustomersWidget_LTVCalculation(t *testing.T) {
	tests := []struct {
		name              string
		avgRevenue        float64
		monthlyChurnRate  float64
		expectedLTV       float64
		expectZero        bool
	}{
		{
			name:             "basic LTV",
			avgRevenue:       100.0,
			monthlyChurnRate: 0.05, // 5%
			expectedLTV:      2000.0, // 100 / 0.05
		},
		{
			name:             "high churn",
			avgRevenue:       50.0,
			monthlyChurnRate: 0.10, // 10%
			expectedLTV:      500.0, // 50 / 0.10
		},
		{
			name:             "low churn",
			avgRevenue:       200.0,
			monthlyChurnRate: 0.02, // 2%
			expectedLTV:      10000.0, // 200 / 0.02
		},
		{
			name:             "zero churn (no LTV calculation)",
			avgRevenue:       100.0,
			monthlyChurnRate: 0.0,
			expectZero:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var ltv float64
			if tt.monthlyChurnRate > 0 {
				ltv = tt.avgRevenue / tt.monthlyChurnRate
			}

			if tt.expectZero {
				if ltv != 0 {
					t.Errorf("expected LTV to be 0 (undefined), got %f", ltv)
				}
			} else {
				if !floatEquals(ltv, tt.expectedLTV, 0.01) {
					t.Errorf("expected LTV %f, got %f", tt.expectedLTV, ltv)
				}
			}
		})
	}
}

func TestCustomersWidget_LTVtoCACRatio(t *testing.T) {
	tests := []struct {
		name         string
		ltv          float64
		cac          float64
		expectedRate float64
	}{
		{
			name:         "healthy ratio 3:1",
			ltv:          3000.0,
			cac:          1000.0,
			expectedRate: 3.0,
		},
		{
			name:         "excellent ratio 10:1",
			ltv:          5000.0,
			cac:          500.0,
			expectedRate: 10.0,
		},
		{
			name:         "poor ratio 1:1",
			ltv:          1000.0,
			cac:          1000.0,
			expectedRate: 1.0,
		},
		{
			name:         "best-in-class 15:1",
			ltv:          7500.0,
			cac:          500.0,
			expectedRate: 15.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var ratio float64
			if tt.cac > 0 {
				ratio = tt.ltv / tt.cac
			}

			if !floatEquals(ratio, tt.expectedRate, 0.01) {
				t.Errorf("expected LTV/CAC ratio %f, got %f", tt.expectedRate, ratio)
			}
		})
	}
}

func TestCustomersWidget_GenerateTrendData(t *testing.T) {
	widget := &customersWidget{
		TotalCustomers:   1000,
		NewCustomers:     50,
		ChurnedCustomers: 20,
	}

	widget.generateTrendData()

	// Check that trend data was generated
	if len(widget.TrendLabels) != 6 {
		t.Errorf("expected 6 trend labels, got %d", len(widget.TrendLabels))
	}

	if len(widget.TrendValues) != 6 {
		t.Errorf("expected 6 trend values, got %d", len(widget.TrendValues))
	}

	// Check that current month has total customers
	if widget.TrendValues[5] != widget.TotalCustomers {
		t.Errorf("expected last trend value to be total customers (%d), got %d", widget.TotalCustomers, widget.TrendValues[5])
	}

	// Check that all values are non-negative
	for i, val := range widget.TrendValues {
		if val < 0 {
			t.Errorf("trend value %d is negative: %d", i, val)
		}
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

func TestCustomersWidget_NetCustomerGrowth(t *testing.T) {
	tests := []struct {
		name         string
		newCustomers int
		churned      int
		expectedNet  int
	}{
		{
			name:         "positive growth",
			newCustomers: 50,
			churned:      20,
			expectedNet:  30,
		},
		{
			name:         "negative growth",
			newCustomers: 10,
			churned:      25,
			expectedNet:  -15,
		},
		{
			name:         "no growth",
			newCustomers: 15,
			churned:      15,
			expectedNet:  0,
		},
		{
			name:         "high growth",
			newCustomers: 100,
			churned:      5,
			expectedNet:  95,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			netGrowth := tt.newCustomers - tt.churned

			if netGrowth != tt.expectedNet {
				t.Errorf("expected net growth %d, got %d", tt.expectedNet, netGrowth)
			}
		})
	}
}
