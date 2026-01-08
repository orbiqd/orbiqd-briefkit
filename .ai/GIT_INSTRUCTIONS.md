# Agent GIT instructions

Agent MUST follow this instructions when working with local GIT or GitHub.

GitHub repository related to this project is: https://github.com/orbiqd/orbiqd-briefkit

## Commit Messages
1. Format commit messages using Conventional Commits: `type(scope): description`
2. Use these commit types: feat, fix, refactor, docs, test, chore, perf, build, ci
3. Keep commit subject line under 72 characters

## Pre-Commit Checks
4. Run `git status` before committing to review all changes
5. Scan all changes for sensitive data: API keys, passwords, tokens, credentials, local paths

## Branch Workflow
6. Verify current branch with `git status` before making changes
7. Create feature branches using format: `feat/feature-name` or `fix/bug-name`
8. Branch from the default branch unless instructed otherwise

## Pull Requests
9. Check for PR templates in `.github/PULL_REQUEST_TEMPLATE` and follow them when creating pull requests
