{{- define "s3sync.fullname"}}
{{- .Values.fullnameOverride | default .Release.Name -}}
{{- end -}}

{{- define "s3sync.configmapName" -}}
{{- .Values.configmap.name | default (include "s3sync.fullname" .) -}}
{{- end -}}

{{- define "s3sync.serviceAccountName" -}}
{{- .Values.serviceAccountName | default (include "s3sync.fullname" .) -}}
{{- end -}}

{{- define "s3sync.labels" -}}
app: {{ include "s3sync.fullname" . | quote }}
{{ toYaml .Values.labels }}
{{- end -}}

{{- define "s3sync.podAnnotations" -}}
{{ toYaml .Values.podAnnotations }}
{{- if .Values.prometheusExporter.enable }}
prometheus.io/path: {{ .Values.prometheusExporter.path | quote }}
prometheus.io/port: {{ .Values.prometheusExporter.port | quote }}
prometheus.io/scrape: "true"
{{- end }}
{{- end -}}
