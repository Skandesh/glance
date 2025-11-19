package glance

import (
	"context"
	"log/slog"
	"sync"
	"time"
)

// RevenueSnapshot stores historical revenue data
type RevenueSnapshot struct {
	Timestamp    time.Time
	MRR          float64
	ARR          float64
	GrowthRate   float64
	NewMRR       float64
	ChurnedMRR   float64
	Mode         string
}

// CustomerSnapshot stores historical customer data
type CustomerSnapshot struct {
	Timestamp        time.Time
	TotalCustomers   int
	NewCustomers     int
	ChurnedCustomers int
	ChurnRate        float64
	ActiveCustomers  int
	Mode             string
}

// SimpleMetricsDB handles in-memory storage of historical metrics
type SimpleMetricsDB struct {
	revenueHistory  map[string][]*RevenueSnapshot  // key: mode
	customerHistory map[string][]*CustomerSnapshot // key: mode
	mu              sync.RWMutex
	maxHistory      int
}

var (
	globalSimpleDB     *SimpleMetricsDB
	globalSimpleDBOnce sync.Once
)

// GetSimpleMetricsDB returns the global simple metrics database (singleton)
func GetSimpleMetricsDB() *SimpleMetricsDB {
	globalSimpleDBOnce.Do(func() {
		globalSimpleDB = &SimpleMetricsDB{
			revenueHistory:  make(map[string][]*RevenueSnapshot),
			customerHistory: make(map[string][]*CustomerSnapshot),
			maxHistory:      100, // Keep last 100 snapshots per mode
		}
		slog.Info("Simple metrics database initialized")
	})
	return globalSimpleDB
}

// SaveRevenueSnapshot saves a revenue snapshot to memory
func (db *SimpleMetricsDB) SaveRevenueSnapshot(ctx context.Context, snapshot *RevenueSnapshot) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	mode := snapshot.Mode
	if db.revenueHistory[mode] == nil {
		db.revenueHistory[mode] = make([]*RevenueSnapshot, 0)
	}

	db.revenueHistory[mode] = append(db.revenueHistory[mode], snapshot)

	// Keep only last N snapshots
	if len(db.revenueHistory[mode]) > db.maxHistory {
		db.revenueHistory[mode] = db.revenueHistory[mode][len(db.revenueHistory[mode])-db.maxHistory:]
	}

	return nil
}

// SaveCustomerSnapshot saves a customer snapshot to memory
func (db *SimpleMetricsDB) SaveCustomerSnapshot(ctx context.Context, snapshot *CustomerSnapshot) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	mode := snapshot.Mode
	if db.customerHistory[mode] == nil {
		db.customerHistory[mode] = make([]*CustomerSnapshot, 0)
	}

	db.customerHistory[mode] = append(db.customerHistory[mode], snapshot)

	// Keep only last N snapshots
	if len(db.customerHistory[mode]) > db.maxHistory {
		db.customerHistory[mode] = db.customerHistory[mode][len(db.customerHistory[mode])-db.maxHistory:]
	}

	return nil
}

// GetRevenueHistory returns historical revenue data for the specified period
func (db *SimpleMetricsDB) GetRevenueHistory(ctx context.Context, mode string, startTime, endTime time.Time) ([]*RevenueSnapshot, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	history, exists := db.revenueHistory[mode]
	if !exists {
		return nil, nil
	}

	// Filter by time range
	var filtered []*RevenueSnapshot
	for _, snapshot := range history {
		if (snapshot.Timestamp.Equal(startTime) || snapshot.Timestamp.After(startTime)) &&
			(snapshot.Timestamp.Equal(endTime) || snapshot.Timestamp.Before(endTime)) {
			filtered = append(filtered, snapshot)
		}
	}

	return filtered, nil
}

// GetCustomerHistory returns historical customer data for the specified period
func (db *SimpleMetricsDB) GetCustomerHistory(ctx context.Context, mode string, startTime, endTime time.Time) ([]*CustomerSnapshot, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	history, exists := db.customerHistory[mode]
	if !exists {
		return nil, nil
	}

	// Filter by time range
	var filtered []*CustomerSnapshot
	for _, snapshot := range history {
		if (snapshot.Timestamp.Equal(startTime) || snapshot.Timestamp.After(startTime)) &&
			(snapshot.Timestamp.Equal(endTime) || snapshot.Timestamp.Before(endTime)) {
			filtered = append(filtered, snapshot)
		}
	}

	return filtered, nil
}

// GetLatestRevenue returns the most recent revenue snapshot
func (db *SimpleMetricsDB) GetLatestRevenue(ctx context.Context, mode string) (*RevenueSnapshot, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	history, exists := db.revenueHistory[mode]
	if !exists || len(history) == 0 {
		return nil, nil
	}

	return history[len(history)-1], nil
}

// GetLatestCustomers returns the most recent customer snapshot
func (db *SimpleMetricsDB) GetLatestCustomers(ctx context.Context, mode string) (*CustomerSnapshot, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	history, exists := db.customerHistory[mode]
	if !exists || len(history) == 0 {
		return nil, nil
	}

	return history[len(history)-1], nil
}

// GetDatabaseStats returns database statistics
func (db *SimpleMetricsDB) GetDatabaseStats(ctx context.Context) (map[string]interface{}, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	stats := make(map[string]interface{})

	totalRevenue := 0
	for _, history := range db.revenueHistory {
		totalRevenue += len(history)
	}

	totalCustomer := 0
	for _, history := range db.customerHistory {
		totalCustomer += len(history)
	}

	stats["revenue_metrics_count"] = totalRevenue
	stats["customer_metrics_count"] = totalCustomer
	stats["modes"] = len(db.revenueHistory)

	return stats, nil
}

// CleanupOldMetrics removes metrics older than the specified duration
func (db *SimpleMetricsDB) CleanupOldMetrics(ctx context.Context, retentionPeriod time.Duration) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	cutoff := time.Now().Add(-retentionPeriod)

	// Clean revenue history
	for mode, history := range db.revenueHistory {
		filtered := make([]*RevenueSnapshot, 0)
		for _, snapshot := range history {
			if snapshot.Timestamp.After(cutoff) {
				filtered = append(filtered, snapshot)
			}
		}
		db.revenueHistory[mode] = filtered
	}

	// Clean customer history
	for mode, history := range db.customerHistory {
		filtered := make([]*CustomerSnapshot, 0)
		for _, snapshot := range history {
			if snapshot.Timestamp.After(cutoff) {
				filtered = append(filtered, snapshot)
			}
		}
		db.customerHistory[mode] = filtered
	}

	slog.Info("Cleaned up old metrics", "cutoff", cutoff)
	return nil
}

// Close is a no-op for in-memory database
func (db *SimpleMetricsDB) Close() error {
	return nil
}

// GetMetricsDatabase returns the simple metrics database (compatibility wrapper)
func GetMetricsDatabase(dbPath string) (*SimpleMetricsDB, error) {
	return GetSimpleMetricsDB(), nil
}
