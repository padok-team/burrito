## :burrito: Burrito Report

{{ len .Layers }} layer(s) affected with {{ .Commit }} commit.

{{ range .Layers }}

### Layer {{ .Path }}

`{{ .ShortDiff }}`

<details>
<summary>Plan</summary>

```terraform
{{ .PrettyPlan }}
```
</details>

{{ end }}
