# Quality Gates (Mandatory v3-Asset)

## QG-1: Think Before Coding
**Timing**: After receiving task, before any action.
- [ ] Assumptions: What unconfirmed assumptions am I acting on?
- [ ] Risks: Worst-case scenario if assumption is wrong?
- [ ] Clarification: Points to ask before proceeding?
> Stop if unsure.

## QG-2: Simplicity First
**Timing**: Self-check before every output.
- [ ] Can it be solved with less code?
- [ ] Am I adding unrequested features?
- [ ] Is there over-engineering?

## QG-3: Surgical Changes
**Timing**: During code modification.
- [ ] Only change what is necessary for the task.
- [ ] Do not "optimize" nearby unrelated code.
- [ ] Clean up your own orphan code.

## QG-4: Goal-Driven Execution
**Timing**: Before starting task.
- [ ] Define `goal_plan` with specific `verify` criteria for each step.
- [ ] Success criteria must be measurable.
- [ ] Rollback plan defined.
