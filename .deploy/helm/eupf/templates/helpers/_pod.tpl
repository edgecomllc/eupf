{{- define "helpers.pod.containerImage" -}}

{{- if .Values.image.tag }}
{{- printf "%s/%s/%s:%s" .Values.image.registry .Values.image.repository .Values.image.name .Values.image.tag }}
{{- else }}
{{- printf "%s/%s/%s:%s" .Values.image.registry .Values.image.repository .Values.image.name .Chart.AppVersion }}
{{- end }}

{{- end -}}

{{- define "helpers.pod.uiImage" -}}

{{- if .Values.image.tag }}
{{- printf "%s/%s/%s:%s" .Values.ui.image.registry .Values.ui.image.repository .Values.ui.image.name .Values.ui.image.tag }}
{{- else }}
{{- printf "%s/%s/%s:%s" .Values.ui.image.registry .Values.ui.image.repository .Values.ui.image.name .Chart.AppVersion }}
{{- end }}

{{- end -}}

{{- define "helpers.pod.initContainers" -}}

{{- $all := . -}}
{{- range $k, $v := .Values.initContainers }}
- name: {{ $v.name }}
  {{- if $v.image }}
  image: {{ $v.image }}
  {{- else }}
  image: {{ include "helpers.pod.containerImage" $all | quote }}
  {{- end }}
  {{- if $v.imagePullPolicy }}
  imagePullPolicy: {{ $v.imagePullPolicy }}
  {{- else }}
  imagePullPolicy: {{ $all.Values.global.imagePullPolicy }}
  {{- end }}
  {{- if $all.Values.env }}
  env:
    {{- include "helpers.pod.envs" $all | nindent 4 }}
  {{- end }}
  command: {{- include "helpers.common.tplvalues.render" (dict "value" $v.command "context" $) | nindent 4 }}

{{- end }}

{{- end -}}

{{- define "helpers.pod.envs" -}}

{{- $env := deepCopy (.Values.env | default dict) }}
{{- if not .Values.config.enabled }}
{{- $env = merge $env (deepCopy (.Values.defaultEnv | default dict)) }}
{{- end }}
{{- range $name, $value := $env }}
- name: {{ $name }}
  value: {{ $value | quote }}
{{- end }}

{{- end -}}

{{/*
Container args: user-provided args, prefixed with `--config <config.path>`
when the ConfigMap-based configuration mode is enabled.
*/}}
{{- define "helpers.pod.args" -}}
{{- $args := .Values.args | default list }}
{{- if .Values.config.enabled }}
{{- $args = prepend (prepend $args .Values.config.path) "--config" }}
{{- end }}
{{- include "helpers.common.tplvalues.render" (dict "value" $args "context" $) }}
{{- end -}}

{{- define "helpers.pod.probes" -}}

{{- with .Values.livenessProbe }}
livenessProbe:
  {{- toYaml . | nindent 2 }}
{{- end }}
{{- with .Values.readinessProbe }}
readinessProbe:
  {{- toYaml . | nindent 2 }}
{{- end }}
{{- with .Values.startupProbe }}
startupProbe:
  {{- toYaml . | nindent 2 }}
{{- end }}

{{- end -}}

{{- define "helpers.pod.volumes" -}}

{{- $volumes := .Values.volumes | default list }}
{{- if .Values.config.enabled }}
{{- $configVolume := dict "name" "config" "configMap" (dict "name" (printf "%s-config" (include "eupf.fullname" .))) }}
{{- $volumes = append $volumes $configVolume }}
{{- end }}
{{- with $volumes }}
volumes:
  {{- toYaml . | nindent 2 }}
{{- end }}

{{- end -}}

{{- define "helpers.pod.volumeMounts" -}}

{{- $mounts := .Values.volumeMounts | default list }}
{{- if .Values.config.enabled }}
{{- $configMount := dict "name" "config" "mountPath" .Values.config.mountPath }}
{{- $mounts = append $mounts $configMount }}
{{- end }}
{{- with $mounts }}
volumeMounts:
  {{- toYaml . | nindent 2 }}
{{- end }}

{{- end -}}
