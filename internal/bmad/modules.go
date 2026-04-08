package bmad

// ModuleDef describes a BMAD module — a named group of processes.
type ModuleDef struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Version     string   `json:"version"`
	Processes   []string `json:"processes"`
	UpgradePath string   `json:"upgradePath,omitempty"`
}

// moduleSpecs defines the static metadata for each BMAD module.
// Process lists are populated dynamically from the registry.
var moduleSpecs = []ModuleDef{
	{ID: "core", Name: "Core BMAD", Version: "1.0.0"},
	{ID: "bmb", Name: "BMAD Module Builder", Version: "1.0.0"},
	{ID: "tea", Name: "Testing & Engineering Automation", Version: "1.0.0"},
	{ID: "bmgd", Name: "BMAD Game Design", Version: "1.0.0"},
	{ID: "cis", Name: "CI/CD & Infrastructure", Version: "1.0.0"},
}

// GetModules returns all available BMAD modules with their process lists
// dynamically populated from the registry.
func GetModules() []ModuleDef {
	modules := make([]ModuleDef, len(moduleSpecs))
	for i, spec := range moduleSpecs {
		procs := ProcessesByModule(spec.ID)
		ids := make([]string, len(procs))
		for j, p := range procs {
			ids[j] = p.ID
		}
		modules[i] = ModuleDef{
			ID:          spec.ID,
			Name:        spec.Name,
			Version:     spec.Version,
			Processes:   ids,
			UpgradePath: spec.UpgradePath,
		}
	}
	return modules
}
