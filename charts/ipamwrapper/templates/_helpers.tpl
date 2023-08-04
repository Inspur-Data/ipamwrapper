{{/*
Create chart name and version as used by the chart label.
*/}}
{{- define "ipamwrapper.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Expand the name of ipamwrapper .
*/}}
{{- define "ipamwrapper.name" -}}
{{- default "ipamwrapper" .Values.global.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "ipamwrapper.ipamwrapperController.labels" -}}
helm.sh/chart: {{ include "ipamwrapper.chart" . }}
{{ include "ipamwrapper.ipamwrapperController.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "ipamwrapper.ipamwrapperInit.labels" -}}
helm.sh/chart: {{ include "ipamwrapper.chart" . }}
{{ include "ipamwrapper.ipamwrapperInit.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
ipamwrapperAgent Common labels
*/}}
{{- define "ipamwrapper.ipamwrapperAgent.labels" -}}
helm.sh/chart: {{ include "ipamwrapper.chart" . }}
{{ include "ipamwrapper.ipamwrapperAgent.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}


{{/*
ipamwrapperController Selector labels
*/}}
{{- define "ipamwrapper.ipamwrapperController.selectorLabels" -}}
app.kubernetes.io/name: {{ include "ipamwrapper.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
app.kubernetes.io/component: {{ .Values.ipamwrapperController.name | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
ipamwrapperAgent Selector labels
*/}}
{{- define "ipamwrapper.ipamwrapperAgent.selectorLabels" -}}
app.kubernetes.io/name: {{ include "ipamwrapper.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
app.kubernetes.io/component: {{ .Values.ipamwrapperAgent.name | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
ipamwrapperInit Selector labels
*/}}
{{- define "ipamwrapper.ipamwrapperInit.selectorLabels" -}}
app.kubernetes.io/name: {{ include "ipamwrapper.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
app.kubernetes.io/component: {{ .Values.ipamwrapperInit.name | trunc 63 | trimSuffix "-" }}
{{- end }}



{{/* vim: set filetype=mustache: */}}
{{/*
Renders a value that contains template.
Usage:
{{ include "tplvalues.render" ( dict "value" .Values.path.to.the.Value "context" $) }}
*/}}
{{- define "tplvalues.render" -}}
    {{- if typeIs "string" .value }}
        {{- tpl .value .context }}
    {{- else }}
        {{- tpl (.value | toYaml) .context }}
    {{- end }}
{{- end -}}




{{/*
Return the appropriate apiVersion for poddisruptionbudget.
*/}}
{{- define "capabilities.policy.apiVersion" -}}
{{- if semverCompare "<1.21-0" .Capabilities.KubeVersion.Version -}}
{{- print "policy/v1beta1" -}}
{{- else -}}
{{- print "policy/v1" -}}
{{- end -}}
{{- end -}}

{{/*
Return the appropriate apiVersion for deployment.
*/}}
{{- define "capabilities.deployment.apiVersion" -}}
{{- if semverCompare "<1.14-0" .Capabilities.KubeVersion.Version -}}
{{- print "extensions/v1beta1" -}}
{{- else -}}
{{- print "apps/v1" -}}
{{- end -}}
{{- end -}}


{{/*
Return the appropriate apiVersion for RBAC resources.
*/}}
{{- define "capabilities.rbac.apiVersion" -}}
{{- if semverCompare "<1.17-0" .Capabilities.KubeVersion.Version -}}
{{- print "rbac.authorization.k8s.io/v1beta1" -}}
{{- else -}}
{{- print "rbac.authorization.k8s.io/v1" -}}
{{- end -}}
{{- end -}}

{{/*
return the ipamwrapperAgent image
*/}}
{{- define "ipamwrapper.ipamwrapperAgent.image" -}}
{{- $registryName := .Values.ipamwrapperAgent.image.registry -}}
{{- $repositoryName := .Values.ipamwrapperAgent.image.repository -}}
{{- if .Values.global.imageRegistryOverride }}
    {{- printf "%s/%s" .Values.global.imageRegistryOverride $repositoryName -}}
{{ else if $registryName }}
    {{- printf "%s/%s" $registryName $repositoryName -}}
{{- else -}}
    {{- printf "%s" $repositoryName -}}
{{- end -}}
{{- if .Values.ipamwrapperAgent.image.digest }}
    {{- print "@" .Values.ipamwrapperAgent.image.digest -}}
{{- else if .Values.ipamwrapperAgent.image.tag -}}
    {{- printf ":%s" .Values.ipamwrapperAgent.image.tag -}}
{{- else -}}
    {{- printf ":v%s" .Chart.AppVersion -}}
{{- end -}}
{{- end -}}

{{/*
return the ipamwrapperController image
*/}}
{{- define "ipamwrapper.ipamwrapperController.image" -}}
{{- $registryName := .Values.ipamwrapperController.image.registry -}}
{{- $repositoryName := .Values.ipamwrapperController.image.repository -}}
{{- if .Values.global.imageRegistryOverride }}
    {{- printf "%s/%s" .Values.global.imageRegistryOverride $repositoryName -}}
{{ else if $registryName }}
    {{- printf "%s/%s" $registryName $repositoryName -}}
{{- else -}}
    {{- printf "%s" $repositoryName -}}
{{- end -}}
{{- if .Values.ipamwrapperController.image.digest }}
    {{- print "@" .Values.ipamwrapperController.image.digest -}}
{{- else if .Values.ipamwrapperController.image.tag -}}
    {{- printf ":%s" .Values.ipamwrapperController.image.tag -}}
{{- else -}}
    {{- printf ":v%s" .Chart.AppVersion -}}
{{- end -}}
{{- end -}}

{{/*
return the ipamwrapperInit image
*/}}
{{- define "ipamwrapper.ipamwrapperInit.image" -}}
{{- $registryName := .Values.ipamwrapperInit.image.registry -}}
{{- $repositoryName := .Values.ipamwrapperInit.image.repository -}}
{{- if .Values.global.imageRegistryOverride }}
    {{- printf "%s/%s" .Values.global.imageRegistryOverride $repositoryName -}}
{{ else if $registryName }}
    {{- printf "%s/%s" $registryName $repositoryName -}}
{{- else -}}
    {{- printf "%s" $repositoryName -}}
{{- end -}}
{{- if .Values.ipamwrapperInit.image.digest }}
    {{- print "@" .Values.ipamwrapperInit.image.digest -}}
{{- else if .Values.ipamwrapperAgent.image.tag -}}
    {{- printf ":%s" .Values.ipamwrapperAgent.image.tag -}}
{{- else -}}
    {{- printf ":v%s" .Chart.AppVersion -}}
{{- end -}}
{{- end -}}

{{/*
generate the CA cert
*/}}
{{- define "generate-ca-certs" }}
    {{- $ca := genCA "spidernet.io" (.Values.ipamwrapperController.tls.auto.caExpiration | int) -}}
    {{- $_ := set . "ca" $ca -}}
{{- end }}

{{/*
insight labels
*/}}
{{- define "insight.labels" -}}
operator.insight.io/managed-by: insight
{{- end}}