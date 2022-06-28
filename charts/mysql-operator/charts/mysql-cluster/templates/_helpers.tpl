{{- define "cluster.name" -}}
{{- default .Release.Name .Values.name }}
{{- end }}

{{- define "images.sidecar" -}}
{{- if (empty .Values.images.sidecar) -}}
{{ ternary (printf "radondb/mysql80-sidecar:%s" .Values.operatorVersion ) (printf "radondb/mysql57-sidecar:%s" .Values.operatorVersion ) (eq .Values.mysqlVersion "8.0") }}
{{- else }}
{{ .Values.images.sidecar }}
{{- end }}
{{- end }}
 
{{- define "images.xenon" -}}
{{- default "radondb/xenon:1.1.5-alpha" .Values.images.xenon -}}
{{- end }}

{{- define "images.metrics" -}}
{{- default "prom/mysqld-exporter:v0.12.1" .Values.images.metrics -}}
{{- end }}

{{- define "images.busybox" -}}
{{- default "busybox:1.32" .Values.images.busybox -}}
{{- end }}

{{- define "user.secret.key" -}}
{{ printf "%s-%s" ( include "cluster.name" . ) .Values.superUser.name }}
{{- end }}

{{- define "user.secret.name" -}}
{{ "radondb-mysql-user-pwd" }}
{{- end }}

{{- define "user.cr.name" -}}
{{ printf "%s-%s-%s" ( include "cluster.name" . ) .Release.Namespace (.Values.superUser.name | replace "_" "-") }}
{{- end }}

{{- define "schedule.disable" }}
{{- and (not .Values.schedule.podAntiaffinity) (not .Values.schedule.nodeSelector) }}
{{- end }}
