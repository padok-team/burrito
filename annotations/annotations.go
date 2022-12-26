package annotations

const (
	LastApplySum    string = "runner.terraform.padok.cloud/apply-sum"
	LastApplyDate   string = "runner.terraform.padok.cloud/apply-date"
	LastApplyCommit string = "runner.terraform.padok.cloud/apply-commit"
	LastPlanCommit  string = "runner.terraform.padok.cloud/plan-commit"
	LastPlanDate    string = "runner.terraform.padok.cloud/plan-date"
	LastPlanSum     string = "runner.terraform.padok.cloud/plan-sum"
	Failure         string = "runner.terraform.padok.cloud/failure"

	LastBranchCommit string = "notifications.terraform.padok.cloud/branch-commit"
	ForceApply       string = "notifications.terraform.padok.cloud/force-apply"
)
