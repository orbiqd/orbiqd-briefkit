# Coding Instructions

## Libraries
1. Use `github.com/iancoleman/strcase` library for string case conversions.
2. Wrap `slog` attributes with helpers like `slog.String` and `slog.Int` to keep types explicit.

## Code style
1. Place package-level error variables at the end of the file.
2. Write log messages as full sentences starting with a capital letter.
3. Use the `.yaml` extension for all YAML files, not `.yml`.
4. Separate each struct field, interface method, const or var block with a blank line when has comment.
5. Add comments to public interfaces, theirs methods, errors or functions.
6. Format error messages as noun phrases describing the failed operation, not as action descriptions.

## Build and executables
1. Use `make build` to build executables.
2. Use executables from `./bin/` directory.

## Unit tests
1. Use the `github.com/stretchr/testify` library, particularly the `assert` and `require` packages, for writing assertions.
2. Always run `make test` after changes to execute unit test.
