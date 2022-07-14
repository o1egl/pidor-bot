package run

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/o1egl/pidor-bot/log"
)

type Service interface {
	Start() error
	Stop(ctx context.Context) error
}

type App struct {
	logger   log.Logger
	services []Service
}

func NewApp(logger log.Logger, services []Service) *App {
	return &App{
		logger:   logger,
		services: services,
	}
}

// Start all application services
func (a *App) Start() error {
	return a.startServices()
}

func (a *App) startServices() error {
	wait := make(chan struct{})

	// run services
	errs := make(chan error, len(a.services))
	for _, svc := range a.services {
		svc := svc
		go func() {
			if err := svc.Start(); err != nil {
				errs <- err
			}
		}()
	}

	// catch interrupt signal
	go func() {
		stop := make(chan os.Signal, 1)
		signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
		<-stop
		a.logger.Warn("interrupt signal received")
		close(wait)
	}()

	select {
	case err := <-errs:
		return err
	case <-wait:
		ctx, cancelFn := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancelFn()
		return a.Stop(ctx)
	}
}

// Stop application
func (a *App) Stop(ctx context.Context) error {
	wg, wgCtx := errgroup.WithContext(ctx)
	for _, svc := range a.services {
		svc := svc
		wg.Go(func() error {
			return svc.Stop(wgCtx)
		})
	}

	done := make(chan struct{})
	errCh := make(chan error)
	go func() {
		if err := wg.Wait(); err != nil {
			errCh <- err
			return
		}
		close(done)
	}()

	select {
	case <-done:
		return nil
	case err := <-errCh:
		return err
	case <-ctx.Done():
		return fmt.Errorf("failed to gracefuly shutdown: %w", ctx.Err())
	}
}
