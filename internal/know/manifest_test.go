package know

import (
	"fmt"
	"testing"
)

// --- helpers ---

// resetManifestMocks restores all manifest git command vars to their defaults.
// Call via defer after overriding in a test.
func resetManifestMocks(
	origFiltered func(string, string, string) ([]string, error),
	origNumstat func(string, string) (int, int, error),
	origLog func(string, string) (string, error),
	origLs func([]string) (int, error),
	origNameStatus func(string, string) ([]string, error),
) {
	gitDiffFiltered = origFiltered
	gitDiffNumstat = origNumstat
	gitLogOneline = origLog
	gitLsFiles = origLs
	gitDiffNameStatus = origNameStatus
}

type manifestMocks struct {
	filtered   func(string, string, string) ([]string, error)
	numstat    func(string, string) (int, int, error)
	log        func(string, string) (string, error)
	lsFiles    func([]string) (int, error)
	nameStatus func(string, string) ([]string, error)
}

// installManifestMocks overrides all manifest git vars and returns a cleanup func.
func installManifestMocks(t *testing.T, m manifestMocks) {
	t.Helper()
	origFiltered := gitDiffFiltered
	origNumstat := gitDiffNumstat
	origLog := gitLogOneline
	origLs := gitLsFiles
	origNameStatus := gitDiffNameStatus

	if m.filtered != nil {
		gitDiffFiltered = m.filtered
	}
	if m.numstat != nil {
		gitDiffNumstat = m.numstat
	}
	if m.log != nil {
		gitLogOneline = m.log
	}
	if m.lsFiles != nil {
		gitLsFiles = m.lsFiles
	}
	if m.nameStatus != nil {
		gitDiffNameStatus = m.nameStatus
	}

	t.Cleanup(func() {
		resetManifestMocks(origFiltered, origNumstat, origLog, origLs, origNameStatus)
	})
}

// noRenames is a stub that returns no renamed files.
func noRenames(_, _ string) ([]string, error) { return nil, nil }

// noLog is a stub returning an empty commit log.
func noLog(_, _ string) (string, error) { return "", nil }

// zeroNumstat is a stub returning zero added/deleted lines.
func zeroNumstat(_, _ string) (int, int, error) { return 0, 0, nil }

// --- test cases ---

// TestComputeChangeManifest_NoChanges verifies that identical hashes return an empty manifest.
func TestComputeChangeManifest_NoChanges(t *testing.T) {
	// No mocks needed: identical hashes short-circuits before any git call.
	m, err := ComputeChangeManifest("abc1234", "abc1234", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if m == nil {
		t.Fatal("manifest is nil, want empty struct")
	}
	if len(m.NewFiles) != 0 || len(m.ModifiedFiles) != 0 ||
		len(m.DeletedFiles) != 0 || len(m.RenamedFiles) != 0 {
		t.Errorf("expected all file lists empty for identical hashes; got %+v", m)
	}
	if m.FromHash != "abc1234" || m.ToHash != "abc1234" {
		t.Errorf("hashes not preserved: got from=%q to=%q", m.FromHash, m.ToHash)
	}
}

// TestComputeChangeManifest_SmallDelta verifies a small set of modifications is
// classified as "incremental" by RecommendedMode.
func TestComputeChangeManifest_SmallDelta(t *testing.T) {
	installManifestMocks(t, manifestMocks{
		filtered: func(_, _, filter string) ([]string, error) {
			if filter == "M" {
				return []string{
					"internal/know/manifest.go",
					"internal/know/know.go",
					"internal/cmd/knows/knows.go",
				}, nil
			}
			return nil, nil
		},
		numstat:    func(_, _ string) (int, int, error) { return 100, 100, nil }, // 200 lines total
		log:        noLog,
		lsFiles:    func(_ []string) (int, error) { return 100, nil },
		nameStatus: noRenames,
	})

	manifest, err := ComputeChangeManifest("aaa0001", "bbb0002", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(manifest.ModifiedFiles) != 3 {
		t.Errorf("ModifiedFiles = %d, want 3", len(manifest.ModifiedFiles))
	}
	if manifest.DeltaLines != 200 {
		t.Errorf("DeltaLines = %d, want 200", manifest.DeltaLines)
	}

	mode := RecommendedMode(manifest, &Meta{})
	if mode != "incremental" {
		t.Errorf("RecommendedMode = %q, want %q", mode, "incremental")
	}
}

// TestComputeChangeManifest_LargeDelta verifies that a large number of files and lines
// causes RecommendedMode to return "full".
func TestComputeChangeManifest_LargeDelta(t *testing.T) {
	// Build 100 file names.
	manyFiles := make([]string, 100)
	for i := range manyFiles {
		manyFiles[i] = fmt.Sprintf("internal/pkg/file%03d.go", i)
	}

	installManifestMocks(t, manifestMocks{
		filtered: func(_, _, filter string) ([]string, error) {
			if filter == "M" {
				return manyFiles, nil
			}
			return nil, nil
		},
		numstat:    func(_, _ string) (int, int, error) { return 3000, 3000, nil }, // 6000 lines
		log:        noLog,
		lsFiles:    func(_ []string) (int, error) { return 200, nil },
		nameStatus: noRenames,
	})

	manifest, err := ComputeChangeManifest("aaa0001", "bbb0002", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if manifest.DeltaLines != 6000 {
		t.Errorf("DeltaLines = %d, want 6000", manifest.DeltaLines)
	}

	mode := RecommendedMode(manifest, &Meta{})
	if mode != "full" {
		t.Errorf("RecommendedMode = %q, want %q", mode, "full")
	}
}

// TestRecommendedMode_RatioAtThreshold verifies that DeltaRatio == 0.5 returns "full".
func TestRecommendedMode_RatioAtThreshold(t *testing.T) {
	// 50 changed files out of 100 total = ratio exactly 0.5.
	installManifestMocks(t, manifestMocks{
		filtered: func(_, _, filter string) ([]string, error) {
			if filter == "M" {
				files := make([]string, 50)
				for i := range files {
					files[i] = fmt.Sprintf("internal/pkg/f%d.go", i)
				}
				return files, nil
			}
			return nil, nil
		},
		numstat:    func(_, _ string) (int, int, error) { return 100, 100, nil },
		log:        noLog,
		lsFiles:    func(_ []string) (int, error) { return 100, nil },
		nameStatus: noRenames,
	})

	manifest, err := ComputeChangeManifest("aaa", "bbb", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if manifest.DeltaRatio != 0.5 {
		t.Errorf("DeltaRatio = %f, want 0.5", manifest.DeltaRatio)
	}

	mode := RecommendedMode(manifest, &Meta{})
	if mode != "full" {
		t.Errorf("RecommendedMode = %q, want %q (ratio >= 0.5 triggers full)", mode, "full")
	}
}

// TestRecommendedMode_LinesAtThreshold verifies that DeltaLines == 5000 returns "full".
func TestRecommendedMode_LinesAtThreshold(t *testing.T) {
	installManifestMocks(t, manifestMocks{
		filtered: func(_, _, filter string) ([]string, error) {
			if filter == "M" {
				return []string{"internal/big_file.go"}, nil
			}
			return nil, nil
		},
		numstat:    func(_, _ string) (int, int, error) { return 2500, 2500, nil }, // exactly 5000 lines
		log:        noLog,
		lsFiles:    func(_ []string) (int, error) { return 1000, nil },
		nameStatus: noRenames,
	})

	manifest, err := ComputeChangeManifest("aaa", "bbb", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if manifest.DeltaLines != 5000 {
		t.Errorf("DeltaLines = %d, want 5000", manifest.DeltaLines)
	}

	mode := RecommendedMode(manifest, &Meta{})
	if mode != "full" {
		t.Errorf("RecommendedMode = %q, want %q (lines >= 5000 triggers full)", mode, "full")
	}
}

// TestRecommendedMode_BelowBothThresholds verifies that a small delta returns "incremental".
func TestRecommendedMode_BelowBothThresholds(t *testing.T) {
	// 2 changed files out of 100 total = ratio 0.02 (< 0.5); 100 lines (< 5000).
	installManifestMocks(t, manifestMocks{
		filtered: func(_, _, filter string) ([]string, error) {
			if filter == "A" {
				return []string{"internal/new_feature.go", "internal/new_feature_test.go"}, nil
			}
			return nil, nil
		},
		numstat:    func(_, _ string) (int, int, error) { return 50, 50, nil },
		log:        func(_, _ string) (string, error) { return "abc1234 feat: add feature", nil },
		lsFiles:    func(_ []string) (int, error) { return 100, nil },
		nameStatus: noRenames,
	})

	manifest, err := ComputeChangeManifest("aaa", "bbb", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	mode := RecommendedMode(manifest, &Meta{})
	if mode != "incremental" {
		t.Errorf("RecommendedMode = %q, want %q", mode, "incremental")
	}
}

// TestComputeChangeManifest_RenamesDetected verifies that renamed files appear
// in RenamedFiles and not in NewFiles or DeletedFiles.
func TestComputeChangeManifest_RenamesDetected(t *testing.T) {
	installManifestMocks(t, manifestMocks{
		// No added/modified/deleted — only renames.
		filtered: func(_, _, _ string) ([]string, error) { return nil, nil },
		numstat:  zeroNumstat,
		log:      noLog,
		lsFiles:  func(_ []string) (int, error) { return 50, nil },
		nameStatus: func(_, _ string) ([]string, error) {
			return []string{
				"R100\tinternal/old_name.go\tinternal/new_name.go",
				"R090\tcmd/oldcmd/main.go\tcmd/newcmd/main.go",
			}, nil
		},
	})

	manifest, err := ComputeChangeManifest("aaa", "bbb", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(manifest.NewFiles) != 0 {
		t.Errorf("NewFiles should be empty for renames-only delta, got: %v", manifest.NewFiles)
	}
	if len(manifest.DeletedFiles) != 0 {
		t.Errorf("DeletedFiles should be empty for renames-only delta, got: %v", manifest.DeletedFiles)
	}
	if len(manifest.RenamedFiles) != 2 {
		t.Fatalf("RenamedFiles = %d, want 2; got: %v", len(manifest.RenamedFiles), manifest.RenamedFiles)
	}

	// Verify rename pairs.
	r0 := manifest.RenamedFiles[0]
	if r0.OldPath != "internal/old_name.go" || r0.NewPath != "internal/new_name.go" {
		t.Errorf("RenamedFiles[0] = {%q, %q}, want {%q, %q}",
			r0.OldPath, r0.NewPath, "internal/old_name.go", "internal/new_name.go")
	}
	r1 := manifest.RenamedFiles[1]
	if r1.OldPath != "cmd/oldcmd/main.go" || r1.NewPath != "cmd/newcmd/main.go" {
		t.Errorf("RenamedFiles[1] = {%q, %q}, want {%q, %q}",
			r1.OldPath, r1.NewPath, "cmd/oldcmd/main.go", "cmd/newcmd/main.go")
	}
}

// TestComputeChangeManifest_SourceScopeFiltering verifies that files outside
// the declared source_scope are excluded from the manifest.
func TestComputeChangeManifest_SourceScopeFiltering(t *testing.T) {
	installManifestMocks(t, manifestMocks{
		filtered: func(_, _, filter string) ([]string, error) {
			if filter == "M" {
				return []string{
					"internal/know/know.go",     // in scope
					"docs/README.md",             // out of scope
					"rites/shared/mena/foo.md",  // out of scope
					"internal/cmd/knows/cmd.go", // in scope
				}, nil
			}
			return nil, nil
		},
		numstat:    func(_, _ string) (int, int, error) { return 50, 50, nil },
		log:        noLog,
		lsFiles:    func(_ []string) (int, error) { return 80, nil },
		nameStatus: noRenames,
	})

	scope := []string{"internal/**/*.go", "cmd/**/*.go"}
	manifest, err := ComputeChangeManifest("aaa", "bbb", scope)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Only the 2 in-scope Go files should appear.
	if len(manifest.ModifiedFiles) != 2 {
		t.Errorf("ModifiedFiles = %v, want 2 in-scope files", manifest.ModifiedFiles)
	}
	for _, f := range manifest.ModifiedFiles {
		if !matchScope("internal/**/*.go", f) && !matchScope("cmd/**/*.go", f) {
			t.Errorf("out-of-scope file leaked into ModifiedFiles: %q", f)
		}
	}
}

// TestRecommendedMode_EmptyManifest verifies that a manifest with no file changes
// returns "time-only".
func TestRecommendedMode_EmptyManifest(t *testing.T) {
	// Build a manifest with no files (all lists nil/empty).
	m := &ChangeManifest{
		FromHash:   "aaa",
		ToHash:     "bbb",
		CommitLog:  "abc1234 chore: update docs only",
		DeltaLines: 10,
		TotalFiles: 100,
		DeltaRatio: 0.0,
	}

	mode := RecommendedMode(m, &Meta{})
	if mode != "time-only" {
		t.Errorf("RecommendedMode = %q, want %q (no file changes)", mode, "time-only")
	}
}

// TestRecommendedMode_NilManifest verifies that a nil manifest returns "time-only".
func TestRecommendedMode_NilManifest(t *testing.T) {
	mode := RecommendedMode(nil, &Meta{})
	if mode != "time-only" {
		t.Errorf("RecommendedMode(nil) = %q, want %q", mode, "time-only")
	}
}

// TestRecommendedMode_CycleLimitHit verifies that when IncrementalCycle >= MaxIncrementalCycles,
// "full" is returned regardless of delta size.
func TestRecommendedMode_CycleLimitHit(t *testing.T) {
	// Small delta that would normally be "incremental".
	m := &ChangeManifest{
		FromHash:      "aaa",
		ToHash:        "bbb",
		ModifiedFiles: []string{"internal/tiny.go"},
		DeltaLines:    5,
		TotalFiles:    200,
		DeltaRatio:    0.005,
	}

	meta := &Meta{
		MaxIncrementalCycles: 3,
		IncrementalCycle:     3, // at limit
	}
	mode := RecommendedMode(m, meta)
	if mode != "full" {
		t.Errorf("RecommendedMode = %q, want %q (cycle limit hit)", mode, "full")
	}

	// Exceeding limit also forces full.
	meta.IncrementalCycle = 10
	mode = RecommendedMode(m, meta)
	if mode != "full" {
		t.Errorf("RecommendedMode = %q, want %q (cycle exceeded)", mode, "full")
	}
}

// TestRecommendedMode_CycleLimitZeroMeansNoLimit verifies that MaxIncrementalCycles == 0
// is treated as "no limit" (backward compatibility).
func TestRecommendedMode_CycleLimitZeroMeansNoLimit(t *testing.T) {
	m := &ChangeManifest{
		FromHash:      "aaa",
		ToHash:        "bbb",
		ModifiedFiles: []string{"internal/tiny.go"},
		DeltaLines:    5,
		TotalFiles:    200,
		DeltaRatio:    0.005,
	}

	// MaxIncrementalCycles == 0: no limit, so even a high cycle count should not force full.
	meta := &Meta{
		MaxIncrementalCycles: 0,
		IncrementalCycle:     999,
	}
	mode := RecommendedMode(m, meta)
	if mode != "incremental" {
		t.Errorf("RecommendedMode = %q, want %q (zero max = no limit)", mode, "incremental")
	}
}

// TestComputeChangeManifest_GitError verifies that a git command failure propagates as error.
func TestComputeChangeManifest_GitError(t *testing.T) {
	installManifestMocks(t, manifestMocks{
		filtered: func(_, _, _ string) ([]string, error) {
			return nil, fmt.Errorf("git: command not found")
		},
		numstat:    zeroNumstat,
		log:        noLog,
		lsFiles:    func(_ []string) (int, error) { return 0, nil },
		nameStatus: noRenames,
	})

	_, err := ComputeChangeManifest("aaa", "bbb", nil)
	if err == nil {
		t.Error("expected error from failed git command, got nil")
	}
}

// TestComputeChangeManifest_TotalFilesZeroForcesFullRatio verifies that when
// TotalFiles is 0, DeltaRatio is set to 1.0.
func TestComputeChangeManifest_TotalFilesZeroForcesFullRatio(t *testing.T) {
	installManifestMocks(t, manifestMocks{
		filtered: func(_, _, filter string) ([]string, error) {
			if filter == "M" {
				return []string{"internal/some.go"}, nil
			}
			return nil, nil
		},
		numstat:    func(_, _ string) (int, int, error) { return 10, 5, nil },
		log:        noLog,
		lsFiles:    func(_ []string) (int, error) { return 0, nil }, // no tracked files
		nameStatus: noRenames,
	})

	manifest, err := ComputeChangeManifest("aaa", "bbb", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if manifest.DeltaRatio != 1.0 {
		t.Errorf("DeltaRatio = %f, want 1.0 when TotalFiles == 0", manifest.DeltaRatio)
	}

	mode := RecommendedMode(manifest, &Meta{})
	if mode != "full" {
		t.Errorf("RecommendedMode = %q, want %q (ratio 1.0 triggers full)", mode, "full")
	}
}

// TestComputeChangeManifest_CommitLogPopulated verifies that the commit log is captured.
func TestComputeChangeManifest_CommitLogPopulated(t *testing.T) {
	installManifestMocks(t, manifestMocks{
		filtered:   func(_, _, _ string) ([]string, error) { return nil, nil },
		numstat:    zeroNumstat,
		log:        func(_, _ string) (string, error) { return "abc1234 feat: something great", nil },
		lsFiles:    func(_ []string) (int, error) { return 10, nil },
		nameStatus: noRenames,
	})

	manifest, err := ComputeChangeManifest("aaa", "bbb", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if manifest.CommitLog != "abc1234 feat: something great" {
		t.Errorf("CommitLog = %q, want %q", manifest.CommitLog, "abc1234 feat: something great")
	}
}

// TestComputeChangeManifest_ScopeFilterRenames verifies that renames where the new
// path is out-of-scope are excluded.
func TestComputeChangeManifest_ScopeFilterRenames(t *testing.T) {
	installManifestMocks(t, manifestMocks{
		filtered: func(_, _, _ string) ([]string, error) { return nil, nil },
		numstat:  zeroNumstat,
		log:      noLog,
		lsFiles:  func(_ []string) (int, error) { return 50, nil },
		nameStatus: func(_, _ string) ([]string, error) {
			return []string{
				// new path is in scope
				"R100\tinternal/old.go\tinternal/new.go",
				// new path is out of scope (docs)
				"R100\tdocs/old.md\tdocs/new.md",
			}, nil
		},
	})

	scope := []string{"internal/**/*.go"}
	manifest, err := ComputeChangeManifest("aaa", "bbb", scope)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(manifest.RenamedFiles) != 1 {
		t.Fatalf("RenamedFiles = %d, want 1 (only in-scope rename); got: %v",
			len(manifest.RenamedFiles), manifest.RenamedFiles)
	}
	if manifest.RenamedFiles[0].NewPath != "internal/new.go" {
		t.Errorf("expected in-scope rename, got: %+v", manifest.RenamedFiles[0])
	}
}
