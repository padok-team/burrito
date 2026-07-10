package status

type Phase string

const (
	PhasePlan  Phase = "plan"
	PhaseApply Phase = "apply"
)

type State string

const (
	StatePending State = "pending"
	StateSuccess State = "success"
	StateFailure State = "failure"
)

type CommitStatus struct {
	Phase       Phase
	State       State
	Description string
	// Commit overrides which commit the status is posted on. When empty,
	// providers fall back to the pull request's LastBranchCommit annotation.
	Commit string
	// Context overrides the default "burrito/<phase>" context/name used to identify the
	// status. Needed when several independent statuses target the same commit, e.g. one
	// per layer for commits pushed directly to the base branch, where there is no pull
	// request to aggregate them under a single context.
	Context string
}
