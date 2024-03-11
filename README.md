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
- Manage Switchers in a GitOps manner
- Multiple and Independent Environments
- Two-way Sync allow you to use the Switcher API Management, Slack App and GitOps at the same time

## Run Project

- Running<br>
    Windows: `$env:GO_ENV="test"; go run ./src/cmd/app/main.go`<br>
    Unit: `GO_ENV=test go run ./src/cmd/app/main.go`
- Testing `go test -coverpkg=./... -v`
- Coverage `go test -coverprofile="coverage.out" ./... && go tool cover -html="coverage.out"`
- Building `go build -o ./bin/app ./src/cmd/app/main.go`