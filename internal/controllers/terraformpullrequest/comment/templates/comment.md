## :burrito: Burrito Report

{{ len .Layers }} layer(s) affected with {{ .Commit }} commit.

{{ range .Layers }}

### Layer {{ .Name }} ({{ .Path }})

`{{ .ShortDiff }}`

<details>
<summary>Plan</summary>

```terraform
{{ .PrettyPlan }}
```
</details>

{{ end }}
