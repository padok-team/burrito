package comment

import (
	"bytes"
	"text/template"

	_ "embed"
)

var (
	//go:embed templates/apply.md
	applyTemplateRaw string
	applyTemplate    = template.Must(template.New("apply-report").Parse(applyTemplateRaw))
)

type ApplyReportedLayer struct {
	Path      string
	Succeeded bool
}

type ApplyComment struct {
	layers []ApplyReportedLayer
}

func NewApplyComment(layers []ApplyReportedLayer) *ApplyComment {
	return &ApplyComment{layers: layers}
}

func (c *ApplyComment) Generate(commit string) (string, error) {
	data := struct {
		Commit string
		Layers []ApplyReportedLayer
	}{
		Commit: commit,
		Layers: c.layers,
	}
	buf := bytes.NewBufferString("")
	err := applyTemplate.Execute(buf, data)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}
