# burrito

![Version: 0.13.0](https://img.shields.io/badge/Version-0.13.0-informational?style=flat-square) ![Type: application](https://img.shields.io/badge/Type-application-informational?style=flat-square) ![AppVersion: v0.13.0](https://img.shields.io/badge/AppVersion-v0.13.0-informational?style=flat-square)

A Helm chart for handling a complete burrito deployment

**Homepage:** <https://docs.burrito.tf/>

## Source Code

* <https://github.com/padok-team/burrito>

## Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| config.burrito.controller.defaultSyncWindows | list | `[]` | Default sync windows for layer reconciliation |
| config.burrito.controller.healthProbeBindAddress | string | `":8081"` | Address to bind the controller health probe |
| config.burrito.controller.kubernetesWebhookPort | int | `9443` | Port used to handle the Kubernetes webhook |
| config.burrito.controller.leaderElection.enabled | bool | `true` | Enable/Disable leader election |
| config.burrito.controller.leaderElection.id | string | `"6d185457.terraform.padok.cloud"` | Leader election lock name |
| config.burrito.controller.logFormat | string | `"text"` | Log format for the controller, either "text" or "json" |
| config.burrito.controller.maxConcurrentReconciles | int | `1` | Maximum number of concurrent reconciles for the controller, increase this value if you have a lot of resources to reconcile |
| config.burrito.controller.maxConcurrentRunnerPods | int | `0` | Maximum number of concurrent runners pods. 0 means no limit |
| config.burrito.controller.metricsBindAddress | string | `":8080"` | Address to bind the controller metrics |
| config.burrito.controller.namespaces | list | `[]` | By default, the controller will only watch the tenants namespaces |
| config.burrito.controller.terraformMaxRetries | int | `3` | Maximum number of retries for Terraform operations (plan, apply...) |
| config.burrito.controller.timers.driftDetection | string | `"10m"` | Drift detection interval |
| config.burrito.controller.timers.failureGracePeriod | string | `"15s"` | Duration to wait before retrying on failure (increases exponentially with the amount of failed retries) |
| config.burrito.controller.timers.onError | string | `"10s"` | Duration to wait before retrying on error |
| config.burrito.controller.timers.repositorySync | string | `"5m"` | Repository polling interval |
| config.burrito.controller.timers.waitAction | string | `"10s"` | Duration to wait before retrying on locked layer |
| config.burrito.controller.types | list | `["layer","repository","run","pullrequest"]` | Resource types to watch for reconciliation. |
| config.burrito.datastore.addr | string | `":8080"` | Datastore exposed port |
| config.burrito.datastore.hostname | string | `"burrito-datastore.burrito-system.svc.cluster.local"` | Datastore hostname, used by controller, server and runner to reach the datastore |
| config.burrito.datastore.logFormat | string | `"text"` | Log format for the datastore, either "text" or "json" |
| config.burrito.datastore.serviceAccounts | list | `[]` | Service accounts that are allowed to access the datastore API in namespace/name format (not the service account used by the datastore pods, check datastore.serviceAccount.metadata for that) |
| config.burrito.datastore.storage.azure.container | string | `""` | Azure storage container name |
| config.burrito.datastore.storage.azure.storageAccount | string | `""` | Azure storage account name |
| config.burrito.datastore.storage.gcs.bucket | string | `""` | GCS bucket name |
| config.burrito.datastore.storage.mock | bool | `false` | Use in-memory storage for testing - not intended for production use, data will be lost on datastore restart |
| config.burrito.datastore.storage.s3.bucket | string | `""` | S3 bucket name |
| config.burrito.datastore.storage.s3.usePathStyle | bool | `false` | S3 option for bucket name in path instead of as subdomain |
| config.burrito.hermitcrab | object | `{}` | Provider cache custom configuration |
| config.burrito.runner.args | list | `["runner","start"]` | Arguments to pass to the Burrito runner container |
| config.burrito.runner.command | list | `["burrito"]` | Command to run in the Burrito runner container |
| config.burrito.runner.image.pullPolicy | string | `"Always"` |  |
| config.burrito.runner.image.repository | string | `"ghcr.io/padok-team/burrito"` | Default image to use for runners, can be overridden with spec.OverrideRunnerSpec in repositories and layer definitions |
| config.burrito.runner.image.tag | string | `"v0.13.0@sha256:e957726fe488c5532d4e510ffd15b0b05694e0ffcaa1831797e5fb01aa699d64"` |  |
| config.burrito.runner.sshKnownHostsConfigMapName | string | `"burrito-ssh-known-hosts"` | Configmap name to store the SSH known hosts in the runner |
| config.burrito.server.addr | string | `":8080"` | Server exposed port |
| config.burrito.server.basicAuth.enabled | bool | `true` | Enable/Disable Basic Authentication for the Burrito server. If both Basic Auth and OIDC are disabled, the server will be publicly accessible. |
| config.burrito.server.oidc.clientId | string | `""` |  |
| config.burrito.server.oidc.enabled | bool | `false` | Enable/Disable OIDC authentication for the Burrito server |
| config.burrito.server.oidc.issuerUrl | string | `""` | OIDC issuer URL |
| config.burrito.server.oidc.redirectUrl | string | `""` | OIDC Redirect URL, should be the Burrito server URL with /auth/callback appended (ex: https://burrito.example.com/auth/callback) |
| config.burrito.server.oidc.scopes | list | `["openid","profile","email"]` | OIDC scopes to request |
| config.burrito.server.session.maxAge | int | `86400` | Session max age in seconds, after which the session will expire |
| config.burrito.server.session.secure | bool | `false` | Cookie secure, set this to true if using HTTPS |
| config.create | bool | `true` | Create ConfigMap with Burrito configuration |
| config.metadata | object | `{"annotations":{},"labels":{"app.kubernetes.io/name":"burrito-config"}}` | Metadata configuration for config map |
| controllers.deployment | object | `{"args":["controllers","start"],"command":["burrito"],"env":[{"name":"SSH_KNOWN_HOSTS","value":"/home/burrito/.ssh/known_hosts"}],"envFrom":[],"extraVolumeMounts":{},"extraVolumes":{},"livenessProbe":{"httpGet":{"path":"/healthz","port":8081},"initialDelaySeconds":5,"periodSeconds":20},"metadata":{"annotations":{},"labels":{}},"podAnnotations":{"kubectl.kubernetes.io/default-container":"burrito"},"podSecurityContext":{},"ports":[{"containerPort":8080,"name":"metrics"}],"readinessProbe":{"httpGet":{"path":"/readyz","port":8081},"initialDelaySeconds":5,"periodSeconds":20},"securityContext":{}}` | Deployment configuration for the Burrito controller |
| controllers.deployment.args | list | `["controllers","start"]` | Arguments to pass to the Burrito controller container |
| controllers.deployment.command | list | `["burrito"]` | Command to run in the Burrito controller container |
| controllers.deployment.env | list | `[{"name":"SSH_KNOWN_HOSTS","value":"/home/burrito/.ssh/known_hosts"}]` | Environment variables to pass to the Burrito controller container |
| controllers.deployment.envFrom | list | `[]` | Environment variables to pass to the Burrito controller container |
| controllers.deployment.extraVolumeMounts | object | `{}` | Additional volume mounts |
| controllers.deployment.extraVolumes | object | `{}` | Additional volumes |
| controllers.deployment.metadata | object | `{"annotations":{},"labels":{}}` | Metadata configuration for the Burrito controller deployment |
| controllers.deployment.podAnnotations | object | `{"kubectl.kubernetes.io/default-container":"burrito"}` | Annotations to be added to the pods generated by the Burrito controller deployment |
| controllers.deployment.podSecurityContext | object | `{}` | Pod security context for the Burrito controller. Merged with (and overrides) global.deployment.podSecurityContext |
| controllers.deployment.ports | list | `[{"containerPort":8080,"name":"metrics"}]` | Controller exposed ports |
| controllers.deployment.readinessProbe | object | `{"httpGet":{"path":"/readyz","port":8081},"initialDelaySeconds":5,"periodSeconds":20}` | Controller readiness probe configuration |
| controllers.deployment.securityContext | object | `{}` | Security context for the Burrito controller container. Merged with (and overrides) global.deployment.securityContext |
| controllers.metadata | object | `{"annotations":{},"labels":{"app.kubernetes.io/component":"controllers","app.kubernetes.io/name":"burrito-controllers"}}` | Metadata configuration for the Burrito controller |
| controllers.metricsHttproute.apiVersion | string | `"gateway.networking.k8s.io/v1"` | Gateway API version to use for the Burrito controller metrics HTTPRoute |
| controllers.metricsHttproute.enabled | bool | `false` | Enable/Disable Gateway API HTTPRoute creation for the Burrito controller metrics endpoint (requires the Gateway API CRDs and controllers.service enabled) |
| controllers.metricsHttproute.hostnames | list | `[]` | Hostnames the HTTPRoute matches. Defaults to the controller metrics ingress host when left empty |
| controllers.metricsHttproute.metadata | object | `{"annotations":{},"labels":{}}` | Metadata configuration for the Burrito controller metrics HTTPRoute |
| controllers.metricsHttproute.parentRefs | list | `[]` | Gateways the HTTPRoute attaches to (required when metricsHttproute is enabled). Each entry follows the Gateway API parentRef schema (name, optionally namespace/sectionName) |
| controllers.metricsHttproute.rules | list | `[]` | HTTPRoute rules. Defaults to a single rule routing "/" (PathPrefix) to the burrito-controllers service when left empty |
| controllers.metricsIngress.enabled | bool | `false` | Enable/Disable ingress creation for the Burrito controller metrics endpoint (controllers.service MUST be enabled) |
| controllers.metricsIngress.host | string | `"burrito-controllers.example.com"` | Hostname for the Burrito controller metrics ingress |
| controllers.metricsIngress.ingressClassName | string | `"nginx"` | Ingress class name to use for the Burrito controller metrics ingress |
| controllers.metricsIngress.metadata | object | `{"annotations":{},"labels":{}}` | Metadata configuration for the Burrito controller metrics ingress |
| controllers.metricsIngress.tls | list | `[]` | TLS configuration for the Burrito controller metrics ingress |
| controllers.service.enabled | bool | `false` | Enable/Disable service creation for the Burrito controller |
| controllers.service.metadata | object | `{"annotations":{},"labels":{}}` | Metadata configuration for the Burrito controller service |
| controllers.service.ports[0].name | string | `"metrics"` |  |
| controllers.service.ports[0].port | int | `80` |  |
| controllers.service.ports[0].targetPort | string | `"metrics"` |  |
| controllers.serviceAccount | object | `{"metadata":{"annotations":{},"labels":{}}}` | Service account configuration for the Burrito controller deployment |
| controllers.serviceAccount.metadata | object | `{"annotations":{},"labels":{}}` | Metadata configuration for the Burrito controller service account |
| controllers.storage.emptyDir.enabled | bool | `true` | Use emptyDir for Burrito repositories storage |
| controllers.storage.emptyDir.medium | string | `""` | EmptyDir medium |
| controllers.storage.emptyDir.sizeLimit | string | `"2Gi"` | EmptyDir size limit |
| controllers.storage.ephemeral.enabled | bool | `false` | Use ephemeral storage for Burrito repositories storage |
| controllers.storage.ephemeral.size | string | `"2Gi"` | Ephemeral storage size |
| controllers.storage.ephemeral.storageClassName | string | `""` | Ephemeral storage class name |
| datastore.deployment.affinity | object | `{}` | Datastore affinity |
| datastore.deployment.args | list | `["datastore","start"]` | Arguments to pass to the Burrito datastore container |
| datastore.deployment.command | list | `["burrito"]` | Command to run in the Burrito datastore container |
| datastore.deployment.envFrom | list | `[]` | Environment variables to pass to the Burrito datastore container |
| datastore.deployment.extraVolumeMounts | object | `{}` | Additional volume mounts |
| datastore.deployment.extraVolumes | object | `{}` | Additional volumes |
| datastore.deployment.livenessProbe | object | `{"httpGet":{"path":"/healthz","port":8080},"initialDelaySeconds":5,"periodSeconds":20}` | Datastore liveness probe configuration |
| datastore.deployment.metadata | object | `{"annotations":{},"labels":{}}` | Metadata configuration for the Burrito datastore deployment |
| datastore.deployment.nodeSelector | object | `{}` | Datastore node selector |
| datastore.deployment.podAnnotations | object | `{"kubectl.kubernetes.io/default-container":"burrito"}` | Annotations to be added to the pods generated by the Burrito datastore deployment |
| datastore.deployment.podSecurityContext | object | `{}` | Pod security context for the Burrito datastore. Merged with (and overrides) global.deployment.podSecurityContext |
| datastore.deployment.ports | list | `[{"containerPort":8080,"name":"http"}]` | Datastore exposed port |
| datastore.deployment.readinessProbe | object | `{"httpGet":{"path":"/healthz","port":8080},"initialDelaySeconds":5,"periodSeconds":20}` | Datastore readiness probe configuration |
| datastore.deployment.securityContext | object | `{}` | Security context for the Burrito datastore container. Merged with (and overrides) global.deployment.securityContext |
| datastore.deployment.tolerations | list | `[]` | Datastore tolerations |
| datastore.deployment.topologySpreadConstraints | list | `[]` | Datastore topology spread constraints |
| datastore.metadata | object | `{"annotations":{},"labels":{"app.kubernetes.io/component":"datastore","app.kubernetes.io/name":"burrito-datastore"}}` | Metadata configuration for the Burrito datastore |
| datastore.service | object | `{"metadata":{"annotations":{},"labels":{}},"ports":[{"name":"http","port":80,"targetPort":"http"},{"name":"https","port":443,"targetPort":"http"}],"trafficDistribution":""}` | Service configuration for the Burrito datastore |
| datastore.service.metadata | object | `{"annotations":{},"labels":{}}` | Metadata configuration for the Burrito datastore service |
| datastore.service.trafficDistribution | string | `""` | Datastore service traffic distribution policy |
| datastore.serviceAccount | object | `{"metadata":{"annotations":{},"labels":{}}}` | Service account configuration for the Burrito datastore deployment. Use this to grant permission to the datastore to interact with external storage |
| datastore.serviceAccount.metadata | object | `{"annotations":{},"labels":{}}` | Metadata configuration for the Burrito datastore service account |
| datastore.tls | object | `{"certManager":{"certificate":{"metadata":{"annotations":{},"labels":{}},"spec":{"commonName":"burrito-datastore.burrito-system.svc.cluster.local","dnsNames":["burrito-datastore.burrito-system.svc.cluster.local","burrito-datastore.burrito-system","burrito-datastore"],"issuerRef":{"kind":"Issuer","name":"burrito-ca-issuer"},"secretName":"burrito-datastore-tls"}},"use":true},"enabled":false,"secretName":"burrito-datastore-tls"}` | TLS configuration for the Burrito datastore |
| datastore.tls.certManager.certificate | object | `{"metadata":{"annotations":{},"labels":{}},"spec":{"commonName":"burrito-datastore.burrito-system.svc.cluster.local","dnsNames":["burrito-datastore.burrito-system.svc.cluster.local","burrito-datastore.burrito-system","burrito-datastore"],"issuerRef":{"kind":"Issuer","name":"burrito-ca-issuer"},"secretName":"burrito-datastore-tls"}}` | CertManager certificate configuration |
| datastore.tls.certManager.certificate.metadata | object | `{"annotations":{},"labels":{}}` | Metadata configuration for the Burrito datastore certificate |
| datastore.tls.certManager.certificate.spec.issuerRef.name | string | `"burrito-ca-issuer"` | The default issuer name is "burrito-ca-issuer", packaged with the chart |
| datastore.tls.certManager.use | bool | `true` | Use CertManager for Burrito datastore TLS (recommended - requires cert-manager to be installed on the cluster) |
| datastore.tls.enabled | bool | `false` | Enable/Disable TLS for the Burrito datastore (recommended for production use) |
| datastore.tls.secretName | string | `"burrito-datastore-tls"` | Reference a secret that contains a CA certificate (ca.crt, tls.crt, tls.key) for the Burrito datastore (if not using CertManager) |
| global.crds.install | bool | `true` | Enable/Disable CRD installation through the Helm chart |
| global.deployment.autoscaling.enabled | bool | `false` | Enable/Disable autoscaling for Burrito pods |
| global.deployment.envFrom | list | `[]` | Global environment variables |
| global.deployment.extraVolumeMounts | object | `{}` | Additional volume mounts |
| global.deployment.extraVolumes | object | `{}` | Additional volumes |
| global.deployment.image | object | `{"pullPolicy":"Always","repository":"ghcr.io/padok-team/burrito","tag":"v0.13.0@sha256:e957726fe488c5532d4e510ffd15b0b05694e0ffcaa1831797e5fb01aa699d64"}` | Global image configuration |
| global.deployment.metadata | object | `{"annotations":{},"labels":{}}` | Global metadata configuration for Burrito components deployments |
| global.deployment.mode | string | `"Release"` |  |
| global.deployment.podAnnotations | object | `{}` | Global annotations for pods generated by Burrito deployments |
| global.deployment.podSecurityContext | object | `{"runAsNonRoot":true}` | Global pod security context configuration |
| global.deployment.ports | list | `[]` | Global ports configuration |
| global.deployment.replicas | int | `1` | Number of replicas for Burrito pods |
| global.deployment.resources | object | `{}` | Global resources configuration |
| global.deployment.securityContext | object | `{"allowPrivilegeEscalation":false,"capabilities":{"drop":["ALL"]}}` | Global security context configuration |
| global.metadata | object | `{"annotations":{},"labels":{"app.kubernetes.io/part-of":"burrito"}}` | Global metadata configuration for Burrito components |
| global.service | object | `{"enabled":true,"metadata":{"annotations":{},"labels":{}}}` | Global service configuration |
| global.service.enabled | bool | `true` | Enable/Disable service creation for Burrito components |
| global.service.metadata | object | `{"annotations":{},"labels":{}}` | Global metadata configuration for Burrito components services |
| global.serviceAccount.metadata | object | `{"annotations":{},"labels":{}}` | Global metadata configuration for Burrito components service accounts |
| global.sshKnownHosts | string | `"bitbucket.org ssh-rsa AAAAB3NzaC1yc2EAAAABIwAAAQEAubiN81eDcafrgMeLzaFPsw2kNvEcqTKl/VqLat/MaB33pZy0y3rJZtnqwR2qOOvbwKZYKiEO1O6VqNEBxKvJJelCq0dTXWT5pbO2gDXC6h6QDXCaHo6pOHGPUy+YBaGQRGuSusMEASYiWunYN0vCAI8QaXnWMXNMdFP3jHAJH0eDsoiGnLPBlBp4TNm6rYI74nMzgz3B9IikW4WVK+dc8KZJZWYjAuORU3jc1c/NPskD2ASinf8v3xnfXeukU0sJ5N6m5E8VLjObPEO+mN2t/FZTMZLiFqPWc/ALSqnMnnhwrNi2rbfg/rd/IpL8Le3pSBne8+seeFVBoGqzHM9yXw==\ngithub.com ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQCj7ndNxQowgcQnjshcLrqPEiiphnt+VTTvDP6mHBL9j1aNUkY4Ue1gvwnGLVlOhGeYrnZaMgRK6+PKCUXaDbC7qtbW8gIkhL7aGCsOr/C56SJMy/BCZfxd1nWzAOxSDPgVsmerOBYfNqltV9/hWCqBywINIR+5dIg6JTJ72pcEpEjcYgXkE2YEFXV1JHnsKgbLWNlhScqb2UmyRkQyytRLtL+38TGxkxCflmO+5Z8CSSNY7GidjMIZ7Q4zMjA2n1nGrlTDkzwDCsw+wqFPGQA179cnfGWOWRVruj16z6XyvxvjJwbz0wQZ75XK5tKSb7FNyeIEs4TT4jk+S4dhPeAUC5y+bDYirYgM4GC7uEnztnZyaVWQ7B381AK4Qdrwt51ZqExKbQpTUNn+EjqoTwvqNj4kqx5QUCI0ThS/YkOxJCXmPUWZbhjpCg56i+2aB6CmK2JGhn57K5mj0MNdBXA4/WnwH6XoPWJzK5Nyu2zB3nAZp+S5hpQs+p1vN1/wsjk=\ngitlab.com ecdsa-sha2-nistp256 AAAAE2VjZHNhLXNoYTItbmlzdHAyNTYAAAAIbmlzdHAyNTYAAABBBFSMqzJeV9rUzU4kWitGjeR4PWSa29SPqJ1fVkhtj3Hw9xjLVXVYrU9QlYWrOLXBpQ6KWjbjTDTdDkoohFzgbEY=\ngitlab.com ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIAfuCHKVTjquxvt6CM6tdG4SLp1Btn/nOeHHE5UOzRdf\ngitlab.com ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQCsj2bNKTBSpIYDEGk9KxsGh3mySTRgMtXL583qmBpzeQ+jqCMRgBqB98u3z++J1sKlXHWfM9dyhSevkMwSbhoR8XIq/U0tCNyokEi/ueaBMCvbcTHhO7FcwzY92WK4Yt0aGROY5qX2UKSeOvuP4D6TPqKF1onrSzH9bx9XUf2lEdWT/ia1NEKjunUqu1xOB/StKDHMoX4/OKyIzuS0q/T1zOATthvasJFoPrAjkohTyaDUz2LN5JoH839hViyEG82yB+MjcFV5MU3N1l1QL3cVUCh93xSaua1N85qivl+siMkPGbO5xR/En4iEY6K2XPASUEMaieWVNTRCtJ4S8H+9\nssh.dev.azure.com ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQC7Hr1oTWqNqOlzGJOfGJ4NakVyIzf1rXYd4d7wo6jBlkLvCA4odBlL0mDUyZ0/QUfTTqeu+tm22gOsv+VrVTMk6vwRU75gY/y9ut5Mb3bR5BV58dKXyq9A9UeB5Cakehn5Zgm6x1mKoVyf+FFn26iYqXJRgzIZZcZ5V6hrE0Qg39kZm4az48o0AUbf6Sp4SLdvnuMa2sVNwHBboS7EJkm57XQPVU3/QpyNLHbWDdzwtrlS+ez30S3AdYhLKEOxAG8weOnyrtLJAUen9mTkol8oII1edf7mWWbWVf0nBmly21+nZcmCTISQBtdcyPaEno7fFQMDD26/s0lfKob4Kw8H\nvs-ssh.visualstudio.com ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQC7Hr1oTWqNqOlzGJOfGJ4NakVyIzf1rXYd4d7wo6jBlkLvCA4odBlL0mDUyZ0/QUfTTqeu+tm22gOsv+VrVTMk6vwRU75gY/y9ut5Mb3bR5BV58dKXyq9A9UeB5Cakehn5Zgm6x1mKoVyf+FFn26iYqXJRgzIZZcZ5V6hrE0Qg39kZm4az48o0AUbf6Sp4SLdvnuMa2sVNwHBboS7EJkm57XQPVU3/QpyNLHbWDdzwtrlS+ez30S3AdYhLKEOxAG8weOnyrtLJAUen9mTkol8oII1edf7mWWbWVf0nBmly21+nZcmCTISQBtdcyPaEno7fFQMDD26/s0lfKob4Kw8H\ngithub.com ecdsa-sha2-nistp256 AAAAE2VjZHNhLXNoYTItbmlzdHAyNTYAAAAIbmlzdHAyNTYAAABBBEmKSENjQEezOmxkZMy7opKgwFB9nkt5YRrYMjNuG5N87uRgg6CLrbo5wAdT/y6v0mKV0U2w0WZ2YB/++Tpockg=\ngithub.com ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIOMqqnkVzrm0SdG6UOoqKLsabgH5C9okWi0dh2l9GKJl"` |  |
| hermitcrab.deployment.affinity | object | `{}` | Hermitcrab affinity |
| hermitcrab.deployment.env | list | `[{"name":"SERVER_TLS_CERT_FILE","value":"/etc/hermitcrab/tls/tls.crt"},{"name":"SERVER_TLS_PRIVATE_KEY_FILE","value":"/etc/hermitcrab/tls/tls.key"}]` | Hermitcrab environment variables |
| hermitcrab.deployment.env[0].value | string | `"/etc/hermitcrab/tls/tls.crt"` | Path to the Hermitcrab TLS certificate |
| hermitcrab.deployment.env[1].value | string | `"/etc/hermitcrab/tls/tls.key"` | Path to the Hermitcrab TLS private key |
| hermitcrab.deployment.extraVolumeMounts | object | `{}` | Additional volume mounts |
| hermitcrab.deployment.extraVolumes | object | `{}` | Additional volumes |
| hermitcrab.deployment.image | object | `{"pullPolicy":"Always","repository":"sealio/hermitcrab","tag":"v0.1.7@sha256:2201a99d97d5f1c8071cc19aca631e3f90a92b3f246dc3a0cfb6d4b3a09d9ffd"}` | Hermitcrab image configuration |
| hermitcrab.deployment.livenessProbe | object | `{"failureThreshold":10,"httpGet":{"httpHeaders":[{"name":"User-Agent","value":""}],"path":"/livez","port":80},"periodSeconds":10,"timeoutSeconds":5}` | Hermitcrab liveness probe configuration |
| hermitcrab.deployment.metadata | object | `{"annotations":{},"labels":{}}` | Metadata configuration for Hermitcrab deployment |
| hermitcrab.deployment.nodeSelector | object | `{}` | Hermitcrab node selector |
| hermitcrab.deployment.podSecurityContext | object | `{}` | Pod security context for Hermitcrab. Merged with (and overrides) global.deployment.podSecurityContext |
| hermitcrab.deployment.ports | list | `[{"containerPort":80,"name":"http"},{"containerPort":443,"name":"https"}]` | Hermitcrab ports configuration |
| hermitcrab.deployment.readinessProbe | object | `{"failureThreshold":3,"httpGet":{"path":"/readyz","port":80},"periodSeconds":5,"timeoutSeconds":5}` | Hermitcrab readiness probe configuration |
| hermitcrab.deployment.replicas | int | `1` | Hermitcrab replicas |
| hermitcrab.deployment.resources | object | `{"limits":{"cpu":"1","memory":"2Gi"},"requests":{"cpu":"300m","memory":"256Mi"}}` | Hermitcrab resources configuration |
| hermitcrab.deployment.securityContext | object | `{}` | Security context for the hermitcrab container. Merged with (and overrides) global.deployment.securityContext |
| hermitcrab.deployment.startupProbe | object | `{"failureThreshold":10,"httpGet":{"path":"/readyz","port":80},"periodSeconds":5}` | Hermitcrab startup probe configuration |
| hermitcrab.deployment.tolerations | list | `[]` | Hermitcrab tolerations |
| hermitcrab.deployment.topologySpreadConstraints | list | `[]` | Hermitcrab topology spread constraints |
| hermitcrab.enabled | bool | `false` | Enable/Disable Hermitcrab (terraform provider cache in cluster) |
| hermitcrab.metadata | object | `{"annotations":{},"labels":{"app.kubernetes.io/component":"hermitcrab","app.kubernetes.io/name":"burrito-hermitcrab"}}` | Metadata configuration for Hermitcrab |
| hermitcrab.service | object | `{"metadata":{"annotations":{},"labels":{}},"trafficDistribution":""}` | Hermitcrab service configuration |
| hermitcrab.service.metadata | object | `{"annotations":{},"labels":{}}` | Metadata configuration for Hermitcrab service |
| hermitcrab.service.trafficDistribution | string | `""` | Hermitcrab service traffic distribution policy |
| hermitcrab.storage.emptyDir.enabled | bool | `true` | Use emptyDir for Hermitcrab storage |
| hermitcrab.storage.emptyDir.medium | string | `""` | EmptyDir medium |
| hermitcrab.storage.emptyDir.sizeLimit | string | `"2Gi"` | EmptyDir size limit |
| hermitcrab.storage.ephemeral.enabled | bool | `false` | Use ephemeral storage for Hermitcrab storage |
| hermitcrab.storage.ephemeral.size | string | `"2Gi"` | Ephemeral storage size |
| hermitcrab.storage.ephemeral.storageClassName | string | `""` | Ephemeral storage class name |
| hermitcrab.tls.certManager.certificate.metadata | object | `{"annotations":{},"labels":{}}` | Metadata configuration for Hermitcrab certificate |
| hermitcrab.tls.certManager.certificate.spec.commonName | string | `"burrito-hermitcrab.burrito-system.svc.cluster.local"` | Common name for the Hermitcrab TLS certificate |
| hermitcrab.tls.certManager.certificate.spec.dnsNames | list | `["burrito-hermitcrab.burrito-system.svc.cluster.local","burrito-hermitcrab.burrito-system","burrito-hermitcrab"]` | DNS names for the Hermitcrab TLS certificate |
| hermitcrab.tls.certManager.certificate.spec.issuerRef.kind | string | `"Issuer"` |  |
| hermitcrab.tls.certManager.certificate.spec.issuerRef.name | string | `"burrito-ca-issuer"` | The default issuer name is "burrito-ca-issuer", packaged with the chart |
| hermitcrab.tls.certManager.certificate.spec.secretName | string | `"burrito-hermitcrab-tls"` | Secret name to store the Hermitcrab TLS certificate |
| hermitcrab.tls.certManager.use | bool | `true` | Use CertManager for Hermitcrab TLS (recommended - requires cert-manager to be installed on the cluster) |
| hermitcrab.tls.secretName | string | `"burrito-hermitcrab-tls"` | Reference a secret that contains a CA cer (ca.crt, tls.crt, tls.key) for Hermitcrab (if not using CertManager) - note: TLS is required for Hermitcrab, as Terraform Provider Mirror protocol requires it |
| networkPolicy.enabled | bool | `false` | Enable/Disable Network Policy creation |
| networkPolicy.ingressFromTenants | object | `{"additionalIngressRules":[],"enabled":true}` | Network policy to allow ingress traffic from all the tenant namespaces to the release namespace |
| networkPolicy.ingressFromTenants.additionalIngressRules | list | `[]` | Additional ingress rules for tenant namespaces network policy |
| networkPolicy.ingressFromTenants.enabled | bool | `true` | Enable/Disable tenant ingress network policy |
| networkPolicy.metadata | object | `{"annotations":{},"labels":{}}` | Metadata configuration for Network Policies |
| runners.rbac.metadata | object | `{"annotations":{},"labels":{"app.kubernetes.io/component":"runner","app.kubernetes.io/name":"burrito-runner"}}` | Metadata configuration for RBAC of the Burrito runners managed by this chart |
| server.deployment | object | `{"affinity":{},"args":["server","start"],"command":["burrito"],"env":[],"envFrom":[],"extraVolumeMounts":{},"extraVolumes":{},"livenessProbe":{"httpGet":{"path":"/healthz","port":8080},"initialDelaySeconds":5,"periodSeconds":20},"metadata":{"annotations":{},"labels":{}},"nodeSelector":{},"podAnnotations":{"kubectl.kubernetes.io/default-container":"burrito"},"podSecurityContext":{},"ports":[{"containerPort":8080,"name":"http"}],"readinessProbe":{"httpGet":{"path":"/healthz","port":8080},"initialDelaySeconds":5,"periodSeconds":20},"securityContext":{},"tolerations":[],"topologySpreadConstraints":[]}` | Deployment configuration for the Burrito server |
| server.deployment.affinity | object | `{}` | Server affinity |
| server.deployment.args | list | `["server","start"]` | Arguments to pass to the Burrito server container |
| server.deployment.command | list | `["burrito"]` | Command to run in the Burrito server container |
| server.deployment.env | list | `[]` | Environment variables to pass to the Burrito server container |
| server.deployment.envFrom | list | `[]` | Environment variables to pass to the Burrito server container |
| server.deployment.extraVolumeMounts | object | `{}` | Additional volume mounts |
| server.deployment.extraVolumes | object | `{}` | Additional volumes |
| server.deployment.livenessProbe | object | `{"httpGet":{"path":"/healthz","port":8080},"initialDelaySeconds":5,"periodSeconds":20}` | Server liveness probe configuration |
| server.deployment.metadata | object | `{"annotations":{},"labels":{}}` | Metadata configuration for the Burrito server deployment |
| server.deployment.nodeSelector | object | `{}` | Server node selector |
| server.deployment.podAnnotations | object | `{"kubectl.kubernetes.io/default-container":"burrito"}` | Annotations to be added to the pods generated by the Burrito server deployment |
| server.deployment.podSecurityContext | object | `{}` | Pod security context for the Burrito server. Merged with (and overrides) global.deployment.podSecurityContext |
| server.deployment.ports | list | `[{"containerPort":8080,"name":"http"}]` | Server exposed port |
| server.deployment.readinessProbe | object | `{"httpGet":{"path":"/healthz","port":8080},"initialDelaySeconds":5,"periodSeconds":20}` | Server readiness probe configuration |
| server.deployment.securityContext | object | `{}` | Security context for the Burrito server container. Merged with (and overrides) global.deployment.securityContext |
| server.deployment.tolerations | list | `[]` | Server tolerations |
| server.deployment.topologySpreadConstraints | list | `[]` | Server topology spread constraints |
| server.httproute.apiVersion | string | `"gateway.networking.k8s.io/v1"` | Gateway API version to use for the Burrito server HTTPRoute |
| server.httproute.enabled | bool | `false` | Enable/Disable Gateway API HTTPRoute creation for the Burrito server (requires the Gateway API CRDs and server.service enabled) |
| server.httproute.hostnames | list | `[]` | Hostnames the HTTPRoute matches. Defaults to the server ingress host when left empty |
| server.httproute.metadata | object | `{"annotations":{},"labels":{}}` | Metadata configuration for the Burrito server HTTPRoute |
| server.httproute.parentRefs | list | `[]` | Gateways the HTTPRoute attaches to (required when httproute is enabled). Each entry follows the Gateway API parentRef schema (name, optionally namespace/sectionName) |
| server.httproute.rules | list | `[]` | HTTPRoute rules. Defaults to a single rule routing "/" (PathPrefix) to the burrito-server service when left empty |
| server.ingress | object | `{"enabled":false,"host":"burrito.example.com","ingressClassName":"nginx","metadata":{"annotations":{},"labels":{}},"tls":[]}` | Ingress configuration for the Burrito server |
| server.ingress.enabled | bool | `false` | Enable/Disable ingress creation for the Burrito server |
| server.ingress.host | string | `"burrito.example.com"` | Hostname for the Burrito server ingress |
| server.ingress.ingressClassName | string | `"nginx"` | Ingress class name to use for the Burrito server ingress |
| server.ingress.metadata | object | `{"annotations":{},"labels":{}}` | Metadata configuration for the Burrito server ingress |
| server.ingress.tls | list | `[]` | TLS configuration for the Burrito server ingress |
| server.metadata | object | `{"annotations":{},"labels":{"app.kubernetes.io/component":"server","app.kubernetes.io/name":"burrito-server"}}` | Metadata configuration for the Burrito server |
| server.service | object | `{"metadata":{"annotations":{},"labels":{}},"ports":[{"name":"http","port":80,"targetPort":"http"}],"trafficDistribution":""}` | Service configuration for the Burrito server |
| server.service.metadata | object | `{"annotations":{},"labels":{}}` | Metadata configuration for the Burrito server service |
| server.service.trafficDistribution | string | `""` | Server service traffic distribution policy |
| server.serviceAccount | object | `{"metadata":{"annotations":{},"labels":{}}}` | Service account configuration for the Burrito server deployment |
| server.serviceAccount.metadata | object | `{"annotations":{},"labels":{}}` | Metadata configuration for the Burrito server service account |
| tenants | list | `[]` | List of tenants to create to manage Terraform resources |

----------------------------------------------
Autogenerated from chart metadata using [helm-docs v1.14.2](https://github.com/norwoodj/helm-docs/releases/v1.14.2)
