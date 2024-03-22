# Caching Terraform Providers

By caching Terraform providers, Burrito can avoid downloading them from outside the cluster every time a runner initializes a terraform layer. This can significantly reduce the ingress traffic to the infrastructure running Burrito.

The Burrito helm chart is packaged with [Hermitcrab](https://github.com/seal-io/hermitcrab) which leverages the [Provider Network Mirror Protocol](https://developer.hashicorp.com/terraform/internals/provider-network-mirror-protocol) from Terraform to cache providers.

## 1. Activate Hermitcrab on Burrito

Using the Burrito helm chart, set the `hermitcrab.enabled` value to `true` to deploy Hermitcrab along with Burrito.  
As the Provider Network Mirror Protocol only allows HTTPS traffic, it is required to update Hermitcrab's TLS configuration.

### Option 1: Mount a custom certificate

Mount your custom certificate to `/etc/hermitcrab/tls/tls.crt` and private key to `/etc/hermitcrab/tls/tls.key` by using the `hermitcrab.deployment.extraVolumeMounts` and `hermitcrab.deployment.extraVolumeMounts` values.  
(You can change the certificate and key paths with the `SERVER_TLS_CERT_FILE` and `SERVER_TLS_PRIVATE_KEY_FILE` environment variables)

### Option 2: Use cert-manager

Assuming that cert-manager is installed on your cluster, set the `tls.certManager.use` value to true, and specify an Issuer or ClusterIssuer with `tls.certManager.certificate.issuer.kind` and `tls.certManager.certificate.issuer.name` values.
This will create a [Certificate](https://cert-manager.io/docs/usage/certificate/) custom ressource that will be used to ensure TLS between runners and Hermitcrab.

//TODO: Update the runner code to automatically mount the cert manager certificate in the pod with a runner.hermitcrab.certificate value

## 2. Configure the runners to use provider caching

//TODO: Update the runner code to automatically create a .terraformrc file in $HOME with the following configuration:
provider_installation {
  network_mirror {
   url = "https://burrito-hermitcrab.{namespace}.svc.cluster.local/v1/providers/"
  }
}
