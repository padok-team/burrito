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
}
