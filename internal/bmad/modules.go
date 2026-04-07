package bmad

// ModuleDef describes a BMAD module — a named group of processes.
type ModuleDef struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Version     string   `json:"version"`
	Processes   []string `json:"processes"`
	UpgradePath string   `json:"upgradePath,omitempty"`
}

// GetModules returns all available BMAD modules.
func GetModules() []ModuleDef {
	// Build core process IDs from the live registry.
	coreProcs := ProcessesByModule("core")
	coreIDs := make([]string, len(coreProcs))
	for i, p := range coreProcs {
		coreIDs[i] = p.ID
	}

	return []ModuleDef{
		{ID: "core", Name: "Core BMAD", Version: "1.0.0", Processes: coreIDs},
		{ID: "bmb", Name: "BMAD Module Builder", Version: "1.0.0", Processes: []string{}},
		{ID: "tea", Name: "Testing & Engineering Automation", Version: "1.0.0", Processes: []string{}},
		{ID: "bmgd", Name: "BMAD Game Design", Version: "1.0.0", Processes: []string{}},
		{ID: "cis", Name: "CI/CD & Infrastructure", Version: "1.0.0", Processes: []string{}},
	}
}
