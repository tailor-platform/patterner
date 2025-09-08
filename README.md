# Patterner

`patterner` is a tool to analyze and present best practices (patterns) for Tailor Platform applications.

## Overview

Patterner is a command-line tool designed to help developers identify patterns and best practices within Tailor Platform workspaces. It provides linting capabilities and metrics collection to ensure your Tailor Platform application follows recommended patterns.

## Features

- **Lint** - Analyze resources in your workspace and identify potential issues or deviations from best practices
- **Metrics** - Collect and display metrics about resources in your workspace
- **Configuration** - Flexible configuration system with YAML-based settings

## Installation

```bash
go install github.com/tailor-platform/patterner@latest
```

## Usage

### Initialize Configuration

Create a `.patterner.yml` configuration file in your project:

```bash
patterner init
```

You can also specify a workspace ID during initialization:

```bash
patterner init -w YOUR_WORKSPACE_ID
```

### Lint Your Workspace

Check your workspace resources against configured patterns:

```bash
patterner lint
```

Or specify a workspace ID directly:

```bash
patterner lint -w YOUR_WORKSPACE_ID
```

### View Metrics

Display metrics about resources in your workspace:

```bash
patterner metrics
```

The metrics command outputs detailed JSON data about your workspace resources.

## Configuration

Patterner uses a `.patterner.yml` file for configuration. The configuration includes various lint rules for different Tailor Platform components:

```yaml
lint:
  pipeline:
    deprecatedFeature:
      enabled: true
    insecureAuthorization:
      enabled: false
    stepLength:
      enabled: true
      max: 30
    multipleMutations:
      enabled: true
    queryBeforeMutation:
      enabled: true
    legacyScript:
      enabled: true
  tailordb:
    deprecatedFeature:
      enabled: true
    legacyPermission:
      enabled: true
  stateflow:
    deprecatedFeature:
      enabled: true
```

### Lint Rules

#### Pipeline Rules
- **deprecatedFeature** - Check for deprecated features in pipelines
- **insecureAuthorization** - Detect insecure authorization patterns
- **stepLength** - Ensure pipeline steps don't exceed maximum length
- **multipleMutations** - Identify multiple mutations in a single operation
- **queryBeforeMutation** - Check for queries before mutations
- **legacyScript** - Identify legacy script/validation patterns and recommend modern hook alternatives
  - Detects `pre_validation` and recommends `pre_hook`
  - Detects `pre_script` and recommends `pre_hook`
  - Detects `post_script` and recommends `post_hook`
  - Detects `post_validation` and recommends `post_hook`
  - Enabled by default to promote migration to modern hook patterns

#### TailorDB Rules
- **deprecatedFeature** - Check for deprecated TailorDB features
- **legacyPermission** - Identify legacy permission patterns

#### Stateflow Rules
- **deprecatedFeature** - Check for deprecated Stateflow features

## Command Reference

### Global Flags

- `-w, --workspace-id string` - Workspace ID (can be set in configuration file)

### Commands

- `patterner init` - Initialize configuration file
- `patterner lint` - Lint workspace resources
- `patterner metrics` - Display workspace metrics

## License

MIT License

Copyright © 2025 Tailor Inc.
