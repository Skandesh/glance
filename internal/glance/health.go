package glance

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"runtime"
	"sync"
	"time"
)

// HealthChecker performs health checks on various system components
type HealthChecker struct {
	checks   map[string]HealthCheckFunc
	mu       sync.RWMutex
	lastRun  map[string]time.Time
	results  map[string]*HealthCheckResult
	cacheTTL time.Duration
}

// HealthCheckFunc is a function that performs a health check
type HealthCheckFunc func(ctx context.Context) *HealthCheckResult

// HealthCheckResult represents the result of a health check
type HealthCheckResult struct {
	Status    HealthStatus           `json:"status"`
	Message   string                 `json:"message,omitempty"`
	Details   map[string]interface{} `json:"details,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
	Duration  time.Duration          `json:"duration"`
}

// HealthStatus represents the health status
type HealthStatus string

const (
	HealthStatusHealthy   HealthStatus = "healthy"
	HealthStatusDegraded  HealthStatus = "degraded"
	HealthStatusUnhealthy HealthStatus = "unhealthy"
)

// HealthResponse is the overall health response
type HealthResponse struct {
	Status    HealthStatus                  `json:"status"`
	Timestamp time.Time                     `json:"timestamp"`
	Uptime    time.Duration                 `json:"uptime"`
	Version   string                        `json:"version"`
	Checks    map[string]*HealthCheckResult `json:"checks"`
}

var (
	globalHealthChecker *HealthChecker
	healthCheckerOnce   sync.Once
	startTime           = time.Now()
)

// GetHealthChecker returns the global health checker (singleton)
func GetHealthChecker() *HealthChecker {
	healthCheckerOnce.Do(func() {
		globalHealthChecker = &HealthChecker{
			checks:   make(map[string]HealthCheckFunc),
			lastRun:  make(map[string]time.Time),
			results:  make(map[string]*HealthCheckResult),
			cacheTTL: 30 * time.Second,
		}

		// Register default health checks
		globalHealthChecker.RegisterCheck("database", checkDatabaseHealth)
		globalHealthChecker.RegisterCheck("memory", checkMemoryHealth)
		globalHealthChecker.RegisterCheck("stripe_pool", checkStripePoolHealth)
	})
	return globalHealthChecker
}

// RegisterCheck registers a new health check
func (hc *HealthChecker) RegisterCheck(name string, check HealthCheckFunc) {
	hc.mu.Lock()
	defer hc.mu.Unlock()
	hc.checks[name] = check
}

// RunChecks runs all registered health checks
func (hc *HealthChecker) RunChecks(ctx context.Context) *HealthResponse {
	hc.mu.RLock()
	checks := make(map[string]HealthCheckFunc, len(hc.checks))
	for k, v := range hc.checks {
		checks[k] = v
	}
	hc.mu.RUnlock()

	results := make(map[string]*HealthCheckResult)
	var wg sync.WaitGroup

	for name, check := range checks {
		// Check if cached result is still valid
		hc.mu.RLock()
		lastRun, hasLastRun := hc.lastRun[name]
		cachedResult, hasCached := hc.results[name]
		hc.mu.RUnlock()

		if hasLastRun && hasCached && time.Since(lastRun) < hc.cacheTTL {
			results[name] = cachedResult
			continue
		}

		wg.Add(1)
		go func(n string, c HealthCheckFunc) {
			defer wg.Done()

			start := time.Now()
			checkCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
			defer cancel()

			result := c(checkCtx)
			result.Duration = time.Since(start)
			result.Timestamp = time.Now()

			hc.mu.Lock()
			hc.results[n] = result
			hc.lastRun[n] = time.Now()
			hc.mu.Unlock()

			results[n] = result
		}(name, check)
	}

	wg.Wait()

	// Determine overall status
	overallStatus := HealthStatusHealthy
	for _, result := range results {
		if result.Status == HealthStatusUnhealthy {
			overallStatus = HealthStatusUnhealthy
			break
		} else if result.Status == HealthStatusDegraded && overallStatus == HealthStatusHealthy {
			overallStatus = HealthStatusDegraded
		}
	}

	return &HealthResponse{
		Status:    overallStatus,
		Timestamp: time.Now(),
		Uptime:    time.Since(startTime),
		Version:   "1.0.0",
		Checks:    results,
	}
}

// checkDatabaseHealth checks database connectivity and performance
func checkDatabaseHealth(ctx context.Context) *HealthCheckResult {
	db, err := GetMetricsDatabase("")
	if err != nil {
		return &HealthCheckResult{
			Status:  HealthStatusDegraded,
			Message: "Database not initialized",
		}
	}

	// Try a simple query
	stats, err := db.GetDatabaseStats(ctx)
	if err != nil {
		return &HealthCheckResult{
			Status:  HealthStatusUnhealthy,
			Message: fmt.Sprintf("Database query failed: %v", err),
		}
	}

	return &HealthCheckResult{
		Status:  HealthStatusHealthy,
		Message: "Database operational",
		Details: stats,
	}
}

// checkMemoryHealth checks memory usage
func checkMemoryHealth(ctx context.Context) *HealthCheckResult {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	memUsedMB := m.Alloc / 1024 / 1024
	memThresholdMB := uint64(512) // 512 MB threshold

	status := HealthStatusHealthy
	if memUsedMB > memThresholdMB*2 {
		status = HealthStatusUnhealthy
	} else if memUsedMB > memThresholdMB {
		status = HealthStatusDegraded
	}

	return &HealthCheckResult{
		Status:  status,
		Message: fmt.Sprintf("Memory usage: %d MB", memUsedMB),
		Details: map[string]interface{}{
			"alloc_mb":      memUsedMB,
			"sys_mb":        m.Sys / 1024 / 1024,
			"num_gc":        m.NumGC,
			"goroutines":    runtime.NumGoroutine(),
			"threshold_mb":  memThresholdMB,
		},
	}
}

// checkStripePoolHealth checks Stripe client pool health
func checkStripePoolHealth(ctx context.Context) *HealthCheckResult {
	pool := GetStripeClientPool()
	metrics := pool.GetMetrics()

	circuitStates := metrics["circuit_states"].(map[string]int)
	openCircuits := circuitStates["open"]

	status := HealthStatusHealthy
	message := "Stripe pool operational"

	if openCircuits > 0 {
		status = HealthStatusDegraded
		message = fmt.Sprintf("%d circuit(s) open", openCircuits)
	}

	return &HealthCheckResult{
		Status:  status,
		Message: message,
		Details: metrics,
	}
}

// HealthHandler returns an HTTP handler for health checks
func HealthHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		checker := GetHealthChecker()
		response := checker.RunChecks(r.Context())

		w.Header().Set("Content-Type", "application/json")

		// Set status code based on health
		statusCode := http.StatusOK
		if response.Status == HealthStatusUnhealthy {
			statusCode = http.StatusServiceUnavailable
		} else if response.Status == HealthStatusDegraded {
			statusCode = http.StatusOK // Return 200 but indicate degraded in body
		}

		w.WriteHeader(statusCode)
		json.NewEncoder(w).Encode(response)
	}
}

// ReadinessHandler returns an HTTP handler for readiness checks
func ReadinessHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		checker := GetHealthChecker()
		response := checker.RunChecks(r.Context())

		w.Header().Set("Content-Type", "application/json")

		// Readiness requires all checks to be healthy
		if response.Status != HealthStatusHealthy {
			w.WriteHeader(http.StatusServiceUnavailable)
		} else {
			w.WriteHeader(http.StatusOK)
		}

		json.NewEncoder(w).Encode(map[string]interface{}{
			"ready":  response.Status == HealthStatusHealthy,
			"status": response.Status,
		})
	}
}

// LivenessHandler returns an HTTP handler for liveness checks
func LivenessHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"alive":  true,
			"uptime": time.Since(startTime).String(),
		})
	}
}

// MetricsHandler returns an HTTP handler for Prometheus-style metrics
func MetricsHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var m runtime.MemStats
		runtime.ReadMemStats(&m)

		metrics := []string{
			fmt.Sprintf("# HELP glance_uptime_seconds Application uptime in seconds"),
			fmt.Sprintf("# TYPE glance_uptime_seconds counter"),
			fmt.Sprintf("glance_uptime_seconds %d", int64(time.Since(startTime).Seconds())),
			"",
			fmt.Sprintf("# HELP glance_memory_alloc_bytes Memory allocated in bytes"),
			fmt.Sprintf("# TYPE glance_memory_alloc_bytes gauge"),
			fmt.Sprintf("glance_memory_alloc_bytes %d", m.Alloc),
			"",
			fmt.Sprintf("# HELP glance_goroutines Number of goroutines"),
			fmt.Sprintf("# TYPE glance_goroutines gauge"),
			fmt.Sprintf("glance_goroutines %d", runtime.NumGoroutine()),
			"",
		}

		// Add Stripe pool metrics
		pool := GetStripeClientPool()
		poolMetrics := pool.GetMetrics()
		circuitStates := poolMetrics["circuit_states"].(map[string]int)

		metrics = append(metrics,
			"# HELP glance_stripe_clients_total Total number of Stripe clients",
			"# TYPE glance_stripe_clients_total gauge",
			fmt.Sprintf("glance_stripe_clients_total %d", poolMetrics["total_clients"]),
			"",
			"# HELP glance_stripe_circuit_breaker_state State of circuit breakers (0=closed, 1=half-open, 2=open)",
			"# TYPE glance_stripe_circuit_breaker_state gauge",
			fmt.Sprintf("glance_stripe_circuit_breaker_state{state=\"closed\"} %d", circuitStates["closed"]),
			fmt.Sprintf("glance_stripe_circuit_breaker_state{state=\"half_open\"} %d", circuitStates["half_open"]),
			fmt.Sprintf("glance_stripe_circuit_breaker_state{state=\"open\"} %d", circuitStates["open"]),
			"",
		)

		// Add database metrics if available
		db, err := GetMetricsDatabase("")
		if err == nil {
			dbStats, err := db.GetDatabaseStats(context.Background())
			if err == nil {
				metrics = append(metrics,
					"# HELP glance_db_records_total Total records in database",
					"# TYPE glance_db_records_total gauge",
				)
				for key, value := range dbStats {
					if count, ok := value.(int); ok && key != "db_size_bytes" {
						metrics = append(metrics, fmt.Sprintf("glance_db_records_total{table=\"%s\"} %d", key, count))
					}
				}
				if size, ok := dbStats["db_size_bytes"].(int); ok {
					metrics = append(metrics,
						"",
						"# HELP glance_db_size_bytes Database size in bytes",
						"# TYPE glance_db_size_bytes gauge",
						fmt.Sprintf("glance_db_size_bytes %d", size),
					)
				}
			}
		}

		w.Header().Set("Content-Type", "text/plain; version=0.0.4")
		w.WriteHeader(http.StatusOK)
		for _, metric := range metrics {
			fmt.Fprintln(w, metric)
		}
	}
}

// StartHealthChecks starts periodic health checks
func StartHealthChecks(interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for range ticker.C {
			checker := GetHealthChecker()
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			response := checker.RunChecks(ctx)
			cancel()

			if response.Status != HealthStatusHealthy {
				slog.Warn("Health check failed",
					"status", response.Status,
					"checks", len(response.Checks))
			}
		}
	}()
}
