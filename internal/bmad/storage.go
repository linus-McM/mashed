package bmad

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
)

var validID = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)

// Storage provides file-based persistence for workflows and agent configs.
type Storage struct {
	workflowDir string
	agentDir    string
	mu          sync.RWMutex
}

// NewStorage creates a Storage rooted at baseDir, creating subdirectories if needed.
func NewStorage(baseDir string) (*Storage, error) {
	wfDir := filepath.Join(baseDir, "workflows")
	agDir := filepath.Join(baseDir, "bmad-agents")
	if err := os.MkdirAll(wfDir, 0755); err != nil {
		return nil, fmt.Errorf("creating workflow dir: %w", err)
	}
	if err := os.MkdirAll(agDir, 0755); err != nil {
		return nil, fmt.Errorf("creating agent dir: %w", err)
	}
	return &Storage{workflowDir: wfDir, agentDir: agDir}, nil
}

func validateID(id string) error {
	if id == "" || !validID.MatchString(id) {
		return ErrInvalidID
	}
	return nil
}

// SaveWorkflow writes a WorkflowDef to disk using an atomic rename.
func (s *Storage) SaveWorkflow(wf WorkflowDef) error {
	if err := validateID(wf.ID); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	return atomicWriteJSON(filepath.Join(s.workflowDir, wf.ID+".json"), wf)
}

// LoadWorkflow reads a single workflow by ID.
func (s *Storage) LoadWorkflow(id string) (WorkflowDef, error) {
	if err := validateID(id); err != nil {
		return WorkflowDef{}, err
	}
	s.mu.RLock()
	defer s.mu.RUnlock()

	data, err := os.ReadFile(filepath.Join(s.workflowDir, id+".json"))
	if os.IsNotExist(err) {
		return WorkflowDef{}, ErrWorkflowNotFound
	}
	if err != nil {
		return WorkflowDef{}, fmt.Errorf("reading workflow %s: %w", id, err)
	}
	var wf WorkflowDef
	if err := json.Unmarshal(data, &wf); err != nil {
		return WorkflowDef{}, fmt.Errorf("decoding workflow %s: %w", id, err)
	}
	return wf, nil
}

// ListWorkflows returns all valid workflows on disk, skipping malformed files.
func (s *Storage) ListWorkflows() ([]WorkflowDef, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	entries, err := os.ReadDir(s.workflowDir)
	if err != nil {
		return nil, fmt.Errorf("listing workflows: %w", err)
	}
	var out []WorkflowDef
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".json") {
			continue
		}
		data, err := os.ReadFile(filepath.Join(s.workflowDir, e.Name()))
		if err != nil {
			log.Printf("bmad: skipping unreadable workflow %s: %v", e.Name(), err)
			continue
		}
		var wf WorkflowDef
		if err := json.Unmarshal(data, &wf); err != nil {
			log.Printf("bmad: skipping malformed workflow %s: %v", e.Name(), err)
			continue
		}
		out = append(out, wf)
	}
	return out, nil
}

// ListWorkflowsByRepo returns workflows whose RepoPath matches the given path.
// Trailing slashes are normalized before comparison.
func (s *Storage) ListWorkflowsByRepo(repoPath string) ([]WorkflowDef, error) {
	all, err := s.ListWorkflows()
	if err != nil {
		return nil, fmt.Errorf("listing workflows by repo: %w", err)
	}
	repoPath = strings.TrimRight(repoPath, "/")
	var filtered []WorkflowDef
	for _, wf := range all {
		if strings.TrimRight(wf.RepoPath, "/") == repoPath {
			filtered = append(filtered, wf)
		}
	}
	return filtered, nil
}

// DeleteWorkflow removes a workflow file by ID.
func (s *Storage) DeleteWorkflow(id string) error {
	if err := validateID(id); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	path := filepath.Join(s.workflowDir, id+".json")
	if err := os.Remove(path); os.IsNotExist(err) {
		return ErrWorkflowNotFound
	} else if err != nil {
		return fmt.Errorf("deleting workflow %s: %w", id, err)
	}
	return nil
}

// SaveAgent writes a BmadAgentConfig to disk using an atomic rename.
func (s *Storage) SaveAgent(agent BmadAgentConfig) error {
	if err := validateID(agent.ID); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	return atomicWriteJSON(filepath.Join(s.agentDir, agent.ID+".json"), agent)
}

// ListAgents returns all valid agent configs on disk.
func (s *Storage) ListAgents() ([]BmadAgentConfig, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	entries, err := os.ReadDir(s.agentDir)
	if err != nil {
		return nil, fmt.Errorf("listing agents: %w", err)
	}
	var out []BmadAgentConfig
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".json") {
			continue
		}
		data, err := os.ReadFile(filepath.Join(s.agentDir, e.Name()))
		if err != nil {
			log.Printf("bmad: skipping unreadable agent %s: %v", e.Name(), err)
			continue
		}
		var ag BmadAgentConfig
		if err := json.Unmarshal(data, &ag); err != nil {
			log.Printf("bmad: skipping malformed agent %s: %v", e.Name(), err)
			continue
		}
		out = append(out, ag)
	}
	return out, nil
}

// DeleteAgent removes an agent config file by ID.
func (s *Storage) DeleteAgent(id string) error {
	if err := validateID(id); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	path := filepath.Join(s.agentDir, id+".json")
	if err := os.Remove(path); os.IsNotExist(err) {
		return ErrAgentNotFound
	} else if err != nil {
		return fmt.Errorf("deleting agent %s: %w", id, err)
	}
	return nil
}

// ListClaudeAgents scans a .claude/agents/ directory and returns AgentInfo
// entries for each .md file found. Returns nil (not error) if the directory
// does not exist.
func ListClaudeAgents(dir string) []AgentInfo {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}
	var out []AgentInfo
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".md") {
			continue
		}
		id := strings.TrimSuffix(e.Name(), ".md")
		name := strings.ReplaceAll(id, "-", " ")
		// Title-case each word
		words := strings.Fields(name)
		for i, w := range words {
			if len(w) > 0 {
				words[i] = strings.ToUpper(w[:1]) + w[1:]
			}
		}
		out = append(out, AgentInfo{
			ID:   id,
			Name: strings.Join(words, " "),
		})
	}
	return out
}

// atomicWriteJSON marshals v as indented JSON and writes it atomically.
func atomicWriteJSON(path string, v any) error {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling JSON: %w", err)
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0644); err != nil {
		return fmt.Errorf("writing temp file: %w", err)
	}
	if err := os.Rename(tmp, path); err != nil {
		os.Remove(tmp)
		return fmt.Errorf("renaming temp file: %w", err)
	}
	return nil
}
