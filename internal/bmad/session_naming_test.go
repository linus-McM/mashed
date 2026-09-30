package bmad

import (
	"regexp"
	"strings"
	"testing"
)

// hashTailRe matches the 8-char lowercase hex hash suffix at the end of a name.
var hashTailRe = regexp.MustCompile(`^[0-9a-f]{8}$`)

// allowedSlugRe matches a single session-name component containing only slug-safe chars.
var allowedSlugRe = regexp.MustCompile(`^[a-z0-9_-]*$`)

// fullNameRe validates the overall descriptive session-name shape.
var fullNameRe = regexp.MustCompile(`^bmad-[a-z0-9_-]+-[a-z0-9_-]+-[a-z0-9_-]+-[0-9a-f]{8}$`)

// allEmptyFallbackRe matches the all-empty-input fallback from AC-3.
var allEmptyFallbackRe = regexp.MustCompile(`^bmad-unknown-unknown-unlabeled-[0-9a-f]{8}$`)

func buildAndSplit(t *testing.T, repoPath, branch, label, nodeID string, nowNanos int64) (full, repo, br, lab, hash string) {
	t.Helper()
	full = BuildSessionName(repoPath, branch, label, nodeID, nowNanos)
	if !strings.HasPrefix(full, SessionNamePrefix) {
		t.Fatalf("BuildSessionName(%q,%q,%q,%q,%d)=%q: missing prefix %q",
			repoPath, branch, label, nodeID, nowNanos, full, SessionNamePrefix)
	}
	body := strings.TrimPrefix(full, SessionNamePrefix)
	parts := strings.Split(body, "-")
	if len(parts) < 4 {
		t.Fatalf("BuildSessionName=%q: expected at least 4 dash-separated components, got %d", full, len(parts))
	}
	hash = parts[len(parts)-1]
	repo = parts[0]
	br = parts[1]
	lab = strings.Join(parts[2:len(parts)-1], "-")
	return full, repo, br, lab, hash
}

func assertValidHash(t *testing.T, context, hash string) {
	t.Helper()
	if !hashTailRe.MatchString(hash) {
		t.Errorf("%s: hash tail %q is not 8 lowercase hex", context, hash)
	}
}

// ─── AC-1: happy path ───────────────────────────────────────────────────────

func TestBuildSessionName_AC1_HappyPath(t *testing.T) {
	const (
		repoPath = "/Users/dev/Development/surfseer"
		branch   = "main"
		label    = "Create Story"
		nodeID   = "node-1775795467345"
		nowNanos = int64(1712600000000000000)
	)
	got := BuildSessionName(repoPath, branch, label, nodeID, nowNanos)

	wantPrefix := "bmad-surfseer-main-create-story-"
	if !strings.HasPrefix(got, wantPrefix) {
		t.Errorf("BuildSessionName = %q, want prefix %q", got, wantPrefix)
	}
	if len(got) < 8 {
		t.Fatalf("BuildSessionName = %q, too short for hash suffix", got)
	}
	assertValidHash(t, "AC1 happy path", got[len(got)-8:])
	if !fullNameRe.MatchString(got) {
		t.Errorf("BuildSessionName = %q, does not match full-name shape %s", got, fullNameRe)
	}
}

// ─── AC-2: slugification rules ──────────────────────────────────────────────

func TestBuildSessionName_AC2_SlugificationRules(t *testing.T) {
	cases := []struct {
		name       string
		repoPath   string
		branch     string
		label      string
		nodeID     string
		wantBranch string // optional exact branch assertion
		wantLabel  string // optional exact label assertion
	}{
		{
			// Branch with slashes slugifies to "feature-foo-bar" — but because the
			// name format uses '-' as separator, the assembled name cannot be split
			// deterministically back into branch and label when either contains
			// dashes. Exact-match assertions live in TestSlugifyComponent instead.
			name:     "feature branch with slashes collapses",
			repoPath: "/tmp/myrepo",
			branch:   "feature/foo/bar",
			label:    "Create Story",
			nodeID:   "n1",
		},
		{
			name:     "unicode runes stripped",
			repoPath: "/tmp/myrepo",
			branch:   "main",
			label:    "Créate Störy 日本語",
			nodeID:   "n1",
		},
		{
			name:      "runs of whitespace collapse to single dash",
			repoPath:  "/tmp/myrepo",
			branch:    "main",
			label:     "foo   bar",
			nodeID:    "n1",
			wantLabel: "foo-bar",
		},
		{
			name:      "leading and trailing whitespace trimmed",
			repoPath:  "/tmp/myrepo",
			branch:    "main",
			label:     "  create  ",
			nodeID:    "n1",
			wantLabel: "create",
		},
		{
			name:     "mixed case lowercased",
			repoPath: "/tmp/MyRepo",
			branch:   "Main",
			label:    "MAKE-IT-GO",
			nodeID:   "n1",
		},
		{
			name:      "underscores preserved",
			repoPath:  "/tmp/my_repo",
			branch:    "main",
			label:     "draft_prd",
			nodeID:    "n1",
			wantLabel: "draft_prd",
		},
		{
			name:      "runs of dashes collapse",
			repoPath:  "/tmp/myrepo",
			branch:    "main",
			label:     "foo---bar---baz",
			nodeID:    "n1",
			wantLabel: "foo-bar-baz",
		},
		{
			name:     "dots and symbols stripped",
			repoPath: "/tmp/my.repo",
			branch:   "main",
			label:    "v1.2.3 (final)",
			nodeID:   "n1",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			full, repo, br, lab, hash := buildAndSplit(t, tc.repoPath, tc.branch, tc.label, tc.nodeID, 1)

			for _, comp := range []struct{ name, val string }{
				{"repo", repo},
				{"branch", br},
				{"label", lab},
			} {
				if comp.val == "" {
					t.Errorf("component %q is empty for full=%q", comp.name, full)
				}
				if !allowedSlugRe.MatchString(comp.val) {
					t.Errorf("%s component %q contains illegal chars (want [a-z0-9_-])", comp.name, comp.val)
				}
				if strings.HasPrefix(comp.val, "-") || strings.HasSuffix(comp.val, "-") {
					t.Errorf("%s component %q has leading/trailing dash", comp.name, comp.val)
				}
				if strings.Contains(comp.val, "--") {
					t.Errorf("%s component %q contains a run of dashes", comp.name, comp.val)
				}
			}
			if tc.wantBranch != "" && br != tc.wantBranch {
				t.Errorf("branch = %q, want %q", br, tc.wantBranch)
			}
			if tc.wantLabel != "" && lab != tc.wantLabel {
				t.Errorf("label = %q, want %q", lab, tc.wantLabel)
			}
			assertValidHash(t, tc.name, hash)
		})
	}
}

// ─── AC-3: empty-input fallbacks ────────────────────────────────────────────

func TestBuildSessionName_AC3_AllEmptyFallback(t *testing.T) {
	got := BuildSessionName("", "", "", "", 123)
	if !allEmptyFallbackRe.MatchString(got) {
		t.Errorf("BuildSessionName(all empty) = %q, want match %s", got, allEmptyFallbackRe)
	}
}

func TestBuildSessionName_AC3_EmptyComponentFallbacks(t *testing.T) {
	cases := []struct {
		name       string
		repoPath   string
		branch     string
		label      string
		nodeID     string
		wantRepo   string
		wantBranch string
		wantLabel  string
	}{
		{
			name:       "empty repo falls back to unknown",
			repoPath:   "",
			branch:     "main",
			label:      "create-story",
			nodeID:     "n1",
			wantRepo:   "unknown",
			wantBranch: "main",
			wantLabel:  "create-story",
		},
		{
			name:       "empty branch falls back to unknown",
			repoPath:   "/tmp/myrepo",
			branch:     "",
			label:      "create-story",
			nodeID:     "n1",
			wantRepo:   "myrepo",
			wantBranch: "unknown",
			wantLabel:  "create-story",
		},
		{
			name:       "empty label uses slugified nodeID",
			repoPath:   "/tmp/myrepo",
			branch:     "main",
			label:      "",
			nodeID:     "fallback-id",
			wantRepo:   "myrepo",
			wantBranch: "main",
			wantLabel:  "fallback-id",
		},
		{
			name:       "empty label and empty nodeID → unlabeled",
			repoPath:   "/tmp/myrepo",
			branch:     "main",
			label:      "",
			nodeID:     "",
			wantRepo:   "myrepo",
			wantBranch: "main",
			wantLabel:  "unlabeled",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, repo, br, lab, hash := buildAndSplit(t, tc.repoPath, tc.branch, tc.label, tc.nodeID, 42)
			if repo != tc.wantRepo {
				t.Errorf("repo = %q, want %q", repo, tc.wantRepo)
			}
			if br != tc.wantBranch {
				t.Errorf("branch = %q, want %q", br, tc.wantBranch)
			}
			if lab != tc.wantLabel {
				t.Errorf("label = %q, want %q", lab, tc.wantLabel)
			}
			assertValidHash(t, tc.name, hash)
		})
	}
}

// ─── AC-4: length cap ───────────────────────────────────────────────────────

func TestBuildSessionName_AC4_LengthCapWith500ByteInputs(t *testing.T) {
	longRepo := "/tmp/" + strings.Repeat("a", 500)
	longBranch := strings.Repeat("b", 500)
	longLabel := strings.Repeat("c", 500)
	longNodeID := strings.Repeat("d", 500)

	got := BuildSessionName(longRepo, longBranch, longLabel, longNodeID, 999)

	if got == "" {
		t.Fatal("BuildSessionName returned empty string for long inputs")
	}
	if len(got) > MaxSessionNameBytes {
		t.Errorf("len(result)=%d, want <= %d (MaxSessionNameBytes); got=%q", len(got), MaxSessionNameBytes, got)
	}
	if !strings.HasPrefix(got, SessionNamePrefix) {
		t.Errorf("result %q does not start with %q", got, SessionNamePrefix)
	}
	if len(got) < 8 {
		t.Fatalf("result %q too short for hash suffix", got)
	}
	assertValidHash(t, "AC4 length cap", got[len(got)-8:])
	if got[len(got)-9] != '-' {
		t.Errorf("expected '-' separator before 8-char hash in %q", got)
	}
}

func TestBuildSessionName_AC4_HashPreservedUnderTruncation(t *testing.T) {
	longRepo := strings.Repeat("x", 500)
	longBranch := strings.Repeat("y", 500)
	longLabel := strings.Repeat("z", 500)
	longNodeID := strings.Repeat("q", 500)

	a := BuildSessionName(longRepo, longBranch, longLabel, longNodeID, 42)
	b := BuildSessionName(longRepo, longBranch, longLabel, longNodeID, 42)
	if a != b {
		t.Errorf("non-deterministic under truncation:\n a=%q\n b=%q", a, b)
	}
	assertValidHash(t, "AC4 truncated hash", a[len(a)-8:])
}

// ─── Hash determinism and uniqueness ────────────────────────────────────────

func TestBuildSessionName_HashDeterminismAndUniqueness(t *testing.T) {
	const (
		repoPath = "/tmp/myrepo"
		branch   = "main"
		label    = "Create Story"
	)
	cases := []struct {
		name          string
		aNodeID       string
		aNanos        int64
		bNodeID       string
		bNanos        int64
		wantHashEqual bool
		wantBodyEqual bool // prefix-before-hash equality
	}{
		{
			name:          "identical inputs are deterministic",
			aNodeID:       "n1", aNanos: 1000,
			bNodeID: "n1", bNanos: 1000,
			wantHashEqual: true,
			wantBodyEqual: true,
		},
		{
			name:          "different nanos produce different hash, same body",
			aNodeID:       "n1", aNanos: 1000,
			bNodeID: "n1", bNanos: 2000,
			wantHashEqual: false,
			wantBodyEqual: true,
		},
		{
			name:          "different nodeIDs produce different hash",
			aNodeID:       "n1", aNanos: 1000,
			bNodeID: "n2", bNanos: 1000,
			wantHashEqual: false,
			wantBodyEqual: true,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			a := BuildSessionName(repoPath, branch, label, tc.aNodeID, tc.aNanos)
			b := BuildSessionName(repoPath, branch, label, tc.bNodeID, tc.bNanos)

			aHash, bHash := a[len(a)-8:], b[len(b)-8:]
			aBody, bBody := a[:len(a)-8], b[:len(b)-8]

			if tc.wantHashEqual && aHash != bHash {
				t.Errorf("expected equal hashes, got a=%q b=%q", aHash, bHash)
			}
			if !tc.wantHashEqual && aHash == bHash {
				t.Errorf("expected different hashes, both=%q", aHash)
			}
			if tc.wantBodyEqual && aBody != bBody {
				t.Errorf("expected equal prefixes:\n a=%q\n b=%q", aBody, bBody)
			}
		})
	}
}

// ─── AC-5: ParseSessionName round-trip ──────────────────────────────────────

func TestParseSessionName_AC5_RoundTrip(t *testing.T) {
	name := BuildSessionName("/x/myrepo", "main", "Create Story", "n1", 100)
	repo, branch, label, shortHash, ok := ParseSessionName(name)
	if !ok {
		t.Fatalf("ParseSessionName(%q) ok=false, want true", name)
	}
	if repo != "myrepo" {
		t.Errorf("repo = %q, want %q", repo, "myrepo")
	}
	if branch != "main" {
		t.Errorf("branch = %q, want %q", branch, "main")
	}
	if label != "create-story" {
		t.Errorf("label = %q, want %q", label, "create-story")
	}
	assertValidHash(t, "AC5 round trip", shortHash)
}

// Feature branches with slashes slugify to multi-dash strings (feature-foo-bar).
// The name format uses '-' as separator, so round-tripping branch/label through
// ParseSessionName is ambiguous when either contains dashes. We only assert that
// the slugified branch and label tokens appear somewhere in the assembled name
// and that parsing still succeeds with a valid hash tail.
func TestParseSessionName_AC5_FeatureBranchAppearsInName(t *testing.T) {
	name := BuildSessionName("/x/myrepo", "feature/foo/bar", "Draft PRD", "n1", 200)
	if !strings.Contains(name, "feature-foo-bar") {
		t.Errorf("name %q does not contain slugified branch %q", name, "feature-foo-bar")
	}
	if !strings.Contains(name, "draft-prd") {
		t.Errorf("name %q does not contain slugified label %q", name, "draft-prd")
	}
	_, _, _, shortHash, ok := ParseSessionName(name)
	if !ok {
		t.Fatalf("ParseSessionName(%q) ok=false, want true", name)
	}
	assertValidHash(t, "AC5 feature branch name", shortHash)
}

// ─── AC-6: ParseSessionName rejection ───────────────────────────────────────

func TestParseSessionName_AC6_InvalidInputs(t *testing.T) {
	cases := []struct {
		name  string
		input string
	}{
		{"empty string", ""},
		{"no bmad prefix", "not-a-bmad-session"},
		{"unrelated dashed name", "foo-bar-baz-12345678"},
		{"too short", "bmad-short"},
		{"non-hex hash tail", "bmad-repo-branch-label-XYZNOTHEX"},
		{"hash wrong length (7 chars)", "bmad-repo-branch-label-abcdef1"},
		{"hash wrong length (9 chars)", "bmad-repo-branch-label-abcdef123"},
		{"uppercase hex tail", "bmad-repo-branch-label-ABCDEF12"},
		{"no hash component at all", "bmad-repo-branch-label"},
		{"missing label component", "bmad-repo-branch-abcdef12"},
		{"prefix only", "bmad-"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			repo, branch, label, shortHash, ok := ParseSessionName(tc.input)
			if ok {
				t.Errorf("ParseSessionName(%q) ok=true, want false", tc.input)
			}
			if repo != "" || branch != "" || label != "" || shortHash != "" {
				t.Errorf("ParseSessionName(%q) fields should be empty on failure, got repo=%q branch=%q label=%q shortHash=%q",
					tc.input, repo, branch, label, shortHash)
			}
		})
	}
}

// ─── slugifyComponent direct coverage ───────────────────────────────────────
//
// slugifyComponent is package-private (internal/bmad), but these tests are in
// package bmad so they can call it directly. This gives unambiguous coverage
// of the slug rules independently of the BuildSessionName assembly format,
// which cannot round-trip branches/labels containing dashes.

func TestSlugifyComponent(t *testing.T) {
	const max = 24
	cases := []struct {
		name   string
		input  string
		maxLen int
		want   string
	}{
		{"simple lowercase phrase", "Create Story", max, "create-story"},
		{"slashes become dashes", "feature/foo/bar", max, "feature-foo-bar"},
		{"runs of spaces collapse", "foo   bar", max, "foo-bar"},
		{"leading and trailing spaces trimmed", "  create  ", max, "create"},
		{"unicode runes stripped, collapsed, trimmed", "Créate Störy 日本語", max, "cr-ate-st-ry"},
		{"empty string stays empty", "", max, ""},
		{"runs of dashes collapse", "foo---bar---baz", max, "foo-bar-baz"},
		{"underscores preserved", "draft_prd", max, "draft_prd"},
		{"mixed case lowercased", "MAKE-IT-GO", max, "make-it-go"},
		{"dots and punctuation stripped", "v1.2.3 (final)", max, "v1-2-3-final"},
		{"long input truncated to maxLen", strings.Repeat("a", 300), max, strings.Repeat("a", max)},
		{"trim leading dashes after substitution", "---leading", max, "leading"},
		{"trim trailing dashes after substitution", "trailing---", max, "trailing"},
		{"only punctuation collapses to empty", "!!!", max, ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := slugifyComponent(tc.input, tc.maxLen)
			if got != tc.want {
				t.Errorf("slugifyComponent(%q, %d) = %q, want %q", tc.input, tc.maxLen, got, tc.want)
			}
			if len(got) > tc.maxLen {
				t.Errorf("slugifyComponent(%q, %d) returned %d bytes, exceeds maxLen", tc.input, tc.maxLen, len(got))
			}
			if got != "" && !allowedSlugRe.MatchString(got) {
				t.Errorf("slugifyComponent(%q) = %q contains illegal characters", tc.input, got)
			}
		})
	}
}

// ─── Exported constants sanity ──────────────────────────────────────────────

func TestSessionNaming_ConstantsSanity(t *testing.T) {
	if SessionNamePrefix != "bmad-" {
		t.Errorf("SessionNamePrefix = %q, want %q", SessionNamePrefix, "bmad-")
	}
	if MaxSessionNameBytes != 100 {
		t.Errorf("MaxSessionNameBytes = %d, want 100", MaxSessionNameBytes)
	}
}
