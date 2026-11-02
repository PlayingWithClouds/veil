// Package notify stores notifications for the in-app inbox and delivers them
// to the user's channels (ntfy, webhooks). Everything outside the app is
// treated as public: deliveries carry neutral wording unless a channel opts
// in to the details ("detail": true in its config).
package notify

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/playingwithclouds/veil/internal/db"
	"github.com/playingwithclouds/veil/internal/outbound"
	"github.com/playingwithclouds/veil/internal/schema"
)

// Table names, exactly as the schema names the models.
const (
	notificationTable = "Notification"
	channelTable      = "NotificationChannel"
)

// deliveryTimeout bounds one delivery to one channel.
const deliveryTimeout = 15 * time.Second

// subscriberBuffer is how many notifications a slow subscriber may lag
// behind before it misses some.
const subscriberBuffer = 16

// Service stores notifications and fans them out to channels and
// subscribers.
type Service struct {
	database *db.DB
	client   *http.Client

	mutex          sync.Mutex
	subscribers    map[int]chan schema.Notification
	nextSubscriber int

	// deliveries tracks deliveries still in flight, for Wait.
	deliveries sync.WaitGroup
}

// New returns a notification service on database.
func New(database *db.DB) *Service {
	return &Service{
		database:    database,
		client:      &http.Client{Transport: outbound.NewTransport()},
		subscribers: map[int]chan schema.Notification{},
	}
}

// SetHTTPClient replaces the client deliveries are sent with.
func (s *Service) SetHTTPClient(client *http.Client) {
	s.client = client
}

// NewNotification is a notification to send. Kind names what happened
// ("newMedia", "camLive", "downloadDone", "automationFailed", …); title and
// body may name the media, since only channels that opt in see them.
type NewNotification struct {
	Kind    string
	Title   string
	Body    string
	Subject *schema.RecordRef
}

// Notify stores a notification in the inbox, passes it to subscribers and
// delivers it to every enabled channel in the background, so callers never
// wait on a slow server.
func (s *Service) Notify(ctx context.Context, notification NewNotification) (*schema.Notification, error) {
	if strings.TrimSpace(notification.Kind) == "" || strings.TrimSpace(notification.Title) == "" {
		return nil, fmt.Errorf("a notification needs a kind and a title")
	}
	stored, err := s.store(ctx, notification)
	if err != nil {
		return nil, err
	}
	s.broadcast(*stored)
	channels, err := s.enabledChannels(ctx)
	if err != nil {
		return stored, err
	}
	for _, channel := range channels {
		s.deliverLater(channel, *stored)
	}
	return stored, nil
}

// store writes the inbox row of a notification.
func (s *Service) store(ctx context.Context, notification NewNotification) (*schema.Notification, error) {
	row := schema.Notification{
		At:      db.Now(),
		Kind:    notification.Kind,
		Title:   notification.Title,
		Subject: notification.Subject,
	}
	if notification.Body != "" {
		row.Body = &notification.Body
	}
	id, err := s.database.Insert(ctx, notificationTable, row)
	if err != nil {
		return nil, err
	}
	return db.Get[schema.Notification](ctx, s.database, notificationTable, id)
}

// deliverLater sends a notification to one channel without blocking. A
// failed delivery is logged; the inbox row is what stays.
func (s *Service) deliverLater(channel schema.NotificationChannel, notification schema.Notification) {
	s.deliveries.Add(1)
	go func() {
		defer s.deliveries.Done()
		deliveryContext, cancel := context.WithTimeout(context.Background(), deliveryTimeout)
		defer cancel()
		s.deliver(deliveryContext, channel, notification)
	}()
}

// Wait blocks until every delivery in flight is done (tests, shutdown).
func (s *Service) Wait() {
	s.deliveries.Wait()
}

// Subscribe returns a channel that receives every notification stored from
// now on, and the function that ends the subscription. A subscriber that
// falls behind misses notifications rather than holding up Notify.
func (s *Service) Subscribe() (<-chan schema.Notification, func()) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	id := s.nextSubscriber
	s.nextSubscriber++
	received := make(chan schema.Notification, subscriberBuffer)
	s.subscribers[id] = received
	var once sync.Once
	cancel := func() {
		once.Do(func() { s.unsubscribe(id) })
	}
	return received, cancel
}

// unsubscribe removes and closes one subscriber's channel.
func (s *Service) unsubscribe(id int) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	close(s.subscribers[id])
	delete(s.subscribers, id)
}

// broadcast hands a notification to every subscriber that has room for it.
func (s *Service) broadcast(notification schema.Notification) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	for _, subscriber := range s.subscribers {
		select {
		case subscriber <- notification:
		default:
		}
	}
}
