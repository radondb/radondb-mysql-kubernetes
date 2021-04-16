{{/* vim: set filetype=mustache: */}}
{{/*
Expand the name of the chart.
*/}}
{{- define "name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Create a default fully qualified app name.
The format of host is "fullname"-0."fullname"."namespace" (podname.servicename.namespace), 
and the MySQL limits the total length of master_host to 60 byte, 
so the length of "fullname" must be limited to '(60-4-len(namespace))/2'.
*/}}
{{- define "fullname" -}}
{{- $length := div (sub 56 (len .Release.Namespace )) 2 | int }}
{{- if .Values.fullnameOverride -}}
{{- .Values.fullnameOverride | trunc $length | trimSuffix "-" -}}
{{- else -}}
{{- $name := default .Chart.Name .Values.nameOverride -}}
{{- if contains $name .Release.Name -}}
{{- .Release.Name | trunc $length | trimSuffix "-" -}}
{{- else -}}
{{- printf "%s-%s" .Release.Name $name | trunc $length | trimSuffix "-" -}}
{{- end -}}
{{- end -}}
{{- end -}}

{{/*
Create the name of the service account to use
*/}}
{{- define "serviceAccountName" -}}
{{- if .Values.serviceAccount.create -}}
    {{ default (include "fullname" .) .Values.serviceAccount.name }}
{{- else -}}
    {{ default "default" .Values.serviceAccount.name }}
{{- end -}}
{{- end -}}

{{/*
Create chart name and version as used by the chart label.
*/}}
{{- define "krypton.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" -}}
{{- end -}}
