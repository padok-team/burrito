# burrito

![Version: 0.4.1](https://img.shields.io/badge/Version-0.4.1-informational?style=flat-square) ![Type: application](https://img.shields.io/badge/Type-application-informational?style=flat-square) ![AppVersion: v0.4.1](https://img.shields.io/badge/AppVersion-v0.4.1-informational?style=flat-square)

A Helm chart for handling a complete burrito deployment

## Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| config.annotations | object | `{}` | Annotations to be added to the ConfigMap |
| config.burrito.controller.githubConfig.apiToken | string | `""` | Github API token, prefer override with the BURRITO_CONTROLLER_GITHUBCONFIG_APITOKEN environment variable |
| config.burrito.controller.githubConfig.appId | string | `""` | Github app ID, prefer override with the BURRITO_CONTROLLER_GITHUBCONFIG_APPID environment variable |
| config.burrito.controller.githubConfig.installationId | string | `""` | Github app unstallation ID, prefer override with the BURRITO_CONTROLLER_GITHUBCONFIG_INSTALLATIONID environment variable |
| config.burrito.controller.githubConfig.privateKey | string | `""` | Github app private key, prefer override with the BURRITO_CONTROLLER_GITHUBCONFIG_PRIVATEKEY environment variable |
| config.burrito.controller.gitlabConfig.apiToken | string | `""` | Gitlab API token Prefer override with the BURRITO_CONTROLLER_GITLABCONFIG_APITOKEN environment variable |
| config.burrito.controller.gitlabConfig.url | string | `""` | Gitlab URL |
| config.burrito.controller.healthProbeBindAddress | string | `":8081"` | Adress to bind the controller health probe |
| config.burrito.controller.kubernetesWebhookPort | int | `9443` | Port used to handle the Kubernetes webhook |
| config.burrito.controller.leaderElection.enabled | bool | `true` | Enable/Disable leader election |
| config.burrito.controller.leaderElection.id | string | `"6d185457.terraform.padok.cloud"` | Leader election lock name |
| config.burrito.controller.maxConcurrentReconciles | int | `1` | Maximum number of concurrent reconciles for the controller, increse this value if you have a lot of resources to reconcile |
| config.burrito.controller.metricsBindAddress | string | `":8080"` | Adress to bind the controller metrics |
| config.burrito.controller.namespaces | list | `[]` | By default, the controller will only watch the tenants namespaces |
| config.burrito.controller.terraformMaxRetries | int | `3` | Maximum number of retries for Terraform operations (plan, apply...) |
| config.burrito.controller.timers.driftDetection | string | `"10m"` | Drift detection interval |
| config.burrito.controller.timers.failureGracePeriod | int | `30` | Duration to wait before retrying on failure (increases exponentially with the amount of failed retries) |
| config.burrito.controller.timers.onError | string | `"10s"` | Duration to wait before retrying on error |
| config.burrito.controller.timers.waitAction | string | `"1m"` | Duration to wait before retrying on locked layer |
| config.burrito.controller.types | list | `["layer","repository","run","pullrequest"]` | Resource types to watch for reconciliation |
| config.burrito.datastore.addr | string | `":8080"` | Datastore exposed port |
| config.burrito.datastore.serviceAccounts | list | `[]` | Service account to use for datastore operations (e.g. reading/writing to storage) |
| config.burrito.datastore.storage.azure.container | string | `""` | Azure storage container name |
| config.burrito.datastore.storage.azure.storageAccount | string | `""` | Azure storage account name |
| config.burrito.datastore.storage.gcs.bucket | string | `""` | GCS bucket name |
| config.burrito.datastore.storage.mock | bool | `false` | Use in-memory storage for testing - not intended for production use, data will be lost on datastore restart  |
| config.burrito.datastore.storage.s3.bucket | string | `""` | S3 bucket name |
| config.burrito.hermitcrab | object | `{}` | Provider cache custom configuration |
| config.burrito.runner.sshKnownHostsConfigMapName | string | `"burrito-ssh-known-hosts"` | Configmap name to store the SSH known hosts in the runner |
| config.burrito.server.addr | string | `":8080"` | Server exposed port |
| config.burrito.server.webhook.github.secret | string | `""` | Secret to validate webhook payload, prefer override with the BURRITO_SERVER_WEBHOOK_GITHUB_SECRET environment variable |
| config.burrito.server.webhook.gitlab.secret | string | `""` | Secret to validate webhook payload, Prefer override with the BURRITO_SERVER_WEBHOOK_GITLAB_SECRET environment variable |
| config.create | bool | `true` | Create ConfigMap with Burrito configuration |
| controllers.deployment | object | `{"args":["controllers","start"],"command":["burrito"],"env":[],"envFrom":[],"livenessProbe":{"httpGet":{"path":"/healthz","port":8081},"initialDelaySeconds":5,"periodSeconds":20},"podAnnotations":{"kubectl.kubernetes.io/default-container":"burrito"},"readinessProbe":{"httpGet":{"path":"/readyz","port":8081},"initialDelaySeconds":5,"periodSeconds":20}}` | Deployment configuration for the Burrito controller |
| controllers.deployment.args | list | `["controllers","start"]` | Arguments to pass to the Burrito controller container |
| controllers.deployment.command | list | `["burrito"]` | Command to run in the Burrito controller container |
| controllers.deployment.env | list | `[]` | Environment variables to pass to the Burrito controller container |
| controllers.deployment.envFrom | list | `[]` | Environment variables to pass to the Burrito controller container |
| controllers.deployment.livenessProbe | object | `{"httpGet":{"path":"/healthz","port":8081},"initialDelaySeconds":5,"periodSeconds":20}` | Controller liveness probe configuration |
| controllers.deployment.podAnnotations | object | `{"kubectl.kubernetes.io/default-container":"burrito"}` | Annotations to be added to the pods generated by the Burrito controller deployment |
| controllers.deployment.readinessProbe | object | `{"httpGet":{"path":"/readyz","port":8081},"initialDelaySeconds":5,"periodSeconds":20}` | Controller readiness probe configuration |
| controllers.metadata | object | `{"labels":{"app.kubernetes.io/component":"controllers","app.kubernetes.io/name":"burrito-controllers"}}` | Metadata configuration for the Burrito controller |
| controllers.service.enabled | bool | `false` | Enable/Disable service creation for the Burrito controller |
| datastore.deployment.args | list | `["datastore","start"]` | Arguments to pass to the Burrito datastore container |
| datastore.deployment.command | list | `["burrito"]` | Command to run in the Burrito datastore container |
| datastore.deployment.envFrom | list | `[]` | Environment variables to pass to the Burrito datastore container |
| datastore.deployment.livenessProbe | object | `{"httpGet":{"path":"/healthz","port":8080},"initialDelaySeconds":5,"periodSeconds":20}` | Datastore liveness probe configuration |
| datastore.deployment.podAnnotations | object | `{"kubectl.kubernetes.io/default-container":"burrito"}` | Annotations to be added to the pods generated by the Burrito datastore deployment |
| datastore.deployment.ports | list | `[{"containerPort":8080,"name":"http"}]` | Datastore exposed port |
| datastore.deployment.readinessProbe | object | `{"httpGet":{"path":"/healthz","port":8080},"initialDelaySeconds":5,"periodSeconds":20}` | Datastore readiness probe configuration |
| datastore.metadata | object | `{"labels":{"app.kubernetes.io/component":"datastore","app.kubernetes.io/name":"burrito-datastore"}}` | Metadata configuration for the Burrito datastore |
| datastore.service | object | `{"ports":[{"name":"http","port":80,"targetPort":"http"},{"name":"https","port":443,"targetPort":"http"}]}` | Service configuration for the Burrito datastore |
| datastore.tls | object | `{"certManager":{"certificate":{"spec":{"commonName":"burrito-datastore.burrito-system.svc.cluster.local","dnsNames":["burrito-datastore.burrito-system.svc.cluster.local","burrito-datastore.burrito-system","burrito-datastore"],"issuerRef":{"kind":"Issuer","name":"burrito-ca-issuer"},"secretName":"burrito-datastore-tls"}},"use":false}}` | TLS configuration for the Burrito datastore |
| datastore.tls.certManager.certificate | object | `{"spec":{"commonName":"burrito-datastore.burrito-system.svc.cluster.local","dnsNames":["burrito-datastore.burrito-system.svc.cluster.local","burrito-datastore.burrito-system","burrito-datastore"],"issuerRef":{"kind":"Issuer","name":"burrito-ca-issuer"},"secretName":"burrito-datastore-tls"}}` | CertManager certificate configuration |
| datastore.tls.certManager.certificate.spec.issuerRef.name | string | `"burrito-ca-issuer"` | The default issuer name is "burrito-ca-issuer", packaged with the chart |
| datastore.tls.certManager.use | bool | `false` | Use CertManager for Burrito datastore TLS (requires cert-manager to be installed on the cluster) |
| global.deployment.autoscaling.enabled | bool | `false` | Enable/Disable autoscaling for Burrito pods |
| global.deployment.envFrom | list | `[]` | Global environment variables  |
| global.deployment.image | object | `{"pullPolicy":"Always","repository":"ghcr.io/padok-team/burrito","tag":""}` | Global image configuration |
| global.deployment.podAnnotations | object | `{}` | Global annotations for pods generated by Burrito deployments |
| global.deployment.podSecurityContext | object | `{"runAsNonRoot":true}` | Global pod security context configuration |
| global.deployment.ports | list | `[]` | Global ports configuration |
| global.deployment.replicas | int | `1` | Number of replicas for Burrito pods |
| global.deployment.resources | object | `{}` | Global resources configuration |
| global.deployment.securityContext | object | `{"allowPrivilegeEscalation":false,"capabilities":{"drop":["ALL"]}}` | Global security context configuration |
| global.metadata | object | `{"annotations":{},"labels":{"app.kubernetes.io/part-of":"burrito"}}` | Global metadata configuration  |
| global.service | object | `{"enabled":true}` | Global service configuration |
| global.service.enabled | bool | `true` | Enable/Disable service creation for Burrito components |
| global.serviceAccount.metadata | object | `{"annotations":{},"labels":{}}` | Global metadata configuration for service accounts used by Burrito components |
| hermitcrab.deployment.affinity | object | `{}` | Hermitcrab affinity |
| hermitcrab.deployment.env | list | `[{"name":"SERVER_TLS_CERT_FILE","value":"/etc/hermitcrab/tls/tls.crt"},{"name":"SERVER_TLS_PRIVATE_KEY_FILE","value":"/etc/hermitcrab/tls/tls.key"}]` | Hermitcrab environment variables |
| hermitcrab.deployment.env[0].value | string | `"/etc/hermitcrab/tls/tls.crt"` | Path to the Hermitcrab TLS certificate |
| hermitcrab.deployment.env[1].value | string | `"/etc/hermitcrab/tls/tls.key"` | Path to the Hermitcrab TLS private key |
| hermitcrab.deployment.image | object | `{"pullPolicy":"Always","repository":"sealio/hermitcrab","tag":"main"}` | Hermitcrab image configuration |
| hermitcrab.deployment.livenessProbe | object | `{"failureThreshold":10,"httpGet":{"httpHeaders":[{"name":"User-Agent","value":""}],"path":"/livez","port":80},"periodSeconds":10,"timeoutSeconds":5}` | Hermitcrab liveness probe configuration |
| hermitcrab.deployment.nodeSelector | object | `{}` | Hermitcrab node selector |
| hermitcrab.deployment.ports | list | `[{"containerPort":80,"name":"http"},{"containerPort":443,"name":"https"}]` | Hermitcrab ports configuration |
| hermitcrab.deployment.readinessProbe | object | `{"failureThreshold":3,"httpGet":{"path":"/readyz","port":80},"periodSeconds":5,"timeoutSeconds":5}` | Hermitcrab readiness probe configuration |
| hermitcrab.deployment.replicas | int | `1` | Hermitcrab replicas |
| hermitcrab.deployment.resources | object | `{"limits":{"cpu":"1","memory":"2Gi"},"requests":{"cpu":"300m","memory":"256Mi"}}` | Hermitcrab resources configuration |
| hermitcrab.deployment.startupProbe | object | `{"failureThreshold":10,"httpGet":{"path":"/readyz","port":80},"periodSeconds":5}` | Hermitcrab startup probe configuration |
| hermitcrab.deployment.tolerations | object | `{}` | Hermitcrab tolerations |
| hermitcrab.enabled | bool | `false` | Enable/Disable Hermitcrab (terraform provider cache in cluster) |
| hermitcrab.metadata.labels | object | `{"app.kubernetes.io/component":"hermitcrab","app.kubernetes.io/name":"burrito-hermitcrab"}` | Labels to be added to the Hermitcrab resources |
| hermitcrab.storage.emptyDir.enabled | bool | `true` | Use emptyDir for Hermitcrab storage |
| hermitcrab.storage.emptyDir.medium | string | `""` | EmptyDir medium |
| hermitcrab.storage.emptyDir.sizeLimit | string | `"2Gi"` | EmptyDir size limit |
| hermitcrab.storage.ephemeral.enabled | bool | `false` | Use ephemeral storage for Hermitcrab storage |
| hermitcrab.storage.ephemeral.size | string | `"2Gi"` | Ephemeral storage size |
| hermitcrab.storage.ephemeral.storageClassName | string | `""` | Ephemeral storage class name |
| hermitcrab.tls.certManager.certificate.spec.commonName | string | `"burrito-hermitcrab.burrito-system.svc.cluster.local"` | Common name for the Hermitcrab TLS certificate |
| hermitcrab.tls.certManager.certificate.spec.dnsNames | list | `["burrito-hermitcrab.burrito-system.svc.cluster.local","burrito-hermitcrab.burrito-system","burrito-hermitcrab"]` | DNS names for the Hermitcrab TLS certificate |
| hermitcrab.tls.certManager.certificate.spec.issuerRef.kind | string | `"Issuer"` |  |
| hermitcrab.tls.certManager.certificate.spec.issuerRef.name | string | `"burrito-ca-issuer"` | The default issuer name is "burrito-ca-issuer", packaged with the chart |
| hermitcrab.tls.certManager.certificate.spec.secretName | string | `"burrito-hermitcrab-tls"` | Secret name to store the Hermitcrab TLS certificate |
| hermitcrab.tls.certManager.use | bool | `true` | Use CertManager for Hermitcrab TLS (requires cert-manager to be installed on the cluster) |
| server.deployment | object | `{"args":["server","start"],"command":["burrito"],"envFrom":[{"secretRef":{"name":"burrito-webhook-secret","optional":true}}],"livenessProbe":{"httpGet":{"path":"/healthz","port":8080},"initialDelaySeconds":5,"periodSeconds":20},"podAnnotations":{"kubectl.kubernetes.io/default-container":"burrito"},"ports":[{"containerPort":8080,"name":"http"}],"readinessProbe":{"httpGet":{"path":"/healthz","port":8080},"initialDelaySeconds":5,"periodSeconds":20}}` | Deployment configuration for the Burrito server |
| server.deployment.args | list | `["server","start"]` | Arguments to pass to the Burrito server container |
| server.deployment.command | list | `["burrito"]` | Command to run in the Burrito server container |
| server.deployment.envFrom | list | `[{"secretRef":{"name":"burrito-webhook-secret","optional":true}}]` | Environment variables to pass to the Burrito server container |
| server.deployment.envFrom[0] | object | `{"secretRef":{"name":"burrito-webhook-secret","optional":true}}` | Reference the webhook secret here, it should define a BURRITO_SERVER_WEBHOOK_GITHUB_SECRET and/or BURRITO_SERVER_WEBHOOK_GITLAB_SECRET key |
| server.deployment.livenessProbe | object | `{"httpGet":{"path":"/healthz","port":8080},"initialDelaySeconds":5,"periodSeconds":20}` | Server liveness probe configuration |
| server.deployment.podAnnotations | object | `{"kubectl.kubernetes.io/default-container":"burrito"}` | Annotations to be added to the pods generated by the Burrito server deployment |
| server.deployment.ports | list | `[{"containerPort":8080,"name":"http"}]` | Server exposed port |
| server.deployment.readinessProbe | object | `{"httpGet":{"path":"/healthz","port":8080},"initialDelaySeconds":5,"periodSeconds":20}` | Server readiness probe configuration |
| server.ingress | object | `{"annotations":{},"enabled":false,"host":"burrito.example.com","ingressClassName":"nginx","tls":[]}` | Ingress configuration for the Burrito server |
| server.ingress.annotations | object | `{}` | Annotations to be added to the Burrito server ingress |
| server.ingress.enabled | bool | `false` | Enable/Disable ingress creation for the Burrito server |
| server.ingress.host | string | `"burrito.example.com"` | Hostname for the Burrito server ingress |
| server.ingress.ingressClassName | string | `"nginx"` | Ingress class name to use for the Burrito server ingress |
| server.ingress.tls | list | `[]` | TLS configuration for the Burrito server ingress |
| server.metadata | object | `{"labels":{"app.kubernetes.io/component":"server","app.kubernetes.io/name":"burrito-server"}}` | Metadata configuration for the Burrito server |
| server.service | object | `{"ports":[{"name":"http","port":80,"targetPort":"http"}]}` | Service configuration for the Burrito server |
| tenants | string | `nil` | List of tenants to create to manage Terraform resources |

----------------------------------------------
Autogenerated from chart metadata using [helm-docs v1.14.2](https://github.com/norwoodj/helm-docs/releases/v1.14.2)
