package task

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestLaneQueue_SerialExecution(t *testing.T) {
	lq := NewLaneQueue(&LaneConfig{
		MaxPerLane: 1,
		QueueSize:  10,
	})
	defer lq.Stop()

	laneKey := "test:serial"
	var order []int
	var mu sync.Mutex

	// Enqueue 3 jobs
	var wg sync.WaitGroup
	for i := 0; i < 3; i++ {
		wg.Add(1)
		idx := i
		done := lq.Enqueue(laneKey, func() {
			time.Sleep(10 * time.Millisecond)
			mu.Lock()
			order = append(order, idx)
			mu.Unlock()
		})
		go func() {
			<-done
			wg.Done()
		}()
	}

	wg.Wait()

	// Verify order is sequential
	mu.Lock()
	defer mu.Unlock()
	if len(order) != 3 {
		t.Errorf("expected 3 completions, got %d", len(order))
	}
	for i, v := range order {
		if v != i {
			t.Errorf("expected order %v, got %v", []int{0, 1, 2}, order)
			break
		}
	}
}

func TestLaneQueue_ParallelLanes(t *testing.T) {
	lq := NewLaneQueue(&LaneConfig{
		MaxPerLane: 1,
		QueueSize:  10,
	})
	defer lq.Stop()

	var running int32
	var maxRunning int32
	var wg sync.WaitGroup

	// Start jobs on different lanes (should run in parallel)
	for i := 0; i < 3; i++ {
		wg.Add(1)
		laneKey := GetLaneKey("provider", string(rune('a'+i)))
		done := lq.Enqueue(laneKey, func() {
			cur := atomic.AddInt32(&running, 1)
			if cur > atomic.LoadInt32(&maxRunning) {
				atomic.StoreInt32(&maxRunning, cur)
			}
			time.Sleep(50 * time.Millisecond)
			atomic.AddInt32(&running, -1)
		})
		go func() {
			<-done
			wg.Done()
		}()
	}

	wg.Wait()

	// Should have seen multiple running at once
	if atomic.LoadInt32(&maxRunning) < 2 {
		t.Errorf("expected parallel execution across lanes, max running was %d", maxRunning)
	}
}

func TestLaneQueue_Stats(t *testing.T) {
	lq := NewLaneQueue(nil)
	defer lq.Stop()

	// Enqueue some jobs
	lq.Enqueue("lane1", func() { time.Sleep(100 * time.Millisecond) })
	lq.Enqueue("lane1", func() { time.Sleep(100 * time.Millisecond) })
	lq.Enqueue("lane2", func() { time.Sleep(100 * time.Millisecond) })

	time.Sleep(10 * time.Millisecond) // Let jobs start

	stats := lq.Stats()
	if stats.TotalLanes != 2 {
		t.Errorf("expected 2 lanes, got %d", stats.TotalLanes)
	}
}

func TestGetLaneKey(t *testing.T) {
	key := GetLaneKey("provider1", "codex")
	expected := "provider1:codex"
	if key != expected {
		t.Errorf("expected %s, got %s", expected, key)
	}
}

func TestLaneQueue_PendingCount(t *testing.T) {
	lq := NewLaneQueue(&LaneConfig{
		MaxPerLane: 1,
		QueueSize:  10,
	})
	defer lq.Stop()

	laneKey := "test:pending"

	// Start a blocking job
	started := make(chan struct{})
	done := lq.Enqueue(laneKey, func() {
		close(started)
		time.Sleep(100 * time.Millisecond)
	})

	<-started // Wait for first job to start

	// Enqueue more jobs
	lq.Enqueue(laneKey, func() {})
	lq.Enqueue(laneKey, func() {})

	pending := lq.PendingCount(laneKey)
	if pending < 1 {
		t.Errorf("expected at least 1 pending job, got %d", pending)
	}

	<-done // Wait for completion
}
