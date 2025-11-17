# BusinessGlance - Build Plan & Development Roadmap

**Project**: BusinessGlance - Business Metrics Dashboard
**Based On**: Glance (forked)
**Target**: SaaS Startups, Digital Agencies, SMBs
**Timeline**: 4-week MVP → 8-week full launch
**Inspired By**: Windsor.ai, Supermetrics (but better UX, self-hostable, broader metrics)

---

## Table of Contents

1. [Project Overview](#project-overview)
2. [What We're Building](#what-were-building)
3. [Competitive Positioning](#competitive-positioning)
4. [Technical Architecture](#technical-architecture)
5. [MVP Widget Specifications](#mvp-widget-specifications)
6. [Week-by-Week Development Plan](#week-by-week-development-plan)
7. [Code Structure](#code-structure)
8. [Implementation Checklist](#implementation-checklist)
9. [Launch Plan](#launch-plan)

---

## Project Overview

### The Problem

Businesses use **15-20 SaaS tools** (Stripe, Google Analytics, CRMs, support tools, etc.) with metrics scattered everywhere:
- **2-5 hours/week wasted** gathering data manually
- **50% of agency time** spent on client reporting
- **$300-5,600/min** lost revenue from downtime blindness
- **Multiple versions of truth** - same metrics calculated differently

### The Solution

**BusinessGlance** - A single dashboard showing all critical business metrics:
- ✅ Revenue (MRR, ARR, growth)
- ✅ Customers (total, new, churned, churn rate)
- ✅ Infrastructure (uptime, performance, costs)
- ✅ Custom integrations (any JSON API)
- ✅ Beautiful, actionable interface

### Market Validated

- **$72.35B SMB software market** (6.98% CAGR)
- **83% of SMBs** need better automation/visibility
- **64% of organizations** lack operational visibility
- **Competitors charge $59-499/mo** - we'll do $29-99/mo

---

## What We're Building

### MVP Scope (4 Weeks)

**5 Core Widgets:**

1. **Revenue Widget** 🆕 (NEW)
   - Stripe integration
   - MRR, ARR, growth rate, trend chart
   - Priority: P0 (highest)

2. **Customer Metrics Widget** 🆕 (NEW)
   - Stripe integration
   - Total/new/churned customers, churn rate, CAC, LTV
   - Priority: P0

3. **Custom API Widget** ✅ (ENHANCE EXISTING)
   - Connect to any JSON API
   - OAuth2 templates
   - Pre-built integrations (Plausible, PostHog)
   - Priority: P0

4. **Monitor Widget** ✅ (ENHANCE EXISTING)
   - Uptime monitoring
   - Response time charts
   - Historical uptime %
   - Priority: P0

5. **Server Stats Widget** ✅ (ENHANCE EXISTING)
   - CPU, memory, disk
   - Multi-server support
   - Cost estimation
   - Priority: P0

**Features:**
- ✅ Multi-page dashboards
- ✅ Responsive 3-column layout
- ✅ YAML configuration
- ✅ Hot reload
- ✅ Light/dark themes
- ✅ Basic authentication
- ✅ Chart visualizations (NEW)
- ✅ API key encryption (NEW)

### What We're NOT Building (Yet)

- ❌ Multi-tenancy (single user for MVP)
- ❌ Historical data storage (in-memory only for MVP)
- ❌ Alerts/notifications (Phase 2)
- ❌ Team collaboration (Phase 2)
- ❌ White-label (Phase 2)
- ❌ CRM integrations (Phase 2)
- ❌ Marketing tool integrations (Phase 2)

### Widgets We're REMOVING from Glance

- ❌ Clock widget (no business value)
- ❌ Weather widget (irrelevant)
- ❌ Bookmarks widget (browser does this)
- ❌ To-do widget (use dedicated tools)
- ❌ Twitch widgets (entertainment, not business)
- ❌ Videos widget (low business priority)
- ❌ Calendar widget (unless business events)
- ❌ Search widget (address bar works)

### Widgets We're KEEPING from Glance

- ✅ Custom API widget (ENHANCE)
- ✅ Monitor widget (ENHANCE)
- ✅ Server Stats widget (ENHANCE)
- ✅ Docker Containers widget
- ✅ GitHub Releases widget
- ✅ Repository widget
- ✅ RSS widget (for industry news only)
- ✅ Hacker News widget (tech industry pulse)
- ✅ HTML/iframe widget (flexibility)

---

## Competitive Positioning

### vs. Supermetrics

| Feature | Supermetrics | BusinessGlance |
|---------|--------------|----------------|
| **Pricing** | €29-$177+/mo | $29-99/mo |
| **Model** | Per-source (scales expensively) | Flat rate |
| **Focus** | Marketing data pipes (ETL) | Visual business dashboards |
| **Setup** | Complex, hours | 10 minutes |
| **Self-host** | No | Yes ✅ |
| **Integrations** | 150+ | Start with 10 essential, grow |

### vs. Windsor.ai

| Feature | Windsor.ai | BusinessGlance |
|---------|------------|----------------|
| **Pricing** | $23-$598/mo | $29-99/mo |
| **Model** | Volume-based (data rows) | Flat rate |
| **Focus** | Marketing attribution | All business metrics |
| **Metrics** | Marketing-only | Revenue, customers, infra, marketing |
| **Self-host** | No | Yes ✅ |
| **Free tier** | Limited (1 source) | Generous (5 widgets, core features) |

### Our Unique Value

1. ✅ **Self-Hostable** - Run on your infrastructure (data privacy, compliance)
2. ✅ **Broader Metrics** - Not just marketing, all business (revenue, infrastructure)
3. ✅ **Visual Dashboards** - Actionable insights, not just data pipes
4. ✅ **Simpler Pricing** - Flat $29-99/mo, no per-source/volume charges
5. ✅ **Faster Setup** - <10 minutes vs. hours
6. ✅ **Beautiful UX** - Modern, clean, professional
7. ✅ **Open Source Option** - Self-host for free (paid cloud option)

**Positioning**: *"The self-hostable business metrics dashboard that founders actually use"*

---

## Technical Architecture

### Technology Stack

**Backend** (Keep from Glance):
- **Language**: Go 1.24+
- **Framework**: Standard library (net/http)
- **Config**: YAML
- **Cache**: In-memory with configurable TTL

**Frontend** (Enhanced):
- **JavaScript**: Vanilla JS
- **Charts**: Chart.js (for metric visualization)
- **CSS**: Enhanced business theme
- **Icons**: Heroicons + business icons

**New Dependencies**:
```go
// Add to go.mod
github.com/stripe/stripe-go/v76  // Stripe SDK
golang.org/x/oauth2              // OAuth2 for future integrations
github.com/joho/godotenv         // .env file support
```

**New NPM (for frontend)**:
```json
// Chart.js for visualizations (CDN or bundled)
{
  "chart.js": "^4.4.0"
}
```

### Architecture Decisions

**KEEP from Glance** ✅
- Single binary deployment
- YAML configuration
- Hot reload
- Widget-based architecture
- In-memory caching
- Static asset embedding

**ADD for Business** 🆕
- Stripe SDK integration
- OAuth2 support (framework)
- Chart.js for visualizations
- Enhanced metric display components
- API key encryption
- Business-focused theme

**REMOVE** ❌
- Personal widgets (clock, weather, bookmarks, to-do)
- Entertainment widgets (Twitch)

### Project Structure

```
businessglance/
├── main.go                      # Entry point (unchanged)
├── go.mod / go.sum              # Dependencies (add Stripe, OAuth2)
├── config.yml                   # Example business config
├── .env.example                 # Environment variables template
│
├── internal/businessglance/     # Renamed from glance
│   ├── main.go                  # CLI & server (minimal changes)
│   ├── glance.go               # HTTP server (minimal changes)
│   ├── config.go               # YAML parsing (minimal changes)
│   ├── widget.go               # Widget interface (add new types)
│   │
│   ├── widget-revenue.go       # 🆕 NEW - Revenue widget
│   ├── widget-customers.go     # 🆕 NEW - Customer metrics widget
│   │
│   ├── widget-custom-api.go    # ✅ ENHANCE - Add OAuth2 templates
│   ├── widget-monitor.go       # ✅ ENHANCE - Add charts
│   ├── widget-server-stats.go  # ✅ ENHANCE - Multi-server, costs
│   │
│   ├── widget-docker-containers.go  # ✅ KEEP
│   ├── widget-releases.go           # ✅ KEEP
│   ├── widget-repository.go         # ✅ KEEP
│   ├── widget-rss.go                # ✅ KEEP
│   ├── widget-hacker-news.go        # ✅ KEEP
│   ├── widget-html.go               # ✅ KEEP
│   ├── widget-iframe.go             # ✅ KEEP
│   │
│   ├── stripe.go                # 🆕 NEW - Stripe API helper
│   ├── charts.go                # 🆕 NEW - Chart data helpers
│   ├── oauth.go                 # 🆕 NEW - OAuth2 framework
│   │
│   ├── auth.go                  # ✅ KEEP - Authentication
│   ├── theme.go                 # ✅ ENHANCE - Business theme
│   ├── embed.go                 # ✅ KEEP - Static assets
│   ├── templates.go             # ✅ KEEP - Template helpers
│   ├── utils.go                 # ✅ KEEP - Utilities
│   │
│   ├── static/                  # Frontend assets
│   │   ├── css/
│   │   │   ├── business.css     # 🆕 NEW - Business theme
│   │   │   ├── charts.css       # 🆕 NEW - Chart styles
│   │   │   └── [other css]      # ✅ KEEP existing
│   │   ├── js/
│   │   │   ├── charts.js        # 🆕 NEW - Chart.js integration
│   │   │   └── [other js]       # ✅ KEEP existing
│   │   └── [icons, fonts]       # ✅ KEEP
│   │
│   └── templates/               # HTML templates
│       ├── widgets/
│       │   ├── revenue.html         # 🆕 NEW
│       │   ├── customers.html       # 🆕 NEW
│       │   └── [other widgets]      # ✅ KEEP/ENHANCE
│       └── [page templates]         # ✅ KEEP
│
├── docs/                        # Documentation
│   ├── BUSINESSGLANCE_BUILD_PLAN.md         # This file
│   ├── BUSINESS_DASHBOARD_MARKET_RESEARCH.md
│   ├── BUSINESS_DASHBOARD_IMPLEMENTATION_PLAN.md
│   └── setup-guide.md           # 🆕 NEW - Quick start for users
│
└── examples/                    # Example configs
    ├── saas-startup.yml         # 🆕 NEW - SaaS startup example
    ├── digital-agency.yml       # 🆕 NEW - Agency example
    └── smb.yml                  # 🆕 NEW - SMB example
```

---

## MVP Widget Specifications

### 1. Revenue Widget 🆕

**Purpose**: Display MRR, ARR, and growth rate

**File**: `internal/businessglance/widget-revenue.go`

**Data Source**: Stripe API

**Configuration**:
```yaml
- type: revenue
  title: Monthly Recurring Revenue
  stripe:
    api-key: ${STRIPE_SECRET_KEY}
    mode: live  # or 'test'
  cache: 1h
```

**Display**:
```
┌─────────────────────────────────────┐
│ Monthly Recurring Revenue    [•••]  │
├─────────────────────────────────────┤
│                                     │
│        $125,450                     │  <- Large number
│        Current MRR                  │  <- Label
│                                     │
│        ↑ 12.5%  vs last month      │  <- Trend indicator
│                                     │
│   ┌──────────────────────────────┐ │
│   │   [Line chart - 12 months]   │ │  <- Trend chart
│   └──────────────────────────────┘ │
│                                     │
│  Metrics:                           │
│  • ARR: $1,505,400                 │
│  • New MRR: $15,230 ↑              │
│  • Churned MRR: $3,200 ↓           │
│  • Net New MRR: $12,030            │
│                                     │
└─────────────────────────────────────┘
```

**Implementation**:

```go
type revenueWidget struct {
    widgetBase      `yaml:",inline"`
    StripeAPIKey    string `yaml:"stripe-api-key"`
    StripeMode      string `yaml:"stripe-mode"`  // 'live' or 'test'

    // Data
    CurrentMRR      float64         `yaml:"-"`
    PreviousMRR     float64         `yaml:"-"`
    GrowthRate      float64         `yaml:"-"`
    ARR             float64         `yaml:"-"`
    NewMRR          float64         `yaml:"-"`
    ChurnedMRR      float64         `yaml:"-"`
    TrendData       []ChartPoint    `yaml:"-"`
}

type ChartPoint struct {
    Month string
    Value float64
}

func (w *revenueWidget) update(ctx context.Context) {
    // 1. Initialize Stripe client
    stripe.Key = w.StripeAPIKey

    // 2. Fetch active subscriptions
    params := &stripe.SubscriptionListParams{}
    params.Status = stripe.String("active")
    subs := subscription.List(params)

    // 3. Calculate MRR
    currentMRR := 0.0
    for subs.Next() {
        sub := subs.Subscription()
        // Sum all subscription amounts (normalized to monthly)
        for _, item := range sub.Items.Data {
            amount := float64(item.Price.UnitAmount) / 100.0
            interval := item.Price.Recurring.Interval

            if interval == "month" {
                currentMRR += amount
            } else if interval == "year" {
                currentMRR += amount / 12.0
            }
        }
    }

    w.CurrentMRR = currentMRR
    w.ARR = currentMRR * 12

    // 4. Calculate growth (compare to last month)
    // TODO: Store historical data or fetch from Stripe events
    w.GrowthRate = ((w.CurrentMRR - w.PreviousMRR) / w.PreviousMRR) * 100

    // 5. Fetch trend data (last 12 months)
    // TODO: Query Stripe for historical subscription data
    w.TrendData = fetchMRRTrend()
}

func (w *revenueWidget) Render() template.HTML {
    return w.render(w, assets.RevenueTemplate)
}
```

**Template** (`templates/widgets/revenue.html`):
```html
{{ template "widget-base" .options }}
    <div class="revenue-widget">
        <div class="metric-primary">
            <div class="metric-value">${{ formatMoney .currentMRR }}</div>
            <div class="metric-label">Current MRR</div>
        </div>

        <div class="metric-trend">
            <span class="trend-indicator {{ if gt .growthRate 0 }}positive{{ else }}negative{{ end }}">
                {{ if gt .growthRate 0 }}↑{{ else }}↓{{ end }}
                {{ formatPercent .growthRate }}%
            </span>
            <span class="trend-label">vs last month</span>
        </div>

        <div class="chart-container">
            <canvas id="mrr-chart-{{ .id }}" width="400" height="200"></canvas>
        </div>

        <div class="metrics-list">
            <div class="metric-item">
                <span class="metric-label">ARR</span>
                <span class="metric-value">${{ formatMoney .arr }}</span>
            </div>
            <div class="metric-item">
                <span class="metric-label">New MRR</span>
                <span class="metric-value positive">${{ formatMoney .newMRR }} ↑</span>
            </div>
            <div class="metric-item">
                <span class="metric-label">Churned MRR</span>
                <span class="metric-value negative">${{ formatMoney .churnedMRR }} ↓</span>
            </div>
            <div class="metric-item">
                <span class="metric-label">Net New MRR</span>
                <span class="metric-value">${{ formatMoney .netNewMRR }}</span>
            </div>
        </div>
    </div>

    <script>
    // Initialize Chart.js
    const ctx = document.getElementById('mrr-chart-{{ .id }}').getContext('2d');
    new Chart(ctx, {
        type: 'line',
        data: {
            labels: {{ .trendLabels }},
            datasets: [{
                label: 'MRR',
                data: {{ .trendValues }},
                borderColor: 'rgb(59, 130, 246)',
                tension: 0.1
            }]
        },
        options: {
            responsive: true,
            plugins: {
                legend: { display: false }
            }
        }
    });
    </script>
{{ template "widget-base-end" .options }}
```

**API Calls**:
1. `stripe.Subscription.List()` - Get active subscriptions
2. Calculate MRR from subscription amounts
3. (Future) Fetch historical data for trends

**Cache**: 1 hour (configurable)

---

### 2. Customer Metrics Widget 🆕

**Purpose**: Display customer count, churn, CAC, LTV

**File**: `internal/businessglance/widget-customers.go`

**Data Source**: Stripe API

**Configuration**:
```yaml
- type: customers
  title: Customer Metrics
  stripe:
    api-key: ${STRIPE_SECRET_KEY}
  cache: 1h
```

**Display**:
```
┌─────────────────────────────────────┐
│ Customer Metrics          [•••]     │
├─────────────────────────────────────┤
│                                     │
│        1,247                        │  <- Total customers
│        Total Customers              │
│                                     │
│  This Month:                        │
│  • New: 45 ↑                       │
│  • Churned: 12 ↓                   │
│  • Churn Rate: 0.96%               │
│                                     │
│   ┌──────────────────────────────┐ │
│   │ [Bar chart - monthly growth] │ │
│   └──────────────────────────────┘ │
│                                     │
│  Key Metrics:                       │
│  • CAC: $125                       │
│  • LTV: $1,850                     │
│  • LTV/CAC: 14.8x                  │
│                                     │
└─────────────────────────────────────┘
```

**Implementation**:
```go
type customersWidget struct {
    widgetBase      `yaml:",inline"`
    StripeAPIKey    string `yaml:"stripe-api-key"`

    // Data
    TotalCustomers  int     `yaml:"-"`
    NewCustomers    int     `yaml:"-"`
    ChurnedCustomers int    `yaml:"-"`
    ChurnRate       float64 `yaml:"-"`
    CAC             float64 `yaml:"-"`
    LTV             float64 `yaml:"-"`
    TrendData       []ChartPoint `yaml:"-"`
}

func (w *customersWidget) update(ctx context.Context) {
    stripe.Key = w.StripeAPIKey

    // 1. Get total active customers
    params := &stripe.CustomerListParams{}
    customers := customer.List(params)

    totalCount := 0
    for customers.Next() {
        totalCount++
    }
    w.TotalCustomers = totalCount

    // 2. Get customers created this month
    now := time.Now()
    startOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)

    newParams := &stripe.CustomerListParams{}
    newParams.Filters.AddFilter("created", "gte", startOfMonth.Unix())
    newCustomers := customer.List(newParams)

    newCount := 0
    for newCustomers.Next() {
        newCount++
    }
    w.NewCustomers = newCount

    // 3. Get churned customers (canceled subscriptions this month)
    subParams := &stripe.SubscriptionListParams{}
    subParams.Status = stripe.String("canceled")
    subParams.Filters.AddFilter("canceled_at", "gte", startOfMonth.Unix())
    canceledSubs := subscription.List(subParams)

    churnedCount := 0
    for canceledSubs.Next() {
        churnedCount++
    }
    w.ChurnedCustomers = churnedCount

    // 4. Calculate churn rate
    if w.TotalCustomers > 0 {
        w.ChurnRate = (float64(w.ChurnedCustomers) / float64(w.TotalCustomers)) * 100
    }

    // 5. Calculate LTV (requires revenue data)
    // LTV = Average Revenue Per Customer / Churn Rate
    // Simplified: LTV = MRR / Total Customers / (Churn Rate / 100)

    // 6. CAC - This requires ad spend data (future integration)
    // For now, allow manual input or calculate from Stripe metadata
}
```

---

### 3. Custom API Widget (Enhanced) ✅

**Enhancements**:
1. Add OAuth2 templates for common services
2. Pre-built integrations (Plausible, PostHog, etc.)
3. Better error handling
4. Data transformation helpers

**New Features**:
```yaml
- type: custom-api
  title: Plausible Analytics
  preset: plausible  # 🆕 NEW - Pre-built integration
  api-key: ${PLAUSIBLE_API_KEY}
  site-id: example.com
  metrics:
    - visitors
    - pageviews
    - bounce_rate
```

**Pre-built Presets**:
- `plausible` - Plausible Analytics
- `posthog` - PostHog
- `fathom` - Fathom Analytics
- `generic` - Any JSON API (existing behavior)

---

### 4. Monitor Widget (Enhanced) ✅

**Enhancements**:
1. Response time chart
2. Historical uptime %
3. Better status indicators

**New Display**:
```yaml
- type: monitor
  title: Uptime Monitor
  sites:
    - url: https://api.example.com
      name: API Server
    - url: https://app.example.com
      name: Web App
  cache: 1m
```

**Enhanced UI**:
- Chart showing response time over last hour
- 30-day uptime percentage
- Last check timestamp

---

### 5. Server Stats Widget (Enhanced) ✅

**Enhancements**:
1. Multi-server support
2. Cost estimation (if cloud provider API)
3. Historical trends

**New Configuration**:
```yaml
- type: server-stats
  title: Infrastructure
  servers:
    - name: Production
      host: prod.example.com
      ssh-key: ${SSH_KEY}  # Optional for remote
    - name: Staging
      host: staging.example.com
  show-costs: true  # 🆕 NEW
  cloud-provider: aws  # 🆕 NEW - For cost estimation
  cache: 5m
```

---

## Week-by-Week Development Plan

### Week 1: Foundation & Revenue Widget

**Goal**: Setup project, build Revenue Widget

**Day 1-2: Project Setup**
- [ ] Fork Glance repository
- [ ] Rename to BusinessGlance
- [ ] Remove personal widgets (clock, weather, bookmarks, to-do, Twitch)
- [ ] Update go.mod (add Stripe SDK, OAuth2)
- [ ] Setup .env file support
- [ ] Update README.md

**Day 3-4: Revenue Widget**
- [ ] Create `widget-revenue.go`
- [ ] Implement Stripe SDK integration
- [ ] Calculate MRR, ARR, growth rate
- [ ] Create revenue widget template
- [ ] Test with Stripe test API

**Day 5: Chart Integration**
- [ ] Add Chart.js to frontend
- [ ] Create chart helper functions
- [ ] Implement MRR trend chart
- [ ] Test chart rendering

**Deliverable**: Revenue Widget working with Stripe

---

### Week 2: Customer Metrics & Business Theme

**Goal**: Complete Customer Metrics Widget, create business theme

**Day 1-2: Customer Metrics Widget**
- [ ] Create `widget-customers.go`
- [ ] Implement customer counting logic
- [ ] Calculate churn rate
- [ ] Create customer widget template
- [ ] Test with Stripe test API

**Day 3-4: Business Theme**
- [ ] Design business color scheme
- [ ] Create `business.css`
- [ ] Update widget templates for business style
- [ ] Design metric display components
- [ ] Add trend indicators (↑↓)

**Day 5: Testing & Polish**
- [ ] Test Revenue + Customer widgets together
- [ ] Fix styling issues
- [ ] Test on mobile
- [ ] Optimize performance

**Deliverable**: 2 working widgets, business theme

---

### Week 3: Enhanced Widgets & Monitoring

**Goal**: Enhance existing widgets, improve monitoring

**Day 1-2: Custom API Widget Enhancement**
- [ ] Add preset system
- [ ] Implement Plausible preset
- [ ] Implement PostHog preset
- [ ] Add OAuth2 framework (basic)
- [ ] Better error handling

**Day 3: Monitor Widget Enhancement**
- [ ] Add response time tracking
- [ ] Create response time chart
- [ ] Calculate 30-day uptime %
- [ ] Improve status indicators

**Day 4: Server Stats Widget Enhancement**
- [ ] Add multi-server support
- [ ] Create historical trend charts
- [ ] Add cost estimation (basic)

**Day 5: Integration Testing**
- [ ] Test all 5 widgets together
- [ ] Create example dashboard configs
- [ ] Performance testing
- [ ] Mobile testing

**Deliverable**: 5 polished widgets

---

### Week 4: Documentation, Testing & Launch Prep

**Goal**: Polish, document, prepare for beta launch

**Day 1-2: Documentation**
- [ ] Write setup guide
- [ ] Create video walkthrough
- [ ] Write integration guides (Stripe)
- [ ] Create example configs:
  - [ ] saas-startup.yml
  - [ ] digital-agency.yml
  - [ ] smb.yml
- [ ] Write troubleshooting guide

**Day 3: Testing**
- [ ] Unit tests for revenue calculations
- [ ] Integration tests for Stripe
- [ ] End-to-end tests
- [ ] Security audit (API keys)
- [ ] Load testing

**Day 4: Demo & Marketing**
- [ ] Create demo dashboard (with fake data)
- [ ] Take screenshots
- [ ] Record demo video
- [ ] Write launch blog post
- [ ] Prepare Product Hunt listing

**Day 5: Beta Launch**
- [ ] Deploy demo instance
- [ ] Recruit 10-20 beta users
- [ ] Setup feedback collection
- [ ] Monitor for bugs

**Deliverable**: MVP ready for beta users

---

## Code Structure

### New Files to Create

**Backend (Go)**:
```
internal/businessglance/
├── widget-revenue.go       # Revenue Widget (NEW)
├── widget-customers.go     # Customer Metrics Widget (NEW)
├── stripe.go               # Stripe API helpers (NEW)
├── charts.go               # Chart data helpers (NEW)
└── oauth.go                # OAuth2 framework (NEW)
```

**Frontend (CSS/JS)**:
```
static/
├── css/
│   ├── business.css        # Business theme (NEW)
│   └── charts.css          # Chart styles (NEW)
└── js/
    └── charts.js           # Chart.js integration (NEW)
```

**Templates**:
```
templates/widgets/
├── revenue.html            # Revenue Widget template (NEW)
└── customers.html          # Customer Metrics template (NEW)
```

**Configuration Examples**:
```
examples/
├── saas-startup.yml        # SaaS startup config (NEW)
├── digital-agency.yml      # Agency config (NEW)
└── smb.yml                 # SMB config (NEW)
```

### Files to Modify

**Minimal Changes**:
- `internal/businessglance/widget.go` - Add new widget types to factory
- `internal/businessglance/theme.go` - Add business theme colors
- `go.mod` - Add new dependencies

**No Changes Needed**:
- `main.go` - Keep as is
- `glance.go` - Keep as is (maybe rename package)
- `config.go` - Keep as is
- `auth.go` - Keep as is

### Files to Remove

**Widgets to Delete**:
```
internal/businessglance/
├── widget-clock.go         # ❌ DELETE
├── widget-weather.go       # ❌ DELETE
├── widget-bookmarks.go     # ❌ DELETE
├── widget-todo.go          # ❌ DELETE
├── widget-twitch-channels.go  # ❌ DELETE
├── widget-twitch-top-games.go # ❌ DELETE
├── widget-videos.go        # ❌ DELETE (low priority)
├── widget-search.go        # ❌ DELETE
└── widget-calendar.go      # ❌ DELETE (or keep for business events)
```

**Templates to Delete**:
```
templates/widgets/
├── clock.html              # ❌ DELETE
├── weather.html            # ❌ DELETE
├── bookmarks.html          # ❌ DELETE
└── [etc...]
```

---

## Implementation Checklist

### Phase 1: Setup (Day 1)

**Project Setup**:
- [ ] Clone/fork Glance repository
- [ ] Create new branch: `feature/businessglance-mvp`
- [ ] Rename package from `glance` to `businessglance`
- [ ] Update module name in go.mod
- [ ] Create .env.example file

**Dependencies**:
- [ ] Add to go.mod:
  ```
  github.com/stripe/stripe-go/v76
  golang.org/x/oauth2
  github.com/joho/godotenv
  ```
- [ ] Run `go mod download`

**Cleanup**:
- [ ] Delete personal widget files (clock, weather, etc.)
- [ ] Delete personal widget templates
- [ ] Update widget factory to remove deleted types

**Testing**:
- [ ] Verify app still compiles
- [ ] Run existing widgets (RSS, GitHub, etc.)
- [ ] Confirm nothing broke

---

### Phase 2: Revenue Widget (Days 2-5)

**Backend**:
- [ ] Create `widget-revenue.go`
- [ ] Define `revenueWidget` struct
- [ ] Implement `initialize()` method
- [ ] Implement `update()` method:
  - [ ] Initialize Stripe client
  - [ ] Fetch active subscriptions
  - [ ] Calculate MRR
  - [ ] Calculate ARR
  - [ ] Calculate growth rate
  - [ ] Fetch trend data (12 months)
- [ ] Implement `Render()` method
- [ ] Add to widget factory in `widget.go`

**Frontend**:
- [ ] Create `templates/widgets/revenue.html`
- [ ] Add large metric display
- [ ] Add trend indicator (↑↓ %)
- [ ] Add Chart.js canvas
- [ ] Add metrics list
- [ ] Style with business theme

**Chart Integration**:
- [ ] Add Chart.js to `static/js/charts.js`
- [ ] Create helper function for MRR chart
- [ ] Pass data from Go to JavaScript
- [ ] Test chart rendering

**Testing**:
- [ ] Test with Stripe test mode
- [ ] Verify MRR calculation
- [ ] Verify chart displays correctly
- [ ] Test on mobile

**Documentation**:
- [ ] Add configuration example to docs
- [ ] Document required Stripe permissions

---

### Phase 3: Customer Metrics Widget (Days 6-10)

**Backend**:
- [ ] Create `widget-customers.go`
- [ ] Define `customersWidget` struct
- [ ] Implement `initialize()` method
- [ ] Implement `update()` method:
  - [ ] Count total customers
  - [ ] Count new customers (this month)
  - [ ] Count churned customers
  - [ ] Calculate churn rate
  - [ ] Calculate CAC (if data available)
  - [ ] Calculate LTV
- [ ] Implement `Render()` method
- [ ] Add to widget factory

**Frontend**:
- [ ] Create `templates/widgets/customers.html`
- [ ] Add customer count display
- [ ] Add churn metrics
- [ ] Add CAC/LTV metrics
- [ ] Add growth chart
- [ ] Style consistently

**Testing**:
- [ ] Test with Stripe test mode
- [ ] Verify calculations
- [ ] Test edge cases (no customers, 100% churn)
- [ ] Mobile testing

---

### Phase 4: Business Theme (Days 11-12)

**Design**:
- [ ] Define color palette (blues, greens, reds for metrics)
- [ ] Design typography scale
- [ ] Create spacing system
- [ ] Design component library

**Implementation**:
- [ ] Create `static/css/business.css`
- [ ] Define CSS variables for colors
- [ ] Style metric displays
- [ ] Style trend indicators
- [ ] Style charts
- [ ] Ensure dark mode support

**Components**:
- [ ] `.metric-primary` - Large primary metric
- [ ] `.metric-trend` - Trend indicator
- [ ] `.metric-list` - List of secondary metrics
- [ ] `.chart-container` - Chart wrapper
- [ ] `.status-indicator` - Status badges

**Testing**:
- [ ] Test light mode
- [ ] Test dark mode
- [ ] Test on different screen sizes
- [ ] Test accessibility (contrast ratios)

---

### Phase 5: Enhanced Widgets (Days 13-17)

**Custom API Widget**:
- [ ] Add preset system
- [ ] Create Plausible preset
- [ ] Create PostHog preset
- [ ] Add OAuth2 framework
- [ ] Improve error messages

**Monitor Widget**:
- [ ] Add response time tracking
- [ ] Create response time chart
- [ ] Calculate uptime percentage
- [ ] Add last check timestamp
- [ ] Improve status indicators

**Server Stats Widget**:
- [ ] Add multi-server configuration
- [ ] Implement remote server monitoring
- [ ] Add historical trend charts
- [ ] Add basic cost estimation
- [ ] Style improvements

---

### Phase 6: Testing & Polish (Days 18-20)

**Testing**:
- [ ] Write unit tests for Stripe calculations
- [ ] Write integration tests
- [ ] End-to-end testing
- [ ] Security audit:
  - [ ] Verify API keys encrypted
  - [ ] Check for injection vulnerabilities
  - [ ] Audit authentication
- [ ] Performance testing:
  - [ ] Load test (100 concurrent users)
  - [ ] Memory profiling
  - [ ] Optimize slow queries

**Polish**:
- [ ] Fix all visual bugs
- [ ] Ensure consistent styling
- [ ] Add loading states
- [ ] Add error states
- [ ] Improve mobile UX

**Example Configs**:
- [ ] Create `examples/saas-startup.yml`
- [ ] Create `examples/digital-agency.yml`
- [ ] Create `examples/smb.yml`

---

### Phase 7: Documentation (Days 21-23)

**Setup Guide**:
- [ ] Write quick start (10 minutes to dashboard)
- [ ] Document installation (binary, Docker)
- [ ] Document Stripe setup
- [ ] Document configuration options
- [ ] Add screenshots

**Integration Guides**:
- [ ] Stripe integration guide
- [ ] Custom API guide
- [ ] Server monitoring guide
- [ ] Troubleshooting guide

**Video Content**:
- [ ] Record setup walkthrough (5 minutes)
- [ ] Record dashboard tour (3 minutes)
- [ ] Record Stripe integration (5 minutes)

**Marketing Content**:
- [ ] Write launch blog post
- [ ] Create Product Hunt description
- [ ] Create demo screenshots
- [ ] Create comparison table (vs. competitors)

---

### Phase 8: Launch Prep (Days 24-28)

**Demo**:
- [ ] Create demo instance with fake data
- [ ] Deploy to demo.businessglance.com
- [ ] Test public access
- [ ] Create demo login credentials

**Beta Recruitment**:
- [ ] Post in SaaS communities
- [ ] Email potential beta users
- [ ] Create waitlist form
- [ ] Setup feedback collection (TypeForm, Google Forms)

**Launch Materials**:
- [ ] Finalize Product Hunt listing
- [ ] Write Hacker News post
- [ ] Prepare Reddit posts
- [ ] Create Twitter/X thread
- [ ] Setup analytics (Plausible)

**Final Checks**:
- [ ] All tests passing
- [ ] No critical bugs
- [ ] Documentation complete
- [ ] Demo working
- [ ] Ready for beta users

---

## Launch Plan

### Pre-Launch (Weeks 1-4)

**Build Audience**:
1. Create landing page with waitlist
2. Share progress on Twitter/X (build in public)
3. Post weekly updates on Indie Hackers
4. Recruit beta users (target: 20)

**Content**:
- Blog: "Why we're building BusinessGlance"
- Video: First working prototype
- Twitter: Development updates

---

### Launch Week (Week 5)

**Day 1 - Product Hunt**:
- Launch at 12:01am PT
- Engage with comments all day
- Share in communities
- Email waitlist

**Day 2 - Hacker News**:
- Post "Show HN: BusinessGlance - Self-hostable business metrics dashboard"
- Respond to every comment
- Be technical and transparent

**Day 3 - Reddit**:
- r/SaaS, r/entrepreneur, r/startups
- Focus on value, not sales
- Share demo, get feedback

**Day 4-5 - Follow-up**:
- Email waitlist with launch details
- Twitter thread with metrics
- Thank beta users
- Collect feedback

---

### Post-Launch (Weeks 6-8)

**Content Marketing**:
1. "How to track SaaS metrics with Stripe"
2. "Self-hosted vs. Cloud: Business dashboards"
3. "5-minute setup: Business metrics dashboard"

**Growth**:
- Setup affiliate program
- Partner with integration providers
- Guest posts on SaaS blogs
- YouTube tutorials

**Goals**:
- Week 6: 50 users, 5 customers
- Week 8: 100 users, 10 customers
- Week 12: 500 users, 50 customers

---

## Success Metrics

### MVP Launch (Week 4)
- ✅ Build completed
- ✅ 5 widgets working
- ✅ Documentation complete
- ✅ Demo deployed

### Beta (Week 6)
- ✅ 20 beta users recruited
- ✅ NPS >30
- ✅ Setup time <10 minutes
- ✅ 5 paying customers ($150 MRR)

### Public Launch (Week 8)
- ✅ 100 active users
- ✅ 10 paying customers ($500 MRR)
- ✅ Product Hunt top 5
- ✅ 50+ GitHub stars

### Growth (Month 3)
- ✅ 500 active users
- ✅ 50 paying customers ($3K MRR)
- ✅ <7% churn
- ✅ Case studies published

---

## Pricing (Reminder)

| Tier | Price | Features |
|------|-------|----------|
| **Free** | $0 | 1 dashboard, 5 widgets, community support |
| **Starter** | $29/mo | 3 dashboards, 20 widgets, email support |
| **Pro** | $99/mo | 10 dashboards, unlimited widgets, priority support |
| **Business** | $299/mo | Unlimited, white-label, dedicated support |

---

## Next Steps - Ready to Code! 🚀

### Right Now (Today)

1. **Fork Glance**
   ```bash
   cd /home/user
   cp -r glance businessglance
   cd businessglance
   ```

2. **Initial Cleanup**
   ```bash
   # Remove personal widgets
   rm internal/glance/widget-clock.go
   rm internal/glance/widget-weather.go
   rm internal/glance/widget-bookmarks.go
   # ... etc
   ```

3. **Add Dependencies**
   ```bash
   go get github.com/stripe/stripe-go/v76
   go get golang.org/x/oauth2
   go get github.com/joho/godotenv
   ```

4. **Start Building Revenue Widget**
   ```bash
   # Create new file
   touch internal/glance/widget-revenue.go

   # Start coding!
   ```

---

**Document Version**: 1.0
**Created**: 2025-11-16
**Status**: ✅ Ready to Code
**Timeline**: 4 weeks to MVP
**Next Action**: Start coding Revenue Widget

---

Let's build this! 🚀
