{{/* vim: set filetype=mustache: */}}
{{/*
Expand the name of the chart.
*/}}
{{- define "vault-issuer.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Create chart name and version as used by the chart label.
*/}}
{{- define "vault-issuer.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Common labels
*/}}
{{- define "vault-issuer.labels" -}}
app.kubernetes.io/name: {{ include "vault-issuer.name" . }}
helm.sh/chart: {{ include "vault-issuer.chart" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- if .Values.commonLabels}}
{{ toYaml .Values.commonLabels }}
{{- end }}
{{- end -}}

{{/*
Util function for generating the image URL based on the provided options.
This function is derived from the one standardized across charts in the
cert-manager GH organization (cert-manager, trust-manager, approver-policy,
see https://github.com/cert-manager/cert-manager/issues/6329), with one
difference: the deprecated `image.registry` backwards-compatibility prefix is
not supported, because this chart never shipped with it.
*/}}
{{- define "vault-issuer.image" -}}
{{- /*
Calling convention:

- (tuple <imageValues> <imageRegistry> <imageNamespace> <defaultReference>)

We intentionally pass imageRegistry/imageNamespace as explicit arguments rather than reading
from `.Values` inside this helper, because `helm-tool lint` does not reliably track `.Values.*`
usage through tuple/variable indirection.
*/ -}}

{{- if ne (len .) 4 -}}
    {{- fail (printf "ERROR: template \"vault-issuer.image\" expects (tuple <imageValues> <imageRegistry> <imageNamespace> <defaultReference>), got %d arguments" (len .)) -}}
{{- end -}}

{{- $image := index . 0 -}}
{{- $imageRegistry := index . 1 | default "" -}}
{{- $imageNamespace := index . 2 | default "" -}}
{{- $defaultReference := index . 3 -}}

{{- $repository := "" -}}
{{- if $image.repository -}}
    {{- $repository = $image.repository -}}
{{- else -}}
    {{- $name := required "ERROR: image.name must be set when image.repository is empty" $image.name -}}
    {{- $repository = $name -}}

    {{- if $imageNamespace -}}
        {{- $repository = printf "%s/%s" $imageNamespace $repository -}}
    {{- end -}}

    {{- if $imageRegistry -}}
        {{- $repository = printf "%s/%s" $imageRegistry $repository -}}
    {{- end -}}
{{- end -}}

{{- $repository -}}
{{- if and $image.tag $image.digest -}}
    {{- printf ":%s@%s" $image.tag $image.digest -}}
{{- else if $image.tag -}}
    {{- printf ":%s" $image.tag -}}
{{- else if $image.digest -}}
    {{- printf "@%s" $image.digest -}}
{{- else -}}
    {{- printf "%s" $defaultReference -}}
{{- end -}}
{{- end -}}

{{/*
Namespace for all resources to be installed into
If not defined in values file then the helm release namespace is used
By default this is not set so the helm release namespace will be used

This gets around an problem within helm discussed here
https://github.com/helm/helm/issues/5358
*/}}
{{- define "vault-issuer.namespace" -}}
    {{ .Values.namespace | default .Release.Namespace }}
{{- end -}}

{{/*
Create the name of the service account to use
*/}}
{{- define "vault-issuer.serviceAccountName" -}}
{{- if .Values.serviceAccount.create -}}
    {{ default (include "vault-issuer.name" .) .Values.serviceAccount.name }}
{{- else -}}
    {{ default "default" .Values.serviceAccount.name }}
{{- end -}}
{{- end -}}

{{/*
The namespace Secrets referenced by a ClusterIssuer are read from.
*/}}
{{- define "vault-issuer.clusterResourceNamespace" -}}
    {{ .Values.app.issuer.clusterResourceNamespace | default (include "vault-issuer.namespace" .) }}
{{- end -}}
