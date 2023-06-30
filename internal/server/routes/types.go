package routes

type Layer struct {
	Id                 string     `json:"id"`
	Name               string     `json:"name"`
	Namespace          string     `json:"namespace"`
	RepoUrl            string     `json:"repoUrl"`
	Branch             string     `json:"branch"`
	Path               string     `json:"path"`
	Status             string     `json:"status"`
	LastPlanCommit     string     `json:"lastPlanCommit"`
	LastApplyCommit    string     `json:"lastApplyCommit"`
	LastRelevantCommit string     `json:"lastRelevantCommit"`
	Resources          []Resource `json:"resources"`
}

type Resource struct {
	Address   string   `json:"address"`
	Type      string   `json:"type"`
	DependsOn []string `json:"dependsOn"`
	Status    string   `json:"status"`
}
