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
      allowCELScript: false
      allowDraft: false
      allowStateFlow: false
    insecureAuthorization:
      enabled: false
    stepLength:
      enabled: true
      max: 30
    multipleMutations:
      enabled: true
    queryBeforeMutation:
      enabled: true
  tailordb:
    deprecatedFeature:
      enabled: true
      allowDraft: false
      allowCELHooks: false
    legacyPermission:
      enabled: true
  stateflow:
    deprecatedFeature:
      enabled: true
```

### Lint Rules

#### Pipeline Rules
- **deprecatedFeature** - Identify deprecated features including legacy script/validation patterns and recommend modern alternatives
  - `enabled` (default: true) - Enable/disable deprecated feature detection
  - `allowCELScript` (default: false) - Allow CEL script usage in pipelines
  - `allowDraft` (default: false) - Allow draft resources in pipeline configurations
  - `allowStateFlow` (default: false) - Allow StateFlow resources in pipeline configurations
  - Detects legacy script patterns (`pre_validation`, `pre_script`, `post_script`, `post_validation`) and recommends modern hook alternatives (`pre_hook`, `post_hook`)
  - Enabled by default to promote migration away from deprecated features
- **insecureAuthorization** - Detect insecure authorization patterns
- **stepLength** - Ensure pipeline steps don't exceed maximum length
- **multipleMutations** - Identify multiple mutations in a single operation
- **queryBeforeMutation** - Check for queries before mutations

#### TailorDB Rules
- **deprecatedFeature** - Identify deprecated TailorDB features and promote modern alternatives
  - `enabled` (default: true) - Enable/disable deprecated feature detection
  - `allowDraft` (default: false) - Allow draft resources in TailorDB configurations
  - `allowCELHooks` (default: false) - Allow CEL hook usage in TailorDB configurations
  - Detects deprecated patterns and recommends modern TailorDB alternatives
  - Enabled by default to promote migration away from deprecated features
- **legacyPermission** - Identify legacy permission patterns

#### StateFlow Rules
- **deprecatedFeature** - Identify deprecated StateFlow features and promote modern alternatives
  - `enabled` (default: true) - Enable/disable deprecated feature detection
  - Detects deprecated StateFlow patterns and recommends modern alternatives
  - Enabled by default to promote migration away from deprecated features

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
