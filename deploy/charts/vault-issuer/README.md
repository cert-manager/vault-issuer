# vault-issuer

<!-- AUTO-GENERATED -->

### Global

#### **global.rbac.create** ~ `bool`
> Default value:
> ```yaml
> true
> ```

Create the ClusterRole, ClusterRoleBinding, Roles and RoleBindings that vault-issuer needs.
#### **replicaCount** ~ `number,string,null`
> Default value:
> ```yaml
> 1
> ```

The number of replicas of vault-issuer to run.  
  
Leader election is enabled, so only one replica reconciles at a time; the others stand by.  
  
For example:  
 Use integer to set a fixed number of replicas

```yaml
replicaCount: 2
```

Use null, if you want to omit the replicas field and use the Kubernetes default value.

```yaml
replicaCount: null
```

Use a string if you want to insert a variable for post-processing of the rendered template.

```yaml
replicaCount: ${REPLICAS_OVERRIDE:=3}
```



#### **revisionHistoryLimit** ~ `number,null`
> Default value:
> ```yaml
> 10
> ```

The number of old ReplicaSets to retain to allow rollback.  
If set to 0, no old ReplicaSets are retained.

#### **nameOverride** ~ `string`
> Default value:
> ```yaml
> ""
> ```
#### **namespace** ~ `string`
> Default value:
> ```yaml
> ""
> ```

The namespace to install vault-issuer into.  
If not set, the namespace of the release is used.  
This is helpful when installing vault-issuer as a chart dependency (sub chart).
#### **imagePullSecrets** ~ `array`
> Default value:
> ```yaml
> []
> ```

For Private docker registries, authentication is needed. Registry secrets are applied to the service account.
#### **imageRegistry** ~ `string`
> Default value:
> ```yaml
> quay.io
> ```

The container registry used for vault-issuer images by default. This can include path prefixes (e.g. "artifactory.example.com/docker").

#### **imageNamespace** ~ `string`
> Default value:
> ```yaml
> jetstack
> ```

The repository namespace used for vault-issuer images by default.  
Examples:  
- jetstack  
- cert-manager

#### **image.repository** ~ `string`
> Default value:
> ```yaml
> ""
> ```

Full repository override (takes precedence over `imageRegistry`,  
`imageNamespace`, and `image.name`).  
Example: quay.io/jetstack/vault-issuer

#### **image.name** ~ `string`
> Default value:
> ```yaml
> vault-issuer
> ```

The image name for vault-issuer.  
This is used (together with `imageRegistry` and `imageNamespace`) to construct the full image reference.

#### **image.tag** ~ `string`

Override the image tag to deploy by setting this variable. If no value is set, the chart's appVersion is used.

#### **image.digest** ~ `string`

Target image digest. Override any tag, if set.  
For example:

```yaml
digest: sha256:0e072dddd1f7f8fc8909a2ca6f65e76c5f0d2fcfb8be47935ae3457e8bbceb20
```

#### **image.pullPolicy** ~ `string`
> Default value:
> ```yaml
> IfNotPresent
> ```

Kubernetes imagePullPolicy on Deployment.
#### **automountServiceAccountToken** ~ `bool`
> Default value:
> ```yaml
> true
> ```

Automounting API credentials for the vault-issuer pod. Required: the controller talks to the Kubernetes API.

#### **serviceAccount.create** ~ `bool`
> Default value:
> ```yaml
> true
> ```

Specifies whether a service account should be created.
#### **serviceAccount.name** ~ `string`

The name of the service account to use.  
If not set and create is true, a name is generated using the name template.

#### **serviceAccount.automountServiceAccountToken** ~ `bool`
> Default value:
> ```yaml
> true
> ```

Automount API credentials for a Service Account.

#### **serviceAccount.annotations** ~ `object`
> Default value:
> ```yaml
> {}
> ```

Annotations to add to the ServiceAccount, for example to bind a cloud identity to it.
#### **volumes** ~ `array`
> Default value:
> ```yaml
> []
> ```

Additional volumes to add to the vault-issuer pod.
#### **volumeMounts** ~ `array`
> Default value:
> ```yaml
> []
> ```

Additional volume mounts to add to the vault-issuer container.
#### **resources** ~ `object`
> Default value:
> ```yaml
> {}
> ```

Kubernetes pod resources for vault-issuer.  
  
For example:

```yaml
resources:
  limits:
    cpu: 100m
    memory: 128Mi
  requests:
    cpu: 100m
    memory: 128Mi
```
#### **podSecurityContext** ~ `object`
> Default value:
> ```yaml
> {}
> ```

Security Context to be set on the vault-issuer Pod. For more information, see [Configure a Security Context for a Pod or Container](https://kubernetes.io/docs/tasks/configure-pod-container/security-context/).
#### **priorityClassName** ~ `string`
> Default value:
> ```yaml
> ""
> ```

Configure the priority class of the pod. For more information, see  
[PriorityClass](https://kubernetes.io/docs/concepts/scheduling-eviction/pod-priority-preemption/#priorityclass).
#### **nodeSelector** ~ `object`
> Default value:
> ```yaml
> kubernetes.io/os: linux
> ```

Configure the nodeSelector; defaults to any Linux node (vault-issuer doesn't support Windows nodes)

#### **affinity** ~ `object`
> Default value:
> ```yaml
> {}
> ```

Kubernetes Affinity. For more information, see [Affinity v1 core](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.27/#affinity-v1-core).  
For example:

```yaml
affinity:
  nodeAffinity:
   requiredDuringSchedulingIgnoredDuringExecution:
     nodeSelectorTerms:
     - matchExpressions:
       - key: foo.bar.com/role
         operator: In
         values:
         - master
```
#### **tolerations** ~ `array`
> Default value:
> ```yaml
> []
> ```

List of Kubernetes Tolerations, if required. For more information, see  
[Toleration v1 core](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.27/#toleration-v1-core).  
For example:

```yaml
tolerations:
- key: foo.bar.com/role
  operator: Equal
  value: master
  effect: NoSchedule
```
#### **topologySpreadConstraints** ~ `array`
> Default value:
> ```yaml
> []
> ```

List of Kubernetes TopologySpreadConstraints. For more information, see  
[TopologySpreadConstraint v1 core](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.27/#topologyspreadconstraint-v1-core).  
For example:

```yaml
topologySpreadConstraints:
- maxSkew: 2
  topologyKey: topology.kubernetes.io/zone
  whenUnsatisfiable: ScheduleAnyway
  labelSelector:
    matchLabels:
      app.kubernetes.io/name: vault-issuer
```
#### **app.logFormat** ~ `string`
> Default value:
> ```yaml
> text
> ```

The format of vault-issuer logging. Accepted values are text or json.
#### **app.logLevel** ~ `number`
> Default value:
> ```yaml
> 1
> ```

The verbosity of vault-issuer logging. This takes a value from 1-5, with the higher value being more verbose.
#### **app.issuer.clusterResourceNamespace** ~ `string`
> Default value:
> ```yaml
> ""
> ```

The namespace Secrets referenced by a ClusterIssuer are read from.  
  
Namespaced Issuers always read their Secrets from their own namespace; this setting only applies to cluster-scoped Issuers, which have no namespace of their own. Defaults to the namespace vault-issuer is installed into.

#### **app.issuer.ambientCredentials** ~ `bool`
> Default value:
> ```yaml
> false
> ```

Allow namespaced Issuers to authenticate with credentials drawn from the controller's environment, such as instance metadata or IRSA, rather than from the Issuer's spec.  
  
This is off by default: anyone who can create an Issuer in any namespace could otherwise borrow the controller's cloud identity.
#### **app.issuer.clusterAmbientCredentials** ~ `bool`
> Default value:
> ```yaml
> true
> ```

Allow ClusterIssuers to authenticate with credentials drawn from the controller's environment. ClusterIssuers are cluster-scoped and therefore managed by cluster administrators, so this is on by default.
#### **app.issuer.maxRetryDuration** ~ `string`
> Default value:
> ```yaml
> 2m
> ```

How long a CertificateRequest is retried after a transient signing error, measured from when the request was created. Once exceeded, the request fails and cert-manager creates a new one.
#### **app.leaderElection.enabled** ~ `bool`
> Default value:
> ```yaml
> true
> ```

Whether to enable leader election for vault-issuer.
#### **app.leaderElection.leaseDuration** ~ `string`
> Default value:
> ```yaml
> 15s
> ```

The duration that non-leader candidates will wait to force acquire leadership. The default should be sufficient in a healthy cluster but can be slightly increased to prevent vault-issuer from restart-looping when the  
API server is overloaded.
#### **app.leaderElection.renewDeadline** ~ `string`
> Default value:
> ```yaml
> 10s
> ```

The interval between attempts by the acting leader to renew a leadership slot before it stops leading. This MUST be less than or equal to the lease duration. The default should be sufficient in a healthy cluster but can be slightly increased to prevent vault-issuer from restart-looping when the  
API server is overloaded.
#### **app.readinessProbe.port** ~ `number`
> Default value:
> ```yaml
> 6060
> ```

The container port on which to expose the vault-issuer HTTP readiness probe using the default network interface.
#### **app.readinessProbe.path** ~ `string`
> Default value:
> ```yaml
> /readyz
> ```

The path on which to expose the vault-issuer HTTP readiness probe using the default network interface.
#### **app.securityContext** ~ `object`
> Default value:
> ```yaml
> allowPrivilegeEscalation: false
> capabilities:
>   drop:
>     - ALL
> readOnlyRootFilesystem: true
> runAsNonRoot: true
> seccompProfile:
>   type: RuntimeDefault
> ```

Security Context to be set on the vault-issuer app container. For more information, see [Configure a Security Context for a Pod or Container](https://kubernetes.io/docs/tasks/configure-pod-container/security-context/).

#### **app.podLabels** ~ `object`
> Default value:
> ```yaml
> {}
> ```

Pod labels to add to vault-issuer pods.
#### **app.podAnnotations** ~ `object`
> Default value:
> ```yaml
> {}
> ```

Pod annotations to add to vault-issuer pods.
#### **app.extraArgs** ~ `array`
> Default value:
> ```yaml
> []
> ```

Extra command line arguments passed to vault-issuer.
### Metrics

#### **app.metrics.port** ~ `number`
> Default value:
> ```yaml
> 9402
> ```

The port for exposing Prometheus metrics on 0.0.0.0 on path '/metrics'.
#### **app.metrics.service.enabled** ~ `bool`
> Default value:
> ```yaml
> true
> ```

Create a Service resource to expose the metrics endpoint.
#### **app.metrics.service.type** ~ `string`
> Default value:
> ```yaml
> ClusterIP
> ```

The Service type to expose metrics.
#### **app.metrics.service.ipFamilyPolicy** ~ `string`
> Default value:
> ```yaml
> ""
> ```

Set the ip family policy to configure dual-stack see [Configure dual-stack](https://kubernetes.io/docs/concepts/services-networking/dual-stack/#services)
#### **app.metrics.service.ipFamilies** ~ `array`
> Default value:
> ```yaml
> []
> ```

Sets the families that should be supported and the order in which they should be applied to ClusterIP as well. Can be IPv4 and/or IPv6.
#### **app.metrics.service.servicemonitor.enabled** ~ `bool`
> Default value:
> ```yaml
> false
> ```

Create a Prometheus ServiceMonitor for vault-issuer.
#### **app.metrics.service.servicemonitor.prometheusInstance** ~ `string`
> Default value:
> ```yaml
> default
> ```

Sets the value of the "prometheus" label on the ServiceMonitor. This is used so that separate Prometheus instances can select different ServiceMonitors using labels.
#### **app.metrics.service.servicemonitor.interval** ~ `string`
> Default value:
> ```yaml
> 10s
> ```

The interval to scrape the metrics.
#### **app.metrics.service.servicemonitor.scrapeTimeout** ~ `string`
> Default value:
> ```yaml
> 5s
> ```

The timeout for a metrics scrape.
#### **app.metrics.service.servicemonitor.labels** ~ `object`
> Default value:
> ```yaml
> {}
> ```

Additional labels to add to the ServiceMonitor.
#### **app.metrics.service.servicemonitor.endpointAdditionalProperties** ~ `object`
> Default value:
> ```yaml
> {}
> ```

EndpointAdditionalProperties allows setting additional properties on the endpoint such as relabelings, metricRelabelings etc.  
  
For example:

```yaml
endpointAdditionalProperties:
 relabelings:
 - action: replace
   sourceLabels:
   - __meta_kubernetes_pod_node_name
   targetLabel: instance
```



#### **podDisruptionBudget.enabled** ~ `bool`
> Default value:
> ```yaml
> false
> ```

Enable or disable the PodDisruptionBudget resource.  
  
This prevents downtime during voluntary disruptions such as during a Node upgrade. For example, the PodDisruptionBudget will block `kubectl drain` if it is used on the Node where the only remaining vault-issuer  
Pod is currently running.
#### **podDisruptionBudget.minAvailable** ~ `unknown`

This configures the minimum available pods for disruptions. It can either be set to an integer (e.g. 1) or a percentage value (e.g. 25%).  
It cannot be used if `maxUnavailable` is set.


#### **podDisruptionBudget.maxUnavailable** ~ `unknown`

This configures the maximum unavailable pods for disruptions. It can either be set to an integer (e.g. 1) or a percentage value (e.g. 25%).  
It cannot be used if `minAvailable` is set.


#### **commonLabels** ~ `object`
> Default value:
> ```yaml
> {}
> ```

Labels to apply to all resources
#### **commonAnnotations** ~ `object`
> Default value:
> ```yaml
> {}
> ```

Annotations to apply to all resources
#### **extraObjects** ~ `array`
> Default value:
> ```yaml
> []
> ```

Extra manifests to be deployed. This is useful for deploying additional resources that are not part of the chart.  
For example:

```yaml
extraObjects:
 - apiVersion: cilium.io/v2
   kind: CiliumNetworkPolicy
   metadata:
     name: vault-issuer
     namespace: cert-manager
   spec:
     endpointSelector:
       matchLabels:
         io.cilium.k8s.policy.serviceaccount: vault-issuer
     egress:
       - toEntities:
           - kube-apiserver
```

<!-- /AUTO-GENERATED -->
