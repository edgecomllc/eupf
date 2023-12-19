{{/*
Expand the name of the chart.
*/}}
{{- define "oai-gnb.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "oai-gnb.fullname" -}}
{{- if .Values.fullnameOverride }}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- $name := default .Chart.Name .Values.nameOverride }}
{{- if contains $name .Release.Name }}
{{- .Release.Name | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" }}
{{- end }}
{{- end }}
{{- end }}

{{/*
Create chart name and version as used by the chart label.
*/}}
{{- define "oai-gnb.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "oai-gnb.labels" -}}
helm.sh/chart: {{ include "oai-gnb.chart" . }}
{{ include "oai-gnb.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "oai-gnb.selectorLabels" -}}
app.kubernetes.io/name: {{ include "oai-gnb.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Create the name of the service account to use
*/}}
{{- define "oai-gnb.serviceAccountName" -}}
{{- if .Values.serviceAccount.create }}
{{- default (include "oai-gnb.fullname" .) .Values.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.serviceAccount.name }}
{{- end }}
{{- end }}

{{- define "oai-gnb.pod.envs" -}}
{{- with .Values.env }}
{{- toYaml . }}
{{- end }}
{{- end }}

{{- define "oai-gnb.pod.initContainers" -}}

{{- $all := . -}}
{{- range $k, $v := .Values.initContainers }}
- name: {{ $k }}
  image: {{ $v.image }}
  env:
    {{- include "oai-gnb.pod.envs" $all | nindent 4 }}
  {{- with $v.command }}
  command: {{ toYaml . | nindent 4 }}
  {{- end }}
  volumeMounts:
    - name: config-template
      mountPath: /cfg/tmpl
    - name: config-rendered
      mountPath: /cfg/files
{{- end }}

{{- end }}
