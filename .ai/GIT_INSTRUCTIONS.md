# Agent GIT instructions

Agent MUST follow this instructions when working with local GIT or GitHub.

GitHub repository related to this project is: https://github.com/orbiqd/orbiqd-briefkit

## Commit Messages
1. Format commit messages using Conventional Commits: `type(scope): description`
2. Use these commit types: feat, fix, refactor, docs, test, chore, perf, build, ci
3. Keep commit subject line under 72 characters
4. Use English as primary language.
5. Write only the subject line without body or additional paragraphs
6. Describe WHY the change was made, not WHAT was changed (the diff shows what)
7. Use `chore(ai):` for changes to AI agent instructions, configuration, or documentation (e.g., AGENTS.md, CLAUDE.md, GEMINI.md, .ai/ directory)

## Pre-Commit Checks
1. Run `git status` before committing to review all changes
2. Scan all changes for sensitive data: API keys, passwords, tokens, credentials, local paths
3. Search for any "TODO" comments in code and STOP if found.

## Branch Workflow
1. Verify current branch with `git status` before making changes
2. Create feature branches using format: `feat/feature-name` or `fix/bug-name`
3. Branch from the default branch unless instructed otherwise

## Pull Requests
1. Check for PR templates in `.github/PULL_REQUEST_TEMPLATE` and follow them when creating pull requests
