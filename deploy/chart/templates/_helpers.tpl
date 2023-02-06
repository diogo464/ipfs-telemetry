{{/*
Expand the name of the chart.
*/}}
{{- define "ipfs-telemetry.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "ipfs-telemetry.fullname" -}}
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
{{- define "ipfs-telemetry.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "ipfs-telemetry.labels" -}}
helm.sh/chart: {{ include "ipfs-telemetry.chart" . }}
{{ include "ipfs-telemetry.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "ipfs-telemetry.selectorLabels" -}}
app.kubernetes.io/name: {{ include "ipfs-telemetry.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Create the name of the service account to use
*/}}
{{- define "ipfs-telemetry.serviceAccountName" -}}
{{- if .Values.serviceAccount.create }}
{{- default (include "ipfs-telemetry.fullname" .) .Values.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.serviceAccount.name }}
{{- end }}
{{- end }}

{{/*
Backend image
*/}}
{{- define "ipfs-telemetry.image.backend.image" -}}
{{ .Values.services.backendImage.repository }}:{{ .Values.services.backendImage.tag }}
{{- end }}

{{/*
Backend image pull policy
*/}}
{{- define "ipfs-telemetry.image.backend.pullPolicy" -}}
{{- default .Values.services.backendImage.pullPolicy }}
{{- end }}


{{/*
Environment variables
*/}}
{{- define "ipfs-telemetry.env" -}}
{{ include "ipfs-telemetry.env.nats" . }}
{{ include "ipfs-telemetry.env.redis" . }}
{{ include "ipfs-telemetry.env.minio" . }}
{{ include "ipfs-telemetry.env.postgres" . }}
{{ include "ipfs-telemetry.env.vm" . }}
{{- end }}

{{/*
Environment variables for nats
*/}}
{{- define "ipfs-telemetry.env.nats" -}}
- name: NATS_ENDPOINT
  {{- tpl (toYaml .Values.services.nats.endpoint) $ | nindent 2 }}
{{- end }}

{{/*
Environment variables for redis
*/}}
{{- define "ipfs-telemetry.env.redis" -}}
- name: REDIS_HOST
  {{- tpl (toYaml .Values.services.redis.host) $ | nindent 2 }}
- name: REDIS_PORT
  {{- tpl (toYaml .Values.services.redis.port) $ | nindent 2 }}
- name: REDIS_PASSWORD
  {{- tpl (toYaml .Values.services.redis.password) $ | nindent 2 }}
{{- end }}

{{/*
Environment variables for minio
*/}}
{{- define "ipfs-telemetry.env.minio" -}}
- name: S3_ENDPOINT
  {{- tpl (toYaml .Values.services.minio.endpoint) $ | nindent 2 }}
- name: S3_ACCESS_KEY
  {{- tpl (toYaml .Values.services.minio.accessKey) $ | nindent 2 }}
- name: S3_SECRET_KEY
  {{- tpl (toYaml .Values.services.minio.secretKey) $ | nindent 2 }}
- name: S3_USE_SSL
  {{- tpl (toYaml .Values.services.minio.useSSL) $ | nindent 2 }}
- name: S3_BUCKET_TELEMETRY
  {{- tpl (toYaml .Values.services.minio.bucketTelemetry) $ | nindent 2 }}
{{- end }}

{{/*
Environment variables for postgres
*/}}
{{- define "ipfs-telemetry.env.postgres" -}}
- name: POSTGRES_HOST
  value: "{{ .Release.Name }}-tsdb"
- name: POSTGRES_PORT
  value: "5432"
- name: POSTGRES_USER
  value: "postgres"
- name: POSTGRES_PASSWORD
  value: {{ .Values.timescaledb.secrets.credentials.PATRONI_SUPERUSER_PASSWORD | quote }}
- name: POSTGRES_DATABASE
  value: "postgres"
{{- end }}

{{/*
Environemnt variables for VictoriaMetrics
*/}}
{{- define "ipfs-telemetry.env.vm" -}}
- name: VM_ENDPOINT
  value: "http://{{ .Release.Name }}-vm-server:8428"
{{- end }}