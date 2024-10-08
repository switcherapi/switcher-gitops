***

<div align="center">
<b>Switcher GitOps</b><br>
GitOps Domain Snapshot Orchestrator for Switcher API
</div>

<div align="center">

[![Master CI](https://github.com/switcherapi/switcher-gitops/actions/workflows/master.yml/badge.svg?branch=master)](https://github.com/switcherapi/switcher-gitops/actions/workflows/master.yml)
[![Quality Gate Status](https://sonarcloud.io/api/project_badges/measure?project=switcherapi_switcher-gitops&metric=alert_status)](https://sonarcloud.io/dashboard?id=switcherapi_switcher-gitops)
[![Known Vulnerabilities](https://snyk.io/test/github/switcherapi/switcher-gitops/badge.svg)](https://snyk.io/test/github/switcherapi/switcher-gitops)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Docker Hub](https://img.shields.io/docker/pulls/trackerforce/switcher-gitops.svg)](https://hub.docker.com/r/trackerforce/switcher-gitops)
[![Slack: Switcher-HQ](https://img.shields.io/badge/slack-@switcher/hq-blue.svg?logo=slack)](https://switcher-hq.slack.com/)

</div>

***

![Switcher API: Cloud-based Feature Flag API](https://github.com/switcherapi/switcherapi-assets/blob/master/logo/switcherapi_gitops.png)

# About  
**Switcher GitOps** is used to orchestrate Domain Snapshots for Switcher API. It allows managing feature flags and configurations lifecycle.

- Manages Switchers with GitOps workflow (repository as a source of truth)
- Repository synchronization allows integrated tools such as Switcher API Management and Switcher Slack App to work in sync
- Flexible settings allow you to define the best workflow for your organization
- Orchestrates accounts per Domain environments allowing seamless integration with any branching strategy

# Getting Started

## Using Swither API Cloud

Switcher GitOps is available as a cloud-hosted service. You can sign up for a free account at [Switcher API Cloud](https://cloud.switcherapi.com).

1. Create and Configure a new Domain
2. Select the Domain and click on the Menu toolbar
3. Under Integrations, select Switcher GitOps
4. Follow the instructions to set up the repository

## Self-hosted: Deploying to Kubernetes

### Requirements
- Kubernetes cluster
- Helm 3
- Switcher API & Switcher Management
- Git Token (read/write access) for the repository

Find detailed instructions on how to deploy Switcher GitOps to Kubernetes [here](https://github.com/switcherapi/helm-charts).

## Development: Deploying locally

### Requirements
- Docker & docker-compose
- Switcher API & Switcher Management
- Git Token (read/write access) for the repository

1. Configure Switcher API to allow Switcher GitOps to access the API<br>
Set SWITCHER_GITOPS_JWT_SECRET for Switcher API and SWITCHER_API_JWT_SECRET for Switcher GitOps.

2. [Start](https://github.com/switcherapi/switcher-api?tab=readme-ov-file#running-switcher-api-from-docker-composer-manifest-file) Switcher API and Switcher Management
3. Start Switcher GitOps `docker-compose -d up`<br>
You might need to remove mongodb setting from docker-compose.yml if launching the full Switcher API stack from step 2.

## Development: Running locally

### Requirements
- Go [check version in go.mod]
- MongoDB +7.0

1. Clone the repository
2. Configure the environment variables in the `.env.test` file
3. `make run:test` to start the application

# Integrated tests

Set up PAT (Personal Access Token) for Switcher GitOps to access the repository. You can either create a fine-grained token with only the necessary permissions such as Content (Read and Write) and Metadata (Read) or use a personal token with full access.

Once you have the token, you can set it up in the `.env.test` environment file by including the following:
```bash
GIT_USER=[YOUR_GIT_USER]
GIT_TOKEN=[YOUR_GIT_TOKEN]
GIT_TOKEN_READ_ONLY=[YOUR_GIT_TOKEN_READ_ONLY]
GIT_REPO_URL=[YOUR_GIT_REPO_URL]
GIT_BRANCH=[YOUR_GIT_BRANCH]
```