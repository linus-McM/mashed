package bmad

import (
	"crypto/sha256"
	"encoding/hex"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

// SessionNamePrefix is the required leading token for every BMAD tmux session
// name. Callers that need to recognise BMAD-managed sessions should match on
// this prefix via strings.HasPrefix or ParseSessionName.
const SessionNamePrefix = "bmad-"

// DetachedBranch is the branch component substituted into a session name
// when the `git rev-parse --abbrev-ref HEAD` lookup fails (e.g. detached
// HEAD, non-repo, or missing git binary). Exposing it as a constant keeps
// the executor and its tests in lockstep.
const DetachedBranch = "detached"

// MaxSessionNameBytes caps the total length of a generated session name.
// The per-component slug cap (maxComponentBytes) is tuned so that a fully
// populated name never exceeds this budget.
const MaxSessionNameBytes = 100

// maxComponentBytes is the per-component (repo/branch/label) slug cap. With
// three components at 24 bytes each plus the "bmad-" prefix, three dash
// separators, and an 8-char hash suffix, a fully-populated name fits in
// 5 + 24 + 1 + 24 + 1 + 24 + 1 + 8 = 88 bytes — comfortably under
// MaxSessionNameBytes, so the per-component cap is the only enforcement
// needed to keep names within the overall budget.
const maxComponentBytes = 24

// shortHashBytes is the number of hex characters in the hash suffix.
const shortHashBytes = 8

// Sentinel components substituted when the corresponding input is empty
// after slugification.
const (
	sentinelUnknownRepo   = "unknown"
	sentinelUnknownBranch = "unknown"
	sentinelUnlabeled     = "unlabeled"
)

// hashValidRe matches the required shape of a session-name hash suffix:
// exactly 8 lowercase hexadecimal characters.
var hashValidRe = regexp.MustCompile(`^[0-9a-f]{8}$`)

// unsafeSlugRun matches any run of characters that are not slug-safe, so
// callers can collapse them into a single dash separator.
var unsafeSlugRun = regexp.MustCompile(`[^a-z0-9_-]+`)

// dashRun collapses runs of '-' into a single dash.
var dashRun = regexp.MustCompile(`-+`)

// BuildSessionName assembles a descriptive tmux session name of the form
//
//	bmad-{repo}-{branch}-{label}-{shortHash}
//
// from the repo path basename, branch, node label (or node ID as fallback),
// and a sha256-derived 8-character hash of "{nodeID}|{nowNanos}". Empty
// components fall back to sentinel values ("unknown" for repo/branch,
// "unlabeled" when both label and nodeID are empty). Each component is
// slugified to [a-z0-9_-] and capped at maxComponentBytes, which keeps the
// total length within MaxSessionNameBytes (the maximum assembled name is
// len("bmad-") + 3*24 + 3 + shortHashBytes = 88 bytes).
func BuildSessionName(repoPath, branch, nodeLabel, nodeID string, nowNanos int64) string {
	repo := slugifyComponent(filepath.Base(repoPath), maxComponentBytes)
	// filepath.Base returns "." for an empty path; treat that as missing.
	if repo == "" || repo == "." {
		repo = sentinelUnknownRepo
	}

	br := slugifyComponent(branch, maxComponentBytes)
	if br == "" {
		br = sentinelUnknownBranch
	}

	lab := slugifyComponent(nodeLabel, maxComponentBytes)
	if lab == "" {
		lab = slugifyComponent(nodeID, maxComponentBytes)
	}
	if lab == "" {
		lab = sentinelUnlabeled
	}

	return SessionNamePrefix + repo + "-" + br + "-" + lab + "-" + shortHashOf(nodeID, nowNanos)
}

// ParseSessionName is the inverse of BuildSessionName. It returns the
// recovered components when name is a well-formed BMAD session name and
// ok=false otherwise. The returned label rejoins any middle-position
// dashes, because user-facing labels commonly slugify to forms like
// "create-story". Branches with slashes are not round-trip-recoverable
// from a dash-separated format; callers that need full round-trip fidelity
// must store the raw branch separately.
func ParseSessionName(name string) (repo, branch, label, shortHash string, ok bool) {
	if !strings.HasPrefix(name, SessionNamePrefix) {
		return "", "", "", "", false
	}
	body := strings.TrimPrefix(name, SessionNamePrefix)
	parts := strings.Split(body, "-")
	// Need at least repo, branch, label, hash.
	if len(parts) < 4 {
		return "", "", "", "", false
	}
	hash := parts[len(parts)-1]
	if !hashValidRe.MatchString(hash) {
		return "", "", "", "", false
	}
	repo = parts[0]
	branch = parts[1]
	label = strings.Join(parts[2:len(parts)-1], "-")
	if repo == "" || branch == "" || label == "" {
		return "", "", "", "", false
	}
	shortHash = hash
	ok = true
	return
}

// slugifyComponent normalises an arbitrary string into the constrained
// session-name alphabet [a-z0-9_-]. It lowercases, replaces each run of
// unsafe runes with a single dash, collapses any resulting run of dashes,
// trims leading and trailing dashes, and truncates to maxLen bytes. The
// returned string is guaranteed to match the slug alphabet and to have no
// leading, trailing, or consecutive dashes.
func slugifyComponent(s string, maxLen int) string {
	if s == "" || maxLen <= 0 {
		return ""
	}
	s = strings.ToLower(s)
	s = unsafeSlugRun.ReplaceAllString(s, "-")
	s = dashRun.ReplaceAllString(s, "-")
	s = strings.Trim(s, "-")
	if len(s) > maxLen {
		s = strings.TrimRight(s[:maxLen], "-")
	}
	return s
}

// shortHashOf returns the first shortHashBytes hex characters of
// sha256("{nodeID}|{nowNanos}"). Using nanoseconds instead of seconds
// minimises collisions when multiple nodes launch within the same second.
func shortHashOf(nodeID string, nowNanos int64) string {
	sum := sha256.Sum256([]byte(nodeID + "|" + strconv.FormatInt(nowNanos, 10)))
	return hex.EncodeToString(sum[:])[:shortHashBytes]
}
