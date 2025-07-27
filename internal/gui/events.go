package gui

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"pushbulleter/internal/pushbullet"
)

type Event struct {
	Timestamp time.Time
	Type      string
	Title     string
	Message   string
	Raw       string
}

func HandleEvent(msg *pushbullet.StreamMessage) {
	event := Event{
		Timestamp: time.Now(),
		Type:      msg.Type,
	}

	switch msg.Type {
	case "push":
		if len(msg.Push) > 0 {
			var push pushbullet.Push
			if err := json.Unmarshal(msg.Push, &push); err == nil {
				event.Title = fmt.Sprintf("Push: %s", push.Type)

				// Special handling for SMS events
				if push.Type == "sms_changed" && len(push.Notifications) > 0 {
					notification := push.Notifications[0] // Show first notification
					if notification.Title != "" {
						event.Message = fmt.Sprintf("SMS from %s: %s", notification.Title, notification.Body)
					} else {
						event.Message = fmt.Sprintf("SMS: %s", notification.Body)
					}
				} else if push.Title != "" {
					event.Message = push.Title
				} else if push.Body != "" {
					event.Message = push.Body
				} else {
					event.Message = "No content"
				}
			}
		}
	case "nop":
		event.Title = "Keep-alive"
		event.Message = "Connection heartbeat"
	case "tickle":
		event.Title = "Data update"
		event.Message = "Server data changed"
	default:
		event.Title = fmt.Sprintf("Unknown: %s", msg.Type)
		event.Message = "Unknown message type"
	}

	// Store raw JSON for debugging
	if rawData, err := json.MarshalIndent(msg, "", "  "); err == nil {
		event.Raw = string(rawData)
	}

	// Log the event
	log.Printf("[%s] %s: %s", event.Timestamp.Format("15:04:05"), event.Title, event.Message)
}
