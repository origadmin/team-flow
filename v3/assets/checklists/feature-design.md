# Feature Design Checklist (v3-Asset)

This is a mandatory quality gate for the `design` phase. All items must be checked before proceeding to `implement`.

## R0: Entry & Navigation Matrix
- [ ] At least 3 entry points defined for core features.
- [ ] User journey map includes "First-time Discovery".
- [ ] Navigation item, action button, and contextual link are mapped.

## SPEC & AC
- [ ] Business background explains the "Why" (Pain point).
- [ ] Acceptance Criteria follows Given/When/Then format.
- [ ] Non-functional requirements (Performance, Security) are defined.

## Technical Design (R1-R3)
- [ ] Data model covers all new entities and relationships.
- [ ] State machine handles all edge cases (Error, Timeout, Cancel).
- [ ] API contract includes request/response examples and error codes.
- [ ] No N+1 query risks identified in the data flow.

## Compliance
- [ ] Module boundaries follow the 5-layer architecture.
- [ ] Sensitive data is handled according to the encryption policy.
- [ ] No circular dependencies introduced between packages.
