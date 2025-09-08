package tailor

import (
	"strings"
	"testing"

	tailorv1 "buf.build/gen/go/tailor-inc/tailor/protocolbuffers/go/tailor/v1"

	"github.com/tailor-platform/patterner/config"
)

func TestClient_Lint(t *testing.T) {
	tests := []struct {
		name      string
		config    *config.Config
		resources *Resources
		wantWarns int
		wantErrs  bool
	}{
		{
			name:   "no warnings with empty resources",
			config: createTestConfig(t),
			resources: &Resources{
				Applications: []*Application{},
				Pipelines:    []*Pipeline{},
				TailorDBs:    []*TailorDB{},
				StateFlows:   []*StateFlow{},
			},
			wantWarns: 0,
			wantErrs:  false,
		},
		{
			name:   "draft feature deprecated warning",
			config: createTestConfig(t),
			resources: &Resources{
				TailorDBs: []*TailorDB{
					{
						NamespaceName: "test-namespace",
						Types: []*TailorDBType{
							{
								Name:           "User",
								Description:    "User type",
								Draft:          true,
								TypePermission: &TailorDBTypePermission{},
								Fields: []*TailorDBField{
									{
										Name:     "id",
										Type:     "ID",
										Required: true,
										Hooks: Hooks{
											CreateExpr: "now()",
											UpdateExpr: "now()",
										},
									},
									{
										Name:     "name",
										Type:     "String",
										Required: true,
									},
								},
							},
						},
					},
				},
			},
			wantWarns: 3, // Draft type + TypePermission + CEL hooks
			wantErrs:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := New(tt.config)
			if err != nil {
				t.Fatalf("Failed to create client: %v", err)
			}
			if client == nil {
				t.Fatal("Client should not be nil")
			}

			warns, err := client.Lint(tt.resources)
			if tt.wantErrs {
				if err == nil {
					t.Error("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if len(warns) != tt.wantWarns {
					t.Errorf("Expected %d warnings, got %d", tt.wantWarns, len(warns))
				}
			}
		})
	}
}

func TestClient_Lint_TailorDB(t *testing.T) {
	tests := []struct {
		name         string
		configMod    func(*config.Config)
		resources    *Resources
		expectedMsgs []string
	}{
		{
			name: "draft feature warning",
			configMod: func(c *config.Config) {
				c.Lint.TailorDB.DeprecatedFeature.Enabled = true
				c.Lint.TailorDB.DeprecatedFeature.AllowDraft = false
			},
			resources: &Resources{
				TailorDBs: []*TailorDB{
					{
						NamespaceName: "test-ns",
						Types: []*TailorDBType{
							{
								Name:  "User",
								Draft: true,
							},
						},
					},
				},
			},
			expectedMsgs: []string{"Draft feature is deprecated"},
		},
		{
			name: "legacy type permission warning",
			configMod: func(c *config.Config) {
				c.Lint.TailorDB.LegacyPermission.Enabled = true
				c.Lint.TailorDB.LegacyPermission.AllowTypePermission = false
			},
			resources: &Resources{
				TailorDBs: []*TailorDB{
					{
						NamespaceName: "test-ns",
						Types: []*TailorDBType{
							{
								Name:           "User",
								TypePermission: &TailorDBTypePermission{},
							},
						},
					},
				},
			},
			expectedMsgs: []string{"Type-level permission is legacy. Use `Permission` or `GQLPermission` instead"},
		},
		{
			name: "legacy record permission warning",
			configMod: func(c *config.Config) {
				c.Lint.TailorDB.LegacyPermission.Enabled = true
				c.Lint.TailorDB.LegacyPermission.AllowRecordPermission = false
			},
			resources: &Resources{
				TailorDBs: []*TailorDB{
					{
						NamespaceName: "test-ns",
						Types: []*TailorDBType{
							{
								Name:             "User",
								RecordPermission: &TailorDBRecordPermission{},
							},
						},
					},
				},
			},
			expectedMsgs: []string{"Record-level permission is legacy. Use `Permission` or `GQLPermission` instead"},
		},
		{
			name: "CEL hooks deprecated warning",
			configMod: func(c *config.Config) {
				c.Lint.TailorDB.DeprecatedFeature.Enabled = true
				c.Lint.TailorDB.DeprecatedFeature.AllowCELHooks = false
			},
			resources: &Resources{
				TailorDBs: []*TailorDB{
					{
						NamespaceName: "test-ns",
						Types: []*TailorDBType{
							{
								Name: "User",
								Fields: []*TailorDBField{
									{
										Name: "testField",
										Hooks: Hooks{
											CreateExpr: "now()",
											UpdateExpr: "now()",
										},
									},
								},
							},
						},
					},
				},
			},
			expectedMsgs: []string{"Hooks `create_expr` and `update_expr` are deprecated. Use `create` or `update` instead"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := createTestConfig(t)
			tt.configMod(cfg)

			client, err := New(cfg)
			if err != nil {
				t.Fatalf("Failed to create client: %v", err)
			}

			warns, err := client.Lint(tt.resources)
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if len(warns) != len(tt.expectedMsgs) {
				t.Errorf("Expected %d warnings, got %d", len(tt.expectedMsgs), len(warns))
			}
			for i, expectedMsg := range tt.expectedMsgs {
				if i >= len(warns) {
					break
				}
				if warns[i].Type != LintTargetTypeTailorDB {
					t.Errorf("Expected warning type %s, got %s", LintTargetTypeTailorDB, warns[i].Type)
				}
				if !strings.Contains(warns[i].Message, expectedMsg) {
					t.Errorf("Expected warning message to contain '%s', got '%s'", expectedMsg, warns[i].Message)
				}
			}
		})
	}
}

func TestClient_Lint_Pipeline(t *testing.T) {
	tests := []struct {
		name         string
		configMod    func(*config.Config)
		resources    *Resources
		expectedMsgs []string
		wantError    bool
	}{
		{
			name: "insecure authorization warning",
			configMod: func(c *config.Config) {
				c.Lint.Pipeline.InsecureAuthorization.Enabled = true
			},
			resources: &Resources{
				Pipelines: []*Pipeline{
					{
						NamespaceName: "test-ns",
						Resolvers: []*PipelineResolver{
							{
								Name:          "testResolver",
								Authorization: "true",
							},
						},
					},
				},
			},
			expectedMsgs: []string{"resolver allows insecure authorization"},
		},
		{
			name: "step length warning",
			configMod: func(c *config.Config) {
				c.Lint.Pipeline.StepLength.Enabled = true
				c.Lint.Pipeline.StepLength.Max = 1
			},
			resources: &Resources{
				Pipelines: []*Pipeline{
					{
						NamespaceName: "test-ns",
						Resolvers: []*PipelineResolver{
							{
								Name: "testResolver",
								Steps: []*PipelineStep{
									{Name: "step1"},
									{Name: "step2"},
								},
							},
						},
					},
				},
			},
			expectedMsgs: []string{"resolver has too many steps (2 > 1)"},
		},
		{
			name: "legacy script warnings",
			configMod: func(c *config.Config) {
				c.Lint.Pipeline.LegacyScript.Enabled = true
			},
			resources: &Resources{
				Pipelines: []*Pipeline{
					{
						NamespaceName: "test-ns",
						Resolvers: []*PipelineResolver{
							{
								Name: "testResolver",
								Steps: []*PipelineStep{
									{
										Name:           "testStep",
										PreValidation:  "validateInput()",
										PreScript:      "console.log('pre');",
										PostScript:     "console.log('post');",
										PostValidation: "validateOutput()",
									},
								},
							},
						},
					},
				},
			},
			expectedMsgs: []string{
				"`pre_validation` is not recommended. Use `pre_hook` instead.",
				"`pre_script` is not recommended. Use `pre_hook` instead.",
				"`post_script` is not recommended. Use `post_hook` instead.",
				"`post_validation` is not recommended. Use `post_hook` instead.",
			},
		},
		{
			name: "invalid GraphQL syntax error",
			configMod: func(c *config.Config) {
				c.Lint.Pipeline.DeprecatedFeature.Enabled = true
			},
			resources: &Resources{
				Pipelines: []*Pipeline{
					{
						NamespaceName: "test-ns",
						Resolvers: []*PipelineResolver{
							{
								Name: "testResolver",
								Steps: []*PipelineStep{
									{
										Name: "testStep",
										Operation: PipelineStepOperation{
											Type:   tailorv1.PipelineResolver_OPERATION_TYPE_GRAPHQL,
											Source: "query { users { id name", // Invalid GraphQL
										},
									},
								},
							},
						},
					},
				},
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := createTestConfig(t)
			tt.configMod(cfg)

			client, err := New(cfg)
			if err != nil {
				t.Fatalf("Failed to create client: %v", err)
			}

			warns, err := client.Lint(tt.resources)

			if tt.wantError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			if len(warns) != len(tt.expectedMsgs) {
				t.Errorf("Expected %d warnings, got %d", len(tt.expectedMsgs), len(warns))
			}

			for i, expectedMsg := range tt.expectedMsgs {
				if i >= len(warns) {
					break
				}
				if warns[i].Type != LintTargetTypePipeline {
					t.Errorf("Expected warning type %s, got %s", LintTargetTypePipeline, warns[i].Type)
				}
				if !strings.Contains(warns[i].Message, expectedMsg) {
					t.Errorf("Expected warning message to contain '%s', got '%s'", expectedMsg, warns[i].Message)
				}
			}
		})
	}
}

func TestClient_Lint_GraphQLParsing(t *testing.T) {
	tests := []struct {
		name         string
		graphqlQuery string
		typeNames    []string
		configMod    func(*config.Config)
		expectedMsgs []string
		wantError    bool
	}{
		{
			name:         "valid GraphQL query",
			graphqlQuery: "query { users { id name } }",
			wantError:    false,
		},
		{
			name:         "valid GraphQL mutation",
			graphqlQuery: "mutation { createUser(input: {name: \"test\"}) { id } }",
			wantError:    false,
		},
		{
			name:         "invalid GraphQL syntax",
			graphqlQuery: "query { users { id name", // Missing closing braces
			wantError:    true,
		},
		{
			name:         "StateFlow mutation warning",
			graphqlQuery: "mutation { newState(input: {status: \"active\"}) { id } }",
			configMod: func(c *config.Config) {
				c.Lint.Pipeline.DeprecatedFeature.Enabled = true
				c.Lint.Pipeline.DeprecatedFeature.AllowStateFlow = false
			},
			expectedMsgs: []string{"StateFlow feature is deprecated (found usage of newState)"},
		},
		{
			name:         "draft mutation warning",
			graphqlQuery: "mutation { appendDraftUser(input: {name: \"test\"}) { id } }",
			typeNames:    []string{"User"},
			configMod: func(c *config.Config) {
				c.Lint.Pipeline.DeprecatedFeature.Enabled = true
				c.Lint.Pipeline.DeprecatedFeature.AllowDraft = false
			},
			expectedMsgs: []string{"Draft feature is deprecated (found usage of appendDraftUser)"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := createTestConfig(t)
			if tt.configMod != nil {
				tt.configMod(cfg)
			}

			resources := &Resources{
				Pipelines: []*Pipeline{
					{
						NamespaceName: "test-ns",
						Resolvers: []*PipelineResolver{
							{
								Name: "testResolver",
								Steps: []*PipelineStep{
									{
										Name: "testStep",
										Operation: PipelineStepOperation{
											Type:   tailorv1.PipelineResolver_OPERATION_TYPE_GRAPHQL,
											Source: tt.graphqlQuery,
										},
									},
								},
							},
						},
					},
				},
				TailorDBs: []*TailorDB{
					{
						Types: func() []*TailorDBType {
							var types []*TailorDBType
							for _, name := range tt.typeNames {
								types = append(types, &TailorDBType{Name: name})
							}
							return types
						}(),
					},
				},
			}

			client, err := New(cfg)
			if err != nil {
				t.Fatalf("Failed to create client: %v", err)
			}

			warns, err := client.Lint(resources)

			if tt.wantError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			if len(warns) != len(tt.expectedMsgs) {
				t.Errorf("Expected %d warnings, got %d", len(tt.expectedMsgs), len(warns))
			}

			for i, expectedMsg := range tt.expectedMsgs {
				if i >= len(warns) {
					break
				}
				if warns[i].Type != LintTargetTypePipeline {
					t.Errorf("Expected warning type %s, got %s", LintTargetTypePipeline, warns[i].Type)
				}
				if !strings.Contains(warns[i].Message, expectedMsg) {
					t.Errorf("Expected warning message to contain '%s', got '%s'", expectedMsg, warns[i].Message)
				}
			}
		})
	}
}

func TestClient_Lint_StateFlow(t *testing.T) {
	cfg := createTestConfig(t)
	cfg.Lint.StateFlow.DeprecatedFeature.Enabled = true

	resources := &Resources{
		StateFlows: []*StateFlow{
			{
				NamespaceName: "test-namespace",
			},
		},
	}

	client, err := New(cfg)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	warns, err := client.Lint(resources)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(warns) != 1 {
		t.Errorf("Expected 1 warning, got %d", len(warns))
	}
	if warns[0].Type != LintTargetTypeStateFlow {
		t.Errorf("Expected warning type %s, got %s", LintTargetTypeStateFlow, warns[0].Type)
	}
	if warns[0].Name != "test-namespace" {
		t.Errorf("Expected warning name 'test-namespace', got '%s'", warns[0].Name)
	}
	if warns[0].Message != "StateFlow is deprecated" {
		t.Errorf("Expected warning message 'StateFlow is deprecated', got '%s'", warns[0].Message)
	}
}

func TestClient_Lint_MultipleMutations(t *testing.T) {
	cfg := createTestConfig(t)
	cfg.Lint.Pipeline.MultipleMutations.Enabled = true

	resources := &Resources{
		Pipelines: []*Pipeline{
			{
				NamespaceName: "test-ns",
				Resolvers: []*PipelineResolver{
					{
						Name: "testResolver",
						Steps: []*PipelineStep{
							{
								Name: "mutation1",
								Operation: PipelineStepOperation{
									Type:   tailorv1.PipelineResolver_OPERATION_TYPE_GRAPHQL,
									Source: "mutation { createUser(input: {name: \"test\"}) { id } }",
								},
							},
							{
								Name: "mutation2",
								Operation: PipelineStepOperation{
									Type:   tailorv1.PipelineResolver_OPERATION_TYPE_GRAPHQL,
									Source: "mutation { updateUser(id: 1, input: {name: \"updated\"}) { id } }",
								},
							},
						},
					},
				},
			},
		},
	}

	client, err := New(cfg)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	warns, err := client.Lint(resources)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(warns) != 1 {
		t.Errorf("Expected 1 warning, got %d", len(warns))
	}
	if warns[0].Type != LintTargetTypePipeline {
		t.Errorf("Expected warning type %s, got %s", LintTargetTypePipeline, warns[0].Type)
	}
	if !strings.Contains(warns[0].Message, "Resolver has multiple mutations") {
		t.Errorf("Expected warning message to contain 'Resolver has multiple mutations', got '%s'", warns[0].Message)
	}
}

func TestClient_Lint_QueryBeforeMutation(t *testing.T) {
	cfg := createTestConfig(t)
	cfg.Lint.Pipeline.QueryBeforeMutation.Enabled = true

	resources := &Resources{
		Pipelines: []*Pipeline{
			{
				NamespaceName: "test-ns",
				Resolvers: []*PipelineResolver{
					{
						Name: "testResolver",
						Steps: []*PipelineStep{
							{
								Name: "queryStep",
								Operation: PipelineStepOperation{
									Type:   tailorv1.PipelineResolver_OPERATION_TYPE_GRAPHQL,
									Source: "query { users { id } }",
								},
							},
							{
								Name: "mutationStep",
								Operation: PipelineStepOperation{
									Type:   tailorv1.PipelineResolver_OPERATION_TYPE_GRAPHQL,
									Source: "mutation { createUser(input: {name: \"test\"}) { id } }",
								},
							},
						},
					},
				},
			},
		},
	}

	client, err := New(cfg)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	warns, err := client.Lint(resources)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(warns) != 1 {
		t.Errorf("Expected 1 warning, got %d", len(warns))
	}
	if warns[0].Type != LintTargetTypePipeline {
		t.Errorf("Expected warning type %s, got %s", LintTargetTypePipeline, warns[0].Type)
	}
	if !strings.Contains(warns[0].Message, "Resolver has query before mutation") {
		t.Errorf("Expected warning message to contain 'Resolver has query before mutation', got '%s'", warns[0].Message)
	}
}

func TestLintWarn_String(t *testing.T) {
	warn := &LintWarn{
		Type:    LintTargetTypePipeline,
		Name:    "test-namespace/test-resolver",
		Message: "Test warning message",
	}

	// Test that the warn struct contains expected values
	if warn.Type != LintTargetTypePipeline {
		t.Errorf("Expected Type %s, got %s", LintTargetTypePipeline, warn.Type)
	}
	if warn.Name != "test-namespace/test-resolver" {
		t.Errorf("Expected Name 'test-namespace/test-resolver', got '%s'", warn.Name)
	}
	if warn.Message != "Test warning message" {
		t.Errorf("Expected Message 'Test warning message', got '%s'", warn.Message)
	}
}

func TestLintTargetTypes(t *testing.T) {
	// Test that lint target types are defined correctly
	if string(LintTargetTypePipeline) != "pipeline" {
		t.Errorf("Expected LintTargetTypePipeline to be 'pipeline', got '%s'", string(LintTargetTypePipeline))
	}
	if string(LintTargetTypeTailorDB) != "tailordb" {
		t.Errorf("Expected LintTargetTypeTailorDB to be 'tailordb', got '%s'", string(LintTargetTypeTailorDB))
	}
	if string(LintTargetTypeStateFlow) != "stateflow" {
		t.Errorf("Expected LintTargetTypeStateFlow to be 'stateflow', got '%s'", string(LintTargetTypeStateFlow))
	}
}
