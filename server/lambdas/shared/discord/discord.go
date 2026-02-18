package discord

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"
)

type Embed struct {
	Title       string `json:"title,omitempty"`
	Description string `json:"description,omitempty"`
	Color       int    `json:"color,omitempty"`
}

type WebhookPayload struct {
	Content string  `json:"content,omitempty"`
	Embeds  []Embed `json:"embeds,omitempty"`
}

type Dispatcher struct {
	client     *http.Client
	webhookURL string
	wg         sync.WaitGroup
	mu         sync.Mutex
	errors     []error
	sent       int
}

func NewDispatcher(webhookURL string) *Dispatcher {
	return &Dispatcher{
		client:     &http.Client{Timeout: 10 * time.Second},
		webhookURL: webhookURL,
	}
}

// async dispatches events, non blocking
func (d *Dispatcher) Send(payload WebhookPayload) {
	d.wg.Go(func() {
		if err := d.sendDiscordWebhook(d.webhookURL, payload); err != nil {
			d.mu.Lock()
			d.errors = append(d.errors, err)
			d.mu.Unlock()
			return
		}

		d.mu.Lock()
		d.sent++
		d.mu.Unlock()
	})
}

// waits for all pending webhooks and logs out success + errors
func (d *Dispatcher) Flush() int {
	d.wg.Wait()

	d.mu.Lock()
	defer d.mu.Unlock()

	total := d.sent + len(d.errors)

	log.Printf("discord webhooks: %d/%d succeeded", d.sent, total)
	for _, err := range d.errors {
		log.Printf("discord webhook error: %v", err)
	}

	failed := len(d.errors)
	d.sent = 0
	d.errors = nil
	return failed
}

func (d *Dispatcher) sendDiscordWebhook(webhookURL string, payload WebhookPayload) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal webhook payload: %w", err)
	}

	resp, err := d.client.Post(webhookURL, "application/json", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("send webhook: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("discord webhook returned status %d", resp.StatusCode)
	}

	return nil
}
