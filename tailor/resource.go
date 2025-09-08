package tailor

import (
	"context"

	tailorv1 "buf.build/gen/go/tailor-inc/tailor/protocolbuffers/go/tailor/v1"
	"connectrpc.com/connect"
)

type Resources struct {
	Applications []*Application
	Pipelines    []*Pipeline
	TailorDBs    []*TailorDB
	StateFlows   []*StateFlow
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
	Name          string
	Description   string
	Authorization string
	SDL           string
	PreHook       string
	PreScript     string
	PostScript    string
	PostHook      string
	Steps         []*PipelineStep
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
}

type TailorDB struct {
	NamespaceName string
	Types         []*TailorDBType
}

type TailorDBType struct {
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

type TailorDBField struct {
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
	Hooks       Hooks
}

type Hooks struct {
	Create     string
	Update     string
	CreateExpr string
	UpdateExpr string
}

type TailorDBPermission struct {
}

type TailorDBGQLPermission struct {
}

type TailorDBTypePermission struct {
}

type TailorDBRecordPermission struct {
}

type StateFlow struct {
	NamespaceName string
	AdminUsers    []*StateFlowAdminUser
}

type StateFlowAdminUser struct {
	UserID string
}

func (c *Client) Resources(ctx context.Context) (*Resources, error) {
	resources := &Resources{}

	// Pipeline Services
	{
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
									},
								}
								resolver.Steps = append(resolver.Steps, step)
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
	{
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
	{
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

// convertTailorDBFields converts proto FieldConfig map to TailorDBField slice
// It handles recursive field structures and preserves all field properties from protobuf
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
			Hooks: Hooks{
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
