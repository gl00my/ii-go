{{ if gt $.Pages 100 }}
<div id="pager" class ="fsmaller">
{{ else }}
<div id="pager">
{{ end }}
{{ if gt .Pages 1 }}
Pages:
{{ end }}
{{ range .Pager }}
{{ if eq . $.Page }}
<span class="selected">{{.}}</span>
{{ else }}
<a href="/{{$.BasePath}}/{{.}}">{{.}}</a>
{{ end }}
{{ end }}
</div>
