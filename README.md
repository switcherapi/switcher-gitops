***

<div align="center">
<b>Switcher GitOps</b><br>
GitOps Domain Snapshot Orchestrator for Switcher API
</div>

<div align="center">

[![Master CI](https://github.com/switcherapi/switcher-gitops/actions/workflows/master.yml/badge.svg?branch=master)](https://github.com/switcherapi/switcher-gitops/actions/workflows/master.yml)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Slack: Switcher-HQ](https://img.shields.io/badge/slack-@switcher/hq-blue.svg?logo=slack)](https://switcher-hq.slack.com/)

</div>

***

![Switcher API: Cloud-based Feature Flag API](https://github.com/switcherapi/switcherapi-assets/blob/master/logo/switcherapi_grey.png)

# About  
**Switcher GitOps** is Domain Snapshot Orchestrator for Switcher API. It allows you to manage your feature flags and configurations in a GitOps manner. It is a simple and easy way to manage your feature flags and configurations in a versioned manner.

# Features
- Manage Switchers in a GitOps manner
- Multiple and Independent Environments
- Two-way Sync allow you to use the Switcher API Management, Slack App and GitOps at the same time

## Run Project

- Running `go run ./src/cmd/app/main.go`
- Testing `go test -coverpkg=./... -v`
- Building `go build -o ./bin/app ./src/cmd/app/main.go`