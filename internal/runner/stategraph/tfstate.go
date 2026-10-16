package stategraph

// Minimal structs to decode terraform.tfstate we care about.

type State struct {
	Version          int        `json:"version"`
	TerraformVersion string     `json:"terraform_version"`
	Resources        []Resource `json:"resources"`
}

type Resource struct {
	Module    string     `json:"module,omitempty"`
	Mode      string     `json:"mode"` // managed|data
	Type      string     `json:"type"`
	Name      string     `json:"name"`
	Provider  string     `json:"provider"`
	Instances []Instance `json:"instances"`
}

type Instance struct {
	IndexKey            any               `json:"index_key,omitempty"`
	Dependencies        []string          `json:"dependencies,omitempty"`
	Attributes          map[string]any    `json:"attributes,omitempty"`
	AttrsFlat           map[string]string `json:"attributes_flat,omitempty"`
	SensitiveAttributes any               `json:"sensitive_attributes,omitempty"`
}
