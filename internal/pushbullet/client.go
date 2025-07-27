package pushbullet

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/gorilla/websocket"
)

const (
	APIBase       = "https://api.pushbullet.com"
	WebSocketURL  = "wss://stream.pushbullet.com/websocket"
)

type Client struct {
	apiKey     string
	httpClient *http.Client
	e2e        *E2EManager
}

type StreamMessage struct {
	Type string          `json:"type"`
	Push json.RawMessage `json:"push,omitempty"`
}

type Push struct {
	Type                string      `json:"type"`
	Title               string      `json:"title,omitempty"`
	Body                string      `json:"body,omitempty"`
	Direction           string      `json:"direction,omitempty"`
	SenderEmail         string      `json:"sender_email,omitempty"`
	SenderName          string      `json:"sender_name,omitempty"`
	ApplicationName     string      `json:"application_name,omitempty"`
	PackageName         string      `json:"package_name,omitempty"`
	NotificationID      interface{} `json:"notification_id,omitempty"`
	NotificationTag     string      `json:"notification_tag,omitempty"`
	ConversationIden    string      `json:"conversation_iden,omitempty"`
	SourceDeviceIden    string      `json:"source_device_iden,omitempty"`
	Dismissable         bool        `json:"dismissable,omitempty"`
	Created             float64     `json:"created,omitempty"`
	Modified            float64     `json:"modified,omitempty"`
	
	// SMS-specific fields
	Notifications       []SMSNotification `json:"notifications,omitempty"`
	
	// Encrypted fields
	Encrypted  bool   `json:"encrypted,omitempty"`
	Ciphertext string `json:"ciphertext,omitempty"`
}

type SMSNotification struct {
	ThreadID    string  `json:"thread_id,omitempty"`
	Title       string  `json:"title,omitempty"`
	Body        string  `json:"body,omitempty"`
	Timestamp   float64 `json:"timestamp,omitempty"`
	ImageURL    string  `json:"image_url,omitempty"`
}

func NewClient(apiKey string, e2eKey string) *Client {
	client := &Client{
		apiKey: apiKey,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
	
	if e2eKey != "" {
		client.e2e = NewE2EManager(e2eKey)
	}
	
	return client
}

func (c *Client) UpdateE2EWithUserIden(e2eKey, userIden string) {
	if e2eKey != "" && userIden != "" {
		c.e2e = NewE2EManagerWithSalt(e2eKey, userIden)
	}
}

func (c *Client) ConnectStream(ctx context.Context, messageHandler func(*StreamMessage)) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		if err := c.connectStreamOnce(ctx, messageHandler); err != nil {
			log.Printf("Stream connection error: %v", err)
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(5 * time.Second):
				// Retry after 5 seconds
			}
		}
	}
}

func (c *Client) connectStreamOnce(ctx context.Context, messageHandler func(*StreamMessage)) error {
	u, err := url.Parse(WebSocketURL + "/" + c.apiKey)
	if err != nil {
		return fmt.Errorf("failed to parse websocket URL: %w", err)
	}

	dialer := websocket.DefaultDialer
	conn, _, err := dialer.Dial(u.String(), nil)
	if err != nil {
		return fmt.Errorf("failed to connect to websocket: %w", err)
	}
	defer conn.Close()

	log.Println("Connected to Pushbullet stream")

	// Set up ping/pong handling
	conn.SetPongHandler(func(string) error {
		return conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	})

	// Start ping ticker
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	go func() {
		for {
			select {
			case <-ticker.C:
				if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
					return
				}
			case <-ctx.Done():
				return
			}
		}
	}()

	// Read messages
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		_, message, err := conn.ReadMessage()
		if err != nil {
			return fmt.Errorf("failed to read message: %w", err)
		}

		var streamMsg StreamMessage
		if err := json.Unmarshal(message, &streamMsg); err != nil {
			log.Printf("Failed to unmarshal stream message: %v", err)
			continue
		}

		// Handle encrypted pushes
		if streamMsg.Type == "push" && len(streamMsg.Push) > 0 {
			var push Push
			if err := json.Unmarshal(streamMsg.Push, &push); err != nil {
				log.Printf("Failed to unmarshal push: %v", err)
				continue
			}

			// Decrypt if necessary
			if push.Encrypted && c.e2e != nil {
				decrypted, err := c.e2e.Decrypt(push.Ciphertext)
				if err != nil {
					log.Printf("Failed to decrypt push: %v", err)
					continue
				}
				
				if err := json.Unmarshal([]byte(decrypted), &push); err != nil {
					log.Printf("Failed to unmarshal decrypted push: %v", err)
					continue
				}
			}

			// Re-marshal the potentially decrypted push
			pushData, _ := json.Marshal(push)
			streamMsg.Push = pushData
		}

		messageHandler(&streamMsg)
	}
}

func (c *Client) GetUser(ctx context.Context) (map[string]interface{}, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", APIBase+"/v2/users/me", nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Access-Token", c.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status %d", resp.StatusCode)
	}

	var user map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return nil, err
	}

	return user, nil
}
