## :burrito: Burrito Apply Report

{{ len .Layers }} layer(s) affected by this pull request have been applied on commit `{{ .Commit }}`.

{{ range .Layers }}
### Layer {{ .Path }}

{{ if .Succeeded }}:white_check_mark: Apply succeeded.{{ else }}:x: Apply failed. Check the Burrito UI or runner logs for details.{{ end }}

{{ end }}