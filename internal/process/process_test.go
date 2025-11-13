package process

import (
	"reflect"
	"testing"
)

// mockProcess implements processProvider for testing
type mockProcess struct {
	pid     int32
	environ []string
	cpu     float64
	mem     float32
	envErr  error
	cpuErr  error
	memErr  error
}

func (m *mockProcess) Environ() ([]string, error) {
	return m.environ, m.envErr
}

func (m *mockProcess) CPUPercent() (float64, error) {
	return m.cpu, m.cpuErr
}

func (m *mockProcess) MemoryPercent() (float32, error) {
	return m.mem, m.memErr
}

func (m *mockProcess) Pid() int32 {
	return m.pid
}

func TestExtractJenkinsInfo(t *testing.T) {
	tests := []struct {
		name    string
		proc    processProvider
		environ []string
		want    *ProcessInfo
	}{
		{
			name: "valid Jenkins process",
			proc: &mockProcess{
				pid: 1234,
				cpu: 10.5,
				mem: 20.2,
			},
			environ: []string{
				"JOB_NAME=build_app",
				"BUILD_ID=42",
				"STAGE_NAME=test",
				"WORKSPACE=/var/lib/jenkins/workspace/build_app",
			},
			want: &ProcessInfo{
				PID:          1234,
				BuildJobName: "build_app",
				BuildId:      "42",
				StageName:    "test",
				WorkSpace:    "/var/lib/jenkins/workspace/build_app",
				CPU:          10.5,
				Mem:          20.2,
			},
		},
		{
			name: "not a Jenkins process (missing JOB_NAME)",
			proc: &mockProcess{
				pid: 5678,
				cpu: 2.5,
				mem: 5.1,
			},
			environ: []string{
				"USER=root",
				"PATH=/usr/bin",
			},
			want: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractJenkinsInfo(tt.proc, tt.environ)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("extractJenkinsInfo() = %+v, want %+v", got, tt.want)
			}
		})
	}
}
