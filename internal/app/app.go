package app

import (
	"context"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"sync"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/SPDG/pace-bms-mqtt-bridge/internal/buildinfo"
	"github.com/SPDG/pace-bms-mqtt-bridge/internal/collector"
	"github.com/SPDG/pace-bms-mqtt-bridge/internal/config"
	"github.com/SPDG/pace-bms-mqtt-bridge/internal/httpapi"
	mqttsvc "github.com/SPDG/pace-bms-mqtt-bridge/internal/mqtt"
	"github.com/SPDG/pace-bms-mqtt-bridge/internal/state"
	"github.com/SPDG/pace-bms-mqtt-bridge/web"
)

type App struct {
	build      buildinfo.Info
	configPath string
	startedAt  time.Time
	assets     fs.FS

	mu           sync.RWMutex
	cfg          config.Config
	runtimeState *state.Store
}

func New(configPath string, build buildinfo.Info) (*App, error) {
	assets, err := web.Assets()
	if err != nil {
		return nil, fmt.Errorf("load embedded web assets: %w", err)
	}
	return &App{
		build:        build,
		configPath:   configPath,
		startedAt:    time.Now(),
		assets:       assets,
		runtimeState: state.New(),
	}, nil
}

func (a *App) Run(ctx context.Context) error {
	cfg, created, err := config.LoadOrCreate(a.configPath)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}
	a.mu.Lock()
	a.cfg = cfg
	a.mu.Unlock()
	if created {
		log.Printf("created default config at %s", a.configPath)
	}

	a.runtimeState.SetServiceStatus("web", "running", true, "", time.Now().UTC())
	handler := httpapi.NewHandler(
		a.build,
		httpapi.StatusSnapshot{StartedAt: a.startedAt, ConfigPath: a.configPath, ConfigReady: true},
		a,
		a,
		a.assets,
	)
	server := &http.Server{
		Addr:              cfg.HTTP.Listen,
		Handler:           handler.Routes(),
		ReadHeaderTimeout: 5 * time.Second,
	}

	paceService := collector.NewService(a, a.runtimeState)
	mqttService := mqttsvc.NewService(a, a.runtimeState, a.build)
	group, groupCtx := errgroup.WithContext(ctx)
	group.Go(func() error {
		return paceService.Run(groupCtx)
	})
	group.Go(func() error {
		return mqttService.Run(groupCtx)
	})
	group.Go(func() error {
		log.Printf("web server listening on http://%s", cfg.HTTP.Listen)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			return err
		}
		return nil
	})
	group.Go(func() error {
		<-groupCtx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		return server.Shutdown(shutdownCtx)
	})
	return group.Wait()
}

func (a *App) GetConfig() config.Config {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.cfg
}

func (a *App) UpdateConfig(cfg config.Config) error {
	if err := config.Save(a.configPath, cfg); err != nil {
		return err
	}
	a.mu.Lock()
	a.cfg = cfg
	a.mu.Unlock()
	return nil
}

func (a *App) GetStateSnapshot() state.Snapshot {
	return a.runtimeState.Snapshot()
}
