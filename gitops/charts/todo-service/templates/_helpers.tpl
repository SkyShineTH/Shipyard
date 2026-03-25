{{/*
Expand the chart name.
*/}}
{{- define "todo-service.name" -}}
{{- .Chart.Name | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Full release name: release-name + chart name unless overridden.
Truncated to 63 chars (Kubernetes label limit).
*/}}
{{- define "todo-service.fullname" -}}
{{- if .Values.fullnameOverride }}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- printf "%s" .Release.Name | trunc 63 | trimSuffix "-" }}
{{- end }}
{{- end }}

{{/*
Common labels applied to every resource.
*/}}
{{- define "todo-service.labels" -}}
helm.sh/chart: {{ printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{ include "todo-service.selectorLabels" . }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Selector labels — used by Deployment and Service selectors.
*/}}
{{- define "todo-service.selectorLabels" -}}
app.kubernetes.io/name: {{ include "todo-service.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}
