package uiadapter

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"log/slog"
	"strings"
)

//go:embed prompts/classify.md
var stageClassifyPrompt string

//go:embed prompts/generate_yn.md
var stageGenerateYN string

//go:embed prompts/generate_menu.md
var stageGenerateMenu string

//go:embed prompts/generate_form.md
var stageGenerateForm string

//go:embed prompts/generate_text.md
var stageGenerateText string

// StageKind is the stage-1 classification. Mirrors backend.Kind but lives
// in the uiadapter package to avoid an import cycle (backend imports this
// package for UIAST).
type StageKind string

const (
	StageKindYN   StageKind = "yn"
	StageKindMenu StageKind = "menu"
	StageKindForm StageKind = "form"
	StageKindText StageKind = "text"
)

// ErrUnknownKind is returned by ParseStageKind when the classifier returns
// a value outside the enum.
var ErrUnknownKind = errors.New("uiadapter: unknown stage kind")

// ParseStageKind validates a classifier output against the enum. Whitespace
// is trimmed; case-insensitive. Unknown values surface ErrUnknownKind
// wrapped so the router can fall back.
func ParseStageKind(s string) (StageKind, error) {
	norm := strings.ToLower(strings.TrimSpace(s))
	switch StageKind(norm) {
	case StageKindYN, StageKindMenu, StageKindForm, StageKindText:
		return StageKind(norm), nil
	default:
		return "", fmt.Errorf("%w: %q", ErrUnknownKind, s)
	}
}

// ClassifyPrompt returns the byte-identical static prefix for the stage-1
// call. Keeping it byte-stable is a precondition of Ollama's KV prefix
// reuse + Claude's prompt cache (Stories v3-07, v3-14).
func ClassifyPrompt() string { return stageClassifyPrompt }

// GeneratePromptFor returns the static per-kind prompt for stage-2.
func GeneratePromptFor(kind StageKind) (string, error) {
	switch kind {
	case StageKindYN:
		return stageGenerateYN, nil
	case StageKindMenu:
		return stageGenerateMenu, nil
	case StageKindForm:
		return stageGenerateForm, nil
	case StageKindText:
		return stageGenerateText, nil
	default:
		return "", fmt.Errorf("%w: %q", ErrUnknownKind, kind)
	}
}

// AssembleStage1 builds the byte-stable stage-1 user message. Shape matches
// the plan §3 Story 5 construction:
//
//	[static_prefix]\n\n---\nRAW CAPTURE:\n[spotlighted_raw]
//
// Stage-1 does not carry a KIND directive — that's the variable Stage 2
// receives after classification.
func AssembleStage1(rawSpotlighted string) string {
	var b strings.Builder
	b.Grow(len(stageClassifyPrompt) + len(rawSpotlighted) + 32)
	b.WriteString(stageClassifyPrompt)
	b.WriteString("\n\n---\nRAW CAPTURE:\n")
	b.WriteString(rawSpotlighted)
	return b.String()
}

// AssembleStage2 builds the stage-2 user message:
//
//	[kind_prompt]\n\n---\nRAW CAPTURE:\n[spotlighted_raw]\n\n---\nKIND: [kind]\n
//
// The kind_prompt carries the narrow per-kind schema + few-shots, so the
// stage-1 classifier directly gates downstream structure.
func AssembleStage2(kind StageKind, rawSpotlighted string) (string, error) {
	kindPrompt, err := GeneratePromptFor(kind)
	if err != nil {
		return "", err
	}
	var b strings.Builder
	b.Grow(len(kindPrompt) + len(rawSpotlighted) + 64)
	b.WriteString(kindPrompt)
	b.WriteString("\n\n---\nRAW CAPTURE:\n")
	b.WriteString(rawSpotlighted)
	b.WriteString("\n\n---\nKIND: ")
	b.WriteString(string(kind))
	b.WriteString("\n")
	return b.String(), nil
}

// RunTwoStage orchestrates the classify → generate pipeline against any
// function pair. Used by the router and backend wrappers; a test can feed
// fake classify/generate funcs to verify call count and argument shape
// without a real HTTP round-trip.
//
// The classifyFn receives the stage-1 assembled prompt and must return one
// of the StageKind enum values. generateFn receives the kind + stage-2
// assembled prompt and returns a *UIAST.
//
// Returns the final UIAST, the decoded kind, and any error. logger may be
// nil; nilSafeLogger normalises it so any future story can emit telemetry
// without an inline guard.
func RunTwoStage(
	ctx context.Context,
	rawSpotlighted string,
	classifyFn func(ctx context.Context, stage1Prompt string) (StageKind, error),
	generateFn func(ctx context.Context, kind StageKind, stage2Prompt string) (*UIAST, error),
	logger *slog.Logger,
) (*UIAST, StageKind, error) {
	_ = nilSafeLogger(logger)
	if ctx.Err() != nil {
		return nil, "", ctx.Err()
	}
	stage1 := AssembleStage1(rawSpotlighted)
	kind, err := classifyFn(ctx, stage1)
	if err != nil {
		return nil, "", fmt.Errorf("uiadapter: stage-1 classify: %w", err)
	}
	stage2, err := AssembleStage2(kind, rawSpotlighted)
	if err != nil {
		return nil, kind, err
	}
	ast, err := generateFn(ctx, kind, stage2)
	if err != nil {
		return nil, kind, fmt.Errorf("uiadapter: stage-2 generate: %w", err)
	}
	return ast, kind, nil
}
