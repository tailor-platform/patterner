package tailor

import "math"

const (
	pageSize = 100
)

type Metric struct {
	Name        string
	Description string
	Value       float64
}

func (c *Client) Metrics(resources *Resources) ([]Metric, error) {
	var metrics []Metric

	// Pipeline Metrics
	metrics = append(metrics, Metric{
		Name:        "pipelines_total",
		Description: "Total number of pipelines",
		Value:       float64(len(resources.Pipelines)),
	})
	resolversTotal := 0
	stepsTotal := 0
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
			}
			executionPathsTotal += len(r.Steps) * int(math.Pow(2, float64(testsCount)))
		}
	}
	metrics = append(metrics, Metric{
		Name:        "pipeline_resolvers_total",
		Description: "Total number of pipeline resolvers",
		Value:       float64(resolversTotal),
	})
	metrics = append(metrics, Metric{
		Name:        "pipeline_resolver_steps_total",
		Description: "Total number of pipeline resolver steps",
		Value:       float64(stepsTotal),
	})
	metrics = append(metrics, Metric{
		Name:        "pipeline_resolver_execution_paths_total",
		Description: "Total number of pipeline resolver execution paths",
		Value:       float64(executionPathsTotal),
	})

	// TailorDB Metrics
	metrics = append(metrics, Metric{
		Name:        "tailordbs_total",
		Description: "Total number of TailorDBs",
		Value:       float64(len(resources.TailorDBs)),
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
		Name:        "tailordb_types_total",
		Description: "Total number of TailorDB types",
		Value:       float64(typesTotal),
	})
	metrics = append(metrics, Metric{
		Name:        "tailordb_type_fields_total",
		Description: "Total number of TailorDB type fields",
		Value:       float64(fieldsTotal),
	})

	// StateFlow Metrics
	metrics = append(metrics, Metric{
		Name:        "stateflows_total",
		Description: "Total number of StateFlows",
		Value:       float64(len(resources.StateFlows)),
	})

	return metrics, nil
}
