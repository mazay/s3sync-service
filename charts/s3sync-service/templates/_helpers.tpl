{{- define "s3sync.fullname"}}
{{- .Values.fullnameOverride | default .Release.Name -}}
{{- end -}}

{{- define "s3sync.configmapName" -}}
{{- .Values.configmap.name | default (include "s3sync.fullname" .) -}}
{{- end -}}

{{- define "s3sync.serviceAccountName" -}}
{{- if .Values.serviceAccountName -}}
{{- .Values.serviceAccountName -}}
{{- else if .Values.createRbac -}}
{{- include "s3sync.fullname" . -}}
{{- else -}}
default
{{- end -}}
{{- end -}}

{{- define "s3sync.labels" -}}
app: {{ include "s3sync.fullname" . | quote }}
{{- if .Values.labels -}}
{{ toYaml .Values.labels }}
{{- end -}}
{{- end -}}

{{- define "s3sync.podAnnotations" -}}
{{- if .Values.podAnnotations -}}
{{ toYaml .Values.podAnnotations }}
{{- end -}}
{{- if .Values.prometheusExporter.enable -}}
prometheus.io/path: {{ .Values.prometheusExporter.path | quote }}
prometheus.io/port: {{ .Values.prometheusExporter.port | quote }}
prometheus.io/scrape: "true"
{{- end -}}
{{- end -}}
