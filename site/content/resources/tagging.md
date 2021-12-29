---
title: Sesame Tagging Policy
layout: page
---

This document describes Sesame's image tagging policy.

## Released versions

`ghcr.io/projectsesame/sesame:<SemVer>`

Sesame follows the [Semantic Versioning][1] standard for releases.
Each tag in the github.com/projectsesame/sesame repository has a matching image. eg. `ghcr.io/projectsesame/sesame:{{< param latest_version >}}`

`ghcr.io/projectsesame/sesame:v<major>.<minor>`

This tag will point to the latest available patch of the release train mentioned.
That is, it's `:latest` where you're guaranteed to not have a minor version bump.

### Latest

`ghcr.io/projectsesame/sesame:latest`

The `latest` tag follows the most recent stable version of Sesame.

## Development

`ghcr.io/projectsesame/sesame:main`

The `main` tag follows the latest commit to land on the `main` branch.

[1]: http://semver.org/
