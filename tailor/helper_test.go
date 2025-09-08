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
			TailorDB: config.TailorDB{
				DeprecatedFeature: config.TailorDBDeprecatedFeature{
					Enabled:       true,
					AllowDraft:    false,
					AllowCELHooks: false,
				},
				LegacyPermission: config.LegacyPermission{
					Enabled:               true,
					AllowTypePermission:   false,
					AllowRecordPermission: false,
				},
			},
			Pipeline: config.Pipeline{
				InsecureAuthorization: config.InsecureAuthorization{
					Enabled: true,
				},
				StepLength: config.StepLength{
					Enabled: true,
					Max:     10,
				},
				DeprecatedFeature: config.PipelineDeprecatedFeature{
					Enabled:        true,
					AllowStateFlow: false,
					AllowDraft:     false,
				},
				LegacyScript: config.LegacyScript{
					Enabled: true,
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
	}
}
