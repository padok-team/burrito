package status

type Phase string

const (
	PhasePlan  Phase = "plan"
	PhaseApply Phase = "apply"
)

type State string

const (
	StatePending State = "pending"
	// StateRunning marks a status as actively in progress, distinct from StatePending. GitLab's
	// commit status is a real state machine (pending -> running -> success/failed) and rejects
	// posting "pending" twice in a row (e.g. once when a run is decided, again once its pod
	// starts) with a 400 ("Cannot transition status via :enqueue from :pending"). GitHub has no
	// such restriction and no "running" state, so its provider maps this back to "pending".
	StateRunning State = "running"
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
	// TargetURL, if set, becomes the status's "Details" link.
	TargetURL string
}
