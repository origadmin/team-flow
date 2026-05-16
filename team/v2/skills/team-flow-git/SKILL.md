---
name: team-flow-git
description: |
  Standardized Git operations with safety checks for AI coding workflows. Create branches with
  issue-backed naming, commit with conventional messages and selective staging, push with base
  branch protection, and create or update pull requests. Use this skill when creating Git
  branches, committing changes, pushing code, or creating PRs. Also use when the user mentions
  "branch", "commit", "push", "PR", "pull request", "git", or "merge request".
---

# team-flow-git — Git Operations

## Status Line (MANDATORY)

Every AI response MUST start with: `[Role: {role} | TaskPool: {beads-id}#{cr-index} | Phase: {phase} | Asset: {PROJECT-basename}]`

## Shared State

All team-flow skills share beads as the single source of truth for task state.

**Essential commands**:

| Operation | Command |
|-----------|---------|
| Show task | `flow task show {id}` |
| Record progress | `flow task update {id} --notes "GIT: branched to feat/xxx, committed, pushed"` |

## git-branch: Create Branch

**Branch naming rules**:
- Issue-backed branches: `<type>/<short-desc>-<issueID>`
- Branches without IssueID: `<type>/<user-provided-name>`
- `type` matches Conventional Commits: `feat`, `fix`, `refactor`, `docs`, `test`, `perf`, `chore`
- `short-desc`: lowercase, brief, hyphen-separated, under 80 characters
- No Chinese, uppercase, or multi-task branch names

**Safety checks before `git switch -c`**:
1. Check for uncommitted changes: `git status --short` — if dirty, ask whether to commit, stash, or continue
2. Check whether target branch already exists: `git branch --list <name>` and `git branch --remotes --list "*/<name>"`
3. Check current branch: `git branch --show-current` — prefer creating from repo's base branch (main/master/develop)
4. If remote available: `git fetch --prune` before checking base freshness
5. Compare local base with upstream — if behind, ask whether to update before branching
6. Validate branch name: `git check-ref-format --branch <name>`

## git-commit: Create Commit

**Core workflow**:

1. **Discover repo conventions** — Check `.github/*`, `.gitmessage`, `CONTRIBUTING.md`, commitlint config, recent commit history
2. **Inspect** — Run `git status --short`, `git diff --stat`, `git diff`. If anything staged, also `git diff --cached`
3. **Group** — Split changes by logical purpose. Split when you see:
   - refactor + behavior change
   - formatting + logic change
   - dependency updates + product code
   - unrelated docs + code
4. **Propose and approve** — Show what changed, recommended commit boundary, candidate message(s), risks. Get user approval before staging
5. **Stage selectively** — Prefer `git add <specific-files>`. Do not default to bulk staging
6. **Validate** — If repository has a native `git` pre-commit hook, it MUST run during commit. Do not skip with `--no-verify` unless user explicitly asks
7. **Commit**

**Commit message format** (unless repo conventions say otherwise):
```
type(scope): summary
```

**Issue linking** — Detect issue ID from user mention, branch name, or commit context:
- Use `Fixes #123` only when there is explicit closing intent
- Use `Refs #123` when the commit is related, partial, or ambiguous
- If no issue ID found, do not invent one

**Guardrails**:
- One logical change per commit
- Keep fix + directly related tests together
- Do not auto-push, auto-force-push, rewrite history, or use `--no-verify` unless explicitly asked
- Do not commit secrets, credentials, conflict markers, local artifacts, or unrelated changes
- Report the final commit hash after a successful commit

## git-push: Push Branch

**Workflow**:

1. **Inspect** — `git status --short`, `git branch --show-current`, `git remote -v`. If worktree dirty, report but continue if user wants to push existing commits
2. **Protect base branches** — Do not push directly to `main`, `master`, `develop`, or release branches unless user explicitly requests
3. **Check upstream** — If upstream exists: `git log --oneline @{u}..HEAD`. If no upstream: prepare to set upstream
4. **Push** — Use normal `git push` so any configured `pre-push` hook runs. If no upstream: `git push -u origin <branch>`
5. **Handle rejection** — Do not force push by default. Fetch, inspect divergence, ask before rebasing, merging, or `git push --force-with-lease`. Never use plain `git push --force` unless user explicitly requests

## create-pr: Create or Update PR

**Workflow**:

1. **Inspect and choose base** — `git status --short`, `git branch --show-current`, `git status -sb`. Prefer repo's default PR base
2. **Review and sync locally** — `git diff --stat <base>...HEAD`, `git diff <base>...HEAD`. Review for accidental files, secrets, conflict markers. Fetch base branch and merge/rebase into current branch
3. **Build title, body, and issue link** — Default title from issue or main commit subject. Keep body concise with summary, validation performed, and issue link
4. **Create or update PR** — If PR already exists for current branch, update instead of creating duplicate. Preserve user-authored content in existing PR body

## Security Rules

- Never commit secrets, credentials, or API keys
- Do not auto-push, auto-force-push, or rewrite history unless explicitly asked
- Protect base branches from direct pushes

## HARD CONSTRAINTS

- **⛔ NEVER commit unless user asks** — Explicit confirmation required
- **⛔ NEVER use PowerShell Set-Content / Out-File** — these add UTF-8 BOM, corrupting source files

## Related Skills

| Skill | When to switch |
|-------|---------------|
| `team-flow` | Task creation, routing, status queries |
| `team-flow-build` | Implementation from specs, bugfix, testing |
| `team-flow-review` | PR review with structured output |
