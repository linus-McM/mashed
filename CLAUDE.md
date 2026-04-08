# OpenWolf

@.wolf/OPENWOLF.md

This project uses OpenWolf for context management. Read and follow .wolf/OPENWOLF.md every session. Check .wolf/cerebrum.md before generating code. Check .wolf/anatomy.md before reading files.


# CLAUDE.md

## Skill routing

When the user's request matches an available skill, ALWAYS invoke it using the Skill
tool as your FIRST action. Do NOT answer directly, do NOT use other tools first.
The skill has specialized workflows that produce better results than ad-hoc answers.

## BMAD Artifact System

- `internal/bmad/artifacts.go` — canonical artifact path map (`artifactPaths`), `ResolveArtifactPath(name, repoPath)`, `VerifyArtifacts(repoPath, outputNames)`
- Artifacts resolve to `{repoPath}/_bmad-output/{category}/{artifact}` via the path map
- Unmapped artifacts (`"code"`, `"tests"`, `"any-doc"`) return `""` from ResolveArtifactPath — handled by callers
- `ArtifactType` and `ArtifactSpec` types in `internal/bmad/types.go`
- Shared constant `bmadOutputDir = "_bmad-output"` in `internal/bmad/sprint.go`

