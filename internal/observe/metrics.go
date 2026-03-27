package observe

import (
	"log/slog"
	"sync"
	"sync/atomic"
	"time"
)

// MetricsRecorder provides a unified interface for recording Clew metrics.
// All 34 metrics from the Thermia consultation are covered.
// Implementation uses CloudWatch Embedded Metric Format (EMF) via structured slog,
// which CloudWatch automatically extracts as metrics from JSON logs on ECS/Fargate.
// The interface allows swapping to Prometheus or OTEL in the future.
type MetricsRecorder interface {
	// Pipeline stage metrics (Sprint 5).
	RecordStageLatency(stage string, source string, duration time.Duration)
	RecordQueryLatency(cmPath string, duration time.Duration)
	RecordHaikuCalls(count int)
	SetConcurrentQueries(count int)
	IncrDropped(reason string)

	// ConversationManager metrics (Sprint 6).
	IncrConversationHit()
	IncrConversationMiss(reason string)
	RecordConversationGetLatency(result string, duration time.Duration)
	SetActiveThreads(count int)
	IncrEvictions()
	SetConversationMemoryBytes(bytes int64)
	SetStartupTimestamp(t time.Time)

	// ConversationManager summarization metrics (Sprint 6).
	IncrSummarization(trigger string)
	RecordSummarizationLatency(trigger string, duration time.Duration)

	// Index build metrics (Sprint 7).
	RecordBuildPhaseLatency(phase string, duration time.Duration)
	IncrBuildFailures(phase string)
	IncrStaleFallbacks()
	SetDomainsReindexed(count int)
	SetDomainsSkipped(count int)
	SetContentDomainsMissing(count int)
	RecordStartupTotal(result string, duration time.Duration)

	// eventDedup metrics (Sprint 6).
	SetDedupMapSize(size int)
	IncrDedupDrops()

	// Pre-pipeline metrics (Sprint 5).
	RecordPrePipelineLatency(cmResult string, duration time.Duration)

	// Contextual-equilibrium metrics (CE follow-up WS-E).
	IncrSectionCandidate()
	IncrGraphInjected(count int)
	IncrDiversityFloorEnforced(domainType string)
	IncrTypeCeilingHit(domainType string)
	RecordAssemblerTypeFraction(domainType string, fraction float64)
}

// emfMetric defines a CloudWatch EMF metric entry.
type emfMetric struct {
	Name string `json:"Name"`
	Unit string `json:"Unit"`
}

// emfCloudWatch defines the CloudWatch EMF metadata block.
type emfCloudWatch struct {
	Namespace  string       `json:"Namespace"`
	Dimensions [][]string   `json:"Dimensions"`
	Metrics    []emfMetric  `json:"Metrics"`
}

// emfAWS is the top-level _aws key in an EMF log line.
type emfAWS struct {
	Timestamp         int64           `json:"Timestamp"`
	CloudWatchMetrics []emfCloudWatch `json:"CloudWatchMetrics"`
}

// EMFRecorder implements MetricsRecorder using CloudWatch Embedded Metric Format.
// Metrics are emitted as structured JSON log lines with the _aws namespace metadata.
// CloudWatch automatically extracts these as metrics when running on ECS/Fargate.
// When running locally or without CloudWatch, they are simply structured log entries.
type EMFRecorder struct {
	logger *slog.Logger

	// Atomic gauges for concurrent-safe gauge reporting.
	concurrentQueries atomic.Int64
	activeThreads     atomic.Int64
	memoryBytes       atomic.Int64
	dedupMapSize      atomic.Int64
	startupTimestamp  atomic.Int64

	// Counters with mutex protection.
	mu             sync.Mutex
	hitTotal       int64
	missTotal      map[string]int64
	evictionTotal  int64
	droppedTotal   map[string]int64
	dedupDrops     int64
	staleFallbacks int64
	buildFailures  map[string]int64
	haikuCalls     int64
	summarizations map[string]int64
}

// NewEMFRecorder creates a MetricsRecorder that emits CloudWatch EMF log lines.
// Uses the default slog logger (configured by ConfigureStructuredLogging).
func NewEMFRecorder() *EMFRecorder {
	return &EMFRecorder{
		logger:         slog.Default(),
		missTotal:      make(map[string]int64),
		droppedTotal:   make(map[string]int64),
		buildFailures:  make(map[string]int64),
		summarizations: make(map[string]int64),
	}
}

// --- Pipeline stage metrics ---

// RecordStageLatency records clew_pipeline_stage{N}_latency_seconds.
// stage: "stage0", "stage1", "stage2", "stage3", "stage4", "stage5"
// source: "memory", "cache_hit", "cache_miss" (Stage 4 only; empty for others).
func (r *EMFRecorder) RecordStageLatency(stage string, source string, duration time.Duration) {
	metricName := "clew_pipeline_" + stage + "_latency_seconds"
	seconds := duration.Seconds()

	dims := [][]string{{"stage"}}
	attrs := []slog.Attr{
		slog.String("stage", stage),
		slog.Float64(metricName, seconds),
	}
	if source != "" {
		dims = [][]string{{"stage", "source"}}
		attrs = append(attrs, slog.String("source", source))
	}

	r.emitEMF(metricName, "Seconds", dims, attrs...)
}

// RecordQueryLatency records clew_pipeline_query_total_latency_seconds.
// cmPath: "hit", "miss_cold_start", "miss_ttl_expired", "miss_deploy_gap", "no_history".
func (r *EMFRecorder) RecordQueryLatency(cmPath string, duration time.Duration) {
	metricName := "clew_pipeline_query_total_latency_seconds"
	seconds := duration.Seconds()
	r.emitEMF(metricName, "Seconds", [][]string{{"cm_path"}},
		slog.String("cm_path", cmPath),
		slog.Float64(metricName, seconds),
	)
}

// RecordHaikuCalls records clew_pipeline_haiku_calls_per_query.
func (r *EMFRecorder) RecordHaikuCalls(count int) {
	r.mu.Lock()
	r.haikuCalls += int64(count)
	r.mu.Unlock()

	metricName := "clew_pipeline_haiku_calls_per_query"
	r.emitEMF(metricName, "Count", nil,
		slog.Int(metricName, count),
	)
}

// SetConcurrentQueries records clew_pipeline_concurrent_queries gauge.
func (r *EMFRecorder) SetConcurrentQueries(count int) {
	r.concurrentQueries.Store(int64(count))

	metricName := "clew_pipeline_concurrent_queries"
	r.emitEMF(metricName, "Count", nil,
		slog.Int(metricName, count),
	)
}

// IncrDropped records clew_pipeline_dropped_total{reason}.
func (r *EMFRecorder) IncrDropped(reason string) {
	r.mu.Lock()
	r.droppedTotal[reason]++
	r.mu.Unlock()

	metricName := "clew_pipeline_dropped_total"
	r.emitEMF(metricName, "Count", [][]string{{"reason"}},
		slog.String("reason", reason),
		slog.Int(metricName, 1),
	)
}

// --- ConversationManager metrics ---

// IncrConversationHit records clew_conversation_hit_total.
func (r *EMFRecorder) IncrConversationHit() {
	r.mu.Lock()
	r.hitTotal++
	r.mu.Unlock()

	metricName := "clew_conversation_hit_total"
	r.emitEMF(metricName, "Count", nil,
		slog.Int(metricName, 1),
	)
}

// IncrConversationMiss records clew_conversation_miss_total{reason}.
// reason: "cold_start", "ttl_expired", "deploy_gap".
func (r *EMFRecorder) IncrConversationMiss(reason string) {
	r.mu.Lock()
	r.missTotal[reason]++
	r.mu.Unlock()

	metricName := "clew_conversation_miss_total"
	r.emitEMF(metricName, "Count", [][]string{{"reason"}},
		slog.String("reason", reason),
		slog.Int(metricName, 1),
	)
}

// RecordConversationGetLatency records clew_conversation_get_latency_seconds{result}.
// result: "hit", "miss".
func (r *EMFRecorder) RecordConversationGetLatency(result string, duration time.Duration) {
	metricName := "clew_conversation_get_latency_seconds"
	seconds := duration.Seconds()
	r.emitEMF(metricName, "Seconds", [][]string{{"result"}},
		slog.String("result", result),
		slog.Float64(metricName, seconds),
	)
}

// SetActiveThreads records clew_conversation_active_threads gauge.
func (r *EMFRecorder) SetActiveThreads(count int) {
	r.activeThreads.Store(int64(count))

	metricName := "clew_conversation_active_threads"
	r.emitEMF(metricName, "Count", nil,
		slog.Int(metricName, count),
	)
}

// IncrEvictions records clew_conversation_evictions_total.
func (r *EMFRecorder) IncrEvictions() {
	r.mu.Lock()
	r.evictionTotal++
	r.mu.Unlock()

	metricName := "clew_conversation_evictions_total"
	r.emitEMF(metricName, "Count", nil,
		slog.Int(metricName, 1),
	)
}

// SetConversationMemoryBytes records clew_conversation_memory_bytes gauge.
func (r *EMFRecorder) SetConversationMemoryBytes(bytes int64) {
	r.memoryBytes.Store(bytes)

	metricName := "clew_conversation_memory_bytes"
	r.emitEMF(metricName, "Bytes", nil,
		slog.Int64(metricName, bytes),
	)
}

// SetStartupTimestamp records clew_startup_timestamp gauge.
// Powers all post-deploy alert suppression.
func (r *EMFRecorder) SetStartupTimestamp(t time.Time) {
	r.startupTimestamp.Store(t.Unix())

	metricName := "clew_startup_timestamp"
	r.emitEMF(metricName, "Seconds", nil,
		slog.Int64(metricName, t.Unix()),
	)
}

// --- ConversationManager summarization metrics ---

// IncrSummarization records clew_conversation_summarization_total{trigger}.
func (r *EMFRecorder) IncrSummarization(trigger string) {
	r.mu.Lock()
	r.summarizations[trigger]++
	r.mu.Unlock()

	metricName := "clew_conversation_summarization_total"
	r.emitEMF(metricName, "Count", [][]string{{"trigger"}},
		slog.String("trigger", trigger),
		slog.Int(metricName, 1),
	)
}

// RecordSummarizationLatency records clew_conversation_summarization_latency_seconds{trigger}.
func (r *EMFRecorder) RecordSummarizationLatency(trigger string, duration time.Duration) {
	metricName := "clew_conversation_summarization_latency_seconds"
	seconds := duration.Seconds()
	r.emitEMF(metricName, "Seconds", [][]string{{"trigger"}},
		slog.String("trigger", trigger),
		slog.Float64(metricName, seconds),
	)
}

// --- Index build metrics ---

// RecordBuildPhaseLatency records clew_build_knowledge_seconds{phase}.
// phase: "summaries", "embeddings", "graph", "persist".
func (r *EMFRecorder) RecordBuildPhaseLatency(phase string, duration time.Duration) {
	metricName := "clew_build_knowledge_seconds"
	seconds := duration.Seconds()
	r.emitEMF(metricName, "Seconds", [][]string{{"phase"}},
		slog.String("phase", phase),
		slog.Float64(metricName, seconds),
	)
}

// IncrBuildFailures records clew_build_knowledge_failures_total{type}.
func (r *EMFRecorder) IncrBuildFailures(phase string) {
	r.mu.Lock()
	r.buildFailures[phase]++
	r.mu.Unlock()

	metricName := "clew_build_knowledge_failures_total"
	r.emitEMF(metricName, "Count", [][]string{{"type"}},
		slog.String("type", phase),
		slog.Int(metricName, 1),
	)
}

// IncrStaleFallbacks records clew_build_knowledge_fallback_stale_total.
func (r *EMFRecorder) IncrStaleFallbacks() {
	r.mu.Lock()
	r.staleFallbacks++
	r.mu.Unlock()

	metricName := "clew_build_knowledge_fallback_stale_total"
	r.emitEMF(metricName, "Count", nil,
		slog.Int(metricName, 1),
	)
}

// SetDomainsReindexed records clew_build_knowledge_domains_reindexed gauge.
func (r *EMFRecorder) SetDomainsReindexed(count int) {
	metricName := "clew_build_knowledge_domains_reindexed"
	r.emitEMF(metricName, "Count", nil,
		slog.Int(metricName, count),
	)
}

// SetDomainsSkipped records clew_build_knowledge_domains_skipped gauge.
func (r *EMFRecorder) SetDomainsSkipped(count int) {
	metricName := "clew_build_knowledge_domains_skipped"
	r.emitEMF(metricName, "Count", nil,
		slog.Int(metricName, count),
	)
}

// SetContentDomainsMissing records clew_build_content_domains_missing gauge.
func (r *EMFRecorder) SetContentDomainsMissing(count int) {
	metricName := "clew_build_content_domains_missing"
	r.emitEMF(metricName, "Count", nil,
		slog.Int(metricName, count),
	)
}

// RecordStartupTotal records clew_startup_total_seconds{result}.
func (r *EMFRecorder) RecordStartupTotal(result string, duration time.Duration) {
	metricName := "clew_startup_total_seconds"
	seconds := duration.Seconds()
	r.emitEMF(metricName, "Seconds", [][]string{{"result"}},
		slog.String("result", result),
		slog.Float64(metricName, seconds),
	)
}

// --- eventDedup metrics ---

// SetDedupMapSize records clew_dedup_map_size gauge.
func (r *EMFRecorder) SetDedupMapSize(size int) {
	r.dedupMapSize.Store(int64(size))

	metricName := "clew_dedup_map_size"
	r.emitEMF(metricName, "Count", nil,
		slog.Int(metricName, size),
	)
}

// IncrDedupDrops records clew_dedup_drops_total.
func (r *EMFRecorder) IncrDedupDrops() {
	r.mu.Lock()
	r.dedupDrops++
	r.mu.Unlock()

	metricName := "clew_dedup_drops_total"
	r.emitEMF(metricName, "Count", nil,
		slog.Int(metricName, 1),
	)
}

// --- Pre-pipeline metrics ---

// RecordPrePipelineLatency records clew_prepipeline_latency_seconds{cm_result}.
// cm_result: "hit", "miss_cold_start", "miss_ttl_expired", "miss_deploy_gap", "no_cm".
func (r *EMFRecorder) RecordPrePipelineLatency(cmResult string, duration time.Duration) {
	metricName := "clew_prepipeline_latency_seconds"
	seconds := duration.Seconds()
	r.emitEMF(metricName, "Seconds", [][]string{{"cm_result"}},
		slog.String("cm_result", cmResult),
		slog.Float64(metricName, seconds),
	)
}

// --- Contextual-equilibrium metrics ---

// IncrSectionCandidate records clew_ce_section_candidate_total.
func (r *EMFRecorder) IncrSectionCandidate() {
	metricName := "clew_ce_section_candidate_total"
	r.emitEMF(metricName, "Count", nil,
		slog.Int(metricName, 1),
	)
}

// IncrGraphInjected records clew_ce_graph_injected_total.
func (r *EMFRecorder) IncrGraphInjected(count int) {
	metricName := "clew_ce_graph_injected_total"
	r.emitEMF(metricName, "Count", nil,
		slog.Int(metricName, count),
	)
}

// IncrDiversityFloorEnforced records clew_ce_diversity_floor_enforced_total{domain_type}.
func (r *EMFRecorder) IncrDiversityFloorEnforced(domainType string) {
	metricName := "clew_ce_diversity_floor_enforced_total"
	r.emitEMF(metricName, "Count", [][]string{{"domain_type"}},
		slog.String("domain_type", domainType),
		slog.Int(metricName, 1),
	)
}

// IncrTypeCeilingHit records clew_ce_type_ceiling_hit_total{domain_type}.
func (r *EMFRecorder) IncrTypeCeilingHit(domainType string) {
	metricName := "clew_ce_type_ceiling_hit_total"
	r.emitEMF(metricName, "Count", [][]string{{"domain_type"}},
		slog.String("domain_type", domainType),
		slog.Int(metricName, 1),
	)
}

// RecordAssemblerTypeFraction records clew_ce_assembler_type_fraction{domain_type}.
func (r *EMFRecorder) RecordAssemblerTypeFraction(domainType string, fraction float64) {
	metricName := "clew_ce_assembler_type_fraction"
	r.emitEMF(metricName, "None", [][]string{{"domain_type"}},
		slog.String("domain_type", domainType),
		slog.Float64(metricName, fraction),
	)
}

// --- Snapshot methods for testing and diagnostics ---

// MetricsSnapshot returns a point-in-time snapshot of counter values.
// Useful for tests and the /metrics health endpoint.
type MetricsSnapshot struct {
	HitTotal       int64
	MissTotal      map[string]int64
	EvictionTotal  int64
	DroppedTotal   map[string]int64
	DedupDrops     int64
	StaleFallbacks int64
	BuildFailures  map[string]int64
	HaikuCalls     int64

	// Gauges.
	ConcurrentQueries int64
	ActiveThreads     int64
	MemoryBytes       int64
	DedupMapSize      int64
	StartupTimestamp  int64
}

// Snapshot returns a point-in-time copy of all counter and gauge values.
func (r *EMFRecorder) Snapshot() MetricsSnapshot {
	r.mu.Lock()
	defer r.mu.Unlock()

	missTotal := make(map[string]int64, len(r.missTotal))
	for k, v := range r.missTotal {
		missTotal[k] = v
	}
	droppedTotal := make(map[string]int64, len(r.droppedTotal))
	for k, v := range r.droppedTotal {
		droppedTotal[k] = v
	}
	buildFailures := make(map[string]int64, len(r.buildFailures))
	for k, v := range r.buildFailures {
		buildFailures[k] = v
	}

	return MetricsSnapshot{
		HitTotal:          r.hitTotal,
		MissTotal:         missTotal,
		EvictionTotal:     r.evictionTotal,
		DroppedTotal:      droppedTotal,
		DedupDrops:        r.dedupDrops,
		StaleFallbacks:    r.staleFallbacks,
		BuildFailures:     buildFailures,
		HaikuCalls:        r.haikuCalls,
		ConcurrentQueries: r.concurrentQueries.Load(),
		ActiveThreads:     r.activeThreads.Load(),
		MemoryBytes:       r.memoryBytes.Load(),
		DedupMapSize:      r.dedupMapSize.Load(),
		StartupTimestamp:  r.startupTimestamp.Load(),
	}
}

// --- EMF emission ---

// emitEMF writes a structured log line in CloudWatch Embedded Metric Format.
// When running on ECS/Fargate with awslogs driver, CloudWatch automatically
// extracts the metric from the _aws metadata block.
// When running locally, this is simply a structured log entry.
func (r *EMFRecorder) emitEMF(metricName, unit string, dimensions [][]string, attrs ...slog.Attr) {
	if dimensions == nil {
		dimensions = [][]string{{}}
	}

	aws := emfAWS{
		Timestamp: time.Now().UnixMilli(),
		CloudWatchMetrics: []emfCloudWatch{{
			Namespace:  "Clew",
			Dimensions: dimensions,
			Metrics: []emfMetric{{
				Name: metricName,
				Unit: unit,
			}},
		}},
	}

	// Build slog args: _aws metadata + metric-specific attributes.
	args := make([]any, 0, len(attrs)+2)
	args = append(args, slog.Any("_aws", aws))
	for _, a := range attrs {
		args = append(args, a)
	}

	r.logger.Info("metric", args...)
}

// NopRecorder is a no-op implementation of MetricsRecorder for testing.
// All methods are safe to call but do nothing.
type NopRecorder struct{}

var _ MetricsRecorder = (*NopRecorder)(nil)

func (NopRecorder) RecordStageLatency(string, string, time.Duration)          {}
func (NopRecorder) RecordQueryLatency(string, time.Duration)                  {}
func (NopRecorder) RecordHaikuCalls(int)                                      {}
func (NopRecorder) SetConcurrentQueries(int)                                  {}
func (NopRecorder) IncrDropped(string)                                        {}
func (NopRecorder) IncrConversationHit()                                      {}
func (NopRecorder) IncrConversationMiss(string)                               {}
func (NopRecorder) RecordConversationGetLatency(string, time.Duration)        {}
func (NopRecorder) SetActiveThreads(int)                                      {}
func (NopRecorder) IncrEvictions()                                            {}
func (NopRecorder) SetConversationMemoryBytes(int64)                          {}
func (NopRecorder) SetStartupTimestamp(time.Time)                             {}
func (NopRecorder) IncrSummarization(string)                                  {}
func (NopRecorder) RecordSummarizationLatency(string, time.Duration)          {}
func (NopRecorder) RecordBuildPhaseLatency(string, time.Duration)             {}
func (NopRecorder) IncrBuildFailures(string)                                  {}
func (NopRecorder) IncrStaleFallbacks()                                       {}
func (NopRecorder) SetDomainsReindexed(int)                                   {}
func (NopRecorder) SetDomainsSkipped(int)                                     {}
func (NopRecorder) SetContentDomainsMissing(int)                              {}
func (NopRecorder) RecordStartupTotal(string, time.Duration)                  {}
func (NopRecorder) SetDedupMapSize(int)                                       {}
func (NopRecorder) IncrDedupDrops()                                           {}
func (NopRecorder) RecordPrePipelineLatency(string, time.Duration)            {}
func (NopRecorder) IncrSectionCandidate()                                     {}
func (NopRecorder) IncrGraphInjected(int)                                     {}
func (NopRecorder) IncrDiversityFloorEnforced(string)                         {}
func (NopRecorder) IncrTypeCeilingHit(string)                                 {}
func (NopRecorder) RecordAssemblerTypeFraction(string, float64)               {}
