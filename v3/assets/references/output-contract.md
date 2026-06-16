# AI Output Contract (Mandatory v3-Asset)

Every response to a `flow proc run` MUST contain these structured sections.

## Analysis
Detailed analysis record. Include everything examined and your reasoning.

- **Files Examined**: <list every file read and purpose>
- **Key Findings**: <what was discovered in each file>
- **Reasoning**: <why decisions were made>
- **Root Cause**: <one-sentence root cause>
- **Evidence**: <proof: file:line or output>
- **Solution**: <concrete fix steps>
- **Trade-offs**: <risks/alternatives>

## Conclusion
- **Decision**: <proceed | block | require-info>
- **Next Action**: <node to advance to and why>
- **Blockers**: <what must be resolved first>

## Traceability Matrix (Quality Gate ONLY)
| Req ID | Code Location | Logic Sync Proof |
|--------|---------------|------------------|
| SPEC-01| handler.go:45 | Implements the retry logic defined in AC. |
| R1-Data| model.go      | Matches schema in R1_DATA_MODEL.md. |

## Analysis/Conclusion Files
1. Run: `flow proc round-path` → Get directory.
2. Write analysis to `{round_dir}/analysis.md`.
3. Write conclusion to `{round_dir}/conclusion.md`.
4. Run: `flow proc run {node-id} --analysis-file=... --conclusion-file=...`

**Forbid**: Empty analysis, one-line summaries, skipping sections.
