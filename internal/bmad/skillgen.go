package bmad

import (
	"fmt"
	"os"
	"path/filepath"
	"text/template"
)

// skillTemplateData holds the resolved data for rendering a SKILL.md file.
type skillTemplateData struct {
	SkillName   string
	Name        string
	Description string
	InputSpecs  []artifactPathSpec
	OutputSpecs []artifactPathSpec
}

// artifactPathSpec pairs an artifact name with its resolved path under _bmad-output.
// Path is empty for unmapped artifacts (e.g. "code", "tests").
type artifactPathSpec struct {
	Name string
	Path string
}

var skillTemplate = template.Must(template.New("skill").Parse(`---
name: {{ .SkillName }}
description: {{ .Description }}
---

# {{ .Name }}

## What You Do
{{ .Description }}

## Inputs
{{- if .InputSpecs }}
{{- range .InputSpecs }}
- Read ` + "`{{ .Path }}`" + ` for {{ .Name }}
{{- end }}
{{- else }}
- No file inputs required
{{- end }}

## Outputs
{{- if .OutputSpecs }}
{{- range .OutputSpecs }}
- Write ` + "`{{ .Path }}`" + ` — {{ .Name }}
{{- end }}
{{- else }}
- No file outputs (operates on working directory)
{{- end }}

## Instructions
1. Read any input artifacts listed above.
2. Perform the task described in "What You Do."
3. Write all output artifacts to the exact paths listed above.
4. All artifacts go under ` + "`_bmad-output/`" + ` in the repository root.

## Output Convention
Write all artifacts to ` + "`_bmad-output/`" + ` using the exact paths above.
`))

// GenerateSkillFiles creates a SKILL.md file for every registered BMAD process.
// Each file is written to baseDir/{skillName}/SKILL.md. The function is idempotent:
// calling it multiple times with the same baseDir overwrites existing files without error.
func GenerateSkillFiles(baseDir string) error {
	for _, proc := range AllProcesses() {
		data := buildSkillData(proc)

		dir := filepath.Join(baseDir, proc.SkillName)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("create skill dir %s: %w", dir, err)
		}

		path := filepath.Join(dir, "SKILL.md")
		f, err := os.Create(path)
		if err != nil {
			return fmt.Errorf("create SKILL.md for %s: %w", proc.SkillName, err)
		}

		if err := skillTemplate.Execute(f, data); err != nil {
			f.Close()
			return fmt.Errorf("render SKILL.md for %s: %w", proc.SkillName, err)
		}
		if err := f.Close(); err != nil {
			return fmt.Errorf("close SKILL.md for %s: %w", proc.SkillName, err)
		}
	}
	return nil
}

// buildSkillData resolves artifact paths for a process definition into template data.
func buildSkillData(proc ProcessDef) skillTemplateData {
	data := skillTemplateData{
		SkillName:   proc.SkillName,
		Name:        proc.Name,
		Description: proc.Description,
	}

	for _, input := range proc.Inputs {
		resolved := ResolveArtifactPath(input, "")
		if resolved != "" {
			data.InputSpecs = append(data.InputSpecs, artifactPathSpec{Name: input, Path: resolved})
		}
	}

	for _, output := range proc.Outputs {
		resolved := ResolveArtifactPath(output, "")
		if resolved != "" {
			data.OutputSpecs = append(data.OutputSpecs, artifactPathSpec{Name: output, Path: resolved})
		}
	}

	return data
}
