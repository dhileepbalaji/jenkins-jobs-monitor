package monitor

import (
	"encoding/csv"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
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
				utils.Fatal(fmt.Sprintf("Failed to start Prometheus metrics server: %v", err))
			}
		}()
	}

	var file *os.File
	var writer *csv.Writer
	var currentDay int

	// Initialize CSV collection if enabled
	if !cfg.DisableCollection {
		// Ensure the output directory exists
		outputDir := utils.GetDir(outputFile)
		if _, err := os.Stat(outputDir); os.IsNotExist(err) {
			if err := os.MkdirAll(outputDir, 0755); err != nil {
				utils.Fatal(fmt.Sprintf("Failed to create output directory: %v", err))
			}
		}

		openCSV := func() {
			var err error
			file, err = os.OpenFile(outputFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			if err != nil {
				utils.Fatal(fmt.Sprintf("Failed to open output file: %v", err))
			}
			writer = csv.NewWriter(file)

			// Write header if the file is new
			info, err := file.Stat()
			if err != nil {
				utils.Fatal(fmt.Sprintf("Failed to get file info: %v", err))
			}
			if info.Size() == 0 {
				writer.Write([]string{"timestamp", "pid", "cpu", "mem", "build_path"})
				writer.Flush()
			}
		}

		openCSV()
		defer func() {
			if file != nil {
				file.Close()
			}
		}()
		currentDay = time.Now().Day()
	} else {
		utils.Info("Collection disabled via config. Only alerting will be active.")
	}

	// Create a channel to receive OS signals
	sigs := make(chan os.Signal, 1)
	// Register the channel to receive SIGINT and SIGTERM signals
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	// Create a ticker that ticks every 30 seconds
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	if !cfg.DisableCollection {
		utils.Info(fmt.Sprintf("Starting process monitoring. Writing to %s", outputFile))
	} else {
		utils.Info("Starting process monitoring (Alerting Only).")
	}
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

			// Log Rotation Logic
			if !cfg.DisableCollection {
				now := time.Now()
				if now.Day() != currentDay {
					utils.Info("Rotating log file...")
					file.Close()

					// Rename old file
					yesterday := now.AddDate(0, 0, -1)
					rotatedName := fmt.Sprintf("%s.%s.csv", outputFile, yesterday.Format("2006-01-02"))
					// Handle if outputFile is just a filename or path
					ext := filepath.Ext(outputFile)
					base := strings.TrimSuffix(outputFile, ext)
					rotatedName = fmt.Sprintf("%s.%s%s", base, yesterday.Format("2006-01-02"), ext)

					if err := os.Rename(outputFile, rotatedName); err != nil {
						utils.Error(fmt.Sprintf("Failed to rotate log file: %v", err))
					}

					// Re-open new file
					var err error
					file, err = os.OpenFile(outputFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
					if err != nil {
						utils.Fatal(fmt.Sprintf("Failed to open output file: %v", err))
					}
					writer = csv.NewWriter(file)
					writer.Write([]string{"timestamp", "pid", "cpu", "mem", "build_path"})
					writer.Flush()

					currentDay = now.Day()
					utils.Info(fmt.Sprintf("Log rotated. New file: %s", outputFile))
				}
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

				// Write to CSV if collection is enabled
				if !cfg.DisableCollection {
					record := []string{
						timestamp,
						fmt.Sprintf("%d", p.PID),
						fmt.Sprintf("%.2f", p.CPU),
						fmt.Sprintf("%.2f", p.Mem),
						p.BuildJobName,
					}
					writer.Write(record)
				}
			}
			if !cfg.DisableCollection {
				writer.Flush()
				utils.Info(fmt.Sprintf("Collected data for %d processes", len(processes)))
			} else {
				utils.Info(fmt.Sprintf("Monitored %d processes (Collection Disabled)", len(processes)))
			}

		case <-sigs:
			utils.Info("Exiting...")
			return
		}
	}
}
