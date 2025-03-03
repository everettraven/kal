package optionalorrequired

import (
	"fmt"
	"go/ast"

	"github.com/JoelSpeed/kal/pkg/analysis/helpers/markers"
	"github.com/JoelSpeed/kal/pkg/config"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

const (
	name = "optionalorrequired"

	// OptionalMarker is the marker that indicates that a field is optional.
	OptionalMarker = "optional"

	// RequiredMarker is the marker that indicates that a field is required.
	RequiredMarker = "required"

	// KubebuilderOptionalMarker is the marker that indicates that a field is optional in kubebuilder.
	KubebuilderOptionalMarker = "kubebuilder:validation:Optional"

	// KubebuilderRequiredMarker is the marker that indicates that a field is required in kubebuilder.
	KubebuilderRequiredMarker = "kubebuilder:validation:Required"
)

type analyzer struct {
	primaryOptionalMarker   string
	secondaryOptionalMarker string

	primaryRequiredMarker   string
	secondaryRequiredMarker string
}

// newAnalyzer creates a new analyzer with the given configuration.
func newAnalyzer(cfg config.OptionalOrRequiredConfig) (*analysis.Analyzer, error) {
	defaultConfig(&cfg)

	a := &analyzer{}

	switch cfg.PreferredOptionalMarker {
	case OptionalMarker:
		a.primaryOptionalMarker = OptionalMarker
		a.secondaryOptionalMarker = KubebuilderOptionalMarker
	case KubebuilderOptionalMarker:
		a.primaryOptionalMarker = KubebuilderOptionalMarker
		a.secondaryOptionalMarker = OptionalMarker
	}

	switch cfg.PreferredRequiredMarker {
	case RequiredMarker:
		a.primaryRequiredMarker = RequiredMarker
		a.secondaryRequiredMarker = KubebuilderRequiredMarker
	case KubebuilderRequiredMarker:
		a.primaryRequiredMarker = KubebuilderRequiredMarker
		a.secondaryRequiredMarker = RequiredMarker
	}

	return &analysis.Analyzer{
		Name:     name,
		Doc:      "Checks that all struct fields are marked either with the optional or required markers.",
		Run:      a.run,
		Requires: []*analysis.Analyzer{inspect.Analyzer, markers.Analyzer},
	}, nil
}

func (a *analyzer) run(pass *analysis.Pass) (interface{}, error) {
	inspect := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	markersAccess := pass.ResultOf[markers.Analyzer].(markers.Markers)

	// Filter to structs so that we can iterate over fields in a struct.
	nodeFilter := []ast.Node{
		(*ast.StructType)(nil),
	}

	inspect.Preorder(nodeFilter, func(n ast.Node) {
		sTyp, ok := n.(*ast.StructType)
		if !ok {
			return
		}

		if sTyp.Fields == nil {
			return
		}

		for _, field := range sTyp.Fields.List {
			if field == nil || len(field.Names) == 0 {
				continue
			}

			fieldName := field.Names[0].Name
			fieldMarkers := markersAccess.StructFieldMarkers(sTyp, fieldName)

			a.checkField(pass, field, fieldMarkers)
		}
	})

	return nil, nil
}

func (a *analyzer) checkField(pass *analysis.Pass, field *ast.Field, fieldMarkers markers.MarkerSet) {
	if field == nil || len(field.Names) == 0 {
		return
	}

	fieldName := field.Names[0].Name

	hasPrimaryOptional := fieldMarkers.Has(a.primaryOptionalMarker)
	hasPrimaryRequired := fieldMarkers.Has(a.primaryRequiredMarker)

	hasSecondaryOptional := fieldMarkers.Has(a.secondaryOptionalMarker)
	hasSecondaryRequired := fieldMarkers.Has(a.secondaryRequiredMarker)

	hasEitherOptional := hasPrimaryOptional || hasSecondaryOptional
	hasEitherRequired := hasPrimaryRequired || hasSecondaryRequired

	hasBothOptional := hasPrimaryOptional && hasSecondaryOptional
	hasBothRequired := hasPrimaryRequired && hasSecondaryRequired

	switch {
	case hasEitherOptional && hasEitherRequired:
		pass.Reportf(field.Pos(), "field %s must not be marked as both optional and required", fieldName)
	case hasSecondaryOptional:
		marker := fieldMarkers[a.secondaryOptionalMarker]
		if hasBothOptional {
			pass.Report(reportShouldRemoveSecondaryMarker(field, marker, a.primaryOptionalMarker, a.secondaryOptionalMarker))
		} else {
			pass.Report(reportShouldReplaceSecondaryMarker(field, marker, a.primaryOptionalMarker, a.secondaryOptionalMarker))
		}
	case hasSecondaryRequired:
		marker := fieldMarkers[a.secondaryRequiredMarker]
		if hasBothRequired {
			pass.Report(reportShouldRemoveSecondaryMarker(field, marker, a.primaryRequiredMarker, a.secondaryRequiredMarker))
		} else {
			pass.Report(reportShouldReplaceSecondaryMarker(field, marker, a.primaryRequiredMarker, a.secondaryRequiredMarker))
		}
	case hasPrimaryOptional || hasPrimaryRequired:
		// This is the correct state.
	default:
		pass.Reportf(field.Pos(), "field %s must be marked as %s or %s", fieldName, a.primaryOptionalMarker, a.primaryRequiredMarker)
	}
}

func reportShouldReplaceSecondaryMarker(field *ast.Field, marker markers.Marker, primaryMarker, secondaryMarker string) analysis.Diagnostic {
	fieldName := field.Names[0].Name

	return analysis.Diagnostic{
		Pos:     field.Pos(),
		Message: fmt.Sprintf("field %s should use marker %s instead of %s", fieldName, primaryMarker, secondaryMarker),
		SuggestedFixes: []analysis.SuggestedFix{
			{
				Message: fmt.Sprintf("should replace `%s` with `%s`", secondaryMarker, primaryMarker),
				TextEdits: []analysis.TextEdit{
					{
						Pos:     marker.Pos,
						End:     marker.End,
						NewText: []byte(fmt.Sprintf("// +%s", primaryMarker)),
					},
				},
			},
		},
	}
}

func reportShouldRemoveSecondaryMarker(field *ast.Field, marker markers.Marker, primaryMarker, secondaryMarker string) analysis.Diagnostic {
	fieldName := field.Names[0].Name

	return analysis.Diagnostic{
		Pos:     field.Pos(),
		Message: fmt.Sprintf("field %s should use only the marker %s, %s is not required", fieldName, primaryMarker, secondaryMarker),
		SuggestedFixes: []analysis.SuggestedFix{
			{
				Message: fmt.Sprintf("should remove `// +%s`", secondaryMarker),
				TextEdits: []analysis.TextEdit{
					{
						Pos:     marker.Pos,
						End:     marker.End + 1, // Add 1 to position to include the new line
						NewText: nil,
					},
				},
			},
		},
	}
}

func defaultConfig(cfg *config.OptionalOrRequiredConfig) {
	if cfg.PreferredOptionalMarker == "" {
		cfg.PreferredOptionalMarker = OptionalMarker
	}

	if cfg.PreferredRequiredMarker == "" {
		cfg.PreferredRequiredMarker = RequiredMarker
	}
}
