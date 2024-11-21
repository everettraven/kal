package analysis

import (
	"github.com/JoelSpeed/kal/pkg/analysis/commentstart"
	"github.com/JoelSpeed/kal/pkg/analysis/jsontags"
	"github.com/JoelSpeed/kal/pkg/analysis/optionalorrequired"
	"github.com/JoelSpeed/kal/pkg/config"
	"golang.org/x/tools/go/analysis"
	"k8s.io/apimachinery/pkg/util/sets"
)

// AnalyzerInitializer is used to intializer analyzers.
type AnalyzerInitializer interface {
	// Name returns the name of the analyzer initialized by this intializer.
	Name() string

	// Init returns the newly initialized analyzer.
	// It will be passed the complete LintersConfig and is expected to rely only on its own configuration.
	Init(config.LintersConfig) *analysis.Analyzer

	// Default determines whether the inializer intializes an analyzer that should be
	// on by default, or not.
	Default() bool
}

type Registry interface {
	// DefaultLinters returns the names of linters that are enabled by default.
	DefaultLinters() sets.Set[string]

	// AllLinters returns the names of all registered linters.
	AllLinters() sets.Set[string]

	// InitializeLinters returns a set of newly initialized linters based on the
	// provided configuration.
	InitializeLinters(config.Linters, config.LintersConfig) []*analysis.Analyzer
}

type registry struct {
	initializers []AnalyzerInitializer
}

// NewRegistry returns a new registry, from which analyzers can be fetched.
func NewRegistry() Registry {
	return &registry{
		initializers: []AnalyzerInitializer{
			commentstart.Initializer(),
			jsontags.Initializer(),
			optionalorrequired.Initializer(),
		},
	}
}

// DefaultLinters returns the list of linters that are registered
// as being enabled by default.
func (r *registry) DefaultLinters() sets.Set[string] {
	defaultLinters := sets.New[string]()

	for _, initializer := range r.initializers {
		if initializer.Default() {
			defaultLinters.Insert(initializer.Name())
		}
	}

	return defaultLinters
}

// AllLinters returns the list of all known linters that are known
// to the registry.
func (r *registry) AllLinters() sets.Set[string] {
	linters := sets.New[string]()

	for _, initializer := range r.initializers {
		linters.Insert(initializer.Name())
	}

	return linters
}

// InitializeLinters returns a list of initialized linters based on the provided config.
func (r *registry) InitializeLinters(cfg config.Linters, lintersCfg config.LintersConfig) []*analysis.Analyzer {
	analyzers := []*analysis.Analyzer{}

	enabled := sets.New(cfg.Enable...)
	disabled := sets.New(cfg.Disable...)

	allEnabled := enabled.Len() == 1 && enabled.Has(config.Wildcard)
	allDisabled := disabled.Len() == 1 && disabled.Has(config.Wildcard)

	for _, initializer := range r.initializers {
		if !disabled.Has(initializer.Name()) && (allEnabled || enabled.Has(initializer.Name()) || !allDisabled && initializer.Default()) {
			analyzers = append(analyzers, initializer.Init(lintersCfg))
		}
	}

	return analyzers
}