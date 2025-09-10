// Copyright 2025 Jonas Teuwen. All rights reserved.
//
// This file is part of SlideInsight.
//
// Use of this source code is governed by the terms found in the
// LICENSE file located in the SlideInsight project root.
package handlers

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"sort"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/disk"
	"github.com/shirou/gopsutil/v4/load"
	"github.com/shirou/gopsutil/v4/mem"
	"github.com/shirou/gopsutil/v4/net"
	"github.com/shirou/gopsutil/v4/process"
)

// SystemStats represents the comprehensive system monitoring data structure
type SystemStats struct {
	PID       SystemStatsPID       `json:"pid"`
	OS        SystemStatsOS        `json:"os"`
	HTTP      SystemStatsHTTP      `json:"http"`
	Runtime   SystemStatsRuntime   `json:"runtime"`
	WebSocket SystemStatsWebSocket `json:"websocket"`
	Timestamp int64                `json:"timestamp"`
}

type SystemStatsPID struct {
	CPU             float64 `json:"cpu"`
	RAM             uint64  `json:"ram"`
	Conns           int     `json:"conns"`
	FDs             int     `json:"fds"`
	Threads         int     `json:"threads"`
	Uptime          float64 `json:"uptime"`           // Process uptime in seconds
	CPUTime         float64 `json:"cpu_time"`         // Total CPU time in seconds
	ContextSwitches uint64  `json:"context_switches"` // Voluntary context switches
}

type SystemStatsOS struct {
	CPU              float64 `json:"cpu"`
	RAM              uint64  `json:"ram"`
	TotalRAM         uint64  `json:"total_ram"`
	SwapUsed         uint64  `json:"swap_used"`
	SwapTotal        uint64  `json:"swap_total"`
	LoadAvg          float64 `json:"load_avg"`
	LoadAvg5         float64 `json:"load_avg_5"`
	LoadAvg15        float64 `json:"load_avg_15"`
	Conns            int     `json:"conns"`
	DiskReadBytes    uint64  `json:"disk_read_bytes"`
	DiskWriteBytes   uint64  `json:"disk_write_bytes"`
	NetReadBytes     uint64  `json:"net_read_bytes"`
	NetWriteBytes    uint64  `json:"net_write_bytes"`
	Uptime           float64 `json:"uptime"`            // System uptime in seconds
	BootTime         int64   `json:"boot_time"`         // System boot time timestamp
	TotalProcesses   int     `json:"total_processes"`   // Total number of processes
	RunningProcesses int     `json:"running_processes"` // Number of running processes
}

type SystemStatsHTTP struct {
	RequestsPerSecond  float64           `json:"requests_per_second"`
	TotalRequests      uint64            `json:"total_requests"`
	ActiveRequests     int32             `json:"active_requests"`
	AvgResponseTimeMs  float64           `json:"avg_response_time_ms"`
	P95ResponseTimeMs  float64           `json:"p95_response_time_ms"`
	P99ResponseTimeMs  float64           `json:"p99_response_time_ms"`
	StatusCodes        map[string]uint64 `json:"status_codes"`
	ErrorRate          float64           `json:"error_rate"`
	RequestsByMethod   map[string]uint64 `json:"requests_by_method"`
	RequestsByEndpoint map[string]uint64 `json:"requests_by_endpoint"`
	LastResetTime      int64             `json:"last_reset_time"`
	TotalResponseTime  float64           `json:"total_response_time"` // Sum of all response times
	SlowRequests       uint64            `json:"slow_requests"`       // Requests > 1 second
}

type SystemStatsRuntime struct {
	Goroutines     int    `json:"goroutines"`
	HeapAlloc      uint64 `json:"heap_alloc"`
	HeapSys        uint64 `json:"heap_sys"`
	HeapInuse      uint64 `json:"heap_inuse"`
	HeapObjects    uint64 `json:"heap_objects"`
	StackInuse     uint64 `json:"stack_inuse"`
	GCCycles       uint32 `json:"gc_cycles"`
	GCPauseTotalNs uint64 `json:"gc_pause_total_ns"`
	NextGC         uint64 `json:"next_gc"`
	Version        string `json:"version"`
}

type SystemStatsWebSocket struct {
	TotalConnections   uint64 `json:"total_connections"`
	ActiveConnections  int32  `json:"active_connections"`
	MessagesSent       uint64 `json:"messages_sent"`
	MessagesReceived   uint64 `json:"messages_received"`
	ConnectionErrors   uint64 `json:"connection_errors"`
	LastConnectionTime int64  `json:"last_connection_time"`
}

// HTTP monitoring data
type HTTPMetrics struct {
	totalRequests      atomic.Uint64
	activeRequests     atomic.Int32
	statusCodes        sync.Map // map[int]uint64
	methods            sync.Map // map[string]uint64
	endpoints          sync.Map // map[string]uint64
	responseTimes      []float64
	responseTimesMutex sync.RWMutex
	startTime          time.Time
	lastResetTime      atomic.Int64
	totalResponseTime  atomic.Uint64 // Total response time in nanoseconds
	slowRequests       atomic.Uint64 // Requests > 1 second
}

// WebSocket monitoring data
type WebSocketMetrics struct {
	totalConnections   atomic.Uint64
	activeConnections  atomic.Int32
	messagesSent       atomic.Uint64
	messagesReceived   atomic.Uint64
	connectionErrors   atomic.Uint64
	lastConnectionTime atomic.Int64
}

// SystemMonitorHub manages WebSocket connections for system monitoring
type SystemMonitorHub struct {
	connections      map[*websocket.Conn]bool
	mutex            sync.RWMutex
	process          *process.Process
	numCPU           int
	stopChan         chan struct{}
	stats            SystemStats
	statsMutex       sync.RWMutex
	httpMetrics      *HTTPMetrics
	websocketMetrics *WebSocketMetrics
	lastIOStats      *IOStats
	lastNetStats     []net.IOCountersStat
	lastDiskStats    map[string]disk.IOCountersStat
}

type IOStats struct {
	DiskReadBytes  uint64
	DiskWriteBytes uint64
	NetReadBytes   uint64
	NetWriteBytes  uint64
	Timestamp      time.Time
}

var (
	monitorHub *SystemMonitorHub
	hubOnce    sync.Once
)

func getMonitorHub() *SystemMonitorHub {
	hubOnce.Do(func() {
		p, _ := process.NewProcess(int32(os.Getpid()))
		monitorHub = &SystemMonitorHub{
			connections: make(map[*websocket.Conn]bool),
			process:     p,
			numCPU:      runtime.NumCPU(),
			stopChan:    make(chan struct{}),
			httpMetrics: &HTTPMetrics{
				startTime: time.Now(),
			},
			websocketMetrics: &WebSocketMetrics{},
			lastIOStats: &IOStats{
				Timestamp: time.Now(),
			},
			lastNetStats:  make([]net.IOCountersStat, 0),
			lastDiskStats: make(map[string]disk.IOCountersStat),
		}
		monitorHub.httpMetrics.lastResetTime.Store(time.Now().Unix())
		monitorHub.startMonitoring()
	})
	return monitorHub
}

// HTTP Monitoring Middleware
func HTTPMonitoringMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		hub := getMonitorHub()
		start := time.Now()

		// Increment active requests
		hub.httpMetrics.activeRequests.Add(1)
		hub.httpMetrics.totalRequests.Add(1)

		// Track method
		method := c.Method()
		if val, ok := hub.httpMetrics.methods.Load(method); ok {
			hub.httpMetrics.methods.Store(method, val.(uint64)+1)
		} else {
			hub.httpMetrics.methods.Store(method, uint64(1))
		}

		// Track endpoint (simplified path)
		endpoint := c.Path()
		if val, ok := hub.httpMetrics.endpoints.Load(endpoint); ok {
			hub.httpMetrics.endpoints.Store(endpoint, val.(uint64)+1)
		} else {
			hub.httpMetrics.endpoints.Store(endpoint, uint64(1))
		}

		// Process request
		err := c.Next()

		// Decrement active requests
		hub.httpMetrics.activeRequests.Add(-1)

		// Track response time
		duration := time.Since(start)
		responseTimeMs := float64(duration.Nanoseconds()) / 1e6

		// Track total response time for averages
		hub.httpMetrics.totalResponseTime.Add(uint64(duration.Nanoseconds()))

		// Track slow requests (> 1 second)
		if duration > time.Second {
			hub.httpMetrics.slowRequests.Add(1)
		}

		hub.httpMetrics.responseTimesMutex.Lock()
		hub.httpMetrics.responseTimes = append(hub.httpMetrics.responseTimes, responseTimeMs)
		// Keep only last 1000 response times for memory efficiency
		if len(hub.httpMetrics.responseTimes) > 1000 {
			hub.httpMetrics.responseTimes = hub.httpMetrics.responseTimes[len(hub.httpMetrics.responseTimes)-1000:]
		}
		hub.httpMetrics.responseTimesMutex.Unlock()

		// Track status code
		statusCode := c.Response().StatusCode()
		statusStr := fmt.Sprintf("%d", statusCode)
		if val, ok := hub.httpMetrics.statusCodes.Load(statusStr); ok {
			hub.httpMetrics.statusCodes.Store(statusStr, val.(uint64)+1)
		} else {
			hub.httpMetrics.statusCodes.Store(statusStr, uint64(1))
		}

		return err
	}
}

func (h *SystemMonitorHub) addConnection(conn *websocket.Conn) {
	h.mutex.Lock()
	defer h.mutex.Unlock()
	h.connections[conn] = true

	// Update WebSocket metrics
	h.websocketMetrics.totalConnections.Add(1)
	h.websocketMetrics.activeConnections.Add(1)
	h.websocketMetrics.lastConnectionTime.Store(time.Now().Unix())

	log.Info("System monitor WebSocket connection added", "total_connections", len(h.connections))
}

func (h *SystemMonitorHub) removeConnection(conn *websocket.Conn) {
	h.mutex.Lock()
	defer h.mutex.Unlock()
	delete(h.connections, conn)

	// Update WebSocket metrics
	h.websocketMetrics.activeConnections.Add(-1)

	log.Info("System monitor WebSocket connection removed", "total_connections", len(h.connections))
}

func (h *SystemMonitorHub) broadcast(stats SystemStats) {
	h.mutex.RLock()
	connections := make([]*websocket.Conn, 0, len(h.connections))
	for conn := range h.connections {
		connections = append(connections, conn)
	}
	h.mutex.RUnlock()

	for _, conn := range connections {
		if err := conn.WriteJSON(stats); err != nil {
			log.Warn("Failed to send system stats via WebSocket", "error", err)
			h.removeConnection(conn)
			h.websocketMetrics.connectionErrors.Add(1)
		} else {
			h.websocketMetrics.messagesSent.Add(1)
		}
	}
}

func (h *SystemMonitorHub) startMonitoring() {
	go func() {
		ticker := time.NewTicker(2 * time.Second) // Update every 2 seconds
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				stats := h.collectStats()
				h.statsMutex.Lock()
				h.stats = stats
				h.statsMutex.Unlock()

				// Only broadcast if there are active connections
				h.mutex.RLock()
				hasConnections := len(h.connections) > 0
				h.mutex.RUnlock()

				if hasConnections {
					h.broadcast(stats)
				}
			case <-h.stopChan:
				return
			}
		}
	}()
}

func (h *SystemMonitorHub) collectStats() SystemStats {
	stats := SystemStats{
		Timestamp: time.Now().UnixMilli(),
	}

	// Collect PID stats
	h.collectPIDStats(&stats)

	// Collect OS stats
	h.collectOSStats(&stats)

	// Collect HTTP stats
	h.collectHTTPStats(&stats)

	// Collect Runtime stats
	h.collectRuntimeStats(&stats)

	// Collect WebSocket stats
	h.collectWebSocketStats(&stats)

	return stats
}

func (h *SystemMonitorHub) collectPIDStats(stats *SystemStats) {
	// PID CPU usage
	if pidCPU, err := h.process.Percent(0); err == nil {
		stats.PID.CPU = pidCPU / float64(h.numCPU)
	}

	// PID RAM usage
	if pidRAM, err := h.process.MemoryInfo(); err == nil && pidRAM != nil {
		stats.PID.RAM = pidRAM.RSS
	}

	// PID connections
	if pidConns, err := net.ConnectionsPid("tcp", h.process.Pid); err == nil {
		stats.PID.Conns = len(pidConns)
	}

	// PID file descriptors
	if pidFDs, err := h.process.OpenFiles(); err == nil {
		stats.PID.FDs = len(pidFDs)
	}

	// PID threads
	if pidThreads, err := h.process.NumThreads(); err == nil {
		stats.PID.Threads = int(pidThreads)
	}

	// PID uptime - Calculate from create time
	if createTime, err := h.process.CreateTime(); err == nil {
		stats.PID.Uptime = float64(time.Now().Unix()) - float64(createTime/1000)
	}

	// PID CPU time - Use percent instead as CPUTimes may not be available
	if pidCPUTimes, err := h.process.Times(); err == nil {
		stats.PID.CPUTime = pidCPUTimes.Total()
	}

	// PID context switches
	if pidContextSwitches, err := h.process.NumCtxSwitches(); err == nil {
		stats.PID.ContextSwitches = uint64(pidContextSwitches.Voluntary)
	}
}

func (h *SystemMonitorHub) collectOSStats(stats *SystemStats) {
	// OS CPU usage
	if osCPU, err := cpu.Percent(0, false); err == nil && len(osCPU) > 0 {
		stats.OS.CPU = osCPU[0]
	}

	// OS RAM usage
	if osRAM, err := mem.VirtualMemory(); err == nil && osRAM != nil {
		stats.OS.RAM = osRAM.Used
		stats.OS.TotalRAM = osRAM.Total
	}

	// Load averages (1, 5, and 15 minute)
	if loadAvg, err := load.Avg(); err == nil && loadAvg != nil {
		stats.OS.LoadAvg = loadAvg.Load1
		stats.OS.LoadAvg5 = loadAvg.Load5
		stats.OS.LoadAvg15 = loadAvg.Load15
	}

	// OS connections
	if osConns, err := net.Connections("tcp"); err == nil {
		stats.OS.Conns = len(osConns)
	}

	// I/O stats (calculate rates)
	h.collectIOStats(stats)

	// OS swap usage
	if swapStats, err := mem.SwapMemory(); err == nil && swapStats != nil {
		stats.OS.SwapUsed = swapStats.Used
		stats.OS.SwapTotal = swapStats.Total
	}

	// System uptime - Use a simple calculation from load avg timestamp
	stats.OS.Uptime = 0 // Will be calculated if possible

	// System boot time - Set to 0 for now as BootTime may not be available
	stats.OS.BootTime = 0

	// Process counts
	if allProcesses, err := process.Processes(); err == nil {
		stats.OS.TotalProcesses = len(allProcesses)
		runningCount := 0
		for _, p := range allProcesses {
			if status, err := p.Status(); err == nil {
				// Check if process is in running state (status is array, check first element)
				if len(status) > 0 && (status[0] == "R" || status[0] == "S") {
					runningCount++
				}
			}
		}
		stats.OS.RunningProcesses = runningCount
	}
}

func (h *SystemMonitorHub) collectIOStats(stats *SystemStats) {
	now := time.Now()
	timeDelta := now.Sub(h.lastIOStats.Timestamp).Seconds()

	// Network I/O
	if netStats, err := net.IOCounters(false); err == nil && len(netStats) > 0 {
		currentNet := netStats[0]
		if len(h.lastNetStats) > 0 && timeDelta > 0 {
			lastNet := h.lastNetStats[0]
			stats.OS.NetReadBytes = uint64(float64(currentNet.BytesRecv-lastNet.BytesRecv) / timeDelta)
			stats.OS.NetWriteBytes = uint64(float64(currentNet.BytesSent-lastNet.BytesSent) / timeDelta)
		}
		h.lastNetStats = netStats
	}

	// Disk I/O
	if diskStats, err := disk.IOCounters(); err == nil {
		var totalReadBytes, totalWriteBytes uint64
		var deltaReadBytes, deltaWriteBytes uint64

		for device, current := range diskStats {
			totalReadBytes += current.ReadBytes
			totalWriteBytes += current.WriteBytes

			if last, exists := h.lastDiskStats[device]; exists && timeDelta > 0 {
				deltaReadBytes += current.ReadBytes - last.ReadBytes
				deltaWriteBytes += current.WriteBytes - last.WriteBytes
			}
		}

		if timeDelta > 0 && len(h.lastDiskStats) > 0 {
			stats.OS.DiskReadBytes = uint64(float64(deltaReadBytes) / timeDelta)
			stats.OS.DiskWriteBytes = uint64(float64(deltaWriteBytes) / timeDelta)
		}

		h.lastDiskStats = diskStats
	}

	h.lastIOStats.Timestamp = now
}

func (h *SystemMonitorHub) collectHTTPStats(stats *SystemStats) {
	// Basic counters
	stats.HTTP.TotalRequests = h.httpMetrics.totalRequests.Load()
	stats.HTTP.ActiveRequests = h.httpMetrics.activeRequests.Load()
	stats.HTTP.LastResetTime = h.httpMetrics.lastResetTime.Load()
	stats.HTTP.SlowRequests = h.httpMetrics.slowRequests.Load()

	// Total response time in milliseconds
	totalResponseTimeNs := h.httpMetrics.totalResponseTime.Load()
	stats.HTTP.TotalResponseTime = float64(totalResponseTimeNs) / 1e6

	// Requests per second
	elapsed := time.Since(h.httpMetrics.startTime).Seconds()
	if elapsed > 0 {
		stats.HTTP.RequestsPerSecond = float64(stats.HTTP.TotalRequests) / elapsed
	}

	// Response times
	h.httpMetrics.responseTimesMutex.RLock()
	if len(h.httpMetrics.responseTimes) > 0 {
		// Calculate average
		var sum float64
		for _, rt := range h.httpMetrics.responseTimes {
			sum += rt
		}
		stats.HTTP.AvgResponseTimeMs = sum / float64(len(h.httpMetrics.responseTimes))

		// Calculate P95 and P99
		if len(h.httpMetrics.responseTimes) >= 2 {
			sortedTimes := make([]float64, len(h.httpMetrics.responseTimes))
			copy(sortedTimes, h.httpMetrics.responseTimes)
			sort.Float64s(sortedTimes)

			// P95
			p95Index := int(float64(len(sortedTimes)) * 0.95)
			if p95Index >= len(sortedTimes) {
				p95Index = len(sortedTimes) - 1
			}
			stats.HTTP.P95ResponseTimeMs = sortedTimes[p95Index]

			// P99
			p99Index := int(float64(len(sortedTimes)) * 0.99)
			if p99Index >= len(sortedTimes) {
				p99Index = len(sortedTimes) - 1
			}
			stats.HTTP.P99ResponseTimeMs = sortedTimes[p99Index]
		}
	}
	h.httpMetrics.responseTimesMutex.RUnlock()

	// Status codes
	stats.HTTP.StatusCodes = make(map[string]uint64)
	var errorCount uint64
	h.httpMetrics.statusCodes.Range(func(key, value interface{}) bool {
		statusStr := key.(string)
		count := value.(uint64)
		stats.HTTP.StatusCodes[statusStr] = count
		// Count 4xx and 5xx as errors (check if status code starts with 4 or 5)
		if len(statusStr) > 0 && (statusStr[0] == '4' || statusStr[0] == '5') {
			errorCount += count
		}
		return true
	})

	// Error rate
	if stats.HTTP.TotalRequests > 0 {
		stats.HTTP.ErrorRate = (float64(errorCount) / float64(stats.HTTP.TotalRequests)) * 100
	}

	// Methods
	stats.HTTP.RequestsByMethod = make(map[string]uint64)
	h.httpMetrics.methods.Range(func(key, value interface{}) bool {
		stats.HTTP.RequestsByMethod[key.(string)] = value.(uint64)
		return true
	})

	// Endpoints
	stats.HTTP.RequestsByEndpoint = make(map[string]uint64)
	h.httpMetrics.endpoints.Range(func(key, value interface{}) bool {
		stats.HTTP.RequestsByEndpoint[key.(string)] = value.(uint64)
		return true
	})
}

func (h *SystemMonitorHub) collectRuntimeStats(stats *SystemStats) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	stats.Runtime.Goroutines = runtime.NumGoroutine()
	stats.Runtime.HeapAlloc = m.HeapAlloc
	stats.Runtime.HeapSys = m.HeapSys
	stats.Runtime.HeapInuse = m.HeapInuse
	stats.Runtime.HeapObjects = m.HeapObjects
	stats.Runtime.StackInuse = m.StackInuse
	stats.Runtime.GCCycles = m.NumGC
	stats.Runtime.GCPauseTotalNs = m.PauseTotalNs
	stats.Runtime.NextGC = m.NextGC
	stats.Runtime.Version = runtime.Version()
}

func (h *SystemMonitorHub) collectWebSocketStats(stats *SystemStats) {
	stats.WebSocket.TotalConnections = h.websocketMetrics.totalConnections.Load()
	stats.WebSocket.ActiveConnections = h.websocketMetrics.activeConnections.Load()
	stats.WebSocket.MessagesSent = h.websocketMetrics.messagesSent.Load()
	stats.WebSocket.MessagesReceived = h.websocketMetrics.messagesReceived.Load()
	stats.WebSocket.ConnectionErrors = h.websocketMetrics.connectionErrors.Load()
	stats.WebSocket.LastConnectionTime = h.websocketMetrics.lastConnectionTime.Load()
}

func (h *SystemMonitorHub) getCurrentStats() SystemStats {
	h.statsMutex.RLock()
	defer h.statsMutex.RUnlock()
	return h.stats
}

// SystemMonitorWebSocket handles WebSocket connections for real-time system monitoring
func SystemMonitorWebSocket() fiber.Handler {
	return websocket.New(func(c *websocket.Conn) {
		ctx := context.Background()
		hub := getMonitorHub()

		// Add connection to hub
		hub.addConnection(c)
		defer hub.removeConnection(c)

		log.Info("System monitor WebSocket connection established", "remote_addr", c.RemoteAddr().String())

		// Send initial stats immediately
		initialStats := hub.getCurrentStats()
		if err := c.WriteJSON(initialStats); err != nil {
			log.Error("Failed to send initial system stats", "error", err)
			return
		}

		// Handle incoming messages (ping/pong, etc.)
		for {
			select {
			case <-ctx.Done():
				return
			default:
				var msg map[string]interface{}
				if err := c.ReadJSON(&msg); err != nil {
					log.Debug("System monitor WebSocket connection closed", "error", err)
					return
				}

				// Track received message
				hub.websocketMetrics.messagesReceived.Add(1)

				// Handle ping messages
				if msgType, ok := msg["type"].(string); ok && msgType == "ping" {
					response := map[string]interface{}{
						"type":      "pong",
						"timestamp": time.Now().UnixMilli(),
					}
					if err := c.WriteJSON(response); err != nil {
						log.Error("Failed to send pong", "error", err)
						return
					}
					hub.websocketMetrics.messagesSent.Add(1)
				}
			}
		}
	})
}
