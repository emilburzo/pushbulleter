package notifications

import (
	"fmt"
	"log"
	"strings"

	"github.com/gen2brain/beeep"
	"pushbulleter/internal/pushbullet"
)

type Manager struct {
	enabled     bool
	showMirrors bool
	showSMS     bool
	showCalls   bool
	filters     []string
}

func NewManager(enabled, showMirrors, showSMS, showCalls bool, filters []string) *Manager {
	return &Manager{
		enabled:     enabled,
		showMirrors: showMirrors,
		showSMS:     showSMS,
		showCalls:   showCalls,
		filters:     filters,
	}
}

func (m *Manager) HandlePush(push *pushbullet.Push) {
	if !m.enabled {
		return
	}

	// Check if we should show this notification
	if !m.shouldNotify(push) {
		return
	}

	// Special handling for SMS notifications
	if push.Type == "sms_changed" && len(push.Notifications) > 0 {
		// Show notification for each SMS
		for _, notification := range push.Notifications {
			title := "💬 SMS"
			if notification.Title != "" {
				title = "💬 " + notification.Title
			}

			message := notification.Body
			if message == "" {
				message = "New SMS message"
			}

			if err := beeep.Notify(title, message, ""); err != nil {
				log.Printf("Failed to show SMS notification: %v", err)
			}
		}
		return
	}

	title, message := m.formatNotification(push)
	if title == "" && message == "" {
		return
	}

	// Show desktop notification
	if err := beeep.Notify(title, message, ""); err != nil {
		log.Printf("Failed to show notification: %v", err)
	}
}

func (m *Manager) shouldNotify(push *pushbullet.Push) bool {
	switch push.Type {
	case "mirror":
		if !m.showMirrors {
			return false
		}

		// Check for SMS/call specific filtering
		if push.PackageName == "com.android.phone" && !m.showCalls {
			return false
		}
		if (push.PackageName == "com.android.mms" ||
			push.PackageName == "com.google.android.apps.messaging" ||
			strings.Contains(strings.ToLower(push.ApplicationName), "sms")) && !m.showSMS {
			return false
		}

	case "sms_changed":
		return m.showSMS

	default:
		// For other push types, apply general filtering
		if push.Direction == "self" {
			return false // Don't notify for pushes from ourselves
		}
	}

	// Apply custom filters
	for _, filter := range m.filters {
		if strings.Contains(strings.ToLower(push.PackageName), strings.ToLower(filter)) ||
			strings.Contains(strings.ToLower(push.ApplicationName), strings.ToLower(filter)) {
			return false
		}
	}

	return true
}

func (m *Manager) formatNotification(push *pushbullet.Push) (string, string) {
	switch push.Type {
	case "mirror":
		title := push.ApplicationName
		if push.Title != "" {
			title = fmt.Sprintf("%s: %s", push.ApplicationName, push.Title)
		}

		message := push.Body

		// Special handling for calls
		if push.PackageName == "com.android.phone" {
			if strings.Contains(strings.ToLower(push.Body), "incoming call") {
				title = "📞 Incoming Call"
			} else if strings.Contains(strings.ToLower(push.Body), "missed call") {
				title = "📞 Missed Call"
			}
		}

		// Special handling for SMS
		if push.PackageName == "com.android.mms" ||
			push.PackageName == "com.google.android.apps.messaging" ||
			strings.Contains(strings.ToLower(push.ApplicationName), "sms") {
			title = "💬 " + title
		}

		return title, message

	case "sms_changed":
		// For sms_changed events, we should show the actual SMS content
		// The title and body should contain the SMS details
		title := "💬 SMS"
		message := "New SMS activity"

		if push.Title != "" {
			title = "💬 " + push.Title
		}
		if push.Body != "" {
			message = push.Body
		}

		return title, message

	case "note":
		title := "📝 Note"
		if push.Title != "" {
			title = push.Title
		}
		return title, push.Body

	case "link":
		title := "🔗 Link"
		if push.Title != "" {
			title = push.Title
		}
		return title, push.Body

	case "file":
		return "📎 File Shared", push.Body

	default:
		if push.Title != "" || push.Body != "" {
			title := "Pushbullet"
			if push.Title != "" {
				title = push.Title
			}
			return title, push.Body
		}
	}

	return "", ""
}
