{{/*
Renders a value that contains template.
Usage:
{{ include "helpers.common.tplvalues.render" ( dict "value" .Values.path.to.the.Value "context" $) }}
*/}}
{{- define "helpers.common.tplvalues.render" -}}
    {{- if typeIs "string" .value }}
        {{- tpl .value .context }}
    {{- else }}
        {{- tpl (.value | toYaml) .context }}
    {{- end }}
{{- end -}}

{{/*
Container image full
*/}}
{{- define "helpers.common.containerImage" -}}
{{- if .Values.image.tag }}
{{- printf "%s/%s/%s:%s" .Values.image.registry .Values.image.repository .Values.image.name .Values.image.tag }}
{{- else }}
{{- printf "%s/%s/%s:%s" .Values.image.registry .Values.image.repository .Values.image.name .Chart.AppVersion }}
{{- end }}
{{- end -}}
