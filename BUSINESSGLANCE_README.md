# BusinessGlance

**A self-hosted business metrics dashboard built on Glance, designed for SaaS startups, digital agencies, and SMBs.**

BusinessGlance extends Glance with powerful business intelligence widgets that integrate with Stripe to provide real-time revenue and customer analytics without the complexity and cost of enterprise BI tools.

## Overview

BusinessGlance transforms the popular Glance personal dashboard into a comprehensive business metrics platform. It maintains all of Glance's core features while adding critical business intelligence capabilities focused on SaaS metrics, customer analytics, and revenue tracking.

### Key Features

- **Real-time Revenue Analytics** - Track MRR, ARR, growth rates, and revenue trends
- **Customer Health Metrics** - Monitor total customers, churn rate, new signups, and LTV/CAC ratios
- **Stripe Integration** - Direct integration with Stripe for subscription and customer data
- **Lightweight Charts** - Beautiful trend visualizations without heavy JavaScript dependencies
- **Self-hosted** - Complete data ownership and privacy
- **Configuration-driven** - YAML-based configuration with hot reload support
- **Professional UI** - Clean, modern business theme optimized for metrics display

## Business Widgets

### Revenue Widget

Provides comprehensive revenue analytics powered by Stripe:

- **MRR (Monthly Recurring Revenue)** - Current monthly recurring revenue
- **ARR (Annual Recurring Revenue)** - Annualized revenue calculation
- **Growth Rate** - Month-over-month growth percentage
- **New MRR** - Revenue from new subscriptions this month
- **Churned MRR** - Lost revenue from cancellations
- **Net New MRR** - Net revenue change (new - churned)
- **6-Month Trend Chart** - Visual revenue trend over time

**Supports all Stripe subscription intervals:**
- Monthly subscriptions
- Annual subscriptions (normalized to MRR)
- Weekly subscriptions (4.33 weeks/month)
- Daily subscriptions (30 days/month)
- Custom interval counts (bi-monthly, quarterly, etc.)

### Customers Widget

Tracks customer health and acquisition metrics:

- **Total Customers** - All-time customer count
- **New Customers** - New signups this month
- **Churned Customers** - Customer losses this month
- **Churn Rate** - Percentage of customers lost
- **Active Customers** - Currently active customer count
- **LTV (Lifetime Value)** - Average customer lifetime value
- **CAC (Customer Acquisition Cost)** - Cost to acquire customers
- **LTV/CAC Ratio** - Key SaaS health metric (ideal: 3:1 or higher)
- **6-Month Customer Trend** - Visual customer growth over time

## Installation

### Prerequisites

- Go 1.24.3 or higher
- Stripe account with API access
- Linux/macOS/Windows system

### Build from Source

```bash
# Clone the repository
git clone https://github.com/yourusername/glance.git
cd glance

# Install dependencies
go mod download

# Build the binary
go build -o build/businessglance .

# Run BusinessGlance
./build/businessglance --config business-config.yml
```

### Docker

```bash
# Build Docker image
docker build -t businessglance .

# Run with environment variables
docker run -d \
  -p 8080:8080 \
  -v $(pwd)/business-config.yml:/app/glance.yml \
  -e STRIPE_SECRET_KEY=sk_test_your_key_here \
  businessglance
```

## Configuration

### Environment Variables

Create a `.env` file or set environment variables:

```bash
# Stripe API Key (required for business widgets)
STRIPE_SECRET_KEY=sk_test_your_key_here

# For production, use live keys:
# STRIPE_SECRET_KEY=sk_live_your_key_here
```

### Dashboard Configuration

Create a `business-config.yml` file:

```yaml
server:
  host: 0.0.0.0
  port: 8080

theme:
  light: true
  background-color: 240 13 20    # HSL values
  primary-color: 43 100 50       # Vibrant green for business metrics

pages:
  - name: Revenue & Customers
    slug: home
    columns:
      - size: small
        widgets:
          - type: revenue
            title: Monthly Recurring Revenue
            stripe-api-key: ${STRIPE_SECRET_KEY}
            stripe-mode: test    # Use 'live' for production
            cache: 1h

          - type: customers
            title: Customer Health
            stripe-api-key: ${STRIPE_SECRET_KEY}
            stripe-mode: test    # Use 'live' for production
            cache: 1h

      - size: full
        widgets:
          # Add other widgets like custom-api, calendar, etc.
          - type: custom-api
            title: API Status
            url: https://api.yourdomain.com/health
            cache: 5m
```

### Widget Parameters

#### Revenue Widget

| Parameter | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| `type` | string | Yes | - | Must be `revenue` |
| `title` | string | No | "Revenue" | Widget title |
| `stripe-api-key` | string | Yes | - | Stripe secret key (sk_test_* or sk_live_*) |
| `stripe-mode` | string | No | "live" | Either "live" or "test" |
| `cache` | duration | No | 1h | How long to cache Stripe data |

#### Customers Widget

| Parameter | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| `type` | string | Yes | - | Must be `customers` |
| `title` | string | No | "Customers" | Widget title |
| `stripe-api-key` | string | Yes | - | Stripe secret key |
| `stripe-mode` | string | No | "live" | Either "live" or "test" |
| `cache` | duration | No | 1h | How long to cache Stripe data |

## Usage

### Starting the Dashboard

```bash
# Start with default config
./build/businessglance

# Start with custom config
./build/businessglance --config business-config.yml

# Enable debug logging
./build/businessglance --debug

# The dashboard will be available at http://localhost:8080
```

### Stripe Configuration

1. **Get your Stripe API keys:**
   - Test mode: https://dashboard.stripe.com/test/apikeys
   - Live mode: https://dashboard.stripe.com/apikeys

2. **Set the API key:**
   ```bash
   export STRIPE_SECRET_KEY=sk_test_your_key_here
   ```

3. **Choose the mode:**
   - Use `stripe-mode: test` for development with test data
   - Use `stripe-mode: live` for production with real data

### Metrics Interpretation

#### Revenue Metrics

- **MRR Growth Rate** - Target: 15-20% monthly for early-stage SaaS
- **Churn Rate** - Benchmark: <5% monthly is healthy, <10% acceptable
- **New vs Churned MRR** - New MRR should exceed churned MRR for growth

#### Customer Metrics

- **Churn Rate** - <5% monthly is excellent, >10% needs attention
- **LTV/CAC Ratio** - 3:1 is healthy, 10:1+ is exceptional
- **Net Customer Growth** - Should be positive for sustainable growth

## Testing

### Run All Tests

```bash
# Run all tests
go test ./internal/glance -v

# Run specific widget tests
go test ./internal/glance -v -run="TestRevenueWidget"
go test ./internal/glance -v -run="TestCustomersWidget"

# Run with coverage
go test ./internal/glance -v -cover
```

### Test Coverage

BusinessGlance includes comprehensive unit tests for:

- Widget initialization and configuration validation
- MRR calculation across all Stripe subscription intervals
- Growth rate calculations (positive, negative, zero)
- Churn rate calculations
- LTV (Lifetime Value) calculations
- LTV/CAC ratio calculations
- Customer growth metrics
- Trend data generation

**Test files:**
- `internal/glance/widget-revenue_test.go` - 24+ test cases
- `internal/glance/widget-customers_test.go` - 24+ test cases

## Architecture

### Widget System

BusinessGlance uses Glance's widget plugin architecture:

```go
type revenueWidget struct {
    widgetBase       `yaml:",inline"`
    StripeAPIKey     string `yaml:"stripe-api-key"`
    StripeMode       string `yaml:"stripe-mode"`

    CurrentMRR       float64
    PreviousMRR      float64
    GrowthRate       float64
    ARR              float64
    // ... more fields
}

func (w *revenueWidget) initialize() error {
    // Validation and defaults
}

func (w *revenueWidget) update(ctx context.Context) {
    // Fetch and calculate metrics
}

func (w *revenueWidget) Render() template.HTML {
    // Render the widget HTML
}
```

### MRR Calculation Logic

All subscription intervals are normalized to monthly:

```go
func calculateMRR(amount float64, interval string, intervalCount int64) float64 {
    amountInDollars := amount / 100.0

    switch interval {
    case "month":
        return amountInDollars / float64(intervalCount)
    case "year":
        return amountInDollars / (12.0 * float64(intervalCount))
    case "week":
        return amountInDollars * 4.33 / float64(intervalCount)
    case "day":
        return amountInDollars * 30 / float64(intervalCount)
    }
}
```

### Chart Rendering

BusinessGlance uses a lightweight canvas-based chart system (`charts.js`) instead of heavy libraries:

- **Zero dependencies** - Pure JavaScript using Canvas API
- **Auto-render** - Charts render on page load via data attributes
- **Responsive** - Adapts to container width
- **Theme-aware** - Respects light/dark mode

### File Structure

```
glance/
├── internal/glance/
│   ├── widget-revenue.go           # Revenue widget implementation
│   ├── widget-revenue_test.go      # Revenue widget tests
│   ├── widget-customers.go         # Customer widget implementation
│   ├── widget-customers_test.go    # Customer widget tests
│   ├── templates/
│   │   ├── revenue.html            # Revenue widget template
│   │   └── customers.html          # Customer widget template
│   ├── static/
│   │   ├── css/
│   │   │   └── business.css        # Business theme styles
│   │   └── js/
│   │       └── charts.js           # Chart rendering
│   └── templates.go                # Template helpers
├── business-config.yml             # Example business configuration
├── .env.example                    # Environment variable template
└── build/
    └── businessglance              # Compiled binary
```

## Roadmap

### Phase 1: Core Business Widgets (Completed)
- ✅ Revenue widget with MRR/ARR tracking
- ✅ Customer metrics widget
- ✅ Stripe integration
- ✅ Trend visualizations
- ✅ Comprehensive testing

### Phase 2: Enhanced Analytics (Planned)
- [ ] Revenue cohort analysis
- [ ] Customer segmentation
- [ ] Forecasting and projections
- [ ] Multi-currency support
- [ ] Export to CSV/PDF

### Phase 3: Additional Integrations (Planned)
- [ ] Google Analytics integration
- [ ] HubSpot CRM integration
- [ ] Plausible Analytics widget
- [ ] QuickBooks/Xero integration
- [ ] Custom SQL data sources

### Phase 4: Advanced Features (Future)
- [ ] Alert system for metric thresholds
- [ ] Email reports and digests
- [ ] Team collaboration features
- [ ] Mobile responsive improvements
- [ ] API for programmatic access

## Performance

- **Response Time**: <100ms for cached data
- **Cache Duration**: Configurable per widget (default: 1 hour)
- **Stripe API Calls**: Minimized through intelligent caching
- **Memory Usage**: ~50MB typical, ~100MB with multiple widgets
- **Build Size**: ~21MB compiled binary

## Security

- **API Key Protection**: Environment variables, never committed to git
- **HTTPS Recommended**: Deploy behind reverse proxy with SSL
- **Data Privacy**: All data stays on your infrastructure
- **Test/Live Separation**: Stripe mode prevents accidental production access
- **Input Validation**: All widget configurations validated on startup

## Troubleshooting

### No Revenue Data Showing

1. Verify Stripe API key is correct and has access to subscriptions
2. Check that you have active subscriptions in your Stripe account
3. Confirm `stripe-mode` matches your API key (test vs live)
4. Check logs for Stripe API errors: `./businessglance --debug`

### Charts Not Rendering

1. Ensure `charts.js` is loaded in your template
2. Check browser console for JavaScript errors
3. Verify trend data is being generated (check widget data)
4. Clear browser cache and reload

### High Churn Rate

This may indicate:
- Data quality issues (canceled test subscriptions)
- Actual customer churn requiring attention
- Incorrect time period for calculation
- Mix of test and live mode data

### Build Errors

```bash
# Clean and rebuild
rm -rf build/
go clean -cache
go mod tidy
go build -o build/businessglance .
```

## Contributing

BusinessGlance is built on [Glance](https://github.com/glanceapp/glance). Contributions are welcome!

1. Fork the repository
2. Create a feature branch
3. Make your changes with tests
4. Run tests: `go test ./internal/glance -v`
5. Submit a pull request

## License

BusinessGlance inherits the AGPL-3.0 license from Glance. See LICENSE file for details.

## Support

- **Documentation**: See this README and `BUSINESSGLANCE_BUILD_PLAN.md`
- **Issues**: Report bugs via GitHub Issues
- **Glance Core**: https://github.com/glanceapp/glance

## Credits

Built with [Glance](https://github.com/glanceapp/glance) by the community.

## Changelog

### v1.0.0 (2025-11-17)

**Initial BusinessGlance Release**

- Revenue widget with MRR/ARR tracking
- Customer health metrics widget
- Stripe integration for subscription data
- Lightweight canvas-based charts
- Professional business theme
- Comprehensive test coverage (48+ test cases)
- Docker support
- Example configurations and documentation

---

**Built for business. Powered by Glance.**
