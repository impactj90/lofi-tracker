package git

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestBranchWatcherEmitsInitialEvent(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	watcher := NewBranchWatcher(50 * time.Millisecond)
	watcher.resolve = func(context.Context) (string, error) {
		return "main", nil
	}

	errCh := make(chan error, 1)
	go func() {
		errCh <- watcher.Start(ctx)
	}()

	select {
	case event := <-watcher.Events():
		if event.From != "" {
			t.Fatalf("expected initial event to have empty From, got %q", event.From)
		}
		if event.To != "main" {
			t.Fatalf("expected initial event To to be 'main', got %q", event.To)
		}
	case err := <-errCh:
		if err != nil {
			t.Fatalf("unexpected error starting watcher: %v", err)
		}
	case <-time.After(500 * time.Millisecond):
		t.Fatal("timed out waiting for initial branch event")
	}

	cancel()
	<-errCh // drain goroutine
}

func TestBranchWatcherDetectsBranchChanges(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	labels := []string{"main", "main", "feature/one"}
	var idx int
	var last string

	watcher := NewBranchWatcher(10 * time.Millisecond)
	watcher.resolve = func(context.Context) (string, error) {
		if idx >= len(labels) {
			return last, nil
		}
		last = labels[idx]
		idx++
		return last, nil
	}

	errCh := make(chan error, 1)
	go func() {
		errCh <- watcher.Start(ctx)
	}()

	// Consume initial event
	<-watcher.Events()

	select {
	case event := <-watcher.Events():
		if event.From != "main" || event.To != "feature/one" {
			t.Fatalf("unexpected change event: %+v", event)
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for branch change event")
	}

	cancel()
	<-errCh
}

func TestBranchWatcherReturnsErrorWhenInitialResolveFails(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	watcher := NewBranchWatcher(10 * time.Millisecond)
	watcher.resolve = func(context.Context) (string, error) {
		return "", errors.New("boom")
	}

	errCh := make(chan error, 1)
	go func() {
		errCh <- watcher.Start(ctx)
	}()

	if err := <-errCh; err == nil {
		t.Fatal("expected error when initial resolve fails")
	}
}
