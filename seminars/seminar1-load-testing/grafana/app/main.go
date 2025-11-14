package main

import (
    "fmt"
    "log"
    "net/http"

    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promhttp"
    "github.com/shirou/gopsutil/v3/cpu"
    "github.com/shirou/gopsutil/v3/disk"
    "github.com/shirou/gopsutil/v3/host"
    "github.com/shirou/gopsutil/v3/mem"
)

// Example counter
var requestsTotal = prometheus.NewCounterVec(
    prometheus.CounterOpts{
        Name: "app_requests_total",
        Help: "Total HTTP requests",
    },
    []string{"path"},
)

func init() {
    prometheus.MustRegister(requestsTotal)
}

func handler(w http.ResponseWriter, r *http.Request) {
    requestsTotal.WithLabelValues(r.URL.Path).Inc()
    fmt.Fprintln(w, "OK")
}

func main() {
    // Application HTTP handler
    http.HandleFunc("/", handler)

    // Run main application server on :8081
    go func() {
        fmt.Println("App server listening on :8081")
        if err := http.ListenAndServe(":8081", nil); err != nil {
            log.Fatalf("app server failed: %v", err)
        }
    }()

    // Create a separate Prometheus registry for hardware metrics
    hwRegistry := prometheus.NewRegistry()

    // CPU percent (overall)
    cpuGauge := prometheus.NewGaugeFunc(prometheus.GaugeOpts{
        Name: "host_cpu_percent",
        Help: "Total CPU usage percent",
    }, func() float64 {
        pct, err := cpu.Percent(0, false)
        if err != nil || len(pct) == 0 {
            return 0
        }
        return pct[0]
    })
    hwRegistry.MustRegister(cpuGauge)

    // Memory
    memTotal := prometheus.NewGaugeFunc(prometheus.GaugeOpts{
        Name: "host_memory_total_bytes",
        Help: "Total system memory in bytes",
    }, func() float64 {
        v, err := mem.VirtualMemory()
        if err != nil {
            return 0
        }
        return float64(v.Total)
    })
    memUsed := prometheus.NewGaugeFunc(prometheus.GaugeOpts{
        Name: "host_memory_used_bytes",
        Help: "Used system memory in bytes",
    }, func() float64 {
        v, err := mem.VirtualMemory()
        if err != nil {
            return 0
        }
        return float64(v.Used)
    })
    hwRegistry.MustRegister(memTotal, memUsed)

    // Disk usage for root
    diskTotal := prometheus.NewGaugeFunc(prometheus.GaugeOpts{
        Name: "host_disk_total_bytes",
        Help: "Total disk bytes for root",
    }, func() float64 {
        d, err := disk.Usage("/")
        if err != nil {
            return 0
        }
        return float64(d.Total)
    })
    diskUsed := prometheus.NewGaugeFunc(prometheus.GaugeOpts{
        Name: "host_disk_used_bytes",
        Help: "Used disk bytes for root",
    }, func() float64 {
        d, err := disk.Usage("/")
        if err != nil {
            return 0
        }
        return float64(d.Used)
    })
    hwRegistry.MustRegister(diskTotal, diskUsed)

    // Host uptime
    uptime := prometheus.NewGaugeFunc(prometheus.GaugeOpts{
        Name: "host_uptime_seconds",
        Help: "System uptime in seconds",
    }, func() float64 {
        u, err := host.Uptime()
        if err != nil {
            return 0
        }
        return float64(u)
    })
    hwRegistry.MustRegister(uptime)

    // Start hardware metrics server on :8082
    go func() {
        mux := http.NewServeMux()
        mux.Handle("/metrics", promhttp.HandlerFor(hwRegistry, promhttp.HandlerOpts{}))
        fmt.Println("Hardware metrics server listening on :8082")
        if err := http.ListenAndServe(":8082", mux); err != nil {
            log.Fatalf("hardware metrics server failed: %v", err)
        }
    }()

    // Block forever
    select {}
}
