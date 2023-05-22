# Configure the TerraformLayer to use private modules' repositories

If your stack use Terraform modules that are hosted on private repositories, you can configure the TerraformLayer to use a specific SSH key to clone the repository and a specific SSH known hosts file to verify the host.

To do so, you'll first need to create a **Secret** containing the SSH key and, if necessary, edit the `burrito-ssh-known-hosts` **ConfigMap** to add your known host key:

```yaml
apiVersion: v1
data:
  key: <YOUR_PRIVATE_SSH_KEY_BASE64_ENCODED>
immutable: false
kind: Secret
metadata:
  name: git-private-key–modules
  namespace: burrito
type: Opaque
---
apiVersion: v1
data:
  known_hosts: |-
    bitbucket.org ssh-rsa AAAAB3NzaC1yc2EAAAABIwAAAQEAubiN81eDcafrgMeLzaFPsw2kNvEcqTKl/VqLat/MaB33pZy0y3rJZtnqwR2qOOvbwKZYKiEO1O6VqNEBxKvJJelCq0dTXWT5pbO2gDXC6h6QDXCaHo6pOHGPUy+YBaGQRGuSusMEASYiWunYN0vCAI8QaXnWMXNMdFP3jHAJH0eDsoiGnLPBlBp4TNm6rYI74nMzgz3B9IikW4WVK+dc8KZJZWYjAuORU3jc1c/NPskD2ASinf8v3xnfXeukU0sJ5N6m5E8VLjObPEO+mN2t/FZTMZLiFqPWc/ALSqnMnnhwrNi2rbfg/rd/IpL8Le3pSBne8+seeFVBoGqzHM9yXw==
    github.com ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQCj7ndNxQowgcQnjshcLrqPEiiphnt+VTTvDP6mHBL9j1aNUkY4Ue1gvwnGLVlOhGeYrnZaMgRK6+PKCUXaDbC7qtbW8gIkhL7aGCsOr/C56SJMy/BCZfxd1nWzAOxSDPgVsmerOBYfNqltV9/hWCqBywINIR+5dIg6JTJ72pcEpEjcYgXkE2YEFXV1JHnsKgbLWNlhScqb2UmyRkQyytRLtL+38TGxkxCflmO+5Z8CSSNY7GidjMIZ7Q4zMjA2n1nGrlTDkzwDCsw+wqFPGQA179cnfGWOWRVruj16z6XyvxvjJwbz0wQZ75XK5tKSb7FNyeIEs4TT4jk+S4dhPeAUC5y+bDYirYgM4GC7uEnztnZyaVWQ7B381AK4Qdrwt51ZqExKbQpTUNn+EjqoTwvqNj4kqx5QUCI0ThS/YkOxJCXmPUWZbhjpCg56i+2aB6CmK2JGhn57K5mj0MNdBXA4/WnwH6XoPWJzK5Nyu2zB3nAZp+S5hpQs+p1vN1/wsjk=
    gitlab.com ecdsa-sha2-nistp256 AAAAE2VjZHNhLXNoYTItbmlzdHAyNTYAAAAIbmlzdHAyNTYAAABBBFSMqzJeV9rUzU4kWitGjeR4PWSa29SPqJ1fVkhtj3Hw9xjLVXVYrU9QlYWrOLXBpQ6KWjbjTDTdDkoohFzgbEY=
    gitlab.com ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIAfuCHKVTjquxvt6CM6tdG4SLp1Btn/nOeHHE5UOzRdf
    gitlab.com ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQCsj2bNKTBSpIYDEGk9KxsGh3mySTRgMtXL583qmBpzeQ+jqCMRgBqB98u3z++J1sKlXHWfM9dyhSevkMwSbhoR8XIq/U0tCNyokEi/ueaBMCvbcTHhO7FcwzY92WK4Yt0aGROY5qX2UKSeOvuP4D6TPqKF1onrSzH9bx9XUf2lEdWT/ia1NEKjunUqu1xOB/StKDHMoX4/OKyIzuS0q/T1zOATthvasJFoPrAjkohTyaDUz2LN5JoH839hViyEG82yB+MjcFV5MU3N1l1QL3cVUCh93xSaua1N85qivl+siMkPGbO5xR/En4iEY6K2XPASUEMaieWVNTRCtJ4S8H+9
    github.com ecdsa-sha2-nistp256 AAAAE2VjZHNhLXNoYTItbmlzdHAyNTYAAAAIbmlzdHAyNTYAAABBBEmKSENjQEezOmxkZMy7opKgwFB9nkt5YRrYMjNuG5N87uRgg6CLrbo5wAdT/y6v0mKV0U2w0WZ2YB/++Tpockg=
    github.com ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIOMqqnkVzrm0SdG6UOoqKLsabgH5C9okWi0dh2l9GKJl
    git.yourcompany.ai <HOST_KEY>
kind: ConfigMap
metadata:
  labels:
    app.kubernetes.io/instance: in-cluster-burrito
    app.kubernetes.io/name: burrito-ssh-known-hosts
    app.kubernetes.io/part-of: burrito
  name: burrito-ssh-known-hosts
  namespace: burrito
```

**Hint**: You can find your host's key in your own `~/.ssh/known_hosts` file.

Once the **Secret** created and the **ConfigMap** edited, you can update your **TerraformLayer** to use the new SSH key:

```yaml
apiVersion: config.terraform.padok.cloud/v1alpha1
kind: TerraformLayer
metadata:
  name: terragrunt-private-module-ssh
spec:
  terraform:
    version: "1.3.1"
    terragrunt:
      enabled: true
      version: "0.45.4"
  remediationStrategy: autoApply
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
