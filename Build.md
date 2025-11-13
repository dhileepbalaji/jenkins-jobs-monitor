# Build Instructions for Jenkins Monitor

This document provides instructions on how to build the `jenkins-monitor` application for different architectures.

## Prerequisites

*   Go (version 1.18 or higher) installed on your system.

## Building for Linux x86_64

To build the executable for Linux x86_64 architecture, run the following command from the project root directory:

```bash
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o cmd/jenkins-monitor/jenkins-monitor-x86_64 cmd/jenkins-monitor/main.go
```

This will create an executable named `jenkins-monitor-x86_64` in the `cmd/jenkins-monitor/` directory.

## Building for Linux arm64

To build the executable for Linux arm64 architecture, run the following command from the project root directory:

```bash
CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -o cmd/jenkins-monitor/jenkins-monitor-arm64 cmd/jenkins-monitor/main.go
```

This will create an executable named `jenkins-monitor-arm64` in the `cmd/jenkins-monitor/` directory.

## Building for Current OS and Architecture

To build the executable for your current operating system and architecture, simply run:

```bash
go build -o cmd/jenkins-monitor/jenkins-monitor cmd/jenkins-monitor/main.go
```

This will create an executable named `jenkins-monitor` in the `cmd/jenkins-monitor/` directory.
