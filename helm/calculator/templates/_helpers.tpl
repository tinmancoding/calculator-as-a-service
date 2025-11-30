{{/*
Expand the name of the chart.
*/}}
{{- define "calculator.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
*/}}
{{- define "calculator.fullname" -}}
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
{{- define "calculator.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "calculator.labels" -}}
helm.sh/chart: {{ include "calculator.chart" . }}
{{ include "calculator.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "calculator.selectorLabels" -}}
app.kubernetes.io/name: {{ include "calculator.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Create the image name for a service
*/}}
{{- define "calculator.image" -}}
{{- $registry := .context.Values.global.imageRegistry -}}
{{- $repository := .service.image.repository -}}
{{- $tag := .service.image.tag | default .context.Chart.AppVersion -}}
{{- if .context.Values.openshift.build.enabled }}
{{- printf "image-registry.openshift-image-registry.svc:5000/%s/%s:%s" .context.Release.Namespace $repository $tag }}
{{- else }}
{{- printf "%s/%s:%s" $registry $repository $tag }}
{{- end }}
{{- end }}

{{/*
Deployment annotations for OpenShift ImageStream triggers
*/}}
{{- define "calculator.deploymentAnnotations" -}}
{{- if .context.Values.openshift.build.enabled }}
image.openshift.io/triggers: '[{"from":{"kind":"ImageStreamTag","name":"{{ .service.image.repository }}:{{ .service.image.tag | default .context.Chart.AppVersion }}"},"fieldPath":"spec.template.spec.containers[?(@.name==\"{{ .containerName }}\")].image"}]'
{{- end }}
{{- end }}

{{/*
Service labels for a specific service
*/}}
{{- define "calculator.serviceLabels" -}}
app: calculator
service: {{ .serviceName }}
{{ include "calculator.labels" .context }}
{{- end }}

{{/*
Service selector labels
*/}}
{{- define "calculator.serviceSelectorLabels" -}}
app: calculator
service: {{ .serviceName }}
{{- end }}
