# Agent GIT instructions

Agent MUST follow this instructions when working with local GIT or GitHub.

GitHub repository related to this project is: https://github.com/orbiqd/orbiqd-briefkit

## General
1. Use English as primary language.

## GitHub
1. Use `gh` cli for interaction with GitHub
2. Use `git` cli for local git operations
3. Follow .github/ISSUE_TEMPLATE when creating issues behalf of user
4. Follow .github/PULL_REQUEST_TEMPLATE when creating pull requests behalf of user
5. Verify available labels before posting issue or pull request and suggest labels matching issue or pull request.

## Commit Messages
1. Format commit messages using Conventional Commits: `type(scope): description`
2. Use these commit types: feat, fix, refactor, docs, test, chore, perf, build, ci
3. Keep commit subject line under 72 characters
4. Write only the subject line without body or additional paragraphs
5. Describe WHY the change was made, not WHAT was changed (the diff shows what)
6. Use `chore(ai):` for changes to AI agent instructions, configuration, or documentation (e.g., AGENTS.md, CLAUDE.md, GEMINI.md, .ai/ directory)

## Issues
1. Always show issue to the user before creating it.
2. Always create or update GitHub issue bodies using `gh issue create/edit --body-file` with a temporary file or here-doc to preserve newlines and avoid shell interpolation.
2. Ask user about issue details.
3. Ask user for confirmation before creating issue.

## Pre-Commit Checks
1. Run `git status` before committing to review all changes
2. Scan all changes for sensitive data: API keys, passwords, tokens, credentials, local paths
3. Search for any "TODO" comments in code and STOP if found.

## Branch Workflow
1. Verify current branch with `git status` before making changes
2. Create feature branches using format: `feat/feature-name` or `fix/bug-name`
3. Branch from the default branch unless instructed otherwise
