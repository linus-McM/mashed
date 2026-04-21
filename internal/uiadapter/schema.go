package uiadapter

import (
	"bytes"
	"encoding/json"
)

// UIAST is the root envelope the translator produces (spec §3.1).
type UIAST struct {
	Version             string      `json:"version"`
	GeneratedBy         string      `json:"generated_by"`
	GeneratedAt         int64       `json:"generated_at"`
	TurnSummary         string      `json:"turn_summary"`
	Nodes               []UINode    `json:"nodes"`
	FallbackAnswerShape string      `json:"fallback_answer_shape"`
	Diagnostics         Diagnostics `json:"diagnostics"`
}

// Diagnostics carries counters + trust signals for U4's round loop (spec §3.1).
// CancelReason is deliberately never serialised — it exists only for the
// in-process signalling path between adapter.Translate and executor.step.
type Diagnostics struct {
	InputBytes      int      `json:"input_bytes,omitempty"`
	OutputBytes     int      `json:"output_bytes,omitempty"`
	LatencyMs       int      `json:"latency_ms,omitempty"`
	FallbackReasons []string `json:"fallback_reasons,omitempty"`
	Untrusted       bool     `json:"untrusted,omitempty"`
	Collapsed       bool     `json:"collapsed,omitempty"`
	CancelReason    string   `json:"-"`
}

// UINode is the single-struct discriminated union for every node variant
// (spec §3.2). Per the spec, variant-specific fields carry omitempty so the
// wire shape only shows fields relevant to the node's Type.
type UINode struct {
	Type        string      `json:"type"`
	Content     string      `json:"content,omitempty"`
	Tone        string      `json:"tone,omitempty"`
	Heading     string      `json:"heading,omitempty"`
	Bullets     []string    `json:"bullets,omitempty"`
	Lang        string      `json:"lang,omitempty"`
	Copyable    bool        `json:"copyable,omitempty"`
	Columns     []string    `json:"columns,omitempty"`
	Rows        [][]string  `json:"rows,omitempty"`
	Prompt      string      `json:"prompt,omitempty"`
	Help        string      `json:"help,omitempty"`
	Required    bool        `json:"required,omitempty"`
	Widget      *WidgetNode `json:"widget,omitempty"`
	ResponseKey string      `json:"response_key,omitempty"`
}

// WidgetNode carries the input-control metadata for decision_group nodes
// (spec §3.2). repoRootRelative / maxLength are camelCase on purpose —
// those names are round-tripped through the executor snapshot.
type WidgetNode struct {
	Type        string          `json:"type"`
	Options     []WidgetOption  `json:"options,omitempty"`
	Default     string          `json:"default,omitempty"`
	Min         int             `json:"min,omitempty"`
	Max         int             `json:"max,omitempty"`
	YesLabel    string          `json:"yes_label,omitempty"`
	NoLabel     string          `json:"no_label,omitempty"`
	Placeholder string          `json:"placeholder,omitempty"`
	MaxLength   int             `json:"maxLength,omitempty"`
	Multiline   bool            `json:"multiline,omitempty"`
	Accept      []string        `json:"accept,omitempty"`
	RepoRootRel bool            `json:"repoRootRelative,omitempty"`
	Schema      json.RawMessage `json:"schema,omitempty"`
}

// WidgetOption is a single choice entry. Label is optional — a bare value
// serialises to {"value":"..."} with no label key.
type WidgetOption struct {
	Value string `json:"value"`
	Label string `json:"label,omitempty"`
}

// UnmarshalJSON enforces strict decoding for widget fields (spec §4.7.1):
// envelope/node are permissive (forward-compat), but a widget with an unknown
// field is a validation error — widget semantics are tightly coupled to the
// enumerated type set, so silently ignoring unknown keys would mask model
// drift.
func (w *WidgetNode) UnmarshalJSON(b []byte) error {
	type widgetAlias WidgetNode
	dec := json.NewDecoder(bytes.NewReader(b))
	dec.DisallowUnknownFields()
	var alias widgetAlias
	if err := dec.Decode(&alias); err != nil {
		return err
	}
	*w = WidgetNode(alias)
	return nil
}
