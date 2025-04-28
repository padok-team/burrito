# Manage SSH known hosts

## Defaults

By default, we provide a list of known hosts with public repositories:

- Azure
- Bitbucket
- GitHub
- Gitlab
- Visual Studio

## Override known hosts

If you need to provide your own keys for other repositories, you can override the default value in the chart with:

```yaml
global:
  sshKnownHosts: |-
    git.domain.com ssh-ed25519 AAAAC3Nxxx
    git.domain.com ssh-rsa AAAAB3Nxxx
    git.domain.com ecdsa-sha2-nistp256 AAAAE2Vxxx
```

To get those keys, you can run: `ssh-keyscan git.domain.com 2>&1| grep -vE '^#'`
