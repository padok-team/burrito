# Caching Terraform Providers

By caching Terraform providers, Burrito can avoid downloading them from outside the cluster every time a runner initializes a Terraform layer. This can significantly reduce the ingress traffic to the infrastructure running Burrito.

The Burrito Helm chart is packaged with [Hermitcrab](https://github.com/seal-io/hermitcrab), which leverages the [Provider Network Mirror Protocol](https://developer.hashicorp.com/terraform/internals/provider-network-mirror-protocol) from Terraform to cache providers.

## 1. Activate Hermitcrab on Burrito

Hermitcrab is available to use with Burrito when using the Helm chart.  
Set the `config.burrito.hermitcrab` parameter to true in your values file to activate Hermitcrab.

As the Provider Network Mirror Protocol only supports HTTPS traffic, it is required to provide Burrito runners & the Hermitcrab server with some TLS configuration. By default, the Helm chart expects a secret named `burrito-hermitcrab-tls` to contain TLS configuration: `ca.crt`, `tls.crt`, and `tls.key`.

### Option 1: Use Cert-Manager

The Helm chart is packaged with Cert-Manager configuration to use for Burrito/Hermitcrab TLS encryption.
Assuming that Cert-Manager is installed on your cluster, set the `hermitcrab.tls.certmanager.use` parameter to `true`. This setting adds a Cert-Manager Certificate resource to be used with Burrito.  
Provide Certificate spec with the `hermitcrab.tls.certmanager.spec` value. You **must** set the `secretName` value to the same value specified in `config.burrito.hermitcrab.certificateSecretName` (default `burrito-hermitcrab-tls`)

#### Example configuration with a self-signed issuer

Deploy Cert-Manager resources to generate self-signed certificates:

```yaml
apiVersion: cert-manager.io/v1
kind: ClusterIssuer
metadata:
  name: selfsigned-issuer
spec:
  selfSigned: {}
---
apiVersion: cert-manager.io/v1
kind: Certificate
metadata:
  name: my-selfsigned-ca
  namespace: cert-manager
spec:
  isCA: true
  commonName: my-selfsigned-ca
  secretName: root-secret
  privateKey:
    algorithm: ECDSA
    size: 256
  issuerRef:
    name: selfsigned-issuer
    kind: ClusterIssuer
    group: cert-manager.io
---
apiVersion: cert-manager.io/v1
kind: ClusterIssuer
metadata:
  name: my-ca-issuer
spec:
  ca:
    secretName: root-secret
```

Update the Helm chart values to create a self-signed certificate:

```yaml
config:
  burrito:
    hermitcrab:
      enabled: true
...
hermitcrab:
  tls:
    certManager:
      use: true
      certificate:
        spec:
          secretName: burrito-hermitcrab-tls
          commonName: burrito-hermitcrab.burrito.svc.cluster.local
          dnsNames:
            - burrito-hermitcrab.burrito.svc.cluster.local
          issuerRef:
            name: my-ca-issuer
            kind: ClusterIssuer
```

Burrito runners should now use Hermitcrab as a network mirror for caching providers.

### Option 2: Mount a custom certificate

If Hermitcrab is activated using the Helm chart, Burrito expects a secret named `burrito-hermitcrab-tls` to contain TLS configuration: `ca.crt`, `tls.crt`, and `tls.key`.
Assuming that Cert-Manager is installed on your cluster, set the `tls.certManager.use` value to true and specify an Issuer or ClusterIssuer with `tls.certManager.certificate.issuer.kind` and `tls.certManager.certificate.issuer.name` values.
This will create a [Certificate](https://cert-manager.io/docs/usage/certificate/) custom resource that will be used to ensure TLS between runners and Hermitcrab.

#### Server side

Mount your custom certificate to `/etc/hermitcrab/tls/tls.crt` and the private key to `/etc/hermitcrab/tls/tls.key` by using the `hermitcrab.deployment.extraVolumeMounts` and `hermitcrab.deployment.extraVolumeMounts` values.
Check out [the Hermitcrab documentation](https://github.com/seal-io/hermitcrab/blob/main/README.md#usage) for more information about injecting TLS Configuration.

#### Runner side

If Hermitcrab is activated using the Helm chart, the Burrito controller expects a secret named `burrito-hermitcrab-tls` to contain client TLS configuration in the `ca.crt` key. This private certificate will be trusted by Burrito runners.
