package observe

import (
	"sync"
	"testing"
	"time"
)

func TestNewEMFRecorder(t *testing.T) {
	r := NewEMFRecorder()
	if r == nil {
		t.Fatal("NewEMFRecorder() returned nil")
	}

	snap := r.Snapshot()
	if snap.HitTotal != 0 {
		t.Errorf("initial HitTotal = %d, want 0", snap.HitTotal)
	}
	if snap.EvictionTotal != 0 {
		t.Errorf("initial EvictionTotal = %d, want 0", snap.EvictionTotal)
	}
	if snap.ConcurrentQueries != 0 {
		t.Errorf("initial ConcurrentQueries = %d, want 0", snap.ConcurrentQueries)
	}
}

func TestEMFRecorder_PipelineStageLatency(t *testing.T) {
	r := NewEMFRecorder()

	// Should not panic.
	r.RecordStageLatency("stage4", "memory", 42*time.Millisecond)
	r.RecordStageLatency("stage3", "", 100*time.Millisecond)
	r.RecordStageLatency("stage5", "", 2*time.Second)
}

func TestEMFRecorder_QueryLatency(t *testing.T) {
	r := NewEMFRecorder()
	r.RecordQueryLatency("hit", 500*time.Millisecond)
	r.RecordQueryLatency("miss_cold_start", 1500*time.Millisecond)
}

func TestEMFRecorder_HaikuCalls(t *testing.T) {
	r := NewEMFRecorder()
	r.RecordHaikuCalls(2)
	r.RecordHaikuCalls(1)

	snap := r.Snapshot()
	if snap.HaikuCalls != 3 {
		t.Errorf("HaikuCalls = %d, want 3", snap.HaikuCalls)
	}
}

func TestEMFRecorder_ConcurrentQueries(t *testing.T) {
	r := NewEMFRecorder()
	r.SetConcurrentQueries(5)

	snap := r.Snapshot()
	if snap.ConcurrentQueries != 5 {
		t.Errorf("ConcurrentQueries = %d, want 5", snap.ConcurrentQueries)
	}

	r.SetConcurrentQueries(3)
	snap = r.Snapshot()
	if snap.ConcurrentQueries != 3 {
		t.Errorf("ConcurrentQueries = %d, want 3", snap.ConcurrentQueries)
	}
}

func TestEMFRecorder_Dropped(t *testing.T) {
	r := NewEMFRecorder()
	r.IncrDropped("rate_limited")
	r.IncrDropped("rate_limited")
	r.IncrDropped("timeout")

	snap := r.Snapshot()
	if snap.DroppedTotal["rate_limited"] != 2 {
		t.Errorf("DroppedTotal[rate_limited] = %d, want 2", snap.DroppedTotal["rate_limited"])
	}
	if snap.DroppedTotal["timeout"] != 1 {
		t.Errorf("DroppedTotal[timeout] = %d, want 1", snap.DroppedTotal["timeout"])
	}
}

func TestEMFRecorder_ConversationHitMiss(t *testing.T) {
	r := NewEMFRecorder()
	r.IncrConversationHit()
	r.IncrConversationHit()
	r.IncrConversationMiss("cold_start")
	r.IncrConversationMiss("ttl_expired")
	r.IncrConversationMiss("cold_start")

	snap := r.Snapshot()
	if snap.HitTotal != 2 {
		t.Errorf("HitTotal = %d, want 2", snap.HitTotal)
	}
	if snap.MissTotal["cold_start"] != 2 {
		t.Errorf("MissTotal[cold_start] = %d, want 2", snap.MissTotal["cold_start"])
	}
	if snap.MissTotal["ttl_expired"] != 1 {
		t.Errorf("MissTotal[ttl_expired] = %d, want 1", snap.MissTotal["ttl_expired"])
	}
}

func TestEMFRecorder_ConversationGetLatency(t *testing.T) {
	r := NewEMFRecorder()
	r.RecordConversationGetLatency("hit", 5*time.Millisecond)
	r.RecordConversationGetLatency("miss", 500*time.Millisecond)
}

func TestEMFRecorder_ActiveThreads(t *testing.T) {
	r := NewEMFRecorder()
	r.SetActiveThreads(25)

	snap := r.Snapshot()
	if snap.ActiveThreads != 25 {
		t.Errorf("ActiveThreads = %d, want 25", snap.ActiveThreads)
	}
}

func TestEMFRecorder_Evictions(t *testing.T) {
	r := NewEMFRecorder()
	r.IncrEvictions()
	r.IncrEvictions()
	r.IncrEvictions()

	snap := r.Snapshot()
	if snap.EvictionTotal != 3 {
		t.Errorf("EvictionTotal = %d, want 3", snap.EvictionTotal)
	}
}

func TestEMFRecorder_MemoryBytes(t *testing.T) {
	r := NewEMFRecorder()
	r.SetConversationMemoryBytes(75000)

	snap := r.Snapshot()
	if snap.MemoryBytes != 75000 {
		t.Errorf("MemoryBytes = %d, want 75000", snap.MemoryBytes)
	}
}

func TestEMFRecorder_StartupTimestamp(t *testing.T) {
	r := NewEMFRecorder()
	now := time.Now()
	r.SetStartupTimestamp(now)

	snap := r.Snapshot()
	if snap.StartupTimestamp != now.Unix() {
		t.Errorf("StartupTimestamp = %d, want %d", snap.StartupTimestamp, now.Unix())
	}
}

func TestEMFRecorder_Summarization(t *testing.T) {
	r := NewEMFRecorder()
	r.IncrSummarization("window_overflow")
	r.RecordSummarizationLatency("window_overflow", 200*time.Millisecond)
}

func TestEMFRecorder_BuildMetrics(t *testing.T) {
	r := NewEMFRecorder()
	r.RecordBuildPhaseLatency("summaries", 5*time.Second)
	r.RecordBuildPhaseLatency("embeddings", 2*time.Second)
	r.IncrBuildFailures("summaries")
	r.IncrBuildFailures("summaries")
	r.IncrStaleFallbacks()
	r.SetDomainsReindexed(10)
	r.SetDomainsSkipped(118)
	r.SetContentDomainsMissing(2)
	r.RecordStartupTotal("success", 30*time.Second)

	snap := r.Snapshot()
	if snap.BuildFailures["summaries"] != 2 {
		t.Errorf("BuildFailures[summaries] = %d, want 2", snap.BuildFailures["summaries"])
	}
	if snap.StaleFallbacks != 1 {
		t.Errorf("StaleFallbacks = %d, want 1", snap.StaleFallbacks)
	}
}

func TestEMFRecorder_DedupMetrics(t *testing.T) {
	r := NewEMFRecorder()
	r.SetDedupMapSize(500)
	r.IncrDedupDrops()
	r.IncrDedupDrops()

	snap := r.Snapshot()
	if snap.DedupMapSize != 500 {
		t.Errorf("DedupMapSize = %d, want 500", snap.DedupMapSize)
	}
	if snap.DedupDrops != 2 {
		t.Errorf("DedupDrops = %d, want 2", snap.DedupDrops)
	}
}

func TestEMFRecorder_PrePipelineLatency(t *testing.T) {
	r := NewEMFRecorder()
	r.RecordPrePipelineLatency("hit", 5*time.Millisecond)
	r.RecordPrePipelineLatency("miss_cold_start", 800*time.Millisecond)
	r.RecordPrePipelineLatency("no_cm", 1*time.Millisecond)
}

func TestEMFRecorder_ThreadSafe(t *testing.T) {
	r := NewEMFRecorder()

	const goroutines = 100
	var wg sync.WaitGroup
	wg.Add(goroutines * 3)

	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			r.IncrConversationHit()
		}()
		go func() {
			defer wg.Done()
			r.IncrConversationMiss("cold_start")
		}()
		go func() {
			defer wg.Done()
			r.RecordStageLatency("stage4", "memory", 42*time.Millisecond)
		}()
	}

	wg.Wait()

	snap := r.Snapshot()
	if snap.HitTotal != goroutines {
		t.Errorf("HitTotal = %d, want %d", snap.HitTotal, goroutines)
	}
	if snap.MissTotal["cold_start"] != goroutines {
		t.Errorf("MissTotal[cold_start] = %d, want %d", snap.MissTotal["cold_start"], goroutines)
	}
}

func TestEMFRecorder_SnapshotIsolation(t *testing.T) {
	r := NewEMFRecorder()
	r.IncrConversationMiss("cold_start")

	snap := r.Snapshot()

	// Modify the snapshot map -- should not affect the recorder.
	snap.MissTotal["cold_start"] = 999

	snap2 := r.Snapshot()
	if snap2.MissTotal["cold_start"] != 1 {
		t.Errorf("snapshot mutation leaked: MissTotal[cold_start] = %d, want 1", snap2.MissTotal["cold_start"])
	}
}

func TestNopRecorder_ImplementsInterface(t *testing.T) {
	var r MetricsRecorder = NopRecorder{}

	// All methods should be callable without panic.
	r.RecordStageLatency("stage4", "memory", time.Millisecond)
	r.RecordQueryLatency("hit", time.Second)
	r.RecordHaikuCalls(2)
	r.SetConcurrentQueries(5)
	r.IncrDropped("rate_limited")
	r.IncrConversationHit()
	r.IncrConversationMiss("cold_start")
	r.RecordConversationGetLatency("hit", time.Millisecond)
	r.SetActiveThreads(10)
	r.IncrEvictions()
	r.SetConversationMemoryBytes(1024)
	r.SetStartupTimestamp(time.Now())
	r.IncrSummarization("window_overflow")
	r.RecordSummarizationLatency("window_overflow", time.Millisecond)
	r.RecordBuildPhaseLatency("summaries", time.Second)
	r.IncrBuildFailures("summaries")
	r.IncrStaleFallbacks()
	r.SetDomainsReindexed(5)
	r.SetDomainsSkipped(10)
	r.SetContentDomainsMissing(1)
	r.RecordStartupTotal("success", 30*time.Second)
	r.SetDedupMapSize(100)
	r.IncrDedupDrops()
	r.RecordPrePipelineLatency("hit", time.Millisecond)
}
