***

<div align="center">
<b>Switcher GitOps</b><br>
GitOps Domain Snapshot Orchestrator for Switcher API
</div>

<div align="center">

[![Master CI](https://github.com/switcherapi/switcher-gitops/actions/workflows/master.yml/badge.svg?branch=master)](https://github.com/switcherapi/switcher-gitops/actions/workflows/master.yml)
[![Quality Gate Status](https://sonarcloud.io/api/project_badges/measure?project=switcherapi_switcher-gitops&metric=alert_status)](https://sonarcloud.io/dashboard?id=switcherapi_switcher-gitops)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Slack: Switcher-HQ](https://img.shields.io/badge/slack-@switcher/hq-blue.svg?logo=slack)](https://switcher-hq.slack.com/)

</div>

***

![Switcher API: Cloud-based Feature Flag API](https://github.com/switcherapi/switcherapi-assets/blob/master/logo/switcherapi_grey.png)

# About  
**Switcher GitOps** is Domain Snapshot Orchestrator for Switcher API. It allows you to manage your feature flags and configurations in a GitOps manner. It is a simple and easy way to manage your feature flags and configurations in a versioned manner.

# Features
- Manage Switchers with GitOps driven environment
- Multiple and independent Environments
- Auto Sync enables a fully integrated environment with Switcher API Management, Slack App and GitOps working simultaneously

# Integrated tests

This project does not run integration test without a valid GitOps repository settings.

In order to run integration tests locally, you need to create a valid GitOps repository with the following structure:

```bash
├── resources
│   ├── default.json (*)
```
(*) see a Switcher snapshot sample [here](https://github.com/switcherapi/switcher-gitops/blob/master/resources/default.json)

The next step is to set up PAT (Personal Access Token) for Switcher GitOps to access the repository. You can either create a fine-grained token with only the necessary permissions such as Content (Read and Write) and Metadata (Read) or use a personal token with full access.

Once you have the token, you can set it up in the `.env.test` environment file by including the following:
```bash
GIT_TOKEN=[YOUR_GIT_TOKEN]
GIT_REPO_URL=[YOUR_GIT_REPO_URL]
GIT_BRANCH=[YOUR_GIT_BRANCH]
```