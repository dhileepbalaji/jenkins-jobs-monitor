package analyze

import (
	"encoding/csv"
	"fmt"
	"os"
	"sort"
	"time"

	"jenkins-monitor/internal/utils"
)

func RunAnalyzer(inputFile string) {
	file, err := os.Open(inputFile)
	if err != nil {
		utils.Fatal(fmt.Sprintf("Failed to open input file: %v", err))
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		utils.Fatal(fmt.Sprintf("Failed to read CSV file: %v", err))
	}

	if len(records) <= 1 {
		fmt.Println("No data to analyze.")
		return
	}

	// Skip header
	records = records[1:]

	type JobStats struct {
		PeakCPU     float64
		PeakCPUTime string
		PeakMem     float64
		PeakMemTime string
	}

	jobStats := make(map[string]*JobStats)

	for _, record := range records {
		buildPath := record[2]
		cpu, _ := utils.ParseFloat(record[3])
		mem, _ := utils.ParseFloat(record[4])
		timestamp := record[0]

		if _, ok := jobStats[buildPath]; !ok {
			jobStats[buildPath] = &JobStats{}
		}

		if cpu > jobStats[buildPath].PeakCPU {
			jobStats[buildPath].PeakCPU = cpu
			jobStats[buildPath].PeakCPUTime = timestamp
		}

		if mem > jobStats[buildPath].PeakMem {
			jobStats[buildPath].PeakMem = mem
			jobStats[buildPath].PeakMemTime = timestamp
		}
	}

	type JobPeak struct {
		BuildPath string
		Value     float64
		Timestamp string
	}

	var cpuPeaks []JobPeak
	var memPeaks []JobPeak

	for buildPath, stats := range jobStats {
		cpuPeaks = append(cpuPeaks, JobPeak{BuildPath: buildPath, Value: stats.PeakCPU, Timestamp: stats.PeakCPUTime})
		memPeaks = append(memPeaks, JobPeak{BuildPath: buildPath, Value: stats.PeakMem, Timestamp: stats.PeakMemTime})
	}

	// Sort by CPU peak
	sort.Slice(cpuPeaks, func(i, j int) bool {
		return cpuPeaks[i].Value > cpuPeaks[j].Value
	})

	// Sort by memory peak
	sort.Slice(memPeaks, func(i, j int) bool {
		return memPeaks[i].Value > memPeaks[j].Value
	})

	fmt.Println("Top 5 Jobs by Peak CPU Usage:")
	fmt.Println("-----------------------------------------------------------------")
	for i := 0; i < 5 && i < len(cpuPeaks); i++ {
		fmt.Printf("%-45s %6.2f%%  (at %s)\n", cpuPeaks[i].BuildPath, cpuPeaks[i].Value, cpuPeaks[i].Timestamp)
	}
	fmt.Println()

	fmt.Println("Top 5 Jobs by Peak Memory Usage:")
	fmt.Println("-----------------------------------------------------------------")
	for i := 0; i < 5 && i < len(memPeaks); i++ {
		fmt.Printf("%-45s %6.2f%%  (at %s)\n", memPeaks[i].BuildPath, memPeaks[i].Value, memPeaks[i].Timestamp)
	}
	fmt.Println()

	fmt.Printf("Stats generated at: %s\n", time.Now().Format(time.RFC1123))
}
