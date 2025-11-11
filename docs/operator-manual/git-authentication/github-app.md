# GitHub App Authentication

## Create the GitHub App

You can create and register GitHub Apps in your personal GitHub account or in any GitHub organization where you have administrative access.

Follow the instructions in the GitHub documentation on [Creating a GitHub App](https://docs.github.com/en/apps/creating-github-apps/registering-a-github-app/registering-a-github-app). Populate the settings as follows:

- **GitHub App Name**: Choose a name for your GitHub App. For example, something featuring `burrito`.
- **Homepage URL**: Enter <https://docs.burrito.tf>.
- **Webhook**: Unselect Active. The app doesn't use this webhook events mechanism at the moment.
- **Permissions**: Configure the following **Repository Permissions**:
    - **Metadata:** Select Read-only.
    - **Contents:** Select Read-only.
    - **Pull requests:** Select Read & write. This is required to issue comments on pull requests.
- Where can this GitHub App be installed: Select **Any account**.

## Creating a custom badge for your GitHub App

You can create a custom badge for your GitHub App to display on your GitHub repository. Follow the instructions in the GitHub documentation on [Creating a custom badge for your GitHub App](https://docs.github.com/en/apps/creating-github-apps/registering-a-github-app/creating-a-custom-badge-for-your-github-app).

We suggest using the following one:

<p align="center"><img src="../../../assets/icon/burrito.png" width="200px" /></p>

## Install the GitHub App

Follow the instructions in the GitHub documentation on [Installing your own GitHub App](https://docs.github.com/en/apps/using-github-apps/installing-your-own-github-app), and note the following:

- For Repository access, select **Only select repositories**, and then select the repos you want to connect with Burrito.

## Get the Installation ID and App ID

You need the **Installation ID** and **App ID** to configure Burrito.

<!-- markdownlint-disable MD032 -->
1. Get the **Installation ID** from the URL of the installed app, such as:
  <p align="center"><img src="../../../assets/pr-mr-workflow/github_installation_id.png" /></p>
2. Get the **App ID** from the app's General tab.
  <p align="center"><img src="../../../assets/pr-mr-workflow/github_app_id.png" /></p>
<!-- markdownlint-enable MD032 -->

## Generate a private key

You need a private key for your GitHub app to configure Burrito.

- Follow the instructions in the GitHub documentation for [generating private keys for GitHub Apps](https://docs.github.com/en/apps/creating-github-apps/authenticating-with-a-github-app/managing-private-keys-for-github-apps#generating-private-keys)

- Save the private key file to your local machine. GitHub only stores the public portion of the key.

## Configure credentials with the GitHub App

Add the credentials of your newly created app to a repository-specific or shared credential secret.

### Repository-specific credentials example

```yaml
apiVersion: config.terraform.padok.cloud/v1alpha1
kind: TerraformRepository
metadata:
  name: my-repository
  namespace: burrito-project
spec:
  repository:
    url: https://github.com/owner/repo
  terraform:
    enabled: true
---
apiVersion: v1
kind: Secret
metadata:
  name: burrito-repo
  namespace: burrito-project
type: credentials.burrito.tf/repository
stringData:
  provider: github
  url: https://github.com/owner/repo
  webhookSecret: "my-webhook-secret"
  githubAppID: "123456"
  githubAppInstallationID: "12345678"
  githubAppPrivateKey: |
    -----BEGIN RSA PRIVATE KEY-----
    my-private-key
    -----END RSA PRIVATE KEY-----
```

### Shared credentials example

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: github-app-credentials
  namespace: burrito-system
  annotations:
    burrito.tf/allowed-tenants: "burrito-project"
type: credentials.burrito.tf/shared
stringData:
  provider: github
  url: https://github.com/owner
  webhookSecret: "my-webhook-secret"
  githubAppID: "123456"
  githubAppInstallationID: "12345678"
  githubAppPrivateKey: |
    -----BEGIN RSA PRIVATE KEY-----
    my-private-key
    -----END RSA PRIVATE KEY-----
```
