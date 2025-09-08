package tailor

import (
	"reflect"
	"testing"

	tailorv1 "buf.build/gen/go/tailor-inc/tailor/protocolbuffers/go/tailor/v1"
	"google.golang.org/protobuf/types/known/structpb"
)

func TestClient_Coverage(t *testing.T) {
	tests := []struct {
		name      string
		resources *Resources
		want      []*ResolverCoverage
		wantErr   bool
	}{
		{
			name: "empty resources",
			resources: &Resources{
				Pipelines: []*Pipeline{},
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "single pipeline with resolver but no execution results",
			resources: &Resources{
				Pipelines: []*Pipeline{
					{
						NamespaceName: "test-namespace",
						Resolvers: []*PipelineResolver{
							{
								Name: "test-resolver",
								Steps: []*PipelineStep{
									{Name: "step1"},
									{Name: "step2"},
								},
								ExecutionResults: []*tailorv1.PipelineResolverExecutionResult{},
							},
						},
					},
				},
			},
			want: []*ResolverCoverage{
				{
					PipelineNamespaceName: "test-namespace",
					Name:                  "test-resolver",
					TotalSteps:            2,
					CoveredSteps:          0,
					Steps: []*StepCoverage{
						{Name: "step1", Count: 0},
						{Name: "step2", Count: 0},
					},
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Client{}
			got, err := c.Coverage(tt.resources)
			if (err != nil) != tt.wantErr {
				t.Errorf("Client.Coverage() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			// Handle nil vs empty slice comparison
			if got == nil && tt.want == nil {
				return
			}
			if (got == nil) != (tt.want == nil) {
				t.Errorf("Client.Coverage() got = %+v, want = %+v", got, tt.want)
				return
			}
			if len(got) != len(tt.want) {
				t.Errorf("Client.Coverage() got length = %d, want length = %d", len(got), len(tt.want))
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Client.Coverage() got = %+v, want = %+v", got, tt.want)
			}
		})
	}
}

func TestClient_Coverage_ExecutionResultsWithContext(t *testing.T) {
	// Create a simple execution result with context containing pipeline field
	pipelineContext := &structpb.Struct{
		Fields: map[string]*structpb.Value{
			"pipeline": structpb.NewStructValue(&structpb.Struct{
				Fields: map[string]*structpb.Value{
					"step1": structpb.NewStringValue("executed"),
					"step3": structpb.NewStringValue("executed"),
				},
			}),
		},
	}

	resources := &Resources{
		Pipelines: []*Pipeline{
			{
				NamespaceName: "test-namespace",
				Resolvers: []*PipelineResolver{
					{
						Name: "test-resolver",
						Steps: []*PipelineStep{
							{Name: "step1"},
							{Name: "step2"},
							{Name: "step3"},
						},
						ExecutionResults: []*tailorv1.PipelineResolverExecutionResult{
							{
								Context: pipelineContext,
							},
						},
					},
				},
			},
		},
	}

	c := &Client{}
	got, err := c.Coverage(resources)
	if err != nil {
		t.Errorf("Client.Coverage() error = %v, wantErr false", err)
		return
	}

	expected := []*ResolverCoverage{
		{
			PipelineNamespaceName: "test-namespace",
			Name:                  "test-resolver",
			TotalSteps:            3,
			CoveredSteps:          2,
			Steps: []*StepCoverage{
				{Name: "step1", Count: 1},
				{Name: "step2", Count: 0},
				{Name: "step3", Count: 1},
			},
		},
	}

	if !reflect.DeepEqual(got, expected) {
		t.Errorf("Client.Coverage() = %v, want %v", got, expected)
	}
}

func TestClient_Coverage_ExecutionResultsWithoutContext(t *testing.T) {
	resources := &Resources{
		Pipelines: []*Pipeline{
			{
				NamespaceName: "test-namespace",
				Resolvers: []*PipelineResolver{
					{
						Name: "test-resolver",
						Steps: []*PipelineStep{
							{Name: "step1"},
							{Name: "step2"},
							{Name: "step3"},
						},
						ExecutionResults: []*tailorv1.PipelineResolverExecutionResult{
							{
								Context:          nil,
								LastPipelineName: "step2",
							},
						},
					},
				},
			},
		},
	}

	c := &Client{}
	got, err := c.Coverage(resources)
	if err != nil {
		t.Errorf("Client.Coverage() error = %v, wantErr false", err)
		return
	}

	expected := []*ResolverCoverage{
		{
			PipelineNamespaceName: "test-namespace",
			Name:                  "test-resolver",
			TotalSteps:            3,
			CoveredSteps:          2,
			Steps: []*StepCoverage{
				{Name: "step1", Count: 1},
				{Name: "step2", Count: 1},
				{Name: "step3", Count: 0},
			},
		},
	}

	if !reflect.DeepEqual(got, expected) {
		t.Errorf("Client.Coverage() = %v, want %v", got, expected)
	}
}

func TestResolverCoverage_Struct(t *testing.T) {
	rc := &ResolverCoverage{
		PipelineNamespaceName: "test",
		Name:                  "resolver",
		TotalSteps:            2,
		CoveredSteps:          1,
		Steps: []*StepCoverage{
			{Name: "step1", Count: 1},
			{Name: "step2", Count: 0},
		},
	}

	if rc.PipelineNamespaceName != "test" {
		t.Errorf("Expected PipelineNamespaceName to be 'test', got %s", rc.PipelineNamespaceName)
	}
	if rc.Name != "resolver" {
		t.Errorf("Expected Name to be 'resolver', got %s", rc.Name)
	}
	if rc.TotalSteps != 2 {
		t.Errorf("Expected TotalSteps to be 2, got %d", rc.TotalSteps)
	}
	if rc.CoveredSteps != 1 {
		t.Errorf("Expected CoveredSteps to be 1, got %d", rc.CoveredSteps)
	}
	if len(rc.Steps) != 2 {
		t.Errorf("Expected 2 steps, got %d", len(rc.Steps))
	}
}

func TestStepCoverage_Struct(t *testing.T) {
	sc := &StepCoverage{
		Name:  "test-step",
		Count: 5,
	}

	if sc.Name != "test-step" {
		t.Errorf("Expected Name to be 'test-step', got %s", sc.Name)
	}
	if sc.Count != 5 {
		t.Errorf("Expected Count to be 5, got %d", sc.Count)
	}
}

func TestClient_Coverage_EdgeCases(t *testing.T) {
	c := &Client{}

	t.Run("nil pipelines", func(t *testing.T) {
		resources := &Resources{
			Pipelines: nil,
		}
		got, err := c.Coverage(resources)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if len(got) != 0 {
			t.Errorf("Expected empty slice, got %d items", len(got))
		}
	})

	t.Run("empty pipeline with nil resolvers", func(t *testing.T) {
		resources := &Resources{
			Pipelines: []*Pipeline{
				{
					NamespaceName: "test",
					Resolvers:     nil,
				},
			},
		}
		got, err := c.Coverage(resources)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if len(got) != 0 {
			t.Errorf("Expected empty slice, got %d items", len(got))
		}
	})

	t.Run("resolver with nil steps", func(t *testing.T) {
		resources := &Resources{
			Pipelines: []*Pipeline{
				{
					NamespaceName: "test",
					Resolvers: []*PipelineResolver{
						{
							Name:             "test-resolver",
							Steps:            nil,
							ExecutionResults: []*tailorv1.PipelineResolverExecutionResult{},
						},
					},
				},
			},
		}
		got, err := c.Coverage(resources)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if len(got) != 1 {
			t.Errorf("Expected 1 item, got %d", len(got))
		}
		if got[0].TotalSteps != 0 {
			t.Errorf("Expected TotalSteps to be 0, got %d", got[0].TotalSteps)
		}
	})
}
