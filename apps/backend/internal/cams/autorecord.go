package cams

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/playingwithclouds/veil/internal/db"
)

const (
	// autoRecordInterval is how often followed auto-record models are checked.
	autoRecordInterval = 2 * time.Minute
	// autoRecordStagger spaces the checks of consecutive models, so a long
	// list does not hit a site in one burst.
	autoRecordStagger = 700 * time.Millisecond
)

// AutoRecorder starts a recording whenever a followed auto-record model comes
// online. A recording ends by itself when the model goes offline.
type AutoRecorder struct {
	service  *Service
	recorder *Recorder
	interval time.Duration
	stagger  time.Duration
}

// NewAutoRecorder returns the watcher over the service and recorder.
func NewAutoRecorder(service *Service, recorder *Recorder) *AutoRecorder {
	return &AutoRecorder{service: service, recorder: recorder, interval: autoRecordInterval, stagger: autoRecordStagger}
}

// Run checks the models now and then every interval until ctx ends.
func (watcher *AutoRecorder) Run(ctx context.Context) {
	ticker := time.NewTicker(watcher.interval)
	defer ticker.Stop()
	for {
		watcher.Check(ctx)
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}

// Check probes every followed auto-record model that is not being recorded and
// starts a recording for the ones that are live. The probe is the service's
// cached online check, so rooms the web just looked at cost nothing extra.
func (watcher *AutoRecorder) Check(ctx context.Context) {
	records, err := db.QueryAs[camRecord](ctx, watcher.service.database,
		`SELECT * FROM cam_model WHERE followed = 1 AND auto_record = 1 ORDER BY display_name COLLATE NOCASE`, nil)
	if err != nil {
		log.Printf("cams: auto-record: %v", err)
		return
	}
	for index := range records {
		if ctx.Err() != nil {
			return
		}
		watcher.checkOne(ctx, &records[index])
		sleepContext(ctx, watcher.stagger)
	}
}

// checkOne starts a recording of the model when she is live.
func (watcher *AutoRecorder) checkOne(ctx context.Context, record *camRecord) {
	if watcher.recorder.IsRecording(record.ID) {
		return
	}
	if !watcher.service.checkOnline(ctx, record) {
		return
	}
	if _, err := watcher.recorder.Start(ctx, record.ID); err != nil {
		log.Printf("cams: auto-record %s: %v", record.DisplayName, err)
		return
	}
	log.Printf("cams: auto-record %s: recording started", record.DisplayName)
}

// HTTPFetcher is the Fetcher used against live sites.
type HTTPFetcher struct {
	client *http.Client
}

// NewHTTPFetcher returns a Fetcher with a request timeout.
func NewHTTPFetcher() *HTTPFetcher {
	return &HTTPFetcher{client: &http.Client{Timeout: 30 * time.Second}}
}

// Fetch requests rawURL with a browser user agent and the given headers.
func (fetcher *HTTPFetcher) Fetch(ctx context.Context, rawURL string, headers map[string]string) (*http.Response, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, err
	}
	request.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0.0.0 Safari/537.36")
	for name, value := range headers {
		request.Header.Set(name, value)
	}
	response, err := fetcher.client.Do(request)
	if err != nil {
		return nil, err
	}
	if response.StatusCode != http.StatusOK {
		response.Body.Close()
		return nil, &StatusError{Status: response.StatusCode, URL: rawURL}
	}
	return response, nil
}

// StatusError is a response other than 200.
type StatusError struct {
	Status int
	URL    string
}

// Error describes the failed request.
func (err *StatusError) Error() string {
	return http.StatusText(err.Status) + " fetching " + err.URL
}
