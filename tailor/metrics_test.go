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
				"pipelines_total":               0,
				"pipeline_resolvers_total":      0,
				"pipeline_resolver_steps_total": 0,
				"tailordbs_total":               0,
				"tailordb_types_total":          0,
				"tailordb_type_fields_total":    0,
				"stateflows_total":              0,
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
				"pipelines_total":               1,
				"pipeline_resolvers_total":      1,
				"pipeline_resolver_steps_total": 1,
				"tailordbs_total":               1,
				"tailordb_types_total":          1,
				"tailordb_type_fields_total":    2, // id and name fields
				"stateflows_total":              1,
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
				"pipelines_total":               2, // ns1, ns2
				"pipeline_resolvers_total":      3, // resolver1, resolver2, resolver3
				"pipeline_resolver_steps_total": 6, // 2+3+1 steps
				"tailordbs_total":               2, // two TailorDB instances
				"tailordb_types_total":          3, // User, Post, Comment
				"tailordb_type_fields_total":    9, // 3+2+4 fields
				"stateflows_total":              3, // flow1, flow2, flow3
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
				"pipelines_total":               0,
				"pipeline_resolvers_total":      0,
				"pipeline_resolver_steps_total": 0,
				"tailordbs_total":               1,
				"tailordb_types_total":          1,
				"tailordb_type_fields_total":    2, // Only top-level fields are counted
				"stateflows_total":              0,
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
				metricMap[m.Name] = m.Value
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
		if metric.Description == "" {
			t.Error("Metric description should not be empty")
		}
		if metric.Value < 0 {
			t.Errorf("Metric value should be non-negative, got %f", metric.Value)
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
		metricMap[m.Name] = m
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
			metricMap[m.Name] = m.Value
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
			metricMap[m.Name] = m.Value
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
			metricMap[m.Name] = m.Value
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
			metricMap[m.Name] = m.Value
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

	expectedMetricNames := []string{
		"pipelines_total",
		"pipeline_resolvers_total",
		"pipeline_resolver_steps_total",
		"tailordbs_total",
		"tailordb_types_total",
		"tailordb_type_fields_total",
		"stateflows_total",
	}

	actualNames := make([]string, len(metrics))
	for i, m := range metrics {
		actualNames[i] = m.Name
	}

	// Check that we have all expected metric names
	for _, expectedName := range expectedMetricNames {
		found := false
		for _, actualName := range actualNames {
			if actualName == expectedName {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected metric %s not found", expectedName)
		}
	}

	if len(expectedMetricNames) != len(actualNames) {
		t.Errorf("Expected %d metrics, got %d", len(expectedMetricNames), len(actualNames))
	}
}

func TestClient_Metrics_MetricDescriptions(t *testing.T) {
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

	expectedDescriptions := map[string]string{
		"pipelines_total":               "Total number of pipelines",
		"pipeline_resolvers_total":      "Total number of pipeline resolvers",
		"pipeline_resolver_steps_total": "Total number of pipeline resolver steps",
		"tailordbs_total":               "Total number of TailorDBs",
		"tailordb_types_total":          "Total number of TailorDB types",
		"tailordb_type_fields_total":    "Total number of TailorDB type fields",
		"stateflows_total":              "Total number of StateFlows",
	}

	for _, metric := range metrics {
		expectedDesc, exists := expectedDescriptions[metric.Name]
		if !exists {
			t.Errorf("Unexpected metric: %s", metric.Name)
		}
		if expectedDesc != metric.Description {
			t.Errorf("Wrong description for metric %s: expected '%s', got '%s'",
				metric.Name, expectedDesc, metric.Description)
		}
	}
}

func TestMetric_Fields(t *testing.T) {
	metric := Metric{
		Name:        "test_metric",
		Description: "Test metric description",
		Value:       42.5,
	}

	if metric.Name != "test_metric" {
		t.Errorf("Expected Name to be 'test_metric', got '%s'", metric.Name)
	}
	if metric.Description != "Test metric description" {
		t.Errorf("Expected Description to be 'Test metric description', got '%s'", metric.Description)
	}
	if metric.Value != 42.5 {
		t.Errorf("Expected Value to be 42.5, got %f", metric.Value)
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
		metricMap[m.Name] = m.Value
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
}
