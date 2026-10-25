package app

import (
	"context"
	"errors"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	data "github.com/pheelee/wishmas/internal/household"
)

type pictureRecovery struct {
	store    *data.Store
	wish     data.Wish
	attempts int
	next     time.Time
	running  bool
}
type recoveryResult struct {
	key string
	err error
}

// @impl WSH-008-AC1 WSH-008-AC2 WSH-008-AC3 WSH-008-AC4 WSH-008-AC5 WSH-008-AC8
// Startup owns the lifecycle. Constructors stay side-effect free for embedders;
// production main and the HTTPS harness start this service explicitly.
func StartBackground(ctx context.Context, handler http.Handler) <-chan struct{} {
	done := make(chan struct{})
	a, ok := handler.(*householdApp)
	if !ok {
		close(done)
		return done
	}
	go func() { defer close(done); a.background(ctx) }()
	return done
}
func (a *householdApp) background(ctx context.Context) {
	tasks := map[string]*pictureRecovery{}
	results := make(chan recoveryResult, 3)
	var workers sync.WaitGroup
	defer workers.Wait()
	initialized := false
	nextScan := time.Time{}
	nextMaintenance := time.Time{}
	active := 0
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()
	for {
		now := a.now()
		if !initialized && !now.Before(nextScan) {
			a.mu.Lock()
			h, s, _, err := a.load()
			a.mu.Unlock()
			if err == nil && h != nil {
				initialized = true
				for _, wish := range h.Wishes {
					if wish.Picture == "" || wish.PictureSource == "" {
						continue
					}
					if _, err = os.Lstat(filepath.Join(s.PictureDirectory(), wish.Picture)); os.IsNotExist(err) {
						key := s.PictureDirectory() + "/" + wish.Picture
						if _, exists := tasks[key]; !exists {
							tasks[key] = &pictureRecovery{store: s, wish: wish}
						}
					}
				}
			}
			nextScan = now.Add(time.Minute)
		}
		// Durable logout/cleanup retries run independently of browser traffic. Their
		// persisted deadlines enforce the approved backoff; this tick is not a retry.
		if !now.Before(nextMaintenance) {
			if a.storage != nil {
				a.storage.Maintenance(now)
			}
			nextMaintenance = now.Add(time.Second)
		}
		for key, task := range tasks {
			if active >= 3 {
				break
			}
			if task.running || now.Before(task.next) {
				continue
			}
			task.running = true
			active++
			workers.Add(1)
			go func(key string, task *pictureRecovery) {
				defer workers.Done()
				err := a.recoverPicture(ctx, task)
				results <- recoveryResult{key, err}
			}(key, task)
		}
		select {
		case <-ctx.Done():
			return
		case result := <-results:
			active--
			task := tasks[result.key]
			if result.err == nil {
				delete(tasks, result.key)
			} else {
				task.running = false
				task.next = a.now().Add(data.RetryDelay(task.attempts))
				if task.attempts < 5 {
					task.attempts++
				}
			}
		case <-ticker.C:
		}
	}
}
func (a *householdApp) recoverPicture(ctx context.Context, task *pictureRecovery) error {
	h, _, err := task.store.Load()
	if errors.Is(err, data.ErrStale) || os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}
	needed := false
	for _, w := range h.Wishes {
		if w.ID == task.wish.ID && w.Picture == task.wish.Picture && w.PictureSource == task.wish.PictureSource {
			needed = true
			break
		}
	}
	if !needed {
		return nil
	}
	if _, err = os.Lstat(filepath.Join(task.store.PictureDirectory(), task.wish.Picture)); err == nil {
		return nil
	} else if !os.IsNotExist(err) {
		return err
	}
	raw, err := fetchPictureWithClient(ctx, a.pictureClient, task.wish.PictureSource)
	if err != nil {
		return err
	}
	return task.store.RestorePicture(task.wish.ID, task.wish.Picture, task.wish.PictureSource, raw)
}
