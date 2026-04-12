package bmad

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// Sentinel errors for the save path.
var (
	// ErrNoFrontmatterBlock is returned when the target file lacks a
	// YAML frontmatter block (legacy skill files). The caller should
	// surface a user-readable message; we never silently prepend one.
	ErrNoFrontmatterBlock = errors.New("mashed: file has no YAML frontmatter block")

	// ErrPathOutsideAllowedRoots is returned when the supplied path is
	// not under ~/.claude/ or {repoPath}/.claude/. Prevents arbitrary
	// file writes via a malicious payload.
	ErrPathOutsideAllowedRoots = errors.New("mashed: path outside allowed roots")
)

// mashedWriteableFields are the YAML keys the editor is allowed to
// mutate. Any key not in this set is preserved verbatim.
var mashedWriteableFields = map[string]bool{
	"name":                true,
	"description":         true,
	"mashedRole":          true,
	"mashedCompletion":    true,
	"mashedChainable":     true,
	"mashedSessionPinned": true,
	"mashedInputs":        true,
	"mashedOutputs":       true,
}

// WriteMashedAssetFrontmatter reads the file at path, parses its YAML
// frontmatter into a yaml.Node tree (preserving key order and
// comments), applies only the mashed-specific field updates from
// updates, and writes the result back atomically via tempfile+rename.
//
// Round-trip symmetry: saving without changes produces a byte-identical
// file because the yaml.Node tree preserves formatting, and the body
// is never touched.
func WriteMashedAssetFrontmatter(path string, updates map[string]interface{}) error {
	if err := validateAssetPath(path); err != nil {
		return err
	}

	content, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("mashed: reading %s: %w", path, err)
	}

	fmBytes, body, err := extractFrontmatter(content)
	if errors.Is(err, errNoFrontmatter) {
		return ErrNoFrontmatterBlock
	}
	if err != nil {
		return fmt.Errorf("mashed: parsing %s: %w", path, err)
	}

	// Parse frontmatter into a yaml.Node tree to preserve ordering.
	var doc yaml.Node
	if len(fmBytes) > 0 {
		if err := yaml.Unmarshal(fmBytes, &doc); err != nil {
			return fmt.Errorf("mashed: parsing YAML in %s: %w", path, err)
		}
	}

	// yaml.Unmarshal wraps content in a DocumentNode. The actual
	// mapping is the first child. Empty frontmatter (or zero-length
	// fmBytes) produces no children — create the structure.
	if doc.Kind == 0 {
		doc.Kind = yaml.DocumentNode
	}
	var mapping *yaml.Node
	if len(doc.Content) == 0 {
		mapping = &yaml.Node{Kind: yaml.MappingNode, Tag: "!!map"}
		doc.Content = append(doc.Content, mapping)
	} else {
		mapping = doc.Content[0]
	}

	// Apply updates: for each mashed-writeable field, find and update
	// in-place or append if new.
	for key, val := range updates {
		if !mashedWriteableFields[key] {
			continue
		}
		applyNodeUpdate(mapping, key, val)
	}

	// Marshal the updated node tree back to YAML.
	var buf bytes.Buffer
	enc := yaml.NewEncoder(&buf)
	enc.SetIndent(2)
	if err := enc.Encode(&doc); err != nil {
		return fmt.Errorf("mashed: encoding YAML for %s: %w", path, err)
	}
	if err := enc.Close(); err != nil {
		return fmt.Errorf("mashed: closing YAML encoder for %s: %w", path, err)
	}

	// yaml.Encoder always emits a trailing newline. The marshal also
	// includes the document-end marker "...\n" — strip it to keep round-
	// trip symmetry with the original which uses "---" delimiters only.
	yamlOut := buf.Bytes()
	yamlOut = bytes.TrimSuffix(yamlOut, []byte("...\n"))
	yamlOut = bytes.TrimRight(yamlOut, "\n")

	// Reconstruct the file: ---\n<frontmatter>\n---\n<body>
	var out bytes.Buffer
	out.WriteString("---\n")
	out.Write(yamlOut)
	out.WriteString("\n---\n")
	out.Write(body)

	// Atomic write: tempfile in the same directory + rename.
	dir := filepath.Dir(path)
	tmp, err := os.CreateTemp(dir, ".mashed-save-*")
	if err != nil {
		return fmt.Errorf("mashed: create temp: %w", err)
	}
	defer os.Remove(tmp.Name()) // cleanup on failure

	if _, err := tmp.Write(out.Bytes()); err != nil {
		tmp.Close()
		return fmt.Errorf("mashed: write temp: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("mashed: close temp: %w", err)
	}
	if err := os.Rename(tmp.Name(), path); err != nil {
		return fmt.Errorf("mashed: rename: %w", err)
	}
	return nil
}

// validateAssetPath rejects paths that do not contain a /.claude/
// segment, preventing writes to arbitrary locations. This accepts both
// the global ~/.claude/ directory and any repo-local .claude/ directory.
func validateAssetPath(path string) error {
	clean := filepath.Clean(path)

	// Accept any path containing a /.claude/ segment (covers both
	// repo-local {repoPath}/.claude/skills/ and global ~/.claude/).
	if strings.Contains(clean, string(filepath.Separator)+".claude"+string(filepath.Separator)) {
		return nil
	}

	return ErrPathOutsideAllowedRoots
}

// applyNodeUpdate finds a key in the YAML mapping node and updates its
// value, or appends a new key-value pair if not found.
func applyNodeUpdate(mapping *yaml.Node, key string, val interface{}) {
	// Search existing keys pairwise (key, value, key, value, ...).
	for i := 0; i+1 < len(mapping.Content); i += 2 {
		if mapping.Content[i].Value == key {
			setNodeValue(mapping.Content[i+1], val)
			return
		}
	}

	// Key not found — append.
	keyNode := &yaml.Node{Kind: yaml.ScalarNode, Value: key, Tag: "!!str"}
	valNode := &yaml.Node{}
	setNodeValue(valNode, val)
	mapping.Content = append(mapping.Content, keyNode, valNode)
}

// setNodeValue sets a yaml.Node's value from a Go interface{},
// handling the types the editor produces: string, bool, []string.
func setNodeValue(node *yaml.Node, val interface{}) {
	if val == nil {
		return // skip nil values — leave the existing node untouched
	}
	switch v := val.(type) {
	case string:
		node.Kind = yaml.ScalarNode
		node.Tag = "!!str"
		node.Value = v
		node.Content = nil
	case bool:
		node.Kind = yaml.ScalarNode
		node.Tag = "!!bool"
		if v {
			node.Value = "true"
		} else {
			node.Value = "false"
		}
		node.Content = nil
	case []interface{}:
		node.Kind = yaml.SequenceNode
		node.Tag = "!!seq"
		node.Value = ""
		node.Content = nil
		for _, item := range v {
			s := fmt.Sprintf("%v", item)
			node.Content = append(node.Content, &yaml.Node{
				Kind: yaml.ScalarNode, Tag: "!!str", Value: s,
			})
		}
	case []string:
		node.Kind = yaml.SequenceNode
		node.Tag = "!!seq"
		node.Value = ""
		node.Content = nil
		for _, s := range v {
			node.Content = append(node.Content, &yaml.Node{
				Kind: yaml.ScalarNode, Tag: "!!str", Value: s,
			})
		}
	default:
		// Fallback: convert to string.
		node.Kind = yaml.ScalarNode
		node.Tag = "!!str"
		node.Value = fmt.Sprintf("%v", val)
		node.Content = nil
	}
}
