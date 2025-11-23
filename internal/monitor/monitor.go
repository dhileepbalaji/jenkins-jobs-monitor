package monitor

import (
	"encoding/csv"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"jenkins-monitor/internal/config"
	"jenkins-monitor/internal/notifier"
	"jenkins-monitor/internal/process"
	"jenkins-monitor/internal/utils"
)

var (
	jenkinsCPUUsage = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "jenkins_job_cpu_usage_percent",
			Help: "Current CPU usage percentage of Jenkins jobs.",
		},
		[]string{"job_name", "pid"},
	)
	jenkinsMemoryUsage = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "jenkins_job_memory_usage_percent",
			Help: "Current memory usage percentage of Jenkins jobs.",
		},
		[]string{"job_name", "pid"},
	)
)

func init() {
	// Register the metrics with Prometheus's default registry.
	prometheus.MustRegister(jenkinsCPUUsage)
	prometheus.MustRegister(jenkinsMemoryUsage)
}

func RunMonitor(outputFile string, cfg *config.Config) {
	// Start Prometheus metrics HTTP server
	if cfg.Prometheus.ListenAddress != "" {
		go func() {
			http.Handle("/metrics", promhttp.Handler())
			utils.Info(fmt.Sprintf("Starting Prometheus metrics server on %s", cfg.Prometheus.ListenAddress))
			if err := http.ListenAndServe(cfg.Prometheus.ListenAddress, nil); err != nil {
				utils.Fatal(fmt.Sprintf("Failed to start Prometheus metrics server: %v", err)) // Changed to Fatal
			}
		}()
	}

	// Ensure the output directory exists
	outputDir := utils.GetDir(outputFile)
	if _, err := os.Stat(outputDir); os.IsNotExist(err) {
		if err := os.MkdirAll(outputDir, 0755); err != nil {
			utils.Fatal(fmt.Sprintf("Failed to create output directory: %v", err))
		}
	}

	// Create or open the output file
	file, err := os.OpenFile(outputFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		utils.Fatal(fmt.Sprintf("Failed to open output file: %v", err))
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write header if the file is new
	info, err := file.Stat()
	if err != nil {
		utils.Fatal(fmt.Sprintf("Failed to get file info: %v", err))
	}
	if info.Size() == 0 {
		writer.Write([]string{"timestamp", "pid", "build_path", "cpu", "mem"})
	}

	// Create a channel to receive OS signals
	sigs := make(chan os.Signal, 1)
	// Register the channel to receive SIGINT and SIGTERM signals
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	// Create a ticker that ticks every 30 seconds
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	utils.Info(fmt.Sprintf("Starting process monitoring. Writing to %s", outputFile))
	utils.Info("Press Ctrl+C to stop...")

	// Run the collection logic in a loop
	for {
		select {
		case <-ticker.C:
			processes, err := process.GetJenkinsProcesses()
			if err != nil {
				utils.Error(fmt.Sprintf("Error getting Jenkins processes: %v", err))
				continue
			}

			timestamp := time.Now().UTC().Format(time.RFC3339)
			for _, p := range processes {
				// Update Prometheus metrics
				labels := prometheus.Labels{"job_name": p.BuildJobName, "pid": fmt.Sprintf("%d", p.PID)}
				jenkinsCPUUsage.With(labels).Set(p.CPU)
				jenkinsMemoryUsage.With(labels).Set(float64(p.Mem))

				// Check for thresholds and send Slack notifications
				if cfg.Thresholds.CPUPercent > 0 && p.CPU >= cfg.Thresholds.CPUPercent {
					utils.Info(fmt.Sprintf("High CPU usage detected for job %s (PID %d): %.2f%% (Threshold: %.2f%%)", p.BuildJobName, p.PID, p.CPU, cfg.Thresholds.CPUPercent))
					notifier.SendSlackNotification(cfg, "CPU_HIGH", &p)
				}
				if cfg.Thresholds.MemPercent > 0 && float64(p.Mem) >= cfg.Thresholds.MemPercent {
					utils.Info(fmt.Sprintf("High Memory usage detected for job %s (PID %d): %.2f%% (Threshold: %.2f%%)", p.BuildJobName, p.PID, p.Mem, cfg.Thresholds.MemPercent))
					notifier.SendSlackNotification(cfg, "MEM_HIGH", &p)
				}

				// Write to CSV
				record := []string{
					timestamp,
					fmt.Sprintf("%d", p.PID),
					p.BuildJobName, // Use BuildJobName for build_path in CSV
					fmt.Sprintf("%.2f", p.CPU),
					fmt.Sprintf("%.2f", p.Mem),
				}
				writer.Write(record)
			}
			writer.Flush()
			utils.Info(fmt.Sprintf("Collected data for %d processes", len(processes)))

		case <-sigs:
			utils.Info("Exiting...")
			return
		}
	}
}
