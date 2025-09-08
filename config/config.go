package config

import (
	"os"
	"path/filepath"

	"github.com/creasty/defaults"
	"github.com/goccy/go-yaml"
)

type Config struct {
	WorkspaceID string `default:"" yaml:"workspaceID,omitempty"`
	Lint        Lint   `yaml:"lint,omitempty"`
}

type Lint struct {
	Pipeline  Pipeline  `yaml:"pipeline,omitempty,omitzero"`
	TailorDB  TailorDB  `yaml:"tailordb,omitempty,omitzero"`
	StateFlow StateFlow `yaml:"stateflow,omitempty,omitzero"`
}

type Pipeline struct {
	DeprecatedFeature     PipelineDeprecatedFeature `yaml:"deprecatedFeature,omitempty,omitzero"`
	InsecureAuthorization InsecureAuthorization     `yaml:"insecureAuthorization,omitempty,omitzero"`
	StepLength            StepLength                `yaml:"stepLength,omitempty,omitzero"`
	MultipleMutations     MultipleMutations         `yaml:"multipleMutations,omitempty,omitzero"`
	QueryBeforeMutation   QueryBeforeMutation       `yaml:"queryBeforeMutation,omitempty,omitzero"`
	LegacyScript          LegacyScript              `yaml:"legacyScript,omitempty,omitzero"`
}

type PipelineDeprecatedFeature struct {
	Enabled        bool `default:"true" yaml:"enabled,omitempty"`
	AllowDraft     bool `default:"false" yaml:"allowDraft,omitempty"`
	AllowStateFlow bool `default:"false" yaml:"allowStateFlow,omitempty"`
}

type InsecureAuthorization struct {
	Enabled bool `default:"true" yaml:"enabled,omitempty"`
}

type StepLength struct {
	Enabled bool `default:"true" yaml:"enabled,omitempty"`
	Max     int  `default:"30" yaml:"max,omitempty"`
}

type MultipleMutations struct {
	Enabled bool `default:"true" yaml:"enabled,omitempty"`
}

type QueryBeforeMutation struct {
	Enabled bool `default:"true" yaml:"enabled,omitempty"`
}

type LegacyScript struct {
	Enabled bool `default:"true" yaml:"enabled,omitempty"`
}

type TailorDB struct {
	DeprecatedFeature TailorDBDeprecatedFeature `yaml:"deprecatedFeature,omitempty,omitzero"`
	LegacyPermission  LegacyPermission          `yaml:"legacyPermission,omitempty,omitzero"`
}

type TailorDBDeprecatedFeature struct {
	Enabled       bool `default:"true" yaml:"enabled,omitempty"`
	AllowDraft    bool `default:"false" yaml:"allowDraft,omitempty"`
	AllowCELHooks bool `default:"false" yaml:"allowCELHooks,omitempty"`
}

type LegacyPermission struct {
	Enabled               bool `default:"true" yaml:"enabled,omitempty"`
	AllowTypePermission   bool `default:"false" yaml:"allowTypePermission,omitempty"`
	AllowRecordPermission bool `default:"false" yaml:"allowRecordPermission,omitempty"`
}

type StateFlow struct {
	DeprecatedFeature StateFlowDeprecatedFeature `yaml:"deprecatedFeature,omitempty,omitzero"`
}

type StateFlowDeprecatedFeature struct {
	Enabled bool `default:"true" yaml:"enabled,omitempty"`
}

const Filename = ".patterner.yml"

func New() (*Config, error) {
	c := &Config{}
	if err := defaults.Set(c); err != nil {
		return nil, err
	}
	return c, nil
}

func Load() (*Config, error) {
	c, err := New()
	if err != nil {
		return nil, err
	}
	wd, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	for {
		b, err := os.ReadFile(filepath.Join(wd, Filename))
		if err == nil {
			if err := yaml.Unmarshal(b, c); err != nil {
				return nil, err
			}
			return c, nil
		}
		if !os.IsNotExist(err) {
			return nil, err
		}
		wd := filepath.Dir(wd)
		if wd == "/" || wd == "." {
			break
		}
	}
	return c, nil
}
