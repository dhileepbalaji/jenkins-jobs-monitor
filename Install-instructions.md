# Installation Instructions for Jenkins Monitor

This document provides instructions on how to install and run the `jenkins-monitor` application on a Linux system using `systemd`.

## Prerequisites

*   A built `jenkins-monitor` executable for your target architecture (x86_64 or arm64). Refer to `Build.md` for build instructions.
*   `sudo` privileges on the target Linux machine.

## Installation Steps

1.  **Copy the Executable:**
    Transfer the built `jenkins-monitor` executable to the `/usr/local/bin/` directory on your target Linux machine. Replace `[ARCHITECTURE]` with `x86_64` or `arm64` as appropriate.

    ```bash
    sudo cp cmd/jenkins-monitor/jenkins-monitor-[ARCHITECTURE] /usr/local/bin/jenkins-monitor
    ```

2.  **Create a Systemd Service File:**
    Create a systemd service file named `jenkins-monitor.service` in `/etc/systemd/system/` with the following content:

    ```ini
    [Unit]
    Description=Jenkins Monitor Service
    After=network.target

    [Service]
    ExecStart=/usr/local/bin/jenkins-monitor monitor --output /var/lib/jenkins-monitor/processes.csv
    Restart=always
    User=jenkins # Or any other non-root user
    Group=jenkins # Or any other group
    StandardOutput=journal
    StandardError=journal
    SyslogIdentifier=jenkins-monitor

    [Install]
    WantedBy=multi-user.target
    ```

    **Note:**
    *   Adjust `User` and `Group` to an appropriate non-root user on your system (e.g., `jenkins` if Jenkins is installed, or `nobody`).
    *   The `--output` path `/var/lib/jenkins-monitor/processes.csv` requires the `jenkins` user (or chosen user) to have write permissions to `/var/lib/jenkins-monitor/`. Ensure this directory exists and has correct permissions:
        ```bash
        sudo mkdir -p /var/lib/jenkins-monitor
        sudo chown jenkins:jenkins /var/lib/jenkins-monitor # Adjust user/group as needed
        ```
    *   The log file `jenkinsjobmonitor.log` will be created in the directory where the `jenkins-monitor` executable is run from, which in this case is `/usr/local/bin/`. Ensure the `jenkins` user has write permissions to `/usr/local/bin/` or modify the `logFilePath` variable in `internal/utils/utils.go` to a different location (e.g., `/var/log/jenkins-monitor/jenkinsjobmonitor.log`) and ensure that directory has appropriate permissions.

3.  **Reload Systemd and Enable the Service:**

    ```bash
    sudo systemctl daemon-reload
    sudo systemctl enable jenkins-monitor.service
    sudo systemctl start jenkins-monitor.service
    ```

4.  **Check Service Status:**

    ```bash
    sudo systemctl status jenkins-monitor.service
    journalctl -u jenkins-monitor.service -f
    ```

## Running the Analyzer (Ad-hoc)

To run the `analyze` or `adhoc` commands, you can execute them manually:

```bash
/usr/local/bin/jenkins-monitor analyze --input /var/lib/jenkins-monitor/processes.csv
/usr/local/bin/jenkins-monitor adhoc
