package tailor

import (
	"fmt"
	"regexp"
	"slices"

	tailorv1 "buf.build/gen/go/tailor-inc/tailor/protocolbuffers/go/tailor/v1"
	"github.com/vektah/gqlparser/ast"
	"github.com/vektah/gqlparser/parser"
)

type LintTargetType string

const (
	LintTargetTypePipeline  LintTargetType = "pipeline"
	LintTargetTypeTailorDB  LintTargetType = "tailordb"
	LintTargetTypeStateFlow LintTargetType = "stateflow"
)

type LintWarn struct {
	Type    LintTargetType
	Name    string
	Message string
}

var draftMutationPrefixRe = regexp.MustCompile(`^(appendDraft|confirmDraft|cancelDraft)`)
var stateFlowMutations = []string{"newState", "moveState"}

func (c *Client) Lint(resources *Resources) ([]*LintWarn, error) {
	var warns []*LintWarn
	var typeNames []string

	// TailorDB Linting
	for _, db := range resources.TailorDBs {
		for _, t := range db.Types {
			typeNames = append(typeNames, t.Name)

			if c.cfg.Lint.TailorDB.DeprecatedFeature.Enabled {
				if !c.cfg.Lint.TailorDB.DeprecatedFeature.AllowDraft && t.Draft {
					warns = append(warns, &LintWarn{
						Type:    LintTargetTypeTailorDB,
						Name:    fmt.Sprintf("%s/%s", db.NamespaceName, t.Name),
						Message: "Draft feature is deprecated",
					})
				}
				if !c.cfg.Lint.TailorDB.DeprecatedFeature.AllowTypePermission && t.TypePermission != nil {
					warns = append(warns, &LintWarn{
						Type:    LintTargetTypeTailorDB,
						Name:    fmt.Sprintf("%s/%s", db.NamespaceName, t.Name),
						Message: "Type-level permission is deprecated. Use `Permission` or `GQLPermission` instead",
					})
				}
				if !c.cfg.Lint.TailorDB.DeprecatedFeature.AllowRecordPermission && t.RecordPermission != nil {
					warns = append(warns, &LintWarn{
						Type:    LintTargetTypeTailorDB,
						Name:    fmt.Sprintf("%s/%s", db.NamespaceName, t.Name),
						Message: "Record-level permission is deprecated. Use `Permission` or `GQLPermission` instead",
					})
				}
			}

			if c.cfg.Lint.TailorDB.DeprecatedFeature.Enabled && !c.cfg.Lint.TailorDB.DeprecatedFeature.AllowCELHooks {
				for _, f := range t.Fields {
					if f.Hooks.CreateExpr != "" || f.Hooks.UpdateExpr != "" {
						warns = append(warns, &LintWarn{
							Type:    LintTargetTypeTailorDB,
							Name:    fmt.Sprintf("%s/%s field %s", db.NamespaceName, t.Name, f.Name),
							Message: "Hooks `create_expr` and `update_expr` are deprecated. Use `create` or `update` instead",
						})
					}
				}
			}
		}
	}

	// Pipeline Linting
	for _, p := range resources.Pipelines {
		for _, r := range p.Resolvers {
			// Pipeline/InsecureAuthorization
			if c.cfg.Lint.Pipeline.InsecureAuthorization.Enabled && (r.Authorization == "true" || r.Authorization == "true==true") {
				warns = append(warns, &LintWarn{
					Type:    LintTargetTypePipeline,
					Name:    fmt.Sprintf("%s/%s", p.NamespaceName, r.Name),
					Message: "resolver allows insecure authorization",
				})
			}

			stepLength := len(r.Steps)
			// Pipeline/StepLength
			if c.cfg.Lint.Pipeline.StepLength.Enabled && stepLength > c.cfg.Lint.Pipeline.StepLength.Max {
				warns = append(warns, &LintWarn{
					Type:    LintTargetTypePipeline,
					Name:    fmt.Sprintf("%s/%s", p.NamespaceName, r.Name),
					Message: fmt.Sprintf("resolver has too many steps (%d > %d)", stepLength, c.cfg.Lint.Pipeline.StepLength.Max),
				})
			}

			var operations []string
			for _, s := range r.Steps {
				if s.Operation.Type == tailorv1.PipelineResolver_OPERATION_TYPE_GRAPHQL {
					query, err := parser.ParseQuery(&ast.Source{
						Input: s.Operation.Source,
					})
					if err != nil {
						return nil, fmt.Errorf("failed to parse GraphQL operation in %s/%s step %s: %w", p.NamespaceName, r.Name, s.Name, err)
					}
					for _, op := range query.Operations {
						operations = append(operations, string(op.Operation))
						for _, selection := range op.SelectionSet {
							switch sel := selection.(type) {
							case *ast.Field:
								if c.cfg.Lint.Pipeline.DeprecatedFeature.Enabled {
									// StateFlow
									if !c.cfg.Lint.Pipeline.DeprecatedFeature.AllowStateFlow {
										if slices.Contains(stateFlowMutations, sel.Name) {
											warns = append(warns, &LintWarn{
												Type:    LintTargetTypePipeline,
												Name:    fmt.Sprintf("%s/%s step %s", p.NamespaceName, r.Name, s.Name),
												Message: fmt.Sprintf("StateFlow feature is deprecated (found usage of %s)", sel.Name),
											})
										}
									}

									// Draft
									if !c.cfg.Lint.Pipeline.DeprecatedFeature.AllowDraft {
										if replaced := draftMutationPrefixRe.ReplaceAllString(sel.Name, ""); replaced != sel.Name {
											if slices.Contains(typeNames, replaced) {
												warns = append(warns, &LintWarn{
													Type:    LintTargetTypePipeline,
													Name:    fmt.Sprintf("%s/%s step %s", p.NamespaceName, r.Name, s.Name),
													Message: fmt.Sprintf("Draft feature is deprecated (found usage of %s)", sel.Name),
												})
											}
										}
									}
								}
							}
						}
					}
				}
				if c.cfg.Lint.Pipeline.DeprecatedFeature.Enabled && !c.cfg.Lint.Pipeline.DeprecatedFeature.AllowCELScript {
					if s.PreValidation != "" {
						warns = append(warns, &LintWarn{
							Type:    LintTargetTypePipeline,
							Name:    fmt.Sprintf("%s/%s step %s", p.NamespaceName, r.Name, s.Name),
							Message: "`pre_validation` is deprecated. Use `pre_hook` instead.",
						})
					}
					if s.PreScript != "" {
						warns = append(warns, &LintWarn{
							Type:    LintTargetTypePipeline,
							Name:    fmt.Sprintf("%s/%s step %s", p.NamespaceName, r.Name, s.Name),
							Message: "`pre_script` is deprecated. Use `pre_hook` instead.",
						})
					}
					if s.PostScript != "" {
						warns = append(warns, &LintWarn{
							Type:    LintTargetTypePipeline,
							Name:    fmt.Sprintf("%s/%s step %s", p.NamespaceName, r.Name, s.Name),
							Message: "`post_script` is deprecated. Use `post_hook` instead.",
						})
					}
					if s.PostValidation != "" {
						warns = append(warns, &LintWarn{
							Type:    LintTargetTypePipeline,
							Name:    fmt.Sprintf("%s/%s step %s", p.NamespaceName, r.Name, s.Name),
							Message: "`post_validation` is deprecated. Use `post_hook` instead.",
						})
					}
				}
			}
			if c.cfg.Lint.Pipeline.MultipleMutations.Enabled {
				var count int
				for _, op := range operations {
					if op == "mutation" {
						count++
					}
				}
				if count > 1 {
					warns = append(warns, &LintWarn{
						Type:    LintTargetTypePipeline,
						Name:    fmt.Sprintf("%s/%s", p.NamespaceName, r.Name),
						Message: "Resolver has multiple mutations. Because transactions are not applied between steps, it is recommended to use transaction within function.",
					})
				}
			}
			if c.cfg.Lint.Pipeline.QueryBeforeMutation.Enabled {
				if slices.Contains(operations, "mutation") && slices.Contains(operations, "query") && slices.Index(operations, "mutation") > slices.Index(operations, "query") {
					warns = append(warns, &LintWarn{
						Type:    LintTargetTypePipeline,
						Name:    fmt.Sprintf("%s/%s", p.NamespaceName, r.Name),
						Message: "Resolver has query before mutation. Because transactions are not applied between steps, it is recommended to use transaction within function.",
					})
				}
			}
		}
	}

	// StateFlow Linting
	if c.cfg.Lint.StateFlow.DeprecatedFeature.Enabled {
		for _, sf := range resources.StateFlows {
			if !c.cfg.Lint.StateFlow.DeprecatedFeature.Enabled {
				continue
			}
			warns = append(warns, &LintWarn{
				Type:    LintTargetTypeStateFlow,
				Name:    sf.NamespaceName,
				Message: "StateFlow is deprecated",
			})
		}
	}

	return warns, nil
}
