package comment

import (
	"bytes"
	"text/template"

	configv1alpha1 "github.com/padok-team/burrito/api/v1alpha1"
	datastore "github.com/padok-team/burrito/internal/datastore/client"

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
	layers    []configv1alpha1.TerraformLayer
	datastore datastore.Client
}

type DefaultCommentInput struct {
}

func NewDefaultComment(layers []configv1alpha1.TerraformLayer, datastore datastore.Client) *DefaultComment {
	return &DefaultComment{
		layers:    layers,
		datastore: datastore,
	}
}

func (c *DefaultComment) Generate(commit string) (string, error) {
	var reportedLayers []ReportedLayer
	for _, layer := range c.layers {
		//TODO: handle attempt
		plan, err := c.datastore.GetPlan(layer.Namespace, layer.Name, layer.Status.LastRun, "0", "pretty")
		if err != nil {
			return "", err
		}
		//TODO: handle attempt
		shortDiff, err := c.datastore.GetPlan(layer.Namespace, layer.Name, layer.Status.LastRun, "0", "short")
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
