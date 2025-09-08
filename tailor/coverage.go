package tailor

import (
	"encoding/json"
)

type ResolverCoverage struct {
	PipelineNamespaceName string
	Name                  string
	TotalSteps            int
	CoveredSteps          int
	Steps                 []*StepCoverage
}

type StepCoverage struct {
	Name  string
	Count int
}

func (c *Client) Coverage(resources *Resources) ([]*ResolverCoverage, error) {
	var coverages []*ResolverCoverage
	for _, p := range resources.Pipelines {
		for _, r := range p.Resolvers {
			rc := &ResolverCoverage{
				PipelineNamespaceName: p.NamespaceName,
				Name:                  r.Name,
				TotalSteps:            len(r.Steps),
				CoveredSteps:          0,
			}
			var stepNames []string
			for _, s := range r.Steps {
				stepNames = append(stepNames, s.Name)
				rc.Steps = append(rc.Steps, &StepCoverage{
					Name:  s.Name,
					Count: 0,
				})
			}
			if len(r.ExecutionResults) == 0 {
				coverages = append(coverages, rc)
				continue
			}
			for _, result := range r.ExecutionResults {
				if result.Context == nil {
					// no branch steps
					for i, stepName := range stepNames {
						rc.Steps[i].Count++
						if result.LastPipelineName == stepName {
							break
						}
					}
					continue
				}
				steps, ok := result.Context.Fields["pipeline"]
				if !ok {
					continue
				}
				b, err := json.Marshal(steps)
				if err != nil {
					return nil, err
				}
				var m map[string]any
				if err := json.Unmarshal(b, &m); err != nil {
					return nil, err
				}
				for i, stepName := range stepNames {
					if _, ok := m[stepName]; ok {
						rc.Steps[i].Count++
					}
				}
			}
			for _, s := range rc.Steps {
				if s.Count > 0 {
					rc.CoveredSteps++
				}
			}
			coverages = append(coverages, rc)
		}
	}

	return coverages, nil
}
