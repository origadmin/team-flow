# Session Startup Protocol (v3-Asset)

## Approach A: Focus Anchor
- Output `[{alias} | Session Start({node_id}:{flow}) | - | start]` from memory.
- Use as the "we are here" anchor for every response.
- No command run yet.

## Approach B: First Session / Context Lost
1. `flow project detect` → Detect workspace and project.
2. `flow proc run --new` → Create new session (ONLY on first call).
3. Adopt principal role from output (Alias, Persona, Traits).

## Context Rescue
If context is lost, run `flow proc run` (no args, no --new) to get latest session state.
Never use `--new` if a session is already active in the same conversation.
