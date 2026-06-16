# Traceability Matrix Template (v3-Asset)

Use this matrix during the `Quality Gate` node to prove that every design requirement has been implemented and verified.

## Requirement to Code Alignment

| Req ID | Req Description | Implementation Path | Verification Proof | Logic Sync Status |
|--------|-----------------|---------------------|--------------------|-------------------|
| SPEC-01| <Brief description>| `path/to/file:line` | `test_name` or `output` | ✅ Fully Sync / ⚠️ Partial |
| R1-Data| <Schema/Entity> | `internal/data/xxx` | `TestXXXSave` | ✅ Fully Sync |

## Document Dependency Chain

| Target Doc | Depends-on (Parent) | Relation Type |
|------------|---------------------|---------------|
| SPEC.md    | ARCHITECTURE.md     | Inherits Scope|
| R1_DATA.md | SPEC.md             | Data Mapping  |

## Compliance Assertion
- [ ] No "Shadow Logic" found (code exists but not in SPEC).
- [ ] No "Orphan Specs" found (SPEC exists but not in code).
- [ ] All `Depends-on` anchors are recursively verified.
