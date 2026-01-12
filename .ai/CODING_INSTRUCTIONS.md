# Coding Instructions

## Libraries
1. Use `github.com/iancoleman/strcase` for string case conversions.
2. Use `github.com/reugn/go-quartz` for time-related things.
3. Use `github.com/alecthomas/kong` for handling CLI arguments, commands and options. 
4. Use `github.com/spf13/afero` for file system operations.
5. Wrap `slog` attributes with helpers like `slog.String` and `slog.Int` to keep types explicit.

## Code style
1. Place package-level error variables at the end of the file.
2. Write log messages as full sentences starting with a capital letter.
3. Use the `.yaml` extension for all YAML files, not `.yml`.
4. Separate each struct field, interface method, const or var block with a blank line when has comment.
5. Add comments to public interfaces, theirs methods, errors or functions.
6. Format error messages as noun phrases describing the failed operation, not as action descriptions.
7. Run `make lint` after change to verify code quality.
8. Run `make build` after change to verify compilation.
9. Skip writing unit tests until explicitly requested by the user.

## Build and executables
1. Use `make build` to build executables.
2. Use executables from `./bin/` directory.
