package tailor

import (
	"context"
	"time"

	tailorv1 "buf.build/gen/go/tailor-inc/tailor/protocolbuffers/go/tailor/v1"
	"connectrpc.com/connect"
)

type Resources struct {
	Applications []*Application
	Pipelines    []*Pipeline
	TailorDBs    []*TailorDB
	StateFlows   []*StateFlow

	// Options
	withoutApplications   bool
	withoutTailorDB       bool
	withoutPipeline       bool
	withoutStateFlow      bool
	executionResultsSince *time.Time
}

type Application struct {
	Name string
}

type Pipeline struct {
	NamespaceName string
	CommonSDL     string
	Resolvers     []*PipelineResolver
}

type PipelineResolver struct {
	Name             string
	Description      string
	Authorization    string
	SDL              string
	PreHook          string
	PreScript        string
	PostScript       string
	PostHook         string
	Steps            []*PipelineStep
	ExecutionResults []*tailorv1.PipelineResolverExecutionResult
}

type PipelineStep struct {
	Name           string
	Description    string
	PreValidation  string
	PreScript      string
	PreHook        string
	PostScript     string
	PostValidation string
	PostHook       string
	Operation      PipelineStepOperation
}

type PipelineStepOperation struct {
	Type    tailorv1.PipelineResolver_OperationType
	Name    string
	Invoker *tailorv1.AuthInvoker
	Source  string
	Test    string
}

type TailorDB struct { //nolint:revive
	NamespaceName string
	Types         []*TailorDBType
}

type TailorDBType struct { //nolint:revive
	Name          string
	Description   string
	Fields        []*TailorDBField
	Permission    *TailorDBPermission
	GQLPermission *TailorDBGQLPermission
	// Legacy Permission
	TypePermission   *TailorDBTypePermission
	RecordPermission *TailorDBRecordPermission
	// Draft
	Draft bool
}

type TailorDBField struct { //nolint:revive
	Name        string
	Type        string
	Description string
	Fields      []*TailorDBField
	Required    bool
	Array       bool
	Index       bool
	Unique      bool
	ForeignKey  bool
	Vector      bool
	SourceID    *string
	Hooks       TailorDBFieldHooks
}

type TailorDBFieldHooks struct { //nolint:revive
	Create     string
	Update     string
	CreateExpr string
	UpdateExpr string
}

type TailorDBPermission struct { //nolint:revive
}

type TailorDBGQLPermission struct { //nolint:revive
}

type TailorDBTypePermission struct { //nolint:revive
}

type TailorDBRecordPermission struct { //nolint:revive
}

type StateFlow struct {
	NamespaceName string
	AdminUsers    []*StateFlowAdminUser
}

type StateFlowAdminUser struct {
	UserID string
}

type ResourceOption func(*Resources) error

func WithoutApplications() ResourceOption {
	return func(r *Resources) error {
		r.withoutApplications = true
		return nil
	}
}

func WithoutTailorDB() ResourceOption {
	return func(r *Resources) error {
		r.withoutTailorDB = true
		return nil
	}
}

func WithoutPipeline() ResourceOption {
	return func(r *Resources) error {
		r.withoutPipeline = true
		return nil
	}
}

func WithoutStateFlow() ResourceOption {
	return func(r *Resources) error {
		r.withoutStateFlow = true
		return nil
	}
}

func WithExecutionResults(since *time.Time) ResourceOption {
	return func(r *Resources) error {
		r.withoutPipeline = false
		r.executionResultsSince = since
		return nil
	}
}

func (c *Client) Resources(ctx context.Context, opts ...ResourceOption) (*Resources, error) {
	resources := &Resources{}
	for _, opt := range opts {
		if err := opt(resources); err != nil {
			return nil, err
		}
	}
	// Pipeline Services
	if !resources.withoutPipeline {
		pageToken := ""
		for {
			res, err := c.client.ListPipelineServices(ctx, connect.NewRequest(&tailorv1.ListPipelineServicesRequest{
				WorkspaceId: c.cfg.WorkspaceID,
				PageSize:    pageSize,
				PageToken:   pageToken,
			}))
			if err != nil {
				return nil, err
			}
			for _, p := range res.Msg.GetPipelineServices() {
				pipeline := &Pipeline{
					NamespaceName: p.GetNamespace().GetName(),
					CommonSDL:     p.GetCommonSdl(),
				}
				// Pipeline Resolvers
				{
					pageToken := ""
					for {
						res, err := c.client.ListPipelineResolvers(ctx, connect.NewRequest(&tailorv1.ListPipelineResolversRequest{
							WorkspaceId:   c.cfg.WorkspaceID,
							NamespaceName: p.GetNamespace().GetName(),
							PageSize:      pageSize,
							PageToken:     pageToken,
						}))
						if err != nil {
							return nil, err
						}
						for _, r := range res.Msg.GetPipelineResolvers() {
							res, err := c.client.GetPipelineResolver(ctx, connect.NewRequest(&tailorv1.GetPipelineResolverRequest{
								WorkspaceId:   c.cfg.WorkspaceID,
								NamespaceName: p.GetNamespace().GetName(),
								ResolverName:  r.GetName(),
							}))
							if err != nil {
								return nil, err
							}
							rr := res.Msg.GetPipelineResolver()
							resolver := &PipelineResolver{
								Name:          rr.GetName(),
								Description:   rr.GetDescription(),
								Authorization: rr.GetAuthorization(),
								SDL:           rr.GetSdl(),
								PreHook:       rr.GetPreHook().GetExpr(),
								PreScript:     rr.GetPreScript(),
								PostScript:    rr.GetPostScript(),
								PostHook:      rr.GetPostHook().GetExpr(),
							}
							hasTest := false
							for _, p := range rr.GetPipelines() {
								step := &PipelineStep{
									Name:           p.GetName(),
									Description:    p.GetDescription(),
									PreValidation:  p.GetPreValidation(),
									PreScript:      p.GetPreScript(),
									PreHook:        p.GetPreHook().GetExpr(),
									PostScript:     p.GetPostScript(),
									PostValidation: p.GetPostValidation(),
									PostHook:       p.GetPostHook().GetExpr(),
									Operation: PipelineStepOperation{
										Type:    p.GetOperationType(),
										Name:    p.GetOperationName(),
										Invoker: p.GetInvoker(),
										Source:  p.GetOperationSource(),
										Test:    p.GetTest(),
									},
								}
								if p.GetTest() != "" {
									hasTest = true
								}
								resolver.Steps = append(resolver.Steps, step)
							}

							// Pipeline Resolvers Execution Results
							if resources.executionResultsSince != nil {
								pageToken := ""
								view := tailorv1.PipelineResolverExecutionResultView_PIPELINE_RESOLVER_EXECUTION_RESULT_VIEW_BASIC
								if hasTest {
									// Because branching occurs, context information is required.
									view = tailorv1.PipelineResolverExecutionResultView_PIPELINE_RESOLVER_EXECUTION_RESULT_VIEW_FULL
								}
							L:
								for {
									res, err := c.client.ListPipelineResolverExecutionResults(ctx, connect.NewRequest(&tailorv1.ListPipelineResolverExecutionResultsRequest{
										WorkspaceId:   c.cfg.WorkspaceID,
										NamespaceName: p.GetNamespace().GetName(),
										ResolverName:  r.GetName(),
										View:          view,
										PageSize:      pageSize,
										PageToken:     pageToken,
									}))
									if err != nil {
										return nil, err
									}
									for _, r := range res.Msg.GetResults() {
										if r.GetCreatedAt().AsTime().Before(*resources.executionResultsSince) {
											// Since the results are ordered by CreatedAt descending,
											// we can stop fetching more results once we reach an older entry.
											pageToken = ""
											break L
										}
										resolver.ExecutionResults = append(resolver.ExecutionResults, r)
									}
									if res.Msg.GetNextPageToken() == "" {
										break
									}
									pageToken = res.Msg.GetNextPageToken()
								}
							}

							pipeline.Resolvers = append(pipeline.Resolvers, resolver)
						}
						if res.Msg.GetNextPageToken() == "" {
							break
						}
						pageToken = res.Msg.GetNextPageToken()
					}
				}
				resources.Pipelines = append(resources.Pipelines, pipeline)
			}
			if res.Msg.GetNextPageToken() == "" {
				break
			}
			pageToken = res.Msg.GetNextPageToken()
		}
	}

	// TailorDB Services
	if !resources.withoutTailorDB {
		pageToken := ""
		for {
			res, err := c.client.ListTailorDBServices(ctx, connect.NewRequest(&tailorv1.ListTailorDBServicesRequest{
				WorkspaceId: c.cfg.WorkspaceID,
				PageSize:    pageSize,
				PageToken:   pageToken,
			}))
			if err != nil {
				return nil, err
			}
			for _, t := range res.Msg.GetTailordbServices() {
				tailordb := &TailorDB{
					NamespaceName: t.GetNamespace().GetName(),
				}
				// TailorDB Types
				{
					pageToken := ""
					for {
						res, err := c.client.ListTailorDBTypes(ctx, connect.NewRequest(&tailorv1.ListTailorDBTypesRequest{
							WorkspaceId:   c.cfg.WorkspaceID,
							NamespaceName: t.GetNamespace().GetName(),
							PageSize:      pageSize,
							PageToken:     pageToken,
						}))
						if err != nil {
							return nil, err
						}
						for _, tt := range res.Msg.GetTailordbTypes() {
							res, err := c.client.GetTailorDBType(ctx, connect.NewRequest(&tailorv1.GetTailorDBTypeRequest{
								WorkspaceId:      c.cfg.WorkspaceID,
								NamespaceName:    t.GetNamespace().GetName(),
								TailordbTypeName: tt.GetName(),
							}))
							if err != nil {
								return nil, err
							}
							ttt := res.Msg.GetTailordbType()
							tailordbType := &TailorDBType{
								Name:        ttt.GetName(),
								Description: ttt.GetSchema().GetDescription(),
								Draft:       ttt.GetSchema().GetSettings().GetDraft(),
							}
							tailordbType.Fields = convertTailorDBFields(ttt.GetSchema().GetFields())
							tailordb.Types = append(tailordb.Types, tailordbType)
							if ttt.GetSchema().GetPermission() != nil {
								tailordbType.Permission = &TailorDBPermission{}
							}
							if ttt.GetSchema().GetTypePermission() != nil {
								tailordbType.TypePermission = &TailorDBTypePermission{}
							}
							if ttt.GetSchema().GetRecordPermission() != nil {
								tailordbType.RecordPermission = &TailorDBRecordPermission{}
							}
							if _, err := c.client.GetTailorDBGQLPermission(ctx, connect.NewRequest(&tailorv1.GetTailorDBGQLPermissionRequest{
								WorkspaceId:   c.cfg.WorkspaceID,
								NamespaceName: t.GetNamespace().GetName(),
								TypeName:      tt.GetName(),
							})); err == nil {
								tailordbType.GQLPermission = &TailorDBGQLPermission{}
							}
						}
						if res.Msg.GetNextPageToken() == "" {
							break
						}
						pageToken = res.Msg.GetNextPageToken()
					}
				}
				resources.TailorDBs = append(resources.TailorDBs, tailordb)
			}
			if res.Msg.GetNextPageToken() == "" {
				break
			}
			pageToken = res.Msg.GetNextPageToken()
		}
	}

	// StateFlow Services
	if !resources.withoutStateFlow {
		pageToken := ""
		for {
			res, err := c.client.ListStateflowServices(ctx, connect.NewRequest(&tailorv1.ListStateflowServicesRequest{
				WorkspaceId: c.cfg.WorkspaceID,
				PageSize:    pageSize,
				PageToken:   pageToken,
			}))
			if err != nil {
				return nil, err
			}
			for _, s := range res.Msg.GetStateflowServices() {
				stateflow := &StateFlow{
					NamespaceName: s.GetNamespace().GetName(),
				}
				// StateFlow Admin Users
				for _, admin := range s.GetAdminUsers() {
					adminUser := &StateFlowAdminUser{
						UserID: admin.GetUserId(),
					}
					stateflow.AdminUsers = append(stateflow.AdminUsers, adminUser)
				}
				resources.StateFlows = append(resources.StateFlows, stateflow)
			}
			if res.Msg.GetNextPageToken() == "" {
				break
			}
			pageToken = res.Msg.GetNextPageToken()
		}
	}

	return resources, nil
}

// convertTailorDBFields converts proto FieldConfig map to TailorDBField slice.
func convertTailorDBFields(fields map[string]*tailorv1.TailorDBType_FieldConfig) []*TailorDBField {
	if fields == nil {
		return nil
	}

	// Pre-allocate slice with known capacity for better performance
	result := make([]*TailorDBField, 0, len(fields))

	for name, config := range fields {
		if config == nil {
			// Skip nil configs but continue processing other fields
			continue
		}

		// Validate required fields
		if name == "" {
			// Skip fields with empty names but continue processing
			continue
		}

		field := &TailorDBField{
			Name:        name,
			Type:        config.GetType(),
			Description: config.GetDescription(),
			Required:    config.GetRequired(),
			Array:       config.GetArray(),
			Index:       config.GetIndex(),
			Unique:      config.GetUnique(),
			ForeignKey:  config.GetForeignKey(),
			Vector:      config.GetVector(),
			Hooks: TailorDBFieldHooks{
				Create:     config.GetHooks().GetCreate().GetExpr(),
				Update:     config.GetHooks().GetUpdate().GetExpr(),
				CreateExpr: config.GetHooks().GetCreateExpr(),
				UpdateExpr: config.GetHooks().GetUpdateExpr(),
			},
		}

		// Set SourceID if available
		if config.HasSourceId() {
			sourceID := config.GetSourceId()
			field.SourceID = &sourceID
		}

		// Convert nested fields recursively
		if nestedFields := config.GetFields(); len(nestedFields) > 0 {
			field.Fields = convertTailorDBFields(nestedFields)
		}

		result = append(result, field)
	}

	return result
}
