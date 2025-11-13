package monitor

import (
	"encoding/csv"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"jenkins-monitor/internal/process"
	"jenkins-monitor/internal/utils"
)

func RunMonitor(outputFile string) {
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
				record := []string{
					timestamp,
					fmt.Sprintf("%d", p.PID),
					p.BuildJobName,
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
