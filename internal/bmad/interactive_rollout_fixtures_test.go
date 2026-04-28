// Stable autonomous BMAD process IDs that remain autonomous across the entire
// rollout (plan §"Out of scope"). Tests requiring predictable autonomous
// behaviour reference these constants so the rollout phases don't churn the
// fixture sites every time another process flips to interactive.
package bmad

const (
	// autonomousProcessFixtureID is the canonical stable autonomous process
	// ID for tests that need a single autonomous fixture.
	autonomousProcessFixtureID = "bmad-sprint-status"

	// autonomousProcessFixtureIDB is a second stable autonomous ID for tests
	// that need TWO distinct autonomous IDs in the same graph (e.g.
	// saveThreeNodeWorkflow, branching tests).
	autonomousProcessFixtureIDB = "bmad-sprint-planning"
)
