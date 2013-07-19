package errplane

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"
	"path"
	"sync/atomic"
	"time"
)

var retries = 1 // Don't do this for now
var sharedClient *Client

type Client struct {
	ApplicationId string
	Environment   string
	APIKey        string
	client        *http.Client
	queue         chan []byte
	dropCount     int64
	ticker        *time.Ticker
	throttle      *time.Ticker
}

// Init singleton
func Setup(app_id, environment, key string)*Client {
	sharedClient = NewClient(app_id, environment, key)
	return sharedClient
}

// Enqueue event using singleton client
func EnqueueEvent(e *Event) {
	if sharedClient != nil {
		sharedClient.EnqueueEvent(e)
	}
}

// Helper for quick tracking
func TrackEvent(name string, value float64, context interface{}) {
	if sharedClient != nil {
		sharedClient.EnqueueEvent(&Event{
			Name:    name,
			Value:   value,
			Context: context,
		})
	}
}

func NewClient(app_id, environment, key string) *Client {
	c := &Client{
		ApplicationId: app_id,
		Environment:   environment,
		APIKey:        key,
		client:        new(http.Client),
		queue:         make(chan []byte, 1024),
		ticker:        time.NewTicker(60 * time.Second),
		throttle:      time.NewTicker(50 * time.Millisecond),
	}
	// FIXME: do we need a Close() method to shut this down?
	go c.sender()
	go c.tick()
	return c
}

func (c *Client) Enqueue(data []byte) {
	select {
	case c.queue <- data:
	default:
		atomic.AddInt64(&c.dropCount, 1)
	}
}

func (c *Client) EnqueueEvent(e *Event) {
	c.Enqueue(e.Line())
}

func (c *Client) Post(data []byte) error {
	<-c.throttle.C // don't send more often than the throttle will allow
	res, err := c.client.Post(c.url(), "application/octet-stream", bytes.NewBuffer(data))
	if res != nil && res.Body != nil {
		defer res.Body.Close()
	}
	if err == nil && res != nil && res.StatusCode/100 != 2 {
		err = errors.New(fmt.Sprintf("errplane: send error %v", res.Status))
	}
	return err
}

func (c *Client) sender() {
	for data := range c.queue {
		if sent := c.send(data); !sent {
			atomic.AddInt64(&c.dropCount, 1)
		}
	}
}

// send with retries (internal.  use Enqueue)
func (c *Client) send(data []byte) bool {
	sent := false
	for i := 0; i < retries && !sent; i++ {
		if err := c.Post(data); err == nil {
			sent = true
		}
	}
	return sent
}

func (c *Client) tick() {
	for _ = range c.ticker.C {
		// Track dropped events
		drops := atomic.LoadInt64(&c.dropCount)
		atomic.StoreInt64(&c.dropCount, -drops)

		if drops > 0 {
			e := new(Event)
			e.Name = "meta/errplane_dropped_events"
			e.Value = float64(drops)

			c.send(e.Line())
		}
	}
}

func (c *Client) url() string {
	return fmt.Sprintf("https://%v?api_key=%v", path.Join("apiv2.errplane.com", "databases", c.ApplicationId+c.Environment, "points"), c.APIKey)
}
