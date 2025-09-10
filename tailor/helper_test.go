package tailor

import (
	"testing"

	"github.com/tailor-platform/patterner/config"
)

// createTestConfig creates a test configuration with default settings.
func createTestConfig(t *testing.T) *config.Config {
	t.Helper()
	return &config.Config{
		WorkspaceID: "test-workspace-id",
		Lint: config.Lint{
			Rules: config.Rules{
				TailorDB: config.TailorDB{
					DeprecatedFeature: config.TailorDBDeprecatedFeature{
						Enabled:               true,
						AllowDraft:            false,
						AllowCELHooks:         false,
						AllowTypePermission:   false,
						AllowRecordPermission: false,
					},
				},
				Pipeline: config.Pipeline{
					InsecureAuthorization: config.InsecureAuthorization{
						Enabled: true,
					},
					StepCount: config.StepCount{
						Enabled: true,
						Max:     10,
					},
					DeprecatedFeature: config.PipelineDeprecatedFeature{
						Enabled:        true,
						AllowStateFlow: false,
						AllowDraft:     false,
						AllowCELScript: false,
					},
					MultipleMutations: config.MultipleMutations{
						Enabled: true,
					},
					QueryBeforeMutation: config.QueryBeforeMutation{
						Enabled: true,
					},
				},
				StateFlow: config.StateFlow{
					DeprecatedFeature: config.StateFlowDeprecatedFeature{
						Enabled: true,
					},
				},
			},
		},
	}
}
