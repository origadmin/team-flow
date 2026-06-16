# Bugfix Completion Gate (v3-Asset)

## Output Guard (Pre-check)
- [ ] Output Guard Summary has been output (all ✅).
- [ ] AC Compliance Matrix has no ❌ items.
- [ ] Build/Type check actually executed and output shown.
- [ ] Tests actually executed and pass counts shown.
- [ ] Self-Critique executed (including data flow + real scenario verification).

## Test Verification (Must show actual output)
- [ ] Backend: `go build ./...` passes.
- [ ] Backend: `go test ./...` passes.
- [ ] Frontend: `bun run typecheck` passes.
- [ ] Frontend: `bun run lint` passes.
- [ ] Frontend: `bun run test` passes.
- [ ] Bug reproduction test passes.
- [ ] Full regression test passes.

## Documentation & Process
- [ ] RCA.md exists (Phenomenon, Root Cause, Impact, Prevention).
- [ ] TEST_CASE.md exists (Reproduction, Expected, Actual).
- [ ] No Chinese comments in code.
- [ ] SCOPE.md generated.
- [ ] Data flow trace included in RCA.md (for API/Auth/State bugs).
- [ ] Real scenario verification included in TEST_CASE.md.
- [ ] Runtime verification (UI/CLI) executed and documented in UI_VERIFICATION.md.
- [ ] Screenshots saved to correct path with standard naming.
- [ ] beads task status updated to `closed`.
- [ ] `{DOCS_INTERNAL}/PROJECT.md` synchronized if affected modules changed.
- [ ] User confirmation received before archiving.
