package comment

import (
	"bytes"
	"text/template"

	configv1alpha1 "github.com/padok-team/burrito/api/v1alpha1"
	"github.com/padok-team/burrito/internal/storage"

	_ "embed"
)

var (
	//go:embed templates/comment.md
	defaultTemplateRaw string
	defaultTemplate    = template.Must(template.New("report").Parse(defaultTemplateRaw))
)

type ReportedLayer struct {
	ShortDiff  string
	Path       string
	PrettyPlan string
}

type DefaultComment struct {
	layers  []configv1alpha1.TerraformLayer
	storage storage.Storage
}

type DefaultCommentInput struct {
}

func NewDefaultComment(layers []configv1alpha1.TerraformLayer, storage storage.Storage) *DefaultComment {
	return &DefaultComment{
		layers:  layers,
		storage: storage,
	}
}

func (c *DefaultComment) Generate(commit string) (string, error) {
	var reportedLayers []ReportedLayer
	for _, layer := range c.layers {
		prettyPlanKey := storage.GenerateKey(storage.LastPrettyPlan, &layer)
		plan, err := c.storage.Get(prettyPlanKey)
		if err != nil {
			return "", err
		}
		shortDiffKey := storage.GenerateKey(storage.LastPlanResult, &layer)
		shortDiff, err := c.storage.Get(shortDiffKey)
		if err != nil {
			return "", err
		}
		reportedLayer := ReportedLayer{
			Path:       layer.Spec.Path,
			ShortDiff:  string(shortDiff),
			PrettyPlan: string(plan),
		}
		reportedLayers = append(reportedLayers, reportedLayer)

	}
	data := struct {
		Commit string
		Layers []ReportedLayer
	}{
		Commit: commit,
		Layers: reportedLayers,
	}
	comment := bytes.NewBufferString("")
	err := defaultTemplate.Execute(comment, data)
	if err != nil {
		return "", err
	}
	return comment.String(), nil
}
