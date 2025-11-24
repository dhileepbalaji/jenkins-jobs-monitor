package adhoc

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"text/tabwriter"

	gopsutil "github.com/shirou/gopsutil/v3/process"

	"jenkins-monitor/internal/process"
	"jenkins-monitor/internal/utils"
)

func RunAdhoc() {
	processes, err := process.GetJenkinsProcesses()
	if err != nil {
		utils.Fatal(fmt.Sprintf("Error getting Jenkins processes: %v", err))
	}

	if len(processes) == 0 {
		utils.Info("No processes with JOB_NAME found")
		fmt.Println("No processes with JOB_NAME found")
		return
	}

	// Sort by CPU usage (descending)
	sort.Slice(processes, func(i, j int) bool {
		return processes[i].CPU > processes[j].CPU
	})

	fmt.Println()
	fmt.Println("🔍 Scanning processes for Jenkins Jobs...")
	fmt.Println()

	// Create tabwriter for aligned columns
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)

	fmt.Fprintf(w, "%-10s\t%-8s\t%-8s\t%-20s\t%-35s\t%-15s\t%-40s\t%-10s\n",
		"PID", "CPU%", "MEM%", "PROCESS", "JOB_NAME", "BUILD_ID", "WORKSPACE", "STAGE_NAME")
	fmt.Fprintln(w, strings.Repeat("-", 160))

	for _, p := range processes {
		name := "unknown"
		if proc, err := gopsutil.NewProcess(p.PID); err == nil {
			if n, err := proc.Name(); err == nil {
				name = n
			}
		}

		fmt.Fprintf(w, "%-10d\t%8.1f\t%8.1f\t%-20s\t%-35s\t%-15s\t%-40s\t%-10s\n",
			p.PID, p.CPU, p.Mem, name, p.BuildJobName, p.BuildId, p.WorkSpace, p.StageName)
	}

	w.Flush()
	fmt.Println(strings.Repeat("-", 160))
	fmt.Printf("✅ Total processes found: %d\n", len(processes))
}
