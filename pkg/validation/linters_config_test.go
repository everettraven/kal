package validation_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/JoelSpeed/kal/pkg/config"
	"github.com/JoelSpeed/kal/pkg/validation"

	"k8s.io/apimachinery/pkg/util/validation/field"
)

var _ = Describe("LintersConfig", func() {
	type validateLintersConfigTableInput struct {
		config      config.LintersConfig
		expectedErr string
	}

	DescribeTable("Validate Linters Configuration", func(in validateLintersConfigTableInput) {
		errs := validation.ValidateLintersConfig(in.config, field.NewPath("lintersConfig"))
		if len(in.expectedErr) > 0 {
			Expect(errs.ToAggregate()).To(MatchError(in.expectedErr))
		} else {
			Expect(errs).To(HaveLen(0), "No errors were expected")
		}

	},
		Entry("Empty config", validateLintersConfigTableInput{
			config:      config.LintersConfig{},
			expectedErr: "",
		}),

		// JSONTagsConfig validation
		Entry("With a valid JSONTagsConfig JSONTagRegex", validateLintersConfigTableInput{
			config: config.LintersConfig{
				JSONTags: config.JSONTagsConfig{
					JSONTagRegex: "^[a-z][a-z0-9]*(?:[A-Z][a-z0-9]*)*$",
				},
			},
			expectedErr: "",
		}),
		Entry("With an invalid JSONTagsConfig JSONTagRegex", validateLintersConfigTableInput{
			config: config.LintersConfig{
				JSONTags: config.JSONTagsConfig{
					JSONTagRegex: "^[a-z][a-z0-9]*(?:[A-Z][a-z0-9]*",
				},
			},
			expectedErr: "lintersConfig.jsonTags.jsonTagRegex: Invalid value: \"^[a-z][a-z0-9]*(?:[A-Z][a-z0-9]*\": invalid regex: error parsing regexp: missing closing ): `^[a-z][a-z0-9]*(?:[A-Z][a-z0-9]*`",
		}),

		// OptionalOrRequiredConfig validation
		Entry("With a valid OptionalOrRequiredConfig", validateLintersConfigTableInput{
			config: config.LintersConfig{
				OptionalOrRequired: config.OptionalOrRequiredConfig{
					PreferredOptionalMarker: "optional",
					PreferredRequiredMarker: "required",
				},
			},
			expectedErr: "",
		}),
		Entry("With kubebuilder preferred markers", validateLintersConfigTableInput{
			config: config.LintersConfig{
				OptionalOrRequired: config.OptionalOrRequiredConfig{
					PreferredOptionalMarker: "kubebuilder:validation:Optional",
					PreferredRequiredMarker: "kubebuilder:validation:Required",
				},
			},
			expectedErr: "",
		}),
		Entry("With invalid preferred optional marker", validateLintersConfigTableInput{
			config: config.LintersConfig{
				OptionalOrRequired: config.OptionalOrRequiredConfig{
					PreferredOptionalMarker: "invalid",
				},
			},
			expectedErr: "lintersConfig.optionalOrRequired.preferredOptionalMarker: Invalid value: \"invalid\": invalid value, must be one of \"optional\", \"kubebuilder:validation:Optional\" or omitted",
		}),
		Entry("With invalid preferred required marker", validateLintersConfigTableInput{
			config: config.LintersConfig{
				OptionalOrRequired: config.OptionalOrRequiredConfig{
					PreferredRequiredMarker: "invalid",
				},
			},
			expectedErr: "lintersConfig.optionalOrRequired.preferredRequiredMarker: Invalid value: \"invalid\": invalid value, must be one of \"required\", \"kubebuilder:validation:Required\" or omitted",
		}),
	)
})
