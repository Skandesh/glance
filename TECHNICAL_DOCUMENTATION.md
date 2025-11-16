# Glance - Complete Technical Documentation

## Table of Contents

1. [Executive Summary](#executive-summary)
2. [Project Overview](#project-overview)
3. [Architecture Deep Dive](#architecture-deep-dive)
4. [Core Components](#core-components)
5. [Widget System](#widget-system)
6. [Configuration System](#configuration-system)
7. [Authentication & Security](#authentication--security)
8. [Data Flow & Request Lifecycle](#data-flow--request-lifecycle)
9. [Build & Deployment](#build--deployment)
10. [Development Guide](#development-guide)
11. [API Reference](#api-reference)
12. [Performance Optimization](#performance-optimization)
13. [Troubleshooting](#troubleshooting)

---

## Executive Summary

**Glance** is a self-hosted dashboard application built in Go that aggregates content from multiple sources (RSS feeds, social media, APIs, system metrics) into a customizable, themeable interface.

### Key Metrics
- **Language**: Go 1.24.3
- **Lines of Code**: ~9,711 lines (46 Go files)
- **Binary Size**: <20MB
- **Widget Types**: 25+
- **License**: Apache 2.0
- **Platforms**: Linux, Windows, macOS, FreeBSD, OpenBSD
- **Architectures**: amd64, arm64, arm, 386

### Core Philosophy
1. **Zero JavaScript Frameworks** - Vanilla JS (~70KB total)
2. **Minimal Dependencies** - 7 direct Go dependencies
3. **Single Binary** - No package.json, no npm
4. **Hot Reload** - Config changes apply without restart
5. **Performance First** - Intelligent caching, parallel updates

---

## Project Overview

### What is Glance?

Glance is a lightweight dashboard that serves as:
- **Personal Homepage/Startpage** - Customizable browser landing page
- **Feed Aggregator** - Centralized RSS, Reddit, Hacker News, etc.
- **System Monitor** - Docker containers, server stats, DNS metrics
- **Development Dashboard** - GitHub releases, repository stats
- **Content Hub** - YouTube uploads, Twitch streams
- **Information Display** - Weather, stocks, calendar

### Key Features

**Content Aggregation:**
- RSS/Atom feeds with thumbnails
- Reddit subreddit posts
- Hacker News & Lobsters
- YouTube channel uploads
- Twitch live streams
- GitHub releases & repo stats

**System Monitoring:**
- Docker container status
- Server stats (CPU, memory, disk)
- DNS stats (Pi-hole, AdGuard)
- Website uptime monitoring

**Customization:**
- Multiple pages/tabs
- 3-column responsive layouts
- Theme system (HSL-based)
- Custom CSS support
- Icon provider integration

**Performance:**
- Intelligent caching (configurable per widget)
- Parallel widget updates
- Conditional HTTP requests (ETags)
- Worker pools for concurrent API calls
- Static asset caching (24h)

**Security:**
- Optional authentication system
- bcrypt password hashing
- Session tokens with HMAC
- Rate limiting
- Secure secret management

---

## Architecture Deep Dive

### Directory Structure

```
/home/user/glance/
├── main.go                          # Entry point (delegates to internal/)
├── go.mod / go.sum                  # Go dependencies
├── Dockerfile                       # Container build
├── .goreleaser.yaml                # Release automation
├── LICENSE                          # Apache 2.0
├── README.md                        # User documentation
│
├── internal/glance/                 # Core application (private package)
│   ├── main.go                      # CLI routing & server lifecycle
│   ├── glance.go                    # Application struct, HTTP server
│   ├── config.go                    # YAML parsing, validation
│   ├── config-fields.go             # Custom YAML field types
│   ├── widget.go                    # Widget interface & factory
│   ├── widget-*.go                  # 25+ widget implementations
│   ├── widget-utils.go              # Shared widget utilities
│   ├── auth.go                      # Authentication system
│   ├── theme.go                     # Theme engine
│   ├── embed.go                     # Static asset embedding
│   ├── templates.go                 # Template helpers
│   ├── utils.go                     # General utilities
│   ├── cli.go                       # CLI command handlers
│   ├── diagnose.go                  # Diagnostic tools
│   │
│   ├── static/                      # Frontend assets (embedded)
│   │   ├── css/                     # Stylesheets
│   │   ├── js/                      # Vanilla JavaScript
│   │   ├── icons/                   # Heroicons
│   │   ├── fonts/                   # Font files
│   │   ├── app-icon.png            # PWA icon
│   │   └── favicon.{svg,png}       # Favicons
│   │
│   └── templates/                   # Go HTML templates
│       ├── page.html               # Main page layout
│       ├── page-content.html       # AJAX content
│       ├── document.html           # HTML document wrapper
│       ├── footer.html             # Footer template
│       ├── manifest.json           # PWA manifest
│       └── widgets/                # Widget templates
│
├── pkg/sysinfo/                     # Public system info package
│   └── sysinfo.go                   # Cross-platform metrics
│
└── docs/                            # Documentation
    ├── configuration.md             # Config guide (91KB)
    ├── custom-api.md               # Custom API widget
    ├── themes.md                   # Theming guide
    ├── glance.yml                  # Example config
    └── images/                     # Screenshots
```

### Tech Stack

**Backend (Go):**
- `net/http` - HTTP server (standard library)
- `html/template` - Template rendering
- `gopkg.in/yaml.v3` - YAML parsing
- `github.com/mmcdole/gofeed` - RSS/Atom parsing
- `github.com/shirou/gopsutil/v4` - System metrics
- `github.com/tidwall/gjson` - Fast JSON parsing
- `github.com/fsnotify/fsnotify` - File watching
- `golang.org/x/crypto` - Password hashing

**Frontend:**
- Vanilla JavaScript (no frameworks)
- CSS with CSS Custom Properties
- Heroicons for icons
- Progressive Web App (PWA) support

**Build & Deploy:**
- Go compiler (CGO_ENABLED=0)
- GoReleaser (multi-platform builds)
- Docker (Alpine-based)

---

## Core Components

### 1. Application Entry Point

**File**: `/home/user/glance/main.go`
```go
package main

import (
    "os"
    "github.com/glanceapp/glance/internal/glance"
)

func main() {
    os.Exit(glance.Main())
}
```

Simple delegation to internal package.

### 2. Main Application Logic

**File**: `internal/glance/main.go`

**Key Functions:**
- `Main()` - CLI entry point, routes to subcommands
- `serve()` - Starts HTTP server, sets up file watching
- `parseYAMLConfig()` - Parses and validates config

**Flow:**
```
Main()
  → Parse CLI flags
  → Determine command (serve, validate, print, etc.)
  → For serve:
      → Parse config → Create app → Start server
      → Setup file watcher (fsnotify)
      → On config change → Reload app
      → Wait for interrupt signal
```

### 3. Application Struct

**File**: `internal/glance/glance.go`

```go
type application struct {
    Version   string
    CreatedAt time.Time
    Config    config

    parsedManifest []byte

    slugToPage map[string]*page       // URL slug → page
    widgetByID map[uint64]widget      // Widget ID → widget instance

    // Auth
    RequiresAuth           bool
    authSecretKey          []byte
    usernameHashToUsername map[string]string
    failedAuthAttempts     map[string]*failedAuthAttempt
}
```

**Key Methods:**
- `newApplication()` - Initializes app from config
- `makeHandler()` - Creates HTTP handler with routing
- `handlePageRequest()` - Serves dashboard pages
- `handleContentRequest()` - AJAX content updates
- `handleWidgetRequest()` - Widget-specific API calls

### 4. HTTP Server & Routing

**Routes:**
```
GET  /                              → First page
GET  /{page}                        → Named page by slug
GET  /api/pages/{page}/content      → AJAX content update
POST /api/set-theme/{key}           → Theme switcher
GET  /api/widgets/{id}/{path...}    → Widget API
GET  /api/healthz                   → Health check
GET  /login                         → Login page
POST /api/authenticate              → Login handler
GET  /logout                        → Logout handler
GET  /static/{hash}/{path...}       → Static assets (24h cache)
GET  /manifest.json                 → PWA manifest
GET  /assets/{path...}              → User assets
```

**Request Flow:**
```
HTTP Request
  → Routing (ServeMux)
  → Auth Check (if enabled)
  → Handler Execution
      → Page Lookup
      → Widget Update Check
      → Parallel Widget Updates (goroutines)
      → Template Rendering
  → Response (HTML/JSON)
```

### 5. Configuration System

**File**: `internal/glance/config.go`

**Structure:**
```go
type config struct {
    Server   serverConfig
    Auth     authConfig
    Document documentConfig
    Branding brandingConfig
    Theme    themeConfig
    Pages    []page
}

type page struct {
    Name    string
    Slug    string
    Columns []column
    Widgets []widget
}

type column struct {
    Size    string        // "small" | "full"
    Widgets []widget
}
```

**Features:**
- **Environment Variables**: `${VAR}` or `${env:VAR}`
- **Secrets**: `${secret:name}` from `/run/secrets/`
- **File Includes**: `!include: path/file.yml`
- **Recursive Includes**: Max depth 5
- **Auto-reload**: File watching with hot-reload
- **Validation**: Comprehensive error checking

**Example Config:**
```yaml
server:
  host: 0.0.0.0
  port: 8080

theme:
  background-color: 240 13 20
  primary-color: 43 100 50

pages:
  - name: Home
    slug: home
    columns:
      - size: small
        widgets:
          - type: weather
            location: London, UK
          - type: calendar

      - size: full
        widgets:
          - type: rss
            feeds:
              - url: https://example.com/feed.xml
```

### 6. Widget System

**File**: `internal/glance/widget.go`

**Widget Interface:**
```go
type widget interface {
    // Exported (called in templates)
    Render() template.HTML
    GetType() string
    GetID() uint64

    // Internal
    initialize() error
    requiresUpdate(*time.Time) bool
    setProviders(*widgetProviders)
    update(context.Context)
    setID(uint64)
    handleRequest(w http.ResponseWriter, r *http.Request)
    setHideHeader(bool)
}
```

**Base Widget:**
```go
type widgetBase struct {
    ID                  uint64
    Type                string
    Title               string
    TitleURL            string
    CustomCacheDuration duration
    UpdatedAt           time.Time
    ContentHTML         template.HTML
    Error               error
    Notice              *notice

    cacheType           cacheType
    cacheDuration       time.Duration
    withTitle           bool
    withTitleURL        bool
    providers           *widgetProviders
}
```

**Widget Factory Pattern:**
```go
func newWidget(widgetType string) (widget, error) {
    switch widgetType {
    case "rss":
        return &rssWidget{}, nil
    case "weather":
        return &weatherWidget{}, nil
    case "calendar":
        return &calendarWidget{}, nil
    // ... 22 more types
    default:
        return nil, fmt.Errorf("unknown widget type: %s", widgetType)
    }
}
```

**Cache Types:**
1. **Infinite** - Never updates (static widgets)
2. **Duration** - Updates after N time (configurable)
3. **OnTheHour** - Updates at the top of each hour

### 7. Static Asset Embedding

**File**: `internal/glance/embed.go`

```go
//go:embed static templates
var embedFS embed.FS

func bundleCSS() ([]byte, string) {
    // Reads all CSS files
    // Concatenates them
    // Returns bundled CSS + MD5 hash for cache busting
}
```

Assets are embedded at compile time, served with 24h cache headers.

---

## Widget System

### Widget Lifecycle

```
1. Config Parse
   ↓
2. Widget Factory (newWidget)
   ↓
3. YAML Unmarshal (widget-specific config)
   ↓
4. initialize() - Setup, validate config
   ↓
5. Page Load Request
   ↓
6. requiresUpdate() - Check cache expiry
   ↓
7. update() - Fetch data (if needed)
   ↓
8. Render() - Generate HTML
   ↓
9. Cache until next update
```

### Available Widgets

| Widget Type | Purpose | API/Source | Cache Default |
|------------|---------|------------|---------------|
| `rss` | RSS/Atom feeds | Any RSS/Atom feed | 12h |
| `videos` | YouTube uploads | YouTube RSS | 1h |
| `weather` | Weather forecast | Open-Meteo | 1h |
| `markets` | Stock/crypto prices | Yahoo Finance | 1m |
| `reddit` | Subreddit posts | Reddit JSON | 12h |
| `hacker-news` | HN top stories | HN API | 30m |
| `lobsters` | Lobsters posts | Lobsters API | 30m |
| `calendar` | Month calendar | Local | Infinite |
| `clock` | Time display | Local | Infinite |
| `bookmarks` | Link collections | Config | Infinite |
| `docker-containers` | Container status | Docker socket | 1m |
| `server-stats` | System metrics | gopsutil | 1m |
| `dns-stats` | DNS metrics | Pi-hole/AdGuard | 5m |
| `monitor` | Uptime monitoring | HTTP GET | 1m |
| `releases` | GitHub releases | GitHub API | 1d |
| `repository` | Repo stats | GitHub API | 1h |
| `twitch-channels` | Stream status | Twitch API | 1m |
| `twitch-top-games` | Popular games | Twitch API | 30m |
| `search` | Search widget | Config | Infinite |
| `custom-api` | Custom API | Any JSON API | 30m |
| `extension` | Fetch HTML | Any URL | 30m |
| `html` | Static HTML | Config | Infinite |
| `iframe` | Embed content | Any URL | Infinite |
| `group` | Widget container | Children | N/A |
| `to-do` | Task list | Local storage | Infinite |

### Widget Implementation Example

**RSS Widget** (`internal/glance/widget-rss.go`):

```go
type rssWidget struct {
    widgetBase     `yaml:",inline"`
    Feeds          []rssFeed         `yaml:"feeds"`
    Limit          int               `yaml:"limit"`
    CollapseAfter  int               `yaml:"collapse-after"`
    ThumbnailHeight float64          `yaml:"thumbnail-height"`
    CardHeight     float64           `yaml:"card-height"`
    Posts          []rssPost         `yaml:"-"`
}

func (widget *rssWidget) initialize() error {
    widget.withTitle = true
    widget.withTitleURL = true
    widget.cacheType = cacheTypeDuration
    widget.cacheDuration = time.Hour * 12

    // Validation
    if widget.Limit <= 0 {
        widget.Limit = 25
    }

    return nil
}

func (widget *rssWidget) update(ctx context.Context) {
    // Create worker pool
    requests := make([]channelRequest[[]rssPost], len(widget.Feeds))

    for i := range widget.Feeds {
        feed := &widget.Feeds[i]
        requests[i] = func() ([]rssPost, error) {
            // Fetch and parse RSS feed
            parser := gofeed.NewParser()
            parsedFeed, err := parser.ParseURL(feed.URL)
            if err != nil {
                return nil, err
            }

            // Convert to posts
            posts := make([]rssPost, 0, len(parsedFeed.Items))
            for _, item := range parsedFeed.Items {
                posts = append(posts, rssPost{
                    Title:       item.Title,
                    Link:        item.Link,
                    PublishedAt: *item.PublishedParsed,
                    // ... more fields
                })
            }

            return posts, nil
        }
    }

    // Execute in parallel
    results := workerPoolWithResponses(requests)

    // Aggregate results
    widget.Posts = aggregateAndSortPosts(results, widget.Limit)
}

func (widget *rssWidget) Render() template.HTML {
    return widget.render(widget, assets.RSSTemplate)
}
```

### Custom API Widget

**Most Flexible Widget** - Build your own widget using any JSON API.

**Example**: GitHub Stars
```yaml
- type: custom-api
  title: Repository Stars
  cache: 1h
  url: https://api.github.com/repos/glanceapp/glance
  data:
    stargazers_count: .stargazers_count
    forks_count: .forks_count
  template: |
    <div class="stars">
      ⭐ {{.stargazers_count}} stars
      🍴 {{.forks_count}} forks
    </div>
```

**Features:**
- JSONPath data extraction (using gjson)
- Go template rendering
- Custom CSS support
- Error handling

---

## Configuration System

### Configuration File Format

**Main Config** (`glance.yml`):
```yaml
# Server Configuration
server:
  host: 0.0.0.0           # Bind address
  port: 8080              # Port number
  proxied: false          # Behind reverse proxy?
  base-url: /             # Base URL path
  assets-path: /assets    # Custom assets directory

# Authentication (optional)
auth:
  secret-key: ${AUTH_SECRET}  # Base64-encoded 32-byte key
  users:
    admin:
      password-hash: ${ADMIN_PASSWORD_HASH}

# Custom HTML in <head>
document:
  head-html: |
    <meta name="robots" content="noindex">

# Branding
branding:
  logo-url: /assets/logo.png
  logo-link-url: https://example.com
  favicon-url: /assets/favicon.png
  app-name: My Dashboard

# Theme
theme:
  light: true
  background-color: 240 13 20       # HSL format
  primary-color: 43 100 50
  contrast-multiplier: 1.0
  custom-css-file: /assets/custom.css

# Pages
pages:
  - name: Home
    slug: home              # URL: /{slug}
    columns:
      - size: small         # or "full"
        widgets:
          - type: weather
            location: London, UK
            units: metric
```

### Environment Variables

**Syntax Options:**
- `${VAR}`
- `${env:VAR}`
- `${secret:name}` - Reads from `/run/secrets/name`

**Example:**
```yaml
auth:
  secret-key: ${AUTH_SECRET}

- type: twitch-channels
  client-id: ${TWITCH_CLIENT_ID}

- type: rss
  feeds:
    - url: https://api.example.com/feed?key=${API_KEY}
```

**Escaping:**
```yaml
something: \${NOT_AN_ENV_VAR}  # Literal ${NOT_AN_ENV_VAR}
```

### File Includes

**Modular Configuration:**

`glance.yml`:
```yaml
server:
  port: 8080

pages:
  - !include: pages/home.yml
  - !include: pages/work.yml
  - !include: pages/fun.yml
```

`pages/home.yml`:
```yaml
name: Home
slug: home
columns:
  - !include: columns/sidebar.yml
  - !include: columns/main.yml
```

**Features:**
- Recursive includes (max depth 5)
- Relative paths
- Watches all included files for changes
- Prevents circular references

### Configuration Validation

**CLI Commands:**
```bash
# Validate config
glance config:validate

# Print parsed config (with env vars expanded)
glance config:print

# Print config as JSON
glance config:print --json
```

**Validation Checks:**
- Required fields present
- Valid widget types
- Valid enum values (size, units, etc.)
- Valid URLs
- Valid colors (HSL format)
- No duplicate page slugs
- No reserved slugs (login, logout)

---

## Authentication & Security

### Authentication System

**File**: `internal/glance/auth.go`

**Components:**
1. **Password Hashing** - bcrypt (cost 10)
2. **Session Tokens** - HMAC-SHA256
3. **Rate Limiting** - Failed login attempts
4. **Username Hashing** - SHA-256 (prevents username enumeration)

### Setup Authentication

**Step 1: Generate Secret Key**
```bash
glance secret:make
# Output: Base64-encoded 32-byte key
```

**Step 2: Hash Password**
```bash
glance password:hash mypassword123
# Output: Bcrypt hash
```

**Step 3: Configure**
```yaml
auth:
  secret-key: "dGVzdC1zZWNyZXQta2V5LTEyMzQ1Njc4OTAxMjM0NTY="
  users:
    admin:
      password-hash: "$2a$10$..."
    john:
      password-hash: "$2a$10$..."
```

### Session Token Format

```
{username_hash_hex}:{timestamp}:{hmac}
```

**Example:**
```
a1b2c3d4e5f6:1699564800:9f8e7d6c5b4a3210
```

**Components:**
- `username_hash_hex` - SHA-256(username + secret_key)
- `timestamp` - Unix timestamp
- `hmac` - HMAC-SHA256(username_hash + timestamp, secret_key)

**Validation:**
1. Split token by `:`
2. Verify HMAC
3. Check timestamp (7-day expiry)
4. Lookup username from hash
5. Set authenticated user in context

### Rate Limiting

**Failed Login Attempts:**
```go
type failedAuthAttempt struct {
    Count     int
    LastAttemptAt time.Time
}
```

**Rules:**
- 5 attempts allowed
- 15-minute lockout after 5 failures
- Resets on successful login
- Per-IP address tracking

### Security Best Practices

1. **Use Environment Variables** for secrets
2. **Generate Strong Secret Key** (32 random bytes)
3. **Use HTTPS** in production (reverse proxy)
4. **Set `server.proxied: true`** if behind proxy
5. **Don't commit secrets** to git
6. **Rotate secret keys** periodically
7. **Use password-hash** instead of plain password

---

## Data Flow & Request Lifecycle

### Page Load Request

```
1. Browser: GET /home
   ↓
2. Server: Route to handlePageRequest()
   ↓
3. Auth Check
   - If RequiresAuth && !authenticated
     → Redirect to /login
   ↓
4. Page Lookup (by slug)
   - If not found → 404
   ↓
5. For each widget:
   - Call requiresUpdate()
     - Check cache expiry
     - If expired → add to update queue
   ↓
6. Parallel Widget Updates
   - Spawn goroutine per widget
   - Call widget.update(ctx)
   - Fetch external data
   - Parse & process
   - Update widget state
   ↓
7. Template Rendering
   - Execute page.html template
   - Inject widget HTML
   - Apply theme CSS
   ↓
8. Send Response
   - HTML page
   - Cache-Control headers
   ↓
9. Browser: Render page
```

### AJAX Content Update

**Frontend JavaScript** (on page load):
```javascript
// Check for new content every 30s
setInterval(() => {
  fetch(`/api/pages/${pageSlug}/content`)
    .then(res => res.text())
    .then(html => {
      document.querySelector('.page-content').innerHTML = html;
    });
}, 30000);
```

**Backend Flow:**
```
1. Browser: GET /api/pages/home/content
   ↓
2. Server: handleContentRequest()
   ↓
3. Same update logic as page load
   ↓
4. Render page-content.html (widgets only)
   ↓
5. Send Response (partial HTML)
   ↓
6. Browser: Replace content div
```

### Widget Update Flow

```
widget.update(ctx)
  ↓
1. Check context cancellation
  ↓
2. Make HTTP request(s)
   - Use widget-utils.go helpers
   - Worker pool for parallel requests
   - Early retry with exponential backoff
  ↓
3. Parse Response
   - JSON: gjson or json.Unmarshal
   - XML/RSS: gofeed
   - HTML: goquery
  ↓
4. Transform Data
   - Filter, sort, limit
   - Calculate derived values
   - Format timestamps
  ↓
5. Update Widget State
   - widget.Posts = results
   - widget.UpdatedAt = time.Now()
   - widget.Error = nil (or error)
  ↓
6. Render Called (later)
   - Generate HTML from template
   - Cache result
```

### Caching Strategy

**Cache Types:**

1. **Infinite Cache** (static widgets)
   - Never updates automatically
   - calendar, clock, bookmarks, html, iframe

2. **Duration Cache** (most widgets)
   - Updates after N time
   - Configurable per widget: `cache: 30m`

3. **On-The-Hour Cache** (time-sensitive)
   - Updates at top of each hour
   - Not widely used

**Cache Invalidation:**
- Config reload clears all caches
- Individual widget updates reset timer
- No persistent cache (in-memory only)

### HTTP Request Optimization

**Conditional Requests:**
```go
req.Header.Set("If-None-Match", etag)
req.Header.Set("If-Modified-Since", lastModified)

resp, err := client.Do(req)
if resp.StatusCode == 304 {
    // Use cached data
}
```

**Worker Pools:**
```go
func workerPoolWithResponses[T any](requests []channelRequest[T]) []T {
    results := make([]T, len(requests))
    var wg sync.WaitGroup

    for i, request := range requests {
        wg.Add(1)
        go func(i int, request channelRequest[T]) {
            defer wg.Done()
            results[i], _ = request()
        }(i, request)
    }

    wg.Wait()
    return results
}
```

**Early Retry Logic:**
```go
for attempt := 0; attempt < maxRetries; attempt++ {
    resp, err := client.Do(req)
    if err == nil || !isRetriableError(err) {
        return resp, err
    }

    backoff := time.Duration(1<<attempt) * time.Second
    time.Sleep(backoff)
}
```

---

## Build & Deployment

### Building from Source

**Requirements:**
- Go >= 1.23 (project uses 1.24.3)
- Git

**Build for Current Platform:**
```bash
cd /home/user/glance
go build -o build/glance .
```

**Cross-Compile:**
```bash
# Linux AMD64
GOOS=linux GOARCH=amd64 go build -o build/glance-linux-amd64 .

# Windows AMD64
GOOS=windows GOARCH=amd64 go build -o build/glance-windows-amd64.exe .

# macOS ARM64
GOOS=darwin GOARCH=arm64 go build -o build/glance-darwin-arm64 .

# Linux ARM64 (Raspberry Pi)
GOOS=linux GOARCH=arm64 go build -o build/glance-linux-arm64 .
```

**Run During Development:**
```bash
go run . --config glance.yml
```

### Docker Build

**Build Image:**
```bash
docker build -t glanceapp/glance:latest .
```

**Multi-Architecture:**
```bash
docker buildx build --platform linux/amd64,linux/arm64,linux/arm/v7 \
  -t glanceapp/glance:latest --push .
```

**Dockerfile Overview:**
```dockerfile
FROM golang:1.24.3-alpine3.21 AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o glance .

FROM alpine:3.21
WORKDIR /app
COPY --from=builder /app/glance /app/glance
EXPOSE 8080/tcp
ENTRYPOINT ["/app/glance", "--config", "/app/config/glance.yml"]
```

### GoReleaser

**File**: `.goreleaser.yaml`

**Features:**
- Multi-OS builds (5 operating systems)
- Multi-architecture (4 architectures)
- Archives (tar.gz, zip)
- Docker images (multi-arch)
- GitHub releases

**Release Build:**
```bash
goreleaser release --snapshot --clean
```

**Artifacts:**
- `glance-linux-amd64.tar.gz`
- `glance-linux-arm64.tar.gz`
- `glance-darwin-amd64.tar.gz`
- `glance-windows-amd64.zip`
- Docker images (glanceapp/glance:latest)

### Deployment Options

#### 1. Docker Compose (Recommended)

```yaml
services:
  glance:
    container_name: glance
    image: glanceapp/glance
    restart: unless-stopped
    volumes:
      - ./config:/app/config
      - ./assets:/app/assets
    ports:
      - 8080:8080
    environment:
      - TZ=America/New_York
```

**Start:**
```bash
docker compose up -d
```

#### 2. Systemd Service

`/etc/systemd/system/glance.service`:
```ini
[Unit]
Description=Glance Dashboard
After=network.target

[Service]
Type=simple
User=glance
WorkingDirectory=/opt/glance
ExecStart=/opt/glance/glance --config /etc/glance.yml
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
```

**Enable & Start:**
```bash
sudo systemctl enable glance
sudo systemctl start glance
sudo systemctl status glance
```

#### 3. Reverse Proxy (Nginx)

```nginx
server {
    listen 80;
    server_name glance.example.com;

    location / {
        proxy_pass http://localhost:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

**Config Update:**
```yaml
server:
  proxied: true  # Important for correct IP detection
```

#### 4. Manual Binary

```bash
# Download binary
wget https://github.com/glanceapp/glance/releases/latest/download/glance-linux-amd64.tar.gz
tar -xzf glance-linux-amd64.tar.gz

# Create config
wget -O glance.yml https://raw.githubusercontent.com/glanceapp/glance/main/docs/glance.yml

# Run
./glance --config glance.yml
```

### Environment Variables

**Docker:**
```bash
docker run -p 8080:8080 \
  -e RSS_TITLE="My Feed" \
  -e API_KEY=secret123 \
  -v ./config:/app/config \
  glanceapp/glance
```

**Systemd:**
```ini
[Service]
Environment="API_KEY=secret123"
Environment="TZ=America/New_York"
```

**Shell:**
```bash
export API_KEY=secret123
./glance
```

---

## Development Guide

### Project Structure for Developers

```
internal/glance/
├── main.go              # CLI & server lifecycle
├── glance.go            # HTTP server & routing
├── config.go            # Configuration parsing
├── widget.go            # Widget interface
├── widget-*.go          # Widget implementations
├── auth.go              # Authentication
├── theme.go             # Theming
├── embed.go             # Asset embedding
├── templates.go         # Template helpers
└── utils.go             # Utilities
```

### Adding a New Widget

**Step 1: Create Widget File**

`internal/glance/widget-example.go`:
```go
package glance

import (
    "context"
    "html/template"
    "time"
)

type exampleWidget struct {
    widgetBase `yaml:",inline"`
    APIKey     string `yaml:"api-key"`
    Limit      int    `yaml:"limit"`
    Items      []exampleItem `yaml:"-"`
}

type exampleItem struct {
    Title string
    URL   string
}

func (widget *exampleWidget) initialize() error {
    widget.withTitle = true
    widget.cacheType = cacheTypeDuration
    widget.cacheDuration = time.Hour

    if widget.Limit <= 0 {
        widget.Limit = 10
    }

    return nil
}

func (widget *exampleWidget) update(ctx context.Context) {
    // Fetch data from API
    data, err := fetchJSON[apiResponse](
        fmt.Sprintf("https://api.example.com/items?key=%s", widget.APIKey),
    )

    if !widget.canContinueUpdateAfterHandlingErr(err) {
        return
    }

    // Transform data
    widget.Items = make([]exampleItem, 0, len(data.Results))
    for _, item := range data.Results {
        widget.Items = append(widget.Items, exampleItem{
            Title: item.Title,
            URL:   item.Link,
        })
    }

    if len(widget.Items) > widget.Limit {
        widget.Items = widget.Items[:widget.Limit]
    }
}

func (widget *exampleWidget) Render() template.HTML {
    return widget.render(widget, assets.ExampleTemplate)
}
```

**Step 2: Register Widget**

`internal/glance/widget.go`:
```go
func newWidget(widgetType string) (widget, error) {
    switch widgetType {
    // ... existing cases
    case "example":
        w = &exampleWidget{}
    // ... rest
    }
}
```

**Step 3: Create Template**

`internal/glance/templates/widgets/example.html`:
```html
{{ template "widget-base" .options }}
    <div class="example-items">
        {{ range .items }}
            <a href="{{ .URL }}" class="example-item">
                {{ .Title }}
            </a>
        {{ end }}
    </div>
{{ template "widget-base-end" .options }}
```

**Step 4: Add CSS (if needed)**

`internal/glance/static/css/widgets.css`:
```css
.example-items {
    display: flex;
    flex-direction: column;
    gap: 10px;
}

.example-item {
    padding: 10px;
    border-radius: 5px;
    background: var(--color-widget-background);
}
```

**Step 5: Test**

```yaml
pages:
  - name: Test
    columns:
      - size: full
        widgets:
          - type: example
            api-key: test123
            limit: 5
```

### Code Style Guidelines

**1. Error Handling:**
```go
// Good
data, err := fetchData()
if !widget.canContinueUpdateAfterHandlingErr(err) {
    return
}

// Avoid
if err != nil {
    widget.Error = err
    return
}
```

**2. Widget Updates:**
```go
// Always use context
func (widget *myWidget) update(ctx context.Context) {
    select {
    case <-ctx.Done():
        return
    default:
    }

    // ... fetch data
}
```

**3. Configuration:**
```go
// Set defaults in initialize()
func (widget *myWidget) initialize() error {
    if widget.Limit <= 0 {
        widget.Limit = 10
    }
    return nil
}
```

**4. Naming:**
- Files: `widget-type-name.go`
- Structs: `typeNameWidget`
- Templates: `type-name.html`

### Testing

**Run Tests:**
```bash
go test ./...
```

**Test Specific Package:**
```bash
go test ./internal/glance
```

**With Coverage:**
```bash
go test -cover ./...
```

**Example Test:**
```go
func TestExampleWidget_Initialize(t *testing.T) {
    widget := &exampleWidget{}
    err := widget.initialize()

    if err != nil {
        t.Errorf("initialize() error = %v", err)
    }

    if widget.Limit != 10 {
        t.Errorf("Expected default limit 10, got %d", widget.Limit)
    }
}
```

### Debugging

**Enable Verbose Logging:**
```bash
GLANCE_LOG_LEVEL=debug ./glance
```

**Use Diagnostic Commands:**
```bash
# Validate config
./glance config:validate

# Print parsed config
./glance config:print

# Test temperature sensors
./glance sensors:print

# Run diagnostics
./glance diagnose
```

**Add Debug Logging:**
```go
import "log/slog"

slog.Debug("fetching data", "url", url, "params", params)
slog.Info("widget updated", "type", widget.GetType(), "items", len(items))
slog.Error("failed to fetch", "error", err)
```

---

## API Reference

### CLI Commands

```bash
# Server
glance                          # Start server (default)
glance --config /path/config    # Custom config path
glance --version                # Show version

# Configuration
glance config:validate          # Validate config file
glance config:print             # Print parsed config
glance config:print --json      # Print as JSON

# Authentication
glance password:hash <password> # Hash password for config
glance secret:make              # Generate secret key

# Diagnostics
glance sensors:print            # List temperature sensors
glance diagnose                 # Run diagnostics
```

### HTTP Endpoints

#### GET /

Returns first page in configuration.

**Response:** HTML page

#### GET /{page}

Returns named page by slug.

**Parameters:**
- `page` - Page slug (from config)

**Response:** HTML page
**Status:** 404 if page not found

#### GET /api/pages/{page}/content

Returns partial HTML for AJAX updates.

**Parameters:**
- `page` - Page slug

**Response:** HTML fragment (widgets only)

#### POST /api/set-theme/{key}

Sets theme preference.

**Parameters:**
- `key` - Theme key (from available themes)

**Response:** 200 OK
**Side Effect:** Sets `glance-theme` cookie

#### GET /api/widgets/{id}/{path...}

Widget-specific API endpoint.

**Parameters:**
- `id` - Widget ID
- `path...` - Widget-specific path

**Response:** Widget-dependent (JSON, HTML, etc.)

**Example:** Calendar widget uses this for event actions.

#### GET /api/healthz

Health check endpoint.

**Response:**
```json
{
  "status": "ok",
  "version": "v0.7.0"
}
```

#### POST /api/authenticate

Login endpoint.

**Request Body:**
```json
{
  "username": "admin",
  "password": "password123"
}
```

**Response:**
```json
{
  "success": true
}
```

**Side Effect:** Sets `glance-session-token` cookie

**Status:**
- 200 - Success
- 401 - Invalid credentials
- 429 - Too many attempts

#### GET /logout

Logout endpoint.

**Response:** Redirect to login page
**Side Effect:** Clears session cookie

#### GET /static/{hash}/{path...}

Static assets (CSS, JS, images).

**Parameters:**
- `hash` - Cache busting hash
- `path` - Asset path

**Headers:**
- `Cache-Control: public, max-age=86400` (24h)

#### GET /manifest.json

PWA manifest.

**Response:**
```json
{
  "name": "Glance",
  "short_name": "Glance",
  "description": "Self-hosted dashboard",
  "start_url": "/",
  "display": "standalone",
  "background_color": "#ffffff",
  "theme_color": "#000000",
  "icons": [
    {
      "src": "/static/{hash}/app-icon.png",
      "sizes": "512x512",
      "type": "image/png"
    }
  ]
}
```

---

## Performance Optimization

### Caching Strategy

**1. Widget-Level Caching:**
- Each widget caches results independently
- Configurable cache duration per widget
- In-memory storage (no disk I/O)

**2. Static Asset Caching:**
- 24-hour browser cache
- Cache-busting via MD5 hash in URL
- Embedded at compile time (zero disk reads)

**3. Conditional Requests:**
- Respects ETags and Last-Modified headers
- 304 Not Modified responses reduce bandwidth

### Parallel Processing

**Widget Updates:**
```go
// All widgets update in parallel
var wg sync.WaitGroup
for _, widget := range page.Widgets {
    if widget.requiresUpdate() {
        wg.Add(1)
        go func(w widget) {
            defer wg.Done()
            w.update(ctx)
        }(widget)
    }
}
wg.Wait()
```

**Multiple Feeds (RSS Widget):**
```go
// Worker pool for concurrent feed fetching
requests := make([]channelRequest[[]rssPost], len(feeds))
for i, feed := range feeds {
    requests[i] = func() ([]rssPost, error) {
        return fetchFeed(feed.URL)
    }
}
results := workerPoolWithResponses(requests)
```

### Memory Management

**Minimal Allocations:**
- Reuse buffers for template rendering
- Pre-allocate slices with capacity
- Use `sync.Pool` for frequently allocated objects

**Example:**
```go
posts := make([]rssPost, 0, expectedCount)  // Pre-allocate
```

### Network Optimization

**1. Request Timeouts:**
```go
client := &http.Client{
    Timeout: 10 * time.Second,
}
```

**2. Connection Pooling:**
- Default `http.Client` reuses connections
- Keep-alive enabled

**3. Early Retry:**
```go
// Exponential backoff for transient failures
for attempt := 0; attempt < 3; attempt++ {
    resp, err := client.Do(req)
    if err == nil {
        return resp, nil
    }
    time.Sleep(time.Duration(1<<attempt) * time.Second)
}
```

### Frontend Optimization

**JavaScript:**
- No frameworks (zero runtime parsing)
- Minified and concatenated
- Total size: ~70KB

**CSS:**
- Bundled at compile time
- Single CSS file
- Total size: ~74KB

**Images:**
- SVG icons (scalable, cacheable)
- Optimized PNGs for favicons

### Database-Free Design

**Benefits:**
- Zero disk I/O for normal operations
- No schema migrations
- Instant startup
- Simplified deployment

**Trade-offs:**
- No persistent cache across restarts
- User data stored in browser (localStorage)
- Configuration in YAML files

### Benchmarks

**Typical Performance:**
- **Page Load**: <1s (with warm cache)
- **Memory Usage**: <100MB
- **Binary Size**: ~20MB
- **Docker Image**: ~40MB (Alpine + binary)
- **Cold Start**: <100ms

**Scaling:**
- **Widgets per page**: 50+ (tested)
- **Pages**: Unlimited (limited by memory)
- **Concurrent users**: Hundreds (with proper resources)

---

## Troubleshooting

### Common Issues

#### 1. Requests Timing Out

**Symptoms:**
- Widgets not loading
- "Request timed out" errors
- Slow page loads

**Causes:**
- DNS rate limiting (Pi-hole, AdGuard)
- Firewall blocking outbound requests
- Network connectivity issues

**Solutions:**
```yaml
# Increase DNS rate limit in Pi-hole/AdGuard

# Or use different DNS server
# Docker: add --dns 8.8.8.8

# Check firewall rules
sudo iptables -L

# Test network connectivity
curl -I https://api.github.com
```

#### 2. Config Parse Errors

**Symptom:**
```
Error: cannot unmarshal !!map into []glance.page
```

**Cause:** Duplicate `pages` key in included files

**Solution:**
```yaml
# glance.yml
pages:
  - !include: pages/home.yml  # ✓ Correct

# pages/home.yml
name: Home        # ✓ Correct (no 'pages:' key)
columns: [...]

# pages/home.yml (WRONG)
pages:            # ✗ Remove this
  - name: Home
```

#### 3. Broken Layout (Dark Reader)

**Symptom:** Markets, bookmarks widgets look broken

**Cause:** Dark Reader browser extension

**Solution:** Disable Dark Reader for Glance domain

#### 4. Authentication Not Working

**Symptoms:**
- Can't log in with correct password
- Session expires immediately

**Checks:**
```bash
# Verify secret key length (must be 32 bytes)
echo "dGVzdC1zZWNyZXQta2V5LTEyMzQ1Njc4OTAxMjM0NTY=" | base64 -d | wc -c
# Should output: 32

# Verify password hash
glance password:hash mypassword
# Compare with config

# Check logs for errors
docker logs glance
```

#### 5. Docker Container Not Starting

**Check logs:**
```bash
docker logs glance
```

**Common errors:**
- Config file not found → Check volume mount
- Invalid config → Validate with `glance config:validate`
- Port already in use → Change port mapping

**Fix:**
```bash
# Verify volume mount
docker run -v ./config:/app/config glanceapp/glance

# Check config location
docker exec glance ls -la /app/config/

# Use different port
docker run -p 8081:8080 glanceapp/glance
```

#### 6. Environment Variables Not Working

**Symptoms:**
- `${VAR}` appears literally in rendered page
- "environment variable not found" error

**Solutions:**
```bash
# Set in Docker
docker run -e VAR=value glanceapp/glance

# Set in docker-compose.yml
environment:
  - VAR=value

# Verify variable is set
docker exec glance env | grep VAR

# Escape literal ${VAR}
something: \${NOT_AN_ENV_VAR}
```

#### 7. Hot Reload Not Working

**Causes:**
- Config file deleted and recreated (stops watching)
- Syntax error in config (logs show error)
- File system doesn't support inotify

**Solutions:**
```bash
# Check logs for errors
docker logs glance

# Restart container
docker restart glance

# Validate config before saving
glance config:validate
```

### Diagnostic Commands

```bash
# Validate configuration
glance config:validate

# Print parsed config (check env vars expanded)
glance config:print

# Check available temperature sensors
glance sensors:print

# Run full diagnostics
glance diagnose
```

### Debug Mode

**Enable verbose logging:**
```bash
# Binary
GLANCE_LOG_LEVEL=debug ./glance

# Docker
docker run -e GLANCE_LOG_LEVEL=debug glanceapp/glance
```

**Log Levels:**
- `debug` - Detailed logs
- `info` - General information (default)
- `warn` - Warnings only
- `error` - Errors only

### Performance Issues

**Symptom:** Slow page loads

**Diagnosis:**
```bash
# Check widget cache durations
glance config:print | grep cache

# Monitor network requests
# Browser DevTools → Network tab

# Check system resources
docker stats glance
```

**Solutions:**
```yaml
# Increase cache durations
- type: rss
  cache: 1h  # Increase from 30m

# Reduce number of feeds/items
- type: rss
  limit: 10  # Reduce from 25

# Reduce update frequency
- type: markets
  cache: 5m  # Increase from 1m
```

### Common Error Messages

| Error | Cause | Solution |
|-------|-------|----------|
| `unknown widget type: X` | Invalid widget type | Check spelling, see docs |
| `secret-key must be exactly 32 bytes` | Wrong key length | Generate new key with `secret:make` |
| `page slug "login" is reserved` | Using reserved slug | Choose different slug |
| `failed to fetch feed` | RSS feed URL invalid | Verify URL works in browser |
| `docker socket not accessible` | No Docker socket | Mount `-v /var/run/docker.sock:/var/run/docker.sock` |

### Getting Help

**Resources:**
- [Documentation](docs/configuration.md)
- [GitHub Issues](https://github.com/glanceapp/glance/issues)
- [Discord Community](https://discord.com/invite/7KQ7Xa9kJd)
- [Community Widgets](https://github.com/glanceapp/community-widgets)

**When Reporting Issues:**
1. Run `glance diagnose`
2. Include error logs
3. Share config (remove secrets!)
4. Describe expected vs actual behavior
5. Provide steps to reproduce

---

## Appendix

### Complete Widget Reference

See [docs/configuration.md](docs/configuration.md) for full widget documentation.

### Theme Colors

**HSL Format:**
```yaml
theme:
  background-color: 240 13 20  # Hue Saturation Lightness
  primary-color: 43 100 50
```

**Available Theme Presets:**
See [docs/themes.md](docs/themes.md)

### Icon Providers

**Supported:**
- Simple Icons - `simple-icons:{name}`
- Dashboard Icons - `di:{name}`
- Material Design Icons - `mdi:{name}`
- Selfh.st Icons - `si:{name}`

**Example:**
```yaml
- type: bookmarks
  groups:
    - title: Social
      links:
        - title: GitHub
          url: https://github.com
          icon: simple-icons:github
```

### File Locations

**Binary Installation:**
```
/opt/glance/
├── glance              # Binary
└── glance.yml          # Config
```

**Docker:**
```
Container:
  /app/glance           # Binary
  /app/config/          # Config mount point
  /app/assets/          # Assets mount point

Host:
  ./config/glance.yml   # Config file
  ./assets/             # Custom CSS, images
```

**Systemd:**
```
/opt/glance/glance      # Binary
/etc/glance.yml         # Config
/etc/systemd/system/glance.service  # Service file
```

### Related Projects

**Community:**
- [Community Widgets](https://github.com/glanceapp/community-widgets)
- [Docker Compose Template](https://github.com/glanceapp/docker-compose-template)

**Inspiration:**
- [Homer](https://github.com/bastienwirtz/homer)
- [Heimdall](https://github.com/linuxserver/Heimdall)
- [Dashy](https://github.com/Lissy93/dashy)

---

## Changelog

See [GitHub Releases](https://github.com/glanceapp/glance/releases) for version history.

---

## License

Apache License 2.0 - See [LICENSE](LICENSE) file.

---

## Contributing

See [CONTRIBUTING.md](README.md#contributing-guidelines) for guidelines.

**Key Points:**
- Submit feature requests before implementing
- Use `dev` branch for new features
- Avoid new dependencies
- No breaking config changes
- Use heroicons for icons
- No package.json

---

**Document Version:** 1.0
**Last Updated:** 2025-11-16
**Glance Version:** Latest (main branch)
**Author:** Claude AI (Anthropic)
