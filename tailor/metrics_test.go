package tailor

import (
	"testing"
)

func TestClient_Metrics(t *testing.T) {
	tests := []struct {
		name            string
		resources       *Resources
		expectedMetrics map[string]float64
	}{
		{
			name: "empty resources",
			resources: &Resources{
				Applications: []*Application{},
				Pipelines:    []*Pipeline{},
				TailorDBs:    []*TailorDB{},
				StateFlows:   []*StateFlow{},
			},
			expectedMetrics: map[string]float64{
				"pipeline_resolver_step_coverage_percentage": 0,
				"lint_warnings_total":                        0,
				"pipelines_total":                            0,
				"pipeline_resolvers_total":                   0,
				"pipeline_resolver_steps_total":              0,
				"pipeline_resolver_execution_paths_total":    0, // 0 resolvers = 0 paths
				"tailordbs_total":                            0,
				"tailordb_types_total":                       0,
				"tailordb_type_fields_total":                 0,
				"stateflows_total":                           0,
			},
		},
		{
			name: "test resources with data",
			resources: &Resources{
				Pipelines: []*Pipeline{
					{
						NamespaceName: "test-namespace",
						Resolvers: []*PipelineResolver{
							{
								Name: "testResolver",
								Steps: []*PipelineStep{
									{Name: "step1"},
								},
							},
						},
					},
				},
				TailorDBs: []*TailorDB{
					{
						Types: []*TailorDBType{
							{
								Name: "User",
								Fields: []*TailorDBField{
									{Name: "id"},
									{Name: "name"},
								},
							},
						},
					},
				},
				StateFlows: []*StateFlow{
					{NamespaceName: "test-namespace"},
				},
			},
			expectedMetrics: map[string]float64{
				"pipeline_resolver_step_coverage_percentage": 0, // no coverage data for test
				"lint_warnings_total":                        1,
				"pipelines_total":                            1,
				"pipeline_resolvers_total":                   1,
				"pipeline_resolver_steps_total":              1,
				"pipeline_resolver_execution_paths_total":    1, // 1 * 2^0 = 1 (1 step, no tests)
				"tailordbs_total":                            1,
				"tailordb_types_total":                       1,
				"tailordb_type_fields_total":                 2, // id and name fields
				"stateflows_total":                           1,
			},
		},
		{
			name: "multiple pipelines and resolvers",
			resources: &Resources{
				Pipelines: []*Pipeline{
					{
						NamespaceName: "ns1",
						Resolvers: []*PipelineResolver{
							{
								Name: "resolver1",
								Steps: []*PipelineStep{
									{Name: "step1"},
									{Name: "step2"},
								},
							},
							{
								Name: "resolver2",
								Steps: []*PipelineStep{
									{Name: "step1"},
									{Name: "step2"},
									{Name: "step3"},
								},
							},
						},
					},
					{
						NamespaceName: "ns2",
						Resolvers: []*PipelineResolver{
							{
								Name: "resolver3",
								Steps: []*PipelineStep{
									{Name: "step1"},
								},
							},
						},
					},
				},
				TailorDBs: []*TailorDB{
					{
						Types: []*TailorDBType{
							{
								Name: "User",
								Fields: []*TailorDBField{
									{Name: "id"},
									{Name: "name"},
									{Name: "email"},
								},
							},
							{
								Name: "Post",
								Fields: []*TailorDBField{
									{Name: "id"},
									{Name: "title"},
								},
							},
						},
					},
					{
						Types: []*TailorDBType{
							{
								Name: "Comment",
								Fields: []*TailorDBField{
									{Name: "id"},
									{Name: "content"},
									{Name: "authorId"},
									{Name: "postId"},
								},
							},
						},
					},
				},
				StateFlows: []*StateFlow{
					{NamespaceName: "flow1"},
					{NamespaceName: "flow2"},
					{NamespaceName: "flow3"},
				},
			},
			expectedMetrics: map[string]float64{
				"pipeline_resolver_step_coverage_percentage": 0, // no coverage data for test
				"lint_warnings_total":                        3,
				"pipelines_total":                            2, // ns1, ns2
				"pipeline_resolvers_total":                   3, // resolver1, resolver2, resolver3
				"pipeline_resolver_steps_total":              6, // 2+3+1 steps
				"pipeline_resolver_execution_paths_total":    6, // 2*2^0 + 3*2^0 + 1*2^0 = 2+3+1 (no tests)
				"tailordbs_total":                            2, // two TailorDB instances
				"tailordb_types_total":                       3, // User, Post, Comment
				"tailordb_type_fields_total":                 9, // 3+2+4 fields
				"stateflows_total":                           3, // flow1, flow2, flow3
			},
		},
		{
			name: "nested fields counting",
			resources: &Resources{
				TailorDBs: []*TailorDB{
					{
						Types: []*TailorDBType{
							{
								Name: "ComplexType",
								Fields: []*TailorDBField{
									{
										Name: "basicField",
									},
									{
										Name: "nestedObject",
										Fields: []*TailorDBField{
											{Name: "nestedField1"},
											{Name: "nestedField2"},
											{
												Name: "deeplyNested",
												Fields: []*TailorDBField{
													{Name: "deepField"},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			expectedMetrics: map[string]float64{
				"pipeline_resolver_step_coverage_percentage": 0,
				"lint_warnings_total":                        0,
				"pipelines_total":                            0,
				"pipeline_resolvers_total":                   0,
				"pipeline_resolver_steps_total":              0,
				"pipeline_resolver_execution_paths_total":    0, // 0 resolvers = 0 paths
				"tailordbs_total":                            1,
				"tailordb_types_total":                       1,
				"tailordb_type_fields_total":                 2, // Only top-level fields are counted
				"stateflows_total":                           0,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := createTestConfig(t)
			client, err := New(cfg)
			if err != nil {
				t.Fatalf("Failed to create client: %v", err)
			}
			if client == nil {
				t.Fatal("Client should not be nil")
			}

			metrics, err := client.Metrics(tt.resources)
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			if metrics == nil {
				t.Fatal("Metrics should not be nil")
			}

			// Create a map for easier assertion
			metricMap := make(map[string]float64)
			for _, m := range metrics {
				metricMap[m.Key] = m.Value
			}

			// Verify all expected metrics are present
			for expectedName, expectedValue := range tt.expectedMetrics {
				actualValue, exists := metricMap[expectedName]
				if !exists {
					t.Errorf("Expected metric %s not found", expectedName)
					continue
				}
				if actualValue != expectedValue {
					t.Errorf("Metric %s: expected %.0f, got %.0f", expectedName, expectedValue, actualValue)
				}
			}

			// Verify we have exactly the expected number of metrics
			if len(metrics) != len(tt.expectedMetrics) {
				t.Errorf("Expected %d metrics, got %d", len(tt.expectedMetrics), len(metrics))
			}
		})
	}
}

func TestClient_Metrics_MetricStructure(t *testing.T) {
	cfg := createTestConfig(t)
	client, err := New(cfg)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	resources := &Resources{
		Pipelines: []*Pipeline{
			{
				Resolvers: []*PipelineResolver{
					{
						Steps: []*PipelineStep{{Name: "step1"}},
					},
				},
			},
		},
		TailorDBs: []*TailorDB{
			{
				Types: []*TailorDBType{
					{
						Fields: []*TailorDBField{{Name: "field1"}},
					},
				},
			},
		},
		StateFlows: []*StateFlow{{NamespaceName: "flow1"}},
	}
	metrics, err := client.Metrics(resources)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Test that each metric has the required fields
	for _, metric := range metrics {
		if metric.Name == "" {
			t.Error("Metric name should not be empty")
		}
		if metric.Key == "" {
			t.Error("Metric key should not be empty")
		}
	}
}

func TestClient_Metrics_SpecificMetricValues(t *testing.T) {
	cfg := createTestConfig(t)
	client, err := New(cfg)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	// Test with known data structure
	resources := &Resources{
		Pipelines: []*Pipeline{
			{
				NamespaceName: "test-namespace",
				Resolvers: []*PipelineResolver{
					{
						Name: "testResolver1",
						Steps: []*PipelineStep{
							{Name: "step1"},
							{Name: "step2"},
							{Name: "step3"},
						},
					},
					{
						Name: "testResolver2",
						Steps: []*PipelineStep{
							{Name: "step1"},
							{Name: "step2"},
						},
					},
				},
			},
		},
		TailorDBs: []*TailorDB{
			{
				Types: []*TailorDBType{
					{
						Name: "User",
						Fields: []*TailorDBField{
							{Name: "id"},
							{Name: "name"},
							{Name: "email"},
							{Name: "createdAt"},
						},
					},
				},
			},
		},
		StateFlows: []*StateFlow{
			{NamespaceName: "workflow1"},
		},
	}

	metrics, err := client.Metrics(resources)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Create metric lookup map
	metricMap := make(map[string]Metric)
	for _, m := range metrics {
		metricMap[m.Key] = m
	}

	// Test pipeline metrics
	if metricMap["pipelines_total"].Value != float64(1) {
		t.Errorf("Expected pipelines_total to be 1, got %f", metricMap["pipelines_total"].Value)
	}
	if metricMap["pipeline_resolvers_total"].Value != float64(2) {
		t.Errorf("Expected pipeline_resolvers_total to be 2, got %f", metricMap["pipeline_resolvers_total"].Value)
	}
	if metricMap["pipeline_resolver_steps_total"].Value != float64(5) {
		t.Errorf("Expected pipeline_resolver_steps_total to be 5, got %f", metricMap["pipeline_resolver_steps_total"].Value)
	}
	if metricMap["pipeline_resolver_execution_paths_total"].Value != float64(5) {
		t.Errorf("Expected pipeline_resolver_execution_paths_total to be 5, got %f", metricMap["pipeline_resolver_execution_paths_total"].Value)
	}

	// Test TailorDB metrics
	if metricMap["tailordbs_total"].Value != float64(1) {
		t.Errorf("Expected tailordbs_total to be 1, got %f", metricMap["tailordbs_total"].Value)
	}
	if metricMap["tailordb_types_total"].Value != float64(1) {
		t.Errorf("Expected tailordb_types_total to be 1, got %f", metricMap["tailordb_types_total"].Value)
	}
	if metricMap["tailordb_type_fields_total"].Value != float64(4) {
		t.Errorf("Expected tailordb_type_fields_total to be 4, got %f", metricMap["tailordb_type_fields_total"].Value)
	}

	// Test StateFlow metrics
	if metricMap["stateflows_total"].Value != float64(1) {
		t.Errorf("Expected stateflows_total to be 1, got %f", metricMap["stateflows_total"].Value)
	}
}

func TestClient_Metrics_EdgeCases(t *testing.T) {
	cfg := createTestConfig(t)
	client, err := New(cfg)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	t.Run("nil resources", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("Expected panic when calling Metrics with nil resources")
			}
		}()
		_, _ = client.Metrics(nil)
	})

	t.Run("resources with nil slices", func(t *testing.T) {
		resources := &Resources{
			Pipelines:  nil,
			TailorDBs:  nil,
			StateFlows: nil,
		}

		metrics, err := client.Metrics(resources)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		metricMap := make(map[string]float64)
		for _, m := range metrics {
			metricMap[m.Key] = m.Value
		}

		// All counts should be 0
		if metricMap["pipelines_total"] != float64(0) {
			t.Errorf("Expected pipelines_total to be 0, got %f", metricMap["pipelines_total"])
		}
		if metricMap["pipeline_resolvers_total"] != float64(0) {
			t.Errorf("Expected pipeline_resolvers_total to be 0, got %f", metricMap["pipeline_resolvers_total"])
		}
		if metricMap["pipeline_resolver_steps_total"] != float64(0) {
			t.Errorf("Expected pipeline_resolver_steps_total to be 0, got %f", metricMap["pipeline_resolver_steps_total"])
		}
		if metricMap["pipeline_resolver_execution_paths_total"] != float64(0) {
			t.Errorf("Expected pipeline_resolver_execution_paths_total to be 0, got %f", metricMap["pipeline_resolver_execution_paths_total"])
		}
		if metricMap["tailordbs_total"] != float64(0) {
			t.Errorf("Expected tailordbs_total to be 0, got %f", metricMap["tailordbs_total"])
		}
		if metricMap["tailordb_types_total"] != float64(0) {
			t.Errorf("Expected tailordb_types_total to be 0, got %f", metricMap["tailordb_types_total"])
		}
		if metricMap["tailordb_type_fields_total"] != float64(0) {
			t.Errorf("Expected tailordb_type_fields_total to be 0, got %f", metricMap["tailordb_type_fields_total"])
		}
		if metricMap["stateflows_total"] != float64(0) {
			t.Errorf("Expected stateflows_total to be 0, got %f", metricMap["stateflows_total"])
		}
	})

	t.Run("pipelines with nil resolvers", func(t *testing.T) {
		resources := &Resources{
			Pipelines: []*Pipeline{
				{
					NamespaceName: "test",
					Resolvers:     nil, // nil resolvers
				},
				{
					NamespaceName: "test2",
					Resolvers:     []*PipelineResolver{}, // empty resolvers
				},
			},
		}

		metrics, err := client.Metrics(resources)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		metricMap := make(map[string]float64)
		for _, m := range metrics {
			metricMap[m.Key] = m.Value
		}

		if metricMap["pipelines_total"] != float64(2) {
			t.Errorf("Expected pipelines_total to be 2, got %f", metricMap["pipelines_total"])
		}
		if metricMap["pipeline_resolvers_total"] != float64(0) {
			t.Errorf("Expected pipeline_resolvers_total to be 0, got %f", metricMap["pipeline_resolvers_total"])
		}
		if metricMap["pipeline_resolver_steps_total"] != float64(0) {
			t.Errorf("Expected pipeline_resolver_steps_total to be 0, got %f", metricMap["pipeline_resolver_steps_total"])
		}
		if metricMap["pipeline_resolver_execution_paths_total"] != float64(0) {
			t.Errorf("Expected pipeline_resolver_execution_paths_total to be 0, got %f", metricMap["pipeline_resolver_execution_paths_total"])
		}
	})

	t.Run("tailordb with nil types", func(t *testing.T) {
		resources := &Resources{
			TailorDBs: []*TailorDB{
				{
					Types: nil, // nil types
				},
				{
					Types: []*TailorDBType{}, // empty types
				},
			},
		}

		metrics, err := client.Metrics(resources)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		metricMap := make(map[string]float64)
		for _, m := range metrics {
			metricMap[m.Key] = m.Value
		}

		if metricMap["tailordbs_total"] != float64(2) {
			t.Errorf("Expected tailordbs_total to be 2, got %f", metricMap["tailordbs_total"])
		}
		if metricMap["tailordb_types_total"] != float64(0) {
			t.Errorf("Expected tailordb_types_total to be 0, got %f", metricMap["tailordb_types_total"])
		}
		if metricMap["tailordb_type_fields_total"] != float64(0) {
			t.Errorf("Expected tailordb_type_fields_total to be 0, got %f", metricMap["tailordb_type_fields_total"])
		}
	})

	t.Run("types with nil fields", func(t *testing.T) {
		resources := &Resources{
			TailorDBs: []*TailorDB{
				{
					Types: []*TailorDBType{
						{
							Name:   "TypeWithNilFields",
							Fields: nil,
						},
						{
							Name:   "TypeWithEmptyFields",
							Fields: []*TailorDBField{},
						},
					},
				},
			},
		}

		metrics, err := client.Metrics(resources)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		metricMap := make(map[string]float64)
		for _, m := range metrics {
			metricMap[m.Key] = m.Value
		}

		if metricMap["tailordbs_total"] != float64(1) {
			t.Errorf("Expected tailordbs_total to be 1, got %f", metricMap["tailordbs_total"])
		}
		if metricMap["tailordb_types_total"] != float64(2) {
			t.Errorf("Expected tailordb_types_total to be 2, got %f", metricMap["tailordb_types_total"])
		}
		if metricMap["tailordb_type_fields_total"] != float64(0) {
			t.Errorf("Expected tailordb_type_fields_total to be 0, got %f", metricMap["tailordb_type_fields_total"])
		}
	})
}

func TestClient_Metrics_MetricNames(t *testing.T) {
	cfg := createTestConfig(t)
	client, err := New(cfg)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	resources := &Resources{
		Applications: []*Application{},
		Pipelines:    []*Pipeline{},
		TailorDBs:    []*TailorDB{},
		StateFlows:   []*StateFlow{},
	}
	metrics, err := client.Metrics(resources)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	expectedMetricKeys := []string{
		"pipeline_resolver_step_coverage_percentage",
		"lint_warnings_total",
		"pipelines_total",
		"pipeline_resolvers_total",
		"pipeline_resolver_steps_total",
		"pipeline_resolver_execution_paths_total",
		"tailordbs_total",
		"tailordb_types_total",
		"tailordb_type_fields_total",
		"stateflows_total",
	}

	actualKeys := make([]string, len(metrics))
	for i, m := range metrics {
		actualKeys[i] = m.Key
	}

	// Check that we have all expected metric keys
	for _, expectedKey := range expectedMetricKeys {
		found := false
		for _, actualKey := range actualKeys {
			if actualKey == expectedKey {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected metric %s not found", expectedKey)
		}
	}

	if len(expectedMetricKeys) != len(actualKeys) {
		t.Errorf("Expected %d metrics, got %d", len(expectedMetricKeys), len(actualKeys))
	}
}

func TestClient_Metrics_MetricFields(t *testing.T) {
	cfg := createTestConfig(t)
	client, err := New(cfg)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	resources := &Resources{
		Applications: []*Application{},
		Pipelines:    []*Pipeline{},
		TailorDBs:    []*TailorDB{},
		StateFlows:   []*StateFlow{},
	}
	metrics, err := client.Metrics(resources)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	expectedKeys := map[string]string{
		"pipeline_resolver_step_coverage_percentage": "pipeline_resolver_step_coverage_percentage",
		"lint_warnings_total":                        "lint_warnings_total",
		"pipelines_total":                            "pipelines_total",
		"pipeline_resolvers_total":                   "pipeline_resolvers_total",
		"pipeline_resolver_steps_total":              "pipeline_resolver_steps_total",
		"pipeline_resolver_execution_paths_total":    "pipeline_resolver_execution_paths_total",
		"tailordbs_total":                            "tailordbs_total",
		"tailordb_types_total":                       "tailordb_types_total",
		"tailordb_type_fields_total":                 "tailordb_type_fields_total",
		"stateflows_total":                           "stateflows_total",
	}

	for _, metric := range metrics {
		expectedKey, exists := expectedKeys[metric.Key]
		if !exists {
			t.Errorf("Unexpected metric: %s", metric.Key)
		}
		if expectedKey != metric.Key {
			t.Errorf("Wrong key for metric %s: expected '%s', got '%s'",
				metric.Key, expectedKey, metric.Key)
		}
	}
}

func TestMetric_Fields(t *testing.T) {
	metric := Metric{
		Key:   "test_metric",
		Name:  "Test Metric",
		Value: 42.5,
		Unit:  "count",
	}

	if metric.Key != "test_metric" {
		t.Errorf("Expected Key to be 'test_metric', got '%s'", metric.Key)
	}
	if metric.Name != "Test Metric" {
		t.Errorf("Expected Name to be 'Test Metric', got '%s'", metric.Name)
	}
	if metric.Value != 42.5 {
		t.Errorf("Expected Value to be 42.5, got %f", metric.Value)
	}
	if metric.Unit != "count" {
		t.Errorf("Expected Unit to be 'count', got '%s'", metric.Unit)
	}
}

func TestClient_Metrics_LargeNumbers(t *testing.T) {
	// Test with large numbers to ensure proper handling
	cfg := createTestConfig(t)
	client, err := New(cfg)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	// Create resources with many items
	pipelines := make([]*Pipeline, 100)
	for i := 0; i < 100; i++ {
		resolvers := make([]*PipelineResolver, 10)
		for j := 0; j < 10; j++ {
			steps := make([]*PipelineStep, 5)
			for k := 0; k < 5; k++ {
				steps[k] = &PipelineStep{Name: "step"}
			}
			resolvers[j] = &PipelineResolver{
				Name:  "resolver",
				Steps: steps,
			}
		}

		pipelines[i] = &Pipeline{
			NamespaceName: "namespace",
			Resolvers:     resolvers,
		}
	}

	resources := &Resources{
		Pipelines: pipelines,
	}

	metrics, err := client.Metrics(resources)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	metricMap := make(map[string]float64)
	for _, m := range metrics {
		metricMap[m.Key] = m.Value
	}

	// Verify calculations
	if metricMap["pipelines_total"] != float64(100) {
		t.Errorf("Expected pipelines_total to be 100, got %f", metricMap["pipelines_total"])
	}
	if metricMap["pipeline_resolvers_total"] != float64(1000) {
		t.Errorf("Expected pipeline_resolvers_total to be 1000, got %f", metricMap["pipeline_resolvers_total"])
	}
	if metricMap["pipeline_resolver_steps_total"] != float64(5000) {
		t.Errorf("Expected pipeline_resolver_steps_total to be 5000, got %f", metricMap["pipeline_resolver_steps_total"])
	}
	if metricMap["pipeline_resolver_execution_paths_total"] != float64(5000) {
		t.Errorf("Expected pipeline_resolver_execution_paths_total to be 5000, got %f", metricMap["pipeline_resolver_execution_paths_total"])
	}
}

func TestClient_Metrics_ExecutionPaths(t *testing.T) {
	tests := []struct {
		name            string
		resources       *Resources
		expectedMetrics map[string]float64
	}{
		{
			name: "resolver with tests - basic calculation",
			resources: &Resources{
				Pipelines: []*Pipeline{
					{
						NamespaceName: "test-namespace",
						Resolvers: []*PipelineResolver{
							{
								Name: "resolver_with_tests",
								Steps: []*PipelineStep{
									{
										Name: "step1",
										Operation: PipelineStepOperation{
											Test: "test1",
										},
									},
									{
										Name: "step2",
										Operation: PipelineStepOperation{
											Test: "test2",
										},
									},
									{
										Name: "step3",
										Operation: PipelineStepOperation{
											Test: "", // no test
										},
									},
								},
							},
						},
					},
				},
			},
			expectedMetrics: map[string]float64{
				"pipelines_total":                         1,
				"pipeline_resolvers_total":                1,
				"pipeline_resolver_steps_total":           3,
				"pipeline_resolver_execution_paths_total": 12, // 3 * 2^2 = 12 (3 steps, 2 tests)
			},
		},
		{
			name: "multiple resolvers with different test counts",
			resources: &Resources{
				Pipelines: []*Pipeline{
					{
						NamespaceName: "test-namespace",
						Resolvers: []*PipelineResolver{
							{
								Name: "resolver1",
								Steps: []*PipelineStep{
									{
										Name: "step1",
										Operation: PipelineStepOperation{
											Test: "test1",
										},
									},
									{
										Name: "step2",
										Operation: PipelineStepOperation{
											Test: "test2",
										},
									},
								},
							},
							{
								Name: "resolver2",
								Steps: []*PipelineStep{
									{
										Name: "step1",
										Operation: PipelineStepOperation{
											Test: "test1",
										},
									},
									{
										Name: "step2",
										Operation: PipelineStepOperation{
											Test: "", // no test
										},
									},
									{
										Name: "step3",
										Operation: PipelineStepOperation{
											Test: "", // no test
										},
									},
								},
							},
							{
								Name: "resolver3",
								Steps: []*PipelineStep{
									{
										Name: "step1",
										Operation: PipelineStepOperation{
											Test: "", // no test
										},
									},
								},
							},
						},
					},
				},
			},
			expectedMetrics: map[string]float64{
				"pipelines_total":                         1,
				"pipeline_resolvers_total":                3,
				"pipeline_resolver_steps_total":           6,  // 2+3+1 steps
				"pipeline_resolver_execution_paths_total": 15, // 2*2^2 + 3*2^1 + 1*2^0 = 8+6+1 = 15
			},
		},
		{
			name: "edge case - no tests",
			resources: &Resources{
				Pipelines: []*Pipeline{
					{
						NamespaceName: "test-namespace",
						Resolvers: []*PipelineResolver{
							{
								Name: "resolver_no_tests",
								Steps: []*PipelineStep{
									{
										Name: "step1",
										Operation: PipelineStepOperation{
											Test: "",
										},
									},
									{
										Name: "step2",
										Operation: PipelineStepOperation{
											Test: "",
										},
									},
								},
							},
						},
					},
				},
			},
			expectedMetrics: map[string]float64{
				"pipelines_total":                         1,
				"pipeline_resolvers_total":                1,
				"pipeline_resolver_steps_total":           2,
				"pipeline_resolver_execution_paths_total": 2, // 2 * 2^0 = 2 (2 steps, no tests)
			},
		},
		{
			name: "edge case - no steps",
			resources: &Resources{
				Pipelines: []*Pipeline{
					{
						NamespaceName: "test-namespace",
						Resolvers: []*PipelineResolver{
							{
								Name:  "resolver_no_steps",
								Steps: []*PipelineStep{},
							},
						},
					},
				},
			},
			expectedMetrics: map[string]float64{
				"pipelines_total":                         1,
				"pipeline_resolvers_total":                1,
				"pipeline_resolver_steps_total":           0,
				"pipeline_resolver_execution_paths_total": 0, // 0 * 2^0 = 0 (no steps)
			},
		},
		{
			name: "complex calculation - all steps have tests",
			resources: &Resources{
				Pipelines: []*Pipeline{
					{
						NamespaceName: "test-namespace",
						Resolvers: []*PipelineResolver{
							{
								Name: "complex_resolver",
								Steps: []*PipelineStep{
									{
										Name: "step1",
										Operation: PipelineStepOperation{
											Test: "validation_test",
										},
									},
									{
										Name: "step2",
										Operation: PipelineStepOperation{
											Test: "integration_test",
										},
									},
									{
										Name: "step3",
										Operation: PipelineStepOperation{
											Test: "unit_test",
										},
									},
									{
										Name: "step4",
										Operation: PipelineStepOperation{
											Test: "e2e_test",
										},
									},
								},
							},
						},
					},
				},
			},
			expectedMetrics: map[string]float64{
				"pipelines_total":                         1,
				"pipeline_resolvers_total":                1,
				"pipeline_resolver_steps_total":           4,
				"pipeline_resolver_execution_paths_total": 64, // 4 * 2^4 = 64 (4 steps, 4 tests)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := createTestConfig(t)
			client, err := New(cfg)
			if err != nil {
				t.Fatalf("Failed to create client: %v", err)
			}

			metrics, err := client.Metrics(tt.resources)
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			// Create a map for easier assertion
			metricMap := make(map[string]float64)
			for _, m := range metrics {
				metricMap[m.Key] = m.Value
			}

			// Verify expected metrics
			for expectedName, expectedValue := range tt.expectedMetrics {
				actualValue, exists := metricMap[expectedName]
				if !exists {
					t.Errorf("Expected metric %s not found", expectedName)
					continue
				}
				if actualValue != expectedValue {
					t.Errorf("Metric %s: expected %.0f, got %.0f", expectedName, expectedValue, actualValue)
				}
			}
		})
	}
}

func TestClient_Metrics_ExecutionPaths_EdgeCases(t *testing.T) {
	cfg := createTestConfig(t)
	client, err := New(cfg)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	t.Run("mixed test scenarios", func(t *testing.T) {
		resources := &Resources{
			Pipelines: []*Pipeline{
				{
					NamespaceName: "mixed-namespace",
					Resolvers: []*PipelineResolver{
						{
							Name: "mixed_resolver",
							Steps: []*PipelineStep{
								{
									Name: "step_with_test",
									Operation: PipelineStepOperation{
										Test: "some test",
									},
								},
								{
									Name: "step_without_test",
									Operation: PipelineStepOperation{
										Test: "",
									},
								},
								{
									Name: "step_with_whitespace_test",
									Operation: PipelineStepOperation{
										Test: "   ", // whitespace only should count as non-empty
									},
								},
							},
						},
					},
				},
			},
		}

		metrics, err := client.Metrics(resources)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		metricMap := make(map[string]float64)
		for _, m := range metrics {
			metricMap[m.Key] = m.Value
		}

		// Expected: 3 steps, 2 tests (one empty string doesn't count, whitespace does count)
		expectedExecutionPaths := float64(12) // 3 * 2^2 = 12
		if metricMap["pipeline_resolver_execution_paths_total"] != expectedExecutionPaths {
			t.Errorf("Expected execution paths to be %.0f, got %.0f",
				expectedExecutionPaths, metricMap["pipeline_resolver_execution_paths_total"])
		}
	})

	t.Run("boundary case - single step with test", func(t *testing.T) {
		resources := &Resources{
			Pipelines: []*Pipeline{
				{
					NamespaceName: "boundary-namespace",
					Resolvers: []*PipelineResolver{
						{
							Name: "single_step_resolver",
							Steps: []*PipelineStep{
								{
									Name: "only_step",
									Operation: PipelineStepOperation{
										Test: "single_test",
									},
								},
							},
						},
					},
				},
			},
		}

		metrics, err := client.Metrics(resources)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		metricMap := make(map[string]float64)
		for _, m := range metrics {
			metricMap[m.Key] = m.Value
		}

		// Expected: 1 step, 1 test -> 1 * 2^1 = 2
		expectedExecutionPaths := float64(2)
		if metricMap["pipeline_resolver_execution_paths_total"] != expectedExecutionPaths {
			t.Errorf("Expected execution paths to be %.0f, got %.0f",
				expectedExecutionPaths, metricMap["pipeline_resolver_execution_paths_total"])
		}
	})
}
