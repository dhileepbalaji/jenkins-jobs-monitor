package notifier

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io" // Use io.ReadAll instead of ioutil.ReadAll
	"net/http"
	"time"

	"jenkins-monitor/internal/config"
	"jenkins-monitor/internal/process"
	"jenkins-monitor/internal/utils"
)

// SlackMessage represents the structure of a Slack message
type SlackMessage struct {
	Channel     string       `json:"channel,omitempty"`
	Username    string       `json:"username,omitempty"`
	Text        string       `json:"text,omitempty"`
	IconEmoji   string       `json:"icon_emoji,omitempty"`
	Attachments []Attachment `json:"attachments,omitempty"`
}

// Attachment represents a Slack message attachment
type Attachment struct {
	Color      string   `json:"color,omitempty"`
	Fallback   string   `json:"fallback,omitempty"`
	Text       string   `json:"text,omitempty"`
	Fields     []Field  `json:"fields,omitempty"`
	MarkdownIn []string `json:"mrkdwn_in,omitempty"`
}

// Field represents a field within a Slack message attachment
type Field struct {
	Title string `json:"title,omitempty"`
	Value string `json:"value,omitempty"`
	Short bool   `json:"short,omitempty"`
}

// SendSlackNotification sends a structured notification to Slack
func SendSlackNotification(cfg *config.Config, alertType string, p *process.ProcessInfo) {
	if cfg.Slack.WebhookURL == "" {
		utils.Info("Slack Webhook URL is not configured. Skipping notification.")
		return
	}

	var color string
	var title string
	var emoji string
	var threshold float64 // To display in the message

	switch alertType {
	case "CPU_HIGH":
		color = "#FF0000" // Red
		title = "High CPU Usage Alert"
		emoji = ":fire:"
		threshold = cfg.Thresholds.CPUPercent
	case "MEM_HIGH":
		color = "#FFA500" // Orange
		title = "High Memory Usage Alert"
		emoji = ":warning:"
		threshold = cfg.Thresholds.MemPercent
	default:
		color = "#CCCCCC" // Grey
		title = "Jenkins Monitor Alert"
		emoji = ":bell:"
		threshold = 0 // Default, or handle as appropriate
	}

	msg := SlackMessage{
		Channel:   cfg.Slack.Channel,
		Username:  cfg.Slack.Username,
		IconEmoji: emoji,
		Attachments: []Attachment{
			{
				Color:    color,
				Fallback: fmt.Sprintf("%s: Job %s (PID %d) exceeded %s threshold (%.2f%%)", title, p.BuildJobName, p.PID, alertType, threshold),
				Text:     fmt.Sprintf("*%s: Jenkins Job Performance Alert*", title),
				Fields: []Field{
					{
						Title: "Job Name",
						Value: p.BuildJobName,
						Short: true,
					},
					{
						Title: "PID",
						Value: fmt.Sprintf("%d", p.PID),
						Short: true,
					},
					{
						Title: "CPU Usage",
						Value: fmt.Sprintf("%.2f%% (Threshold: %.2f%%)", p.CPU, cfg.Thresholds.CPUPercent),
						Short: true,
					},
					{
						Title: "Memory Usage",
						Value: fmt.Sprintf("%.2f%% (Threshold: %.2f%%)", p.Mem, cfg.Thresholds.MemPercent),
						Short: true,
					},
					{
						Title: "Timestamp",
						Value: time.Now().Format(time.RFC1123Z),
						Short: false,
					},
				},
				MarkdownIn: []string{"text", "fields"},
			},
		},
	}

	jsonBytes, err := json.MarshalIndent(msg, "", "  ")
	if err != nil {
		utils.Error(fmt.Sprintf("Failed to marshal Slack message: %v", err))
		return
	}

	req, err := http.NewRequest("POST", cfg.Slack.WebhookURL, bytes.NewBuffer(jsonBytes))
	if err != nil {
		utils.Error(fmt.Sprintf("Failed to create Slack request: %v", err))
		return
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		utils.Error(fmt.Sprintf("Failed to send Slack notification: %v", err))
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body) // Use io.ReadAll
		utils.Error(fmt.Sprintf("Received non-OK response from Slack (%d): %s", resp.StatusCode, string(body)))
	} else {
		utils.Info(fmt.Sprintf("Slack notification sent successfully for job %s (PID %d)", p.BuildJobName, p.PID))
	}
}
