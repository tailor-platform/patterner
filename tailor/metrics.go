package tailor

import (
	"errors"
	"math"

	tailorv1 "buf.build/gen/go/tailor-inc/tailor/protocolbuffers/go/tailor/v1"
)

const (
	pageSize = 100
)

type Metric struct {
	Key   string
	Name  string
	Value float64
	Unit  string
	Error error
}

func (c *Client) Metrics(resources *Resources) ([]Metric, error) {
	var metrics []Metric

	// Coverage Metrics
	coverage, err := c.Coverage(resources)
	if err != nil {
		return nil, err
	}
	var total, covered int
	for _, rc := range coverage {
		total += rc.TotalSteps
		covered += rc.CoveredSteps
	}
	var coverTotal float64
	if total > 0 {
		coverTotal = float64(float64(covered)/float64(total)) * 100
	} else {
		coverTotal = 0
	}
	metrics = append(metrics, Metric{
		Key:   "pipeline_resolver_step_coverage_percentage",
		Name:  "Pipeline resolver step coverage",
		Value: coverTotal,
		Unit:  "%",
	})

	// Lint Metrics
	warns, err := c.Lint(resources)
	if err != nil {
		return nil, err
	}
	metrics = append(metrics, Metric{
		Key:   "lint_warnings_total",
		Name:  "Total number of lint warnings",
		Value: float64(len(warns)),
		Unit:  "",
	})

	// Pipeline Metrics
	metrics = append(metrics, Metric{
		Key:   "pipelines_total",
		Name:  "Total number of Pipelines",
		Value: float64(len(resources.Pipelines)),
		Unit:  "",
	})
	resolversTotal := 0
	stepsTotal := 0
	graphQLStepsTotal := 0
	functionStepsTotal := 0
	executionPathsTotal := 0
	for _, p := range resources.Pipelines {
		resolversTotal += len(p.Resolvers)
		for _, r := range p.Resolvers {
			testsCount := 0
			stepsTotal += len(r.Steps)
			for _, s := range r.Steps {
				if s.Operation.Test != "" {
					testsCount++
				}
				switch s.Operation.Type {
				case tailorv1.PipelineResolver_OPERATION_TYPE_GRAPHQL:
					graphQLStepsTotal++
				case tailorv1.PipelineResolver_OPERATION_TYPE_FUNCTION:
					functionStepsTotal++
				}
			}
			executionPathsTotal += len(r.Steps) * int(math.Pow(2, float64(testsCount)))
		}
	}
	metrics = append(metrics, Metric{
		Key:   "pipeline_resolvers_total",
		Name:  "Total number of Pipeline resolvers",
		Value: float64(resolversTotal),
		Unit:  "",
	})
	metrics = append(metrics, Metric{
		Key:   "pipeline_resolver_steps_total",
		Name:  "Total number of Pipeline resolver steps",
		Value: float64(stepsTotal),
		Unit:  "",
	})
	metrics = append(metrics, Metric{
		Key:   "pipeline_resolver_graphql_steps_total",
		Name:  "Total number of Pipeline resolver GraphQL steps",
		Value: float64(graphQLStepsTotal),
		Unit:  "",
	})
	metrics = append(metrics, Metric{
		Key:   "pipeline_resolver_function_steps_total",
		Name:  "Total number of Pipeline resolver Function steps",
		Value: float64(functionStepsTotal),
		Unit:  "",
	})

	pathsMetic := Metric{
		Key:   "pipeline_resolver_execution_paths_total",
		Name:  "Total number of Pipeline resolver execution paths",
		Value: float64(executionPathsTotal),
		Unit:  "",
	}
	if executionPathsTotal < 0 {
		pathsMetic.Error = errors.New("overflow detected")
	}
	metrics = append(metrics, pathsMetic)

	// TailorDB Metrics
	metrics = append(metrics, Metric{
		Key:   "tailordbs_total",
		Name:  "Total number of TailorDBs",
		Value: float64(len(resources.TailorDBs)),
		Unit:  "",
	})
	typesTotal := 0
	fieldsTotal := 0
	for _, db := range resources.TailorDBs {
		typesTotal += len(db.Types)
		for _, t := range db.Types {
			fieldsTotal += len(t.Fields)
		}
	}
	metrics = append(metrics, Metric{
		Key:   "tailordb_types_total",
		Name:  "Total number of TailorDB types",
		Value: float64(typesTotal),
		Unit:  "",
	})
	metrics = append(metrics, Metric{
		Key:   "tailordb_type_fields_total",
		Name:  "Total number of TailorDB type fields",
		Value: float64(fieldsTotal),
		Unit:  "",
	})

	// StateFlow Metrics
	metrics = append(metrics, Metric{
		Key:   "stateflows_total",
		Name:  "Total number of StateFlows",
		Value: float64(len(resources.StateFlows)),
		Unit:  "",
	})

	return metrics, nil
}
