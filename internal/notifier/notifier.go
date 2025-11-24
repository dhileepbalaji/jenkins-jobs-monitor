package notifier

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
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
	IconEmoji   string       `json:"icon_emoji,omitempty"`
	Attachments []Attachment `json:"attachments,omitempty"`
}

// Attachment represents a Slack message attachment
type Attachment struct {
	Color  string  `json:"color,omitempty"`
	Blocks []Block `json:"blocks,omitempty"`
}

// Block represents a generic Slack block
type Block interface {
	isBlock()
}

// SectionBlock represents a section block
type SectionBlock struct {
	Type   string          `json:"type"`
	Text   *MarkdownText   `json:"text,omitempty"`
	Fields []*MarkdownText `json:"fields,omitempty"`
}

func (b SectionBlock) isBlock() {}

// HeaderBlock represents a header block
type HeaderBlock struct {
	Type string    `json:"type"`
	Text PlainText `json:"text"`
}

func (b HeaderBlock) isBlock() {}

// DividerBlock represents a divider block
type DividerBlock struct {
	Type string `json:"type"`
}

func (b DividerBlock) isBlock() {}

// ContextBlock represents a context block
type ContextBlock struct {
	Type     string         `json:"type"`
	Elements []MarkdownText `json:"elements"`
}

func (b ContextBlock) isBlock() {}

// MarkdownText represents a text object with markdown type
type MarkdownText struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// PlainText represents a text object with plain_text type
type PlainText struct {
	Type  string `json:"type"`
	Text  string `json:"text"`
	Emoji bool   `json:"emoji,omitempty"`
}

// SendSlackNotification sends a structured notification to Slack
func SendSlackNotification(cfg *config.Config, alertType string, p *process.ProcessInfo) {
	if cfg.Slack.WebhookURL == "" {
		utils.Info("Slack Webhook URL is not configured. Skipping notification.")
		return
	}

	var color string
	var title string

	switch alertType {
	case "CPU_HIGH":
		color = "#FF0000" // Red
		title = "Jenkins Monitor Alert: High CPU Usage"
	case "MEM_HIGH":
		color = "#FF0000" // Red
		title = "Jenkins Monitor Alert: High Memory Usage"
	default:
		color = "#CCCCCC" // Grey
		title = "Jenkins Monitor Alert"
	}

	// Construct Blocks
	blocks := []Block{
		HeaderBlock{
			Type: "header",
			Text: PlainText{
				Type: "plain_text",
				Text: title,
			},
		},
		DividerBlock{Type: "divider"},
		SectionBlock{
			Type: "section",
			Fields: []*MarkdownText{
				{Type: "mrkdwn", Text: fmt.Sprintf("*Job Name:*\n%s", p.BuildJobName)},
				{Type: "mrkdwn", Text: fmt.Sprintf("*PID:*\n%d", p.PID)},
				{Type: "mrkdwn", Text: fmt.Sprintf("*Build ID:*\n%s", p.BuildId)},
				{Type: "mrkdwn", Text: fmt.Sprintf("*Stage Name:*\n%s", p.StageName)},
			},
		},
		SectionBlock{
			Type: "section",
			Fields: []*MarkdownText{
				{Type: "mrkdwn", Text: fmt.Sprintf("*Workspace:*\n%s", p.WorkSpace)},
			},
		},
		SectionBlock{
			Type: "section",
			Fields: []*MarkdownText{
				{Type: "mrkdwn", Text: fmt.Sprintf("*CPU Usage:*\n%.2f%% (Threshold: %.2f%%)", p.CPU, cfg.Thresholds.CPUPercent)},
				{Type: "mrkdwn", Text: fmt.Sprintf("*Memory Usage:*\n%.2f%% (Threshold: %.2f%%)", p.Mem, cfg.Thresholds.MemPercent)},
			},
		},
		ContextBlock{
			Type: "context",
			Elements: []MarkdownText{
				{Type: "mrkdwn", Text: fmt.Sprintf("Timestamp: %s", time.Now().Format(time.RFC1123))},
			},
		},
	}

	msg := SlackMessage{
		Channel:  cfg.Slack.Channel,
		Username: cfg.Slack.Username,
		Attachments: []Attachment{
			{
				Color:  color,
				Blocks: blocks,
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
		body, _ := io.ReadAll(resp.Body)
		utils.Error(fmt.Sprintf("Received non-OK response from Slack (%d): %s", resp.StatusCode, string(body)))
	} else {
		utils.Info(fmt.Sprintf("Slack notification sent successfully for job %s (PID %d)", p.BuildJobName, p.PID))
	}
}
