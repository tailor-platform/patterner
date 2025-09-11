# Patterner

`patterner` is a tool to analyze and present best practices (patterns) for [Tailor Platform](https://www.tailor.tech/) applications.

## Overview

Patterner is a command-line tool designed to help developers identify patterns and best practices within Tailor Platform workspaces. It provides linting capabilities and metrics collection to ensure your Tailor Platform application follows recommended patterns.

## Features

- **Lint** - Analyze resources in your workspace and identify potential issues or deviations from best practices
- **Metrics** - Collect and display metrics about resources in your workspace
- **Coverage** - Analyze and display pipeline resolver step coverage to monitor execution quality
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

**Note:** The lint command requires a Tailor Platform access token. Set the `TAILOR_TOKEN` environment variable:

```bash
# Using tailorctl to get access token
env TAILOR_TOKEN=$(tailorctl auth get accessToken) patterner lint

# Or set the token directly
export TAILOR_TOKEN=your_access_token
patterner lint
```

### View Metrics

Display metrics about resources in your workspace:

```bash
patterner metrics
```

**Note:** The metrics command requires a Tailor Platform access token. Set the `TAILOR_TOKEN` environment variable:

```bash
# Using tailorctl to get access token
env TAILOR_TOKEN=$(tailorctl auth get accessToken) patterner metrics

# Or set the token directly
export TAILOR_TOKEN=your_access_token
patterner metrics
```

The metrics command displays metrics in a table format. Use the `--out-octocov-path` option to output metrics in octocov custom metrics format.

#### Metrics Options

- `--since, -s` (default: "30min") - Analyze execution results since the specified time period
- `--out-octocov-path` - Output metrics in octocov custom metrics format to the specified file
- `--with-lint-warnings` (default: false) - Display lint warnings along with metrics output
- `--with-coverage-full-report` (default: false) - Display detailed pipeline resolver step coverage along with metrics output

##### Option Details

**--with-coverage-full-report vs --with-lint-warnings:**
- `--with-coverage-full-report`: Displays detailed pipeline resolver step coverage information for execution quality monitoring
  - Shows coverage percentage and executed steps count for each resolver
  - Format: `Coverage Rate% [Executed Steps/Total Steps] Resolver Name`
- `--with-lint-warnings`: Displays detailed lint warnings list for code quality monitoring
  - Shows specific lint issues and recommendations for code improvements

Both options can be used together to provide comprehensive analysis combining execution quality and code quality insights.

#### Usage Examples

```bash
# Basic metrics (last 30 minutes)
patterner metrics

# Metrics for the past hour
patterner metrics --since 1hour

# Display metrics with lint warnings
patterner metrics --with-lint-warnings

# Display metrics with coverage details
patterner metrics --with-coverage-full-report

# Display metrics with both coverage and lint warnings
patterner metrics --with-coverage-full-report --with-lint-warnings

# Output metrics to octocov custom metrics file
patterner metrics --out-octocov-path metrics.json

# Comprehensive analysis with all options
patterner metrics --since 24hours --out-octocov-path metrics.json --with-coverage-full-report --with-lint-warnings
```

**Implementation Notes:**
```
<!-- Note: Current implementation has a typo in variable name 'lint_warnigns_total'
     and potential variable reference issue in metrics.go that should be addressed -->
```

### View Coverage

Display pipeline resolver step coverage for your workspace:

```bash
patterner coverage
```

**Note:** The coverage command requires a Tailor Platform access token. Set the `TAILOR_TOKEN` environment variable:

```bash
# Using tailorctl to get access token
env TAILOR_TOKEN=$(tailorctl auth get accessToken) patterner coverage

# Or set the token directly
export TAILOR_TOKEN=your_access_token
patterner coverage
```

The coverage command analyzes pipeline resolver step execution and displays coverage percentages to help monitor pipeline execution quality.

#### Coverage Options

- `--since, -s` (default: "30min") - Analyze execution results since the specified time period
- `--full-report, -f` (default: false) - Display detailed coverage report including per-resolver breakdown
- `--workspace-id` - Target workspace ID to analyze

#### Usage Examples

```bash
# Basic coverage analysis (last 30 minutes)
patterner coverage

# Coverage for the past hour
patterner coverage --since 1hour

# Detailed coverage report
patterner coverage --full-report

# Detailed report for the past 24 hours
patterner coverage --since 24hours --full-report
```

#### Output Format

**Standard output:**

```
Pipeline Resolver Step Coverage 75.0% [3/4]
```

**Detailed report:**
Shows coverage breakdown per resolver in addition to overall coverage statistics.

#### Available Metrics

The following metrics are collected and displayed:

**Pipeline Metrics:**

- `pipelines_total` - Total number of pipelines (Unit: count)
- `pipeline_resolvers_total` - Total number of pipeline resolvers (Unit: count)
- `pipeline_resolver_steps_total` - Total number of pipeline resolver steps (Unit: count)
- `pipeline_resolver_execution_paths_total` - Total number of pipeline resolver execution paths (Unit: count)
  - Calculation: Based on the number of steps and tests in each resolver (steps \* 2^tests)
  - Includes overflow detection: Reports error if negative values are encountered
  - Used to understand the total number of execution paths based on testable step combinations

**TailorDB Metrics:**

- `tailordbs_total` - Total number of TailorDBs (Unit: count)
- `tailordb_types_total` - Total number of TailorDB types (Unit: count)
- `tailordb_type_fields_total` - Total number of TailorDB type fields (Unit: count)

**StateFlow Metrics:**

- `stateflows_total` - Total number of StateFlows (Unit: count)

**Coverage Metrics:**

- `pipeline_resolver_step_coverage_percentage` - Pipeline resolver step coverage (Unit: %)
  - Calculation: (covered steps / total steps) * 100
  - Provides percentage representation of pipeline resolver step execution coverage

**Lint Metrics:**

- `lint_warnings_total` - Total number of lint warnings (Unit: count)
  - Calculation: Number of warnings returned from the lint function
  - Helps monitor code quality and adherence to best practices

## Configuration

Patterner uses a `.patterner.yml` file for configuration. The configuration includes various lint rules for different Tailor Platform components:

```yaml
workspaceID: xxxxxxxXXxxxxxxxxxxxxx
lint:
  acceptable: 5
  rules:
    pipeline:
      deprecatedFeature:
        enabled: true
        allowCELScript: false
        allowDraft: false
        allowStateFlow: false
      insecureAuthorization:
        enabled: false
      stepCount:
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
    stateflow:
      deprecatedFeature:
        enabled: true
```

### Lint Configuration

#### Acceptable Warnings

- **acceptable** - Set the maximum number of acceptable lint warnings (default: 0)
  - When the number of warnings exceeds this value, the lint command will exit with a failure status
  - This allows you to gradually improve code quality by setting a reasonable warning threshold
  - Example: `acceptable: 5` allows up to 5 warnings before failing

### Lint Rules

#### Pipeline Rules

- **deprecatedFeature** - Identify deprecated features and promote modern alternatives
  - `enabled` (default: true) - Enable/disable deprecated feature detection
  - `allowCELScript` (default: false) - Allow CEL script usage in pipelines
  - `allowDraft` (default: false) - Allow draft resources in pipeline configurations
  - `allowStateFlow` (default: false) - Allow StateFlow resources in pipeline configurations
  - Detects deprecated patterns and recommends modern Pipeline alternatives
    - https://docs.tailor.tech/reference/service-lifecycle-policy
  - Enabled by default to promote migration away from deprecated features
- **insecureAuthorization** - Detect insecure authorization patterns
- **stepCount** - Ensure pipeline steps don't exceed maximum count
- **multipleMutations** - Identify multiple mutations in a single operation
- **queryBeforeMutation** - Check for queries before mutations

#### TailorDB Rules

- **deprecatedFeature** - Identify deprecated TailorDB features and promote modern alternatives
  - `enabled` (default: true) - Enable/disable deprecated feature detection
  - `allowDraft` (default: false) - Allow draft resources in TailorDB configurations
  - `allowCELHooks` (default: false) - Allow CEL hook usage in TailorDB configurations
  - Detects deprecated patterns and recommends modern TailorDB alternatives
    - https://docs.tailor.tech/reference/service-lifecycle-policy
  - Enabled by default to promote migration away from deprecated features

#### StateFlow Rules

- **deprecatedFeature** - Identify deprecated StateFlow features and promote modern alternatives
  - `enabled` (default: true) - Enable/disable deprecated feature detection
  - Detects deprecated StateFlow patterns and recommends modern alternatives
    - https://docs.tailor.tech/reference/service-lifecycle-policy
  - Enabled by default to promote migration away from deprecated features

## Command Reference

### Global Flags

- `-w, --workspace-id string` - Workspace ID (can be set in configuration file)

### Commands

- `patterner init` - Initialize configuration file
- `patterner lint` - Lint workspace resources
- `patterner metrics` - Display workspace metrics
  - `--since, -s` (default: "30min") - Analyze execution results since the specified time period
  - `--out-octocov-path` - Output metrics in octocov custom metrics format to the specified file
  - `--with-lint-warnings` - Display lint warnings along with metrics output
  - `--with-coverage-full-report` - Display detailed pipeline resolver step coverage along with metrics output
- `patterner coverage` - Display pipeline resolver step coverage

---

Copyright © 2025 Tailor Inc.
