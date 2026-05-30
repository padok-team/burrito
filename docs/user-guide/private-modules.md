# Configure the TerraformLayer to use private modules' repositories

If your stack use Terraform modules that are hosted on private repositories, you can configure the `TerraformLayer` to be able to use those private modules by [configuring the `overrideRunnerSpec` in your resource definition](./override-runner.md).

## The layer uses a private module with HTTPS

### 1. Create a secret containing a .git-credentilas file

Create a Kubernetes Secret which looks like the following:

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: git-credentials
stringData:
  .git-credentials: |
    https://<username>:<password | access_token>@github.com
```

!!! info
    You can replace `github.com` with `gitlab.com` or any GitHub or GitLab URL.

### 2. Create a ConfigMap for configuring the git agent

Create a Kubernetes ConfigMap which looks like the following:

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: gitconfig
data:
  .gitconfig: |
    [credential]
        helper = store
```

### 3. Mount those configurations' files in the runners' configuration

You need to mount this Secret and ConfigMap as file with some VolumeMounts:

```yaml
apiVersion: config.terraform.padok.cloud/v1alpha1
kind: TerraformLayer
metadata:
  name: terragrunt-private-module
spec:
  terraform:
    enabled: true
    version: "1.3.1"
  terragrunt:
    enabled: true
    version: "0.45.4"
  remediationStrategy:
    autoApply: true
  path: "terragrunt/random-pets-private-module/test"
  branch: main
  repository:
    name: burrito
    namespace: burrito
  overrideRunnerSpec:
    env:
    volumes:
    - name: gitconfig
      configMap:
        name: gitconfig
    - name: git-credentials
      secret:
        secretName: git-credentials
    volumeMounts:
    - name: gitconfig
      mountPath: /home/burrito/.gitconfig
      subPath: .gitconfig
    - name: git-credentials
      mountPath: /home/burrito/.git-credentials
      subPath: .git-credentials
```

## The layer uses a private module with SSH

### 1. Create a Secret with a SSH private key which can pull the modules' repositories

Create a Kubernetes Secret which looks like the following:

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: git-private-key-modules
  namespace: burrito
type: Opaque
stringSata:
  key: |
    -----BEGIN OPENSSH PRIVATE KEY-----
    ...
    -----END OPENSSH PRIVATE KEY-----
```

!!! info
    You can update the Kubernetes ConfigmMap `burrito-ssh-known-hosts` to add others known hosts. See [Manage SSH known hosts](./ssh-known-hosts.md).

### 2. Mount this Secret in your runner spec

You need to mount this Secret as a volume and add a `GIT_SSH_COMMAND` environements variables:

```yaml
apiVersion: config.terraform.padok.cloud/v1alpha1
kind: TerraformLayer
metadata:
  name: terragrunt-private-module-ssh
spec:
  terraform:
    version: "1.3.1"
    enabled: true
  terragrunt:
    enabled: true
    version: "0.45.4"
  remediationStrategy:
    autoApply: true
  path: "terragrunt/random-pets-private-module-ssh/test"
  branch: main
  repository:
    name: burrito
    namespace: burrito
  overrideRunnerSpec:
    env:
    - name: GIT_SSH_COMMAND
      value: ssh -i /home/burrito/.ssh/key
    volumes:
    - name: private-key
      secret:
        secretName: private-key-ssh-module
    volumeMounts:
    - name: private-key
      mountPath: /home/burrito/.ssh/key
      subPath: key
      readOnly: true
```

As you can see, we added a new `overrideRunnerSpec` field to the `TerraformLayer` spec. This field allows you to override the default runner pod spec.
In this case, we added a new volume and a new environment variable to the runner pod spec:

- The volume is a secret volume that contains the SSH key we created earlier
- The environment variable is used to tell git to use the SSH key we added to the runner pod

## The layer uses a private Terraform provider registry

If your stack uses providers from a private Terraform registry, such as a JFrog Artifactory registry, mount Terraform CLI configuration in the runner pod with `overrideRunnerSpec`.

!!! warning
    Do not enable [Hermitcrab provider caching](../operator-manual/provider-caching.md) for this layer unless you also merge the private registry configuration into the Terraform CLI config generated for Hermitcrab. When Hermitcrab is enabled, Burrito sets `TF_CLI_CONFIG_FILE` in the runner to point to the network mirror configuration, which can override the custom CLI config shown below.

### 1. Create a Secret with Terraform CLI configuration

Create a Kubernetes Secret containing the Terraform CLI configuration file. Replace `terraform-registry.example.com` and the token with the hostname and token for your private registry.

The hostname in the `credentials` block must match the hostname used in the provider `source` address.

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: terraform-cli-config
type: Opaque
stringData:
  private-provider.tfrc: |
    credentials "terraform-registry.example.com" {
      token = "<registry-token>"
    }

    provider_installation {
      direct {}
    }
```

The minimal `provider_installation` configuration keeps the default direct provider installation behavior.

### 2. Mount the file in the runner configuration

Mount the file in the runner home directory and point `TF_CLI_CONFIG_FILE` to the mounted `*.tfrc` file.

```yaml
apiVersion: config.terraform.padok.cloud/v1alpha1
kind: TerraformLayer
metadata:
  name: private-provider-registry
spec:
  terraform:
    enabled: true
    version: "1.6.6"
  path: "terraform/private-provider"
  branch: main
  repository:
    name: burrito
    namespace: burrito
  overrideRunnerSpec:
    env:
    - name: TF_CLI_CONFIG_FILE
      value: /home/burrito/private-provider.tfrc
    volumes:
    - name: terraform-cli-config
      secret:
        secretName: terraform-cli-config
    volumeMounts:
    - name: terraform-cli-config
      mountPath: /home/burrito/private-provider.tfrc
      subPath: private-provider.tfrc
      readOnly: true
```

Terraform will use `/home/burrito/private-provider.tfrc` as its CLI configuration during `terraform init`, including the private registry credentials and provider installation configuration.
