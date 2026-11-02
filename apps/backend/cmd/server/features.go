package main

import (
	"context"
	"encoding/base64"
	"log"

	"github.com/playingwithclouds/veil/internal/accounts"
	"github.com/playingwithclouds/veil/internal/automations"
	"github.com/playingwithclouds/veil/internal/cams"
	"github.com/playingwithclouds/veil/internal/db"
	"github.com/playingwithclouds/veil/internal/devices"
	"github.com/playingwithclouds/veil/internal/embeddings"
	"github.com/playingwithclouds/veil/internal/files"
	"github.com/playingwithclouds/veil/internal/follows"
	"github.com/playingwithclouds/veil/internal/jobs"
	"github.com/playingwithclouds/veil/internal/library"
	"github.com/playingwithclouds/veil/internal/limits"
	"github.com/playingwithclouds/veil/internal/livebuffer"
	"github.com/playingwithclouds/veil/internal/lookup"
	"github.com/playingwithclouds/veil/internal/notify"
	"github.com/playingwithclouds/veil/internal/pipeline"
	"github.com/playingwithclouds/veil/internal/plugins"
	"github.com/playingwithclouds/veil/internal/recommend"
	"github.com/playingwithclouds/veil/internal/resolve"
	"github.com/playingwithclouds/veil/internal/schema"
	"github.com/playingwithclouds/veil/internal/settings"
	"github.com/playingwithclouds/veil/internal/storage"
	"github.com/playingwithclouds/veil/internal/stream"
	"github.com/playingwithclouds/veil/internal/tokens"
	"github.com/playingwithclouds/veil/internal/user"
)

// core is what the features are built on.
type core struct {
	database       *db.DB
	library        *library.Library
	pipeline       *pipeline.Pipeline
	registry       *plugins.Registry
	runner         *plugins.Runner
	queue          *jobs.Queue
	settings       *settings.Service
	devices        *devices.Service
	user           *user.Service
	store          *storage.Client
	streamCache    *stream.CacheService
	streamResolver *resolve.Service
	streamProxy    *stream.Proxy
	dataDir        string
	blobDir        string
	publicURL      string
}

// features are the services past the core library: notifications,
// follows, limits, automations, cams, files, recommendations, embeddings,
// lookups, tokens and accounts.
type features struct {
	notify      *notify.Service
	follows     *follows.Service
	limits      *limits.Service
	automations *automations.Service
	cams        *cams.Service
	buffer      *livebuffer.Manager
	files       *files.Service
	recommend   *recommend.Engine
	embeddings  *embeddings.Service
	lookup      *lookup.Service
	tokens      *tokens.Service
	accounts    *accounts.Service
}

// openFeatures builds the features on the core, connects their hooks and
// ensures their built-in rows (watchlist, queue, wanted hunt).
func openFeatures(ctx context.Context, base core) *features {
	built := &features{
		notify:     notify.New(base.database),
		limits:     limits.New(base.database),
		recommend:  recommend.New(base.database),
		embeddings: embeddings.New(base.database),
		lookup:     lookup.New(base.registry, base.runner, base.library),
		tokens:     tokens.New(base.database),
		accounts:   openAccounts(base),
	}
	base.runner.SetAccountProvider(built.accounts)
	base.pipeline.SetNotifier(built.notify)
	built.automations = automations.New(automations.Options{Library: base.library, Queue: base.queue, Notifier: built.notify, Searcher: base.pipeline})
	built.follows = follows.New(follows.Options{
		Library: base.library, Queue: base.queue, Discoverer: base.pipeline, Notifier: built.notify,
		OnNewMedia: func(ctx context.Context, follow schema.Follow, media []schema.RecordRef) {
			built.fire(ctx, schema.TriggerKindNewMedia, follow, media)
		},
	})
	built.files = openFiles(base, built)
	built.buffer, built.cams = openCams(ctx, base, built)
	built.ensureBuiltIns(ctx, base)
	base.user.SetOnEvent(func(ctx context.Context, event schema.Event) {
		if err := built.automations.HandleEvent(ctx, event); err != nil {
			log.Printf("automations: event %s: %v", event.Kind, err)
		}
	})
	return built
}

// openAccounts builds the account store with its sealing key.
func openAccounts(base core) *accounts.Service {
	hostKey, _ := base64.StdEncoding.DecodeString(envOr("BLOB_ENCRYPTION_KEY", ""))
	key, err := accounts.LoadKey(base.dataDir, hostKey)
	if err != nil {
		log.Fatalf("accounts: %v", err)
	}
	service, err := accounts.New(base.database, base.registry, base.runner, base.devices.Current().Id, key)
	if err != nil {
		log.Fatalf("accounts: %v", err)
	}
	return service
}

// openFiles builds the files service; new library files fire fileAdded.
func openFiles(base core, built *features) *files.Service {
	return files.New(files.Config{
		Database: base.database, Devices: base.devices, Queue: base.queue, Settings: base.settings,
		Store: base.store, BlobDir: base.blobDir, PublicURL: base.publicURL,
		Resolver: base.streamResolver, StreamCache: base.streamCache,
		OnFileAdded: func(file schema.MediaFile) {
			built.fire(context.Background(), schema.TriggerKindFileAdded, file, []schema.RecordRef{{Kind: file.SubjectKind, Id: file.SubjectId}})
		},
	})
}

// openCams builds the live buffer and the cam service; a room going live
// fires camLive and notifies when it's followed or auto-recorded.
func openCams(ctx context.Context, base core, built *features) (*livebuffer.Manager, *cams.Service) {
	buffer := livebuffer.New(ctx, base.store)
	if err := buffer.Reset(ctx); err != nil {
		log.Printf("live buffer reset: %v", err)
	}
	base.streamProxy.SetLiveBuffer(buffer)
	service := cams.New(cams.Services{
		Database: base.database, Library: base.library, Sites: cams.NewPluginSites(base.registry, base.runner),
		Proxy: base.streamProxy, Blobs: base.store, Devices: base.devices, Queue: base.queue,
		Settings: base.settings, Buffer: buffer,
	})
	service.SetOnLive(func(model schema.CamModel) {
		subject := schema.RecordRef{Kind: schema.RecordKindCamModel, Id: model.Id}
		built.fire(ctx, schema.TriggerKindCamLive, model, []schema.RecordRef{subject})
		built.announceLive(ctx, model, subject)
	})
	return buffer, service
}

// announceLive notifies that a followed cam room went live.
func (f *features) announceLive(ctx context.Context, model schema.CamModel, subject schema.RecordRef) {
	follow, err := f.follows.FollowOf(ctx, subject)
	if err != nil || follow == nil || !follow.Notify {
		return
	}
	_, err = f.notify.Notify(ctx, notify.NewNotification{Kind: "camLive", Title: model.DisplayName + " is live", Subject: &subject})
	if err != nil {
		log.Printf("notify camLive: %v", err)
	}
}

// fire starts the automations whose trigger matches, logging failures.
func (f *features) fire(ctx context.Context, kind schema.TriggerKind, data any, subjects []schema.RecordRef) {
	if _, err := f.automations.Fire(ctx, kind, data, subjects); err != nil {
		log.Printf("automations: %s: %v", kind, err)
	}
}

// ensureBuiltIns creates the rows every library has.
func (f *features) ensureBuiltIns(ctx context.Context, base core) {
	if err := base.user.EnsureBuiltInCollections(ctx); err != nil {
		log.Fatalf("collections: %v", err)
	}
	f.registerActions(base)
	if err := f.automations.EnsureBuiltIns(ctx); err != nil {
		log.Fatalf("automations: %v", err)
	}
}

// registerTasks gives the job runner the tasks of every feature, with the
// per-kind limits from settings.
func (f *features) registerTasks(runner *jobs.Runner, configured settings.Settings) {
	tasks := map[jobs.Kind]jobs.Task{
		jobs.KindDownload:           f.files.DownloadTask(),
		schema.JobKindLibraryScan:   f.files.ScanTask(),
		schema.JobKindFingerprint:   f.files.FingerprintTask(),
		schema.JobKindSnapshot:      f.files.SnapshotTask(),
		schema.JobKindRecord:        f.cams.RecordTask(),
		schema.JobKindFollowRun:     f.follows.RunTask(),
		schema.JobKindAutomationRun: f.automations.RunTask(),
	}
	for kind, task := range tasks {
		runner.Register(kind, jobLimits(configured, kind), task)
	}
}

// start runs the features' background loops until ctx ends.
func (f *features) start(ctx context.Context) {
	go f.buffer.Run(ctx)
	go f.cams.Run(ctx)
	go f.files.Run(ctx)
	go f.follows.Run(ctx)
	go f.automations.RunSchedules(ctx)
}
