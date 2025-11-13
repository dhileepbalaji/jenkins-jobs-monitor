package process

import (
	"strings"

	"github.com/shirou/gopsutil/v3/process"
)

// ProcessInfo holds information about a Jenkins process
type ProcessInfo struct {
	Timestamp    string
	PID          int32
	BuildJobName string
	BuildId      string
	StageName    string
	WorkSpace    string
	CPU          float64
	Mem          float32
}

// processProvider defines what methods we need from gopsutil.Process.
// This makes it mockable for testing.
type processProvider interface {
	Environ() ([]string, error)
	CPUPercent() (float64, error)
	MemoryPercent() (float32, error)
	Pid() int32
}

// realProcess wraps gopsutil.Process so it implements processProvider
type realProcess struct {
	*process.Process
}

func (rp *realProcess) Pid() int32 {
	return rp.Process.Pid
}

func GetJenkinsProcesses() ([]ProcessInfo, error) {
	procs, err := process.Processes()
	if err != nil {
		return nil, err
	}

	var jenkinsProcesses []ProcessInfo
	for _, p := range procs {
		rp := &realProcess{p}

		environ, err := rp.Environ()
		if err != nil {
			continue
		}

		info := extractJenkinsInfo(rp, environ)
		if info != nil {
			jenkinsProcesses = append(jenkinsProcesses, *info)
		}
	}

	return jenkinsProcesses, nil
}

// extractJenkinsInfo parses environment variables for Jenkins process info
func extractJenkinsInfo(p processProvider, environ []string) *ProcessInfo {
	var buildJobName, buildID, stageName, workSpace string

	for _, env := range environ {
		if strings.HasPrefix(env, "JOB_NAME=") {
			buildJobName = strings.TrimPrefix(env, "JOB_NAME=")
		}
		if strings.HasPrefix(env, "BUILD_ID=") {
			buildID = strings.TrimPrefix(env, "BUILD_ID=")
		}
		if strings.HasPrefix(env, "STAGE_NAME=") {
			stageName = strings.TrimPrefix(env, "STAGE_NAME=")
		}
		if strings.HasPrefix(env, "WORKSPACE=") {
			workSpace = strings.TrimPrefix(env, "WORKSPACE=")
		}
	}

	if buildJobName == "" {
		return nil
	}

	cpu, err := p.CPUPercent()
	if err != nil {
		return nil
	}
	mem, err := p.MemoryPercent()
	if err != nil {
		return nil
	}

	return &ProcessInfo{
		PID:          p.Pid(),
		BuildJobName: buildJobName,
		BuildId:      buildID,
		StageName:    stageName,
		WorkSpace:    workSpace,
		CPU:          cpu,
		Mem:          mem,
	}
}
