# Burrito UI

<img src="https://raw.githubusercontent.com/padok-team/burrito/main/ui/public/favicon.svg" alt="Burrito Logo" width="600" />

Web UI for [Burrito](https://github.com/padok-team/burrito).

## Getting started

1. Install [NodeJS](https://nodejs.org/en/download/) and [Yarn](https://yarnpkg.com).
2. Run `yarn install` to install local prerequisites.
3. Run `yarn dev` to launch the dev UI server.
4. Run `yarn build` to bundle static resources into the `./dist` directory.

## Build Docker production image

Run the following commands to build the Docker image:

```bash
TAG=latest # or any other tag
BURRITO_API_BASE_URL=https://burrito.example.cloud/api # or any other API base URL
docker build -t burrito-ui:$TAG --build-arg API_BASE_URL=$BURRITO_API_BASE_URL .
```
