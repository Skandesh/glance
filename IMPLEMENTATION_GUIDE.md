# Glance Implementation Guide

This guide shows how to build and run the Glance dashboard application from source.

## Quick Start

### Prerequisites

- Go 1.23 or higher
- Git

### Build and Run

1. **Build the application**:
   ```bash
   go build -o build/glance .
   ```

   This creates a single binary at `build/glance` (~20MB).

2. **Verify the build**:
   ```bash
   ./build/glance --version
   # Output: dev
   ```

3. **Validate the configuration**:
   ```bash
   ./build/glance -config config.yml config:validate
   ```

4. **Run the server**:
   ```bash
   ./build/glance -config config.yml
   ```

5. **Access the dashboard**:
   - Open http://localhost:8080 in your browser
   - The dashboard will load with 3 pages:
     - **Home** - Main dashboard with RSS, news, bookmarks, etc.
     - **Development** - GitHub releases and repository stats
     - **Custom** - Custom widgets and search

## Configuration

The included `config.yml` demonstrates:

### Pages
- **3 pages** with different layouts
- **Multiple columns** (small and full-width)
- **Page slugs** for clean URLs

### Widgets Implemented

#### Home Page
- **Clock** - Multi-timezone clock (UTC, New York)
- **Calendar** - Monthly calendar view
- **Bookmarks** - Organized link collections
- **RSS** - Tech news aggregator
- **Hacker News** - Top stories
- **Reddit** - Multiple subreddits (technology, programming)
- **Server Stats** - CPU, memory, disk usage
- **Monitor** - Website uptime checks
- **Markets** - Stock/crypto prices (SPY, BTC, ETH)

#### Development Page
- **Releases** - Latest GitHub releases
- **Repository** - Repository statistics

#### Custom Page
- **HTML** - Custom HTML content
- **Search** - Search with custom bangs
- **To-do** - Task management

### Customization

Edit `config.yml` to:
- Add/remove widgets
- Change theme colors
- Add more pages
- Configure cache durations
- Add API keys for external services

## CLI Commands

### Configuration Management
```bash
# Validate configuration
./build/glance -config config.yml config:validate

# Print parsed configuration
./build/glance -config config.yml config:print

# Print as JSON
./build/glance -config config.yml config:print --json
```

### Authentication Setup
```bash
# Generate secret key
./build/glance secret:make

# Hash password
./build/glance password:hash mypassword
```

### Diagnostics
```bash
# List temperature sensors
./build/glance sensors:print

# Run diagnostics
./build/glance diagnose
```

## API Endpoints

The server exposes several HTTP endpoints:

### Pages
- `GET /` - First page (home)
- `GET /home` - Home page
- `GET /dev` - Development page
- `GET /custom` - Custom page

### API
- `GET /api/healthz` - Health check
- `GET /api/pages/{slug}/content` - Page content (AJAX)
- `POST /api/set-theme/{key}` - Set theme
- `GET /api/widgets/{id}/{path}` - Widget-specific API

### Assets
- `GET /static/{hash}/{path}` - Static assets (24h cache)
- `GET /manifest.json` - PWA manifest

## Development

### Project Structure
```
glance/
├── build/
│   └── glance              # Built binary
├── config.yml              # Configuration file
├── internal/glance/        # Core application code
│   ├── main.go            # Entry point
│   ├── glance.go          # HTTP server
│   ├── widget-*.go        # Widget implementations
│   └── static/            # Frontend assets
├── pkg/sysinfo/           # System info package
└── docs/                  # Documentation
```

### Hot Reload

The server watches `config.yml` for changes and automatically reloads:

1. Start the server: `./build/glance -config config.yml`
2. Edit `config.yml`
3. Save the file
4. Changes apply immediately (no restart needed)

**Note**: Config errors will be logged, but the server continues with the old config.

### Adding Widgets

1. Edit `config.yml`
2. Add a new widget to any page
3. Save the file (hot reload applies changes)

Example:
```yaml
- type: weather
  location: London, UK
  units: metric
```

See [TECHNICAL_DOCUMENTATION.md](TECHNICAL_DOCUMENTATION.md) for available widgets and options.

## Production Deployment

### Docker

Build Docker image:
```bash
docker build -t glance:latest .
```

Run container:
```bash
docker run -d \
  --name glance \
  -p 8080:8080 \
  -v $(pwd)/config.yml:/app/config/glance.yml \
  glance:latest
```

### Systemd Service

Create `/etc/systemd/system/glance.service`:
```ini
[Unit]
Description=Glance Dashboard
After=network.target

[Service]
Type=simple
User=glance
WorkingDirectory=/opt/glance
ExecStart=/opt/glance/build/glance -config /opt/glance/config.yml
Restart=always

[Install]
WantedBy=multi-user.target
```

Enable and start:
```bash
sudo systemctl enable glance
sudo systemctl start glance
```

### Reverse Proxy (Nginx)

```nginx
server {
    listen 80;
    server_name dashboard.example.com;

    location / {
        proxy_pass http://localhost:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    }
}
```

Update `config.yml`:
```yaml
server:
  proxied: true
```

## Performance

### Build Metrics
- **Binary size**: ~20MB
- **Build time**: ~30 seconds (first build)
- **Dependencies**: 7 direct, 16 indirect

### Runtime Metrics
- **Memory usage**: <100MB typical
- **Startup time**: <100ms
- **Page load**: <1s (with cache)

### Optimization Tips

1. **Cache durations** - Increase for less frequent updates:
   ```yaml
   - type: rss
     cache: 1h  # Reduce API calls
   ```

2. **Limit results** - Reduce data processing:
   ```yaml
   - type: rss
     limit: 10  # Fewer items
   ```

3. **Collapse widgets** - Improve initial load:
   ```yaml
   - type: rss
     collapse-after: 5  # Show 5, hide rest
   ```

## Troubleshooting

### Server won't start
```bash
# Check logs
cat glance.log

# Validate config
./build/glance -config config.yml config:validate

# Check port availability
lsof -i :8080
```

### Widgets not loading
```bash
# Check server logs for errors
tail -f glance.log

# Test external APIs
curl -I https://hnrss.org/frontpage

# Verify network connectivity
ping 8.8.8.8
```

### Configuration errors
```bash
# Validate YAML syntax
./build/glance -config config.yml config:validate

# Print parsed config
./build/glance -config config.yml config:print
```

### Hot reload not working
- Check file permissions
- Verify file watcher support
- Restart the server manually

## Testing

### Manual Testing

1. **Health check**:
   ```bash
   curl http://localhost:8080/api/healthz
   # Should return 200 OK
   ```

2. **Page load**:
   ```bash
   curl http://localhost:8080
   # Should return HTML
   ```

3. **Widget content**:
   ```bash
   curl http://localhost:8080/api/pages/home/content/
   # Should return widget HTML
   ```

### Load Testing

Use `ab` (Apache Bench):
```bash
ab -n 1000 -c 10 http://localhost:8080/
```

## Next Steps

1. **Customize the dashboard** - Edit `config.yml` to add your widgets
2. **Add authentication** - Generate secret key and configure users
3. **Create themes** - Customize colors and appearance
4. **Deploy to production** - Use Docker or systemd
5. **Monitor performance** - Check logs and resource usage

## Resources

- **Technical Documentation**: [TECHNICAL_DOCUMENTATION.md](TECHNICAL_DOCUMENTATION.md)
- **Configuration Guide**: [docs/configuration.md](docs/configuration.md)
- **Theme Guide**: [docs/themes.md](docs/themes.md)
- **GitHub Repository**: https://github.com/glanceapp/glance
- **Discord Community**: https://discord.com/invite/7KQ7Xa9kJd

## License

Apache License 2.0 - See [LICENSE](LICENSE) file.

---

**Implementation Date**: 2025-11-16
**Glance Version**: dev
**Build Status**: ✅ Successful
**Server Status**: ✅ Running on port 8080
