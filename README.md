# KAL - The Kubernetes API Linter

KAL is a Golang based linter for Kubernetes API types. It checks for common mistakes and enforces best practices.
The rules implemented by KAL, are based on the [Kubernetes API Conventions][api-conventions].

KAL is aimed at being an assistant to API review, by catching the mechanical elements of API review, and allowing reviewers to focus on the more complex aspects of API design.

[api-conventions]: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md

## Installation

KAL currently comes in two flavours, a standalone binary, and a golangci-lint plugin.

### Standalone Binary

To install the standalone binary, run the following command:

```shell
go install github.com/JoelSpeed/kal/cmd/kal@latest
```

The standalone binary can be run with the following command:

```shell
kal path/to/api/types
```

`kal` currently accepts no complex configuration, and will run all checks considered to be default.

Individual linters can be disabled with the flag corresponding to the linter name. For example, to disable the `commentstart` linter, run the following command:

```shell
kal -commentstart=false path/to/api/types
```

Where fixes are available, these can be applied automatically with the `-fix` flag.
Note, automatic fixing is currently only available via the standalone binary, and is not available via the `golangci-lint` plugin.

```shell
kal -fix path/to/api/types
```

Other standard Golang linter flags implemented by [multichecker][multichecker] based linters are also supported.

[multichecker]: https://pkg.go.dev/golang.org/x/tools/go/analysis/multichecker

### Golangci-lint Plugin

To install the `golangci-lint` plugin, first you must have `golangci-lint` installed.
If you do not have `golangci-lint` installed, review the `golangci-lint` [install guide][golangci-lint-install].

[golangci-lint-install]: https://golangci-lint.run/welcome/install/

You will need to create a `.custom-gcl.yml` file to describe the custom linters you want to run. The following is an example of a `.custom-gcl.yml` file:

```yaml
version:  v1.62.0
name: golangci-kal
destination: ./bin
plugins:
- module: 'github.com/JoelSpeed/kal'
  version: 'v0.0.0' # Replace with the latest version
```

Once you have created the custom configuration file, you can run the following command to build the custom `golangci-kal` binary:

```shell
golangci-lint custom
```

The output binary will be a combination of the initial `golangci-lint` binary and the KAL linters.
This means that you can use any of the standard `golangci-lint` configuration or flags to run the binary, but may also include the KAL linters.

If you wish to only use the KAL linters, you can configure your `.golangci.yml` file to only run the KAL linters:

```yaml
linters-settings:
  custom:
    kal:
      type: "module"
      description: KAL is the Kube-API-Linter and lints Kube like APIs based on API conventions and best practices.
      settings:
        linters: {}
        lintersConfig: {}
linters:
  disable-all: true
  enable:
    - kal
```

The settings for KAL are based on the [GolangCIConfig][golangci-config-struct] struct and allow for finer control over the linter rules.
The finer control over linter rules is not currently avaialable outside of the plugin based version of KAL.

If you wish to use the KAL linters in conjunction with other linters, you can enable the KAL linters in the `.golangci.yml` file by ensuring that `kal` is in the `linters.enabled` list.
To provide further configuration, add the `custom.kal` section to your `linter-settings` as per the example above.

[golangci-config-struct]: https://pkg.go.dev/github.com/JoelSpeed/kal/pkg/config#GolangCIConfig

#### VSCode integration

Since VSCode already integrates with `golangci-lint` via the [Go][vscode-go] extension, you can use the `golangci-kal` binary as a linter in VSCode.
If your project authors are already using VSCode and have the configuration to lint their code when saving, this can be a seamless integration.

Ensure that your project setup includes building the `golangci-kal` binary, and then configure the `go.lintTool` and `go.alternateTools` settings in your project `.vscode/settings.json` file.

[vscode-go]: https://code.visualstudio.com/docs/languages/go

```json
{
    "go.lintTool": "golangci-lint",
    "go.alternateTools": {
        "golangci-lint": "${workspaceFolder}/bin/golangci-kal",
    }
}
```

Alternatively, you can also replace the binary with a script that runs the `golangci-kal` binary,
allowing for customisation or automatic copmilation of the project should it not already exist.

```json
{
    "go.lintTool": "golangci-lint",
    "go.alternateTools": {
        "golangci-lint": "${workspaceFolder}/hack/golangci-lint.sh",
    }
}
```

# Linters

## CommentStart

The `commentstart` linter checks that all comments in the API types start with the serialized form of the type they are commenting on.
This helps to ensure that generated documentation reflects the most common usage of the field, the serialized YAML form.

### Fixes (via standalone binary only)

The `commentstart` linter can automatically fix comments that do not start with the serialized form of the type.

When the `json` tag is present, and matches the first word of the field comment in all but casing, the linter will suggest that the comment be updated to match the `json` tag.

## JSONTags

The `jsontags` linter checks that all fields in the API types have a `json` tag, and that those tags are correctly formatted.
The `json` tag for a field within a Kubernetes API type should use a camel case version of the field name.

The `jsontags` linter checks the tag name against the regex `"^[a-z][a-z0-9]*(?:[A-Z][a-z0-9]*)*$"` which allows consecutive upper case characters, to allow for acronyms, e.g. `requestTTL`.

### Configuration

```yaml
lintersConfig:
  jsonTags:
    jsonTagRegex: "^[a-z][a-z0-9]*(?:[A-Z][a-z0-9]*)*$" # Provide a custom regex, which the json tag must match.
```

## OptionalOrRequired

The `optionalorrequired` linter checks that all fields in the API types are either optional or required, and are marked explicitly as such.

The linter expects to find a comment marker `// +optional` or `// +required` within the comment for the field.

It also supports the `// +kubebuilder:validation:Optional` and `// +kubebuilder:validation:Required` markers, but will suggest to use the `// +optional` and `// +required` markers instead.

If you prefer to use the Kubebuilder markers instead, you can change the preference in the configuration.

### Configuration

```yaml
lintersConfig:
  optionalOrRequired:
    preferredOptionalMarker: optional | kubebuilder:validation:Optional # The preferred optional marker to use, fixes will suggest to use this marker. Defaults to `optional`.
    preferredRequiredMarker: required | kubebuilder:validation:Required # The preferred required marker to use, fixes will suggest to use this marker. Defaults to `required`.
```

### Fixes (via standalone binary only)

The `optionalorrequired` linter can automatically fix fields that are using the incorrect form of either the optional or required marker.

It will also remove the secondary marker where both the preferred and secondary marker are present on a field.

# Contributing

New linters can be added by following the [New Linter][new-linter] guide.

[new-linter]: docs/new-linter.md

# License

KAL is licensed under the Apache License, Version 2.0. See [LICENSE](LICENSE) for the full license text.
