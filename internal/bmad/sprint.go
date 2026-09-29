package bmad

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"

	"gopkg.in/yaml.v3"
)

// sprintMu serializes concurrent writes to sprint-status.yaml.
var sprintMu sync.Mutex

// StoryStatus represents the status of a sprint story.
type StoryStatus string

const (
	StoryBacklog     StoryStatus = "backlog"
	StoryReadyForDev StoryStatus = "ready-for-dev"
	StoryInProgress  StoryStatus = "in-progress"
	StoryReview      StoryStatus = "review"
	StoryDone        StoryStatus = "done"
	// StorySkip — closed without implementation (merged into another story,
	// retired by course-correction, or scope-deferred). Documented in the
	// sprint-status.yaml header comments and used in the wild.
	StorySkip StoryStatus = "skip"
)

// EpicStatus represents the status of an epic.
type EpicStatus string

const (
	EpicBacklog    EpicStatus = "backlog"
	EpicInProgress EpicStatus = "in-progress"
	EpicDone       EpicStatus = "done"
)

// SprintStory is a single story entry within an epic.
type SprintStory struct {
	ID       string      `json:"id"`
	EpicID   string      `json:"epicId"`
	Status   StoryStatus `json:"status"`
	Sequence int         `json:"sequence"`
}

// SprintEpic groups stories under an epic with its own status.
type SprintEpic struct {
	ID      string        `json:"id"`
	Status  EpicStatus    `json:"status"`
	Stories []SprintStory `json:"stories"`
}

// SprintStatus is the top-level parsed representation of a sprint-status.yaml file.
type SprintStatus struct {
	Generated      string       `json:"generated"`
	LastUpdated    string       `json:"lastUpdated"`
	Project        string       `json:"project"`
	ProjectKey     string       `json:"projectKey"`
	TrackingSystem string       `json:"trackingSystem"`
	StoryLocation  string       `json:"storyLocation"`
	Epics          []SprintEpic `json:"epics"`
}

// sprintStatusYAML is the intermediate struct for YAML unmarshalling.
// The DevelopmentStatus field uses yaml.Node to preserve key ordering.
type sprintStatusYAML struct {
	Generated         string    `yaml:"generated"`
	LastUpdated       string    `yaml:"last_updated"`
	Project           string    `yaml:"project"`
	ProjectKey        string    `yaml:"project_key"`
	TrackingSystem    string    `yaml:"tracking_system"`
	StoryLocation     string    `yaml:"story_location"`
	DevelopmentStatus yaml.Node `yaml:"development_status"`
}

// epicKeyPrefix matches keys that start with "epic-".
const epicKeyPrefix = "epic-"

// storyKeyPattern matches story keys like "1-2-dashboard" or "2-1-notifications".
var storyKeyPattern = regexp.MustCompile(`^\d+-\d+-.+$`)

// retrospectiveSuffix matches keys ending with "-retrospective".
const retrospectiveSuffix = "-retrospective"

// sprintStatusPath returns the expected filesystem path to the sprint-status.yaml
// given a repository root.
func sprintStatusPath(repoPath string) string {
	return filepath.Join(repoPath, bmadOutputDir, "implementation-artifacts", "sprint-status.yaml")
}

// ParseSprintStatus reads and parses the sprint-status.yaml file from the
// standard location within the given repository path.
func ParseSprintStatus(repoPath string) (SprintStatus, error) {
	path := sprintStatusPath(repoPath)

	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return SprintStatus{}, fmt.Errorf("%w: %s", ErrSprintFileNotFound, path)
		}
		return SprintStatus{}, fmt.Errorf("bmad: reading sprint file: %w", err)
	}

	var raw sprintStatusYAML
	if err := yaml.Unmarshal(data, &raw); err != nil {
		return SprintStatus{}, fmt.Errorf("%w: %w", ErrSprintFileMalformed, err)
	}

	epics, err := parseDevelopmentStatus(&raw.DevelopmentStatus)
	if err != nil {
		return SprintStatus{}, err
	}

	return SprintStatus{
		Generated:      raw.Generated,
		LastUpdated:    raw.LastUpdated,
		Project:        raw.Project,
		ProjectKey:     raw.ProjectKey,
		TrackingSystem: raw.TrackingSystem,
		StoryLocation:  raw.StoryLocation,
		Epics:          epics,
	}, nil
}

// parseDevelopmentStatus walks the yaml.Node MappingNode pairwise, grouping
// story entries under the most recently encountered epic entry.
func parseDevelopmentStatus(node *yaml.Node) ([]SprintEpic, error) {
	// An empty or null development_status is valid — just no epics.
	if node == nil || node.Kind == 0 || node.Tag == "!!null" {
		return nil, nil
	}

	if node.Kind != yaml.MappingNode {
		return nil, fmt.Errorf("%w: development_status must be a mapping, got kind %d", ErrSprintFileMalformed, node.Kind)
	}

	var epics []SprintEpic
	var currentEpic *SprintEpic
	storySeq := 0

	for i := 0; i+1 < len(node.Content); i += 2 {
		keyNode := node.Content[i]
		valNode := node.Content[i+1]

		key := keyNode.Value
		val := valNode.Value

		// Check retrospective suffix before epic prefix — "epic-3-retrospective" is a story, not an epic
		if strings.HasSuffix(key, retrospectiveSuffix) {
			if currentEpic == nil {
				return nil, fmt.Errorf("%w: story %q appears before any epic", ErrSprintFileMalformed, key)
			}
			status, ok := ValidateStoryStatus(val)
			if !ok {
				// Retrospectives may have non-standard statuses like "optional" — treat as backlog
				status = StoryBacklog
			}
			currentEpic.Stories = append(currentEpic.Stories, SprintStory{
				ID:       key,
				EpicID:   currentEpic.ID,
				Status:   status,
				Sequence: storySeq,
			})
			storySeq++
			continue
		}

		if strings.HasPrefix(key, epicKeyPrefix) {
			// Flush the current epic before starting a new one.
			if currentEpic != nil {
				epics = append(epics, *currentEpic)
			}

			status, ok := ValidateEpicStatus(val)
			if !ok {
				return nil, fmt.Errorf("%w: invalid epic status %q for %s", ErrSprintFileMalformed, val, key)
			}

			currentEpic = &SprintEpic{
				ID:      key,
				Status:  status,
				Stories: []SprintStory{},
			}
			storySeq = 0
			continue
		}

		// Story entry: matches digit pattern (retrospective keys are handled above).
		if storyKeyPattern.MatchString(key) {
			if currentEpic == nil {
				return nil, fmt.Errorf("%w: story %q appears before any epic", ErrSprintFileMalformed, key)
			}

			status, ok := ValidateStoryStatus(val)
			if !ok {
				return nil, fmt.Errorf("%w: invalid story status %q for %s", ErrSprintFileMalformed, val, key)
			}

			currentEpic.Stories = append(currentEpic.Stories, SprintStory{
				ID:       key,
				EpicID:   currentEpic.ID,
				Status:   status,
				Sequence: storySeq,
			})
			storySeq++
			continue
		}

		// Unknown key format — skip silently for forward compatibility.
	}

	// Flush the last epic.
	if currentEpic != nil {
		epics = append(epics, *currentEpic)
	}

	return epics, nil
}

// ValidateStoryStatus checks whether the given string is a valid StoryStatus.
func ValidateStoryStatus(s string) (StoryStatus, bool) {
	switch StoryStatus(s) {
	case StoryBacklog, StoryReadyForDev, StoryInProgress, StoryReview, StoryDone, StorySkip:
		return StoryStatus(s), true
	default:
		return "", false
	}
}

// ValidateEpicStatus checks whether the given string is a valid EpicStatus.
func ValidateEpicStatus(s string) (EpicStatus, bool) {
	switch EpicStatus(s) {
	case EpicBacklog, EpicInProgress, EpicDone:
		return EpicStatus(s), true
	default:
		return "", false
	}
}

// UpdateStoryStatus reads sprint-status.yaml, modifies the target story's
// status in-place using yaml.Node round-trip to preserve formatting and
// comments, and writes it back atomically.
func UpdateStoryStatus(repoPath, storyID, newStatus string) error {
	// 1. Validate inputs.
	if !storyKeyPattern.MatchString(storyID) {
		return fmt.Errorf("bmad: invalid story ID %q: must match pattern digit-digit-slug", storyID)
	}
	if _, ok := ValidateStoryStatus(newStatus); !ok {
		return fmt.Errorf("bmad: invalid status %q for story %s", newStatus, storyID)
	}

	// 2. Read the YAML file.
	path := sprintStatusPath(repoPath)
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("%w: %s", ErrSprintFileNotFound, path)
		}
		return fmt.Errorf("bmad: reading sprint file: %w", err)
	}

	// 3. Unmarshal into yaml.Node to preserve formatting/comments.
	var doc yaml.Node
	if err := yaml.Unmarshal(data, &doc); err != nil {
		return fmt.Errorf("%w: %w", ErrSprintFileMalformed, err)
	}

	// doc is a Document node; doc.Content[0] is the root mapping.
	if doc.Kind != yaml.DocumentNode || len(doc.Content) == 0 {
		return fmt.Errorf("%w: unexpected YAML structure", ErrSprintFileMalformed)
	}
	root := doc.Content[0]
	if root.Kind != yaml.MappingNode {
		return fmt.Errorf("%w: root is not a mapping", ErrSprintFileMalformed)
	}

	// 4. Find the development_status mapping node.
	var devStatus *yaml.Node
	for i := 0; i+1 < len(root.Content); i += 2 {
		if root.Content[i].Value == "development_status" {
			devStatus = root.Content[i+1]
			break
		}
	}
	if devStatus == nil || devStatus.Kind != yaml.MappingNode {
		return fmt.Errorf("%w: %s", ErrStoryNotFound, storyID)
	}

	// 5. Walk development_status pairwise to find the story key.
	found := false
	for i := 0; i+1 < len(devStatus.Content); i += 2 {
		if devStatus.Content[i].Value == storyID {
			devStatus.Content[i+1].Value = newStatus
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("%w: %s", ErrStoryNotFound, storyID)
	}

	// 6. Marshal back and write atomically.
	out, err := yaml.Marshal(&doc)
	if err != nil {
		return fmt.Errorf("bmad: marshaling sprint status: %w", err)
	}

	sprintMu.Lock()
	defer sprintMu.Unlock()

	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, out, 0644); err != nil {
		return fmt.Errorf("bmad: writing temp file: %w", err)
	}
	if err := os.Rename(tmp, path); err != nil {
		os.Remove(tmp)
		return fmt.Errorf("bmad: renaming temp file: %w", err)
	}

	return nil
}
