#!/usr/bin/env python
"""Enrich all v3 flow JSONs with prompt_directives, content_rules, and skill metadata."""

import json
import os
import sys

FLOWS_DIR = os.path.join(os.path.dirname(os.path.dirname(os.path.abspath(__file__))), "flows")

DIRECTIVES_MAP = {
    "triage": [
        "Dispatch tasks to sub-agents, never execute directly",
        "Classify all user input before action: Feature/Bug/Change/Analysis/Batch",
        "Read .team/project.md for project constraints before dispatch",
        "Use flow task create/show for all task state management",
        "Report deliverables and handoff to next role on completion",
    ],
    "tech-lead": [
        "Design before implement: SPEC.md + AC.md + Data Model mandatory",
        "Ensure R0 Navigation Matrix covers all entry points",
        "Review all API contracts for consistency",
        "Verify design satisfies all acceptance criteria",
        "Output Guard self-check before marking phase complete",
    ],
    "dev": [
        "Read SCOPE.md before starting implementation",
        "Surgical changes only: never modify files outside task scope",
        "Run tests before marking phase complete",
        "No Chinese comments or commit messages",
        "Never delete files unless explicitly asked",
        "Output Guard 5-step self-check mandatory before completion gate",
    ],
    "qa": [
        "Verify implementation matches SPEC.md and AC.md",
        "Run full regression tests, not just new tests",
        "Check front-back interface consistency (paths, params, response)",
        "For bugfixes: data flow trace + real scenario verification required",
        "Output Guard self-check before marking review complete",
    ],
    "devops": [
        "Verify CI pipeline passes before any deployment",
        "Always have rollback plan before deployment",
        "Check Docker image build success",
        "Run smoke tests after deployment",
        "Update CHANGELOG and version number",
    ],
    "pm": [
        "Verify 100% acceptance criteria satisfaction",
        "Confirm business flow end-to-end",
        "Sign off only after QA approval",
    ],
}

DOCS_RULES = {
    "SPEC.md": ["Must contain: overview, requirements, constraints, dependencies"],
    "AC.md": [
        "Must contain: acceptance criteria as testable assertions",
        "Each criterion must be verifiable (pass/fail)",
    ],
    "R0_NAVIGATION_MATRIX.md": [
        "Must define all entry points and navigation paths",
        "Every listed entry must have corresponding code implementation",
    ],
    "SCOPE.md": [
        "Must contain: changed files, added files, deleted files, risk assessment"
    ],
    "RCA.md": [
        "Must contain: phenomenon, root_cause, impact, prevention",
        "Data flow trace required for API/permission/state bugs",
        "Real scenario verification required (not mock-only)",
    ],
    "TEST_CASE.md": [
        "Must contain: reproduction steps, expected result, actual result",
        "Regression test cases required for bugfixes",
    ],
}

SKILL_ENRICHMENTS = {
    # Domain-specific skills
    "api-design": {
        "path": "{TEAM_PATH}/skills/api-design/SKILL.md",
        "description": "Design RESTful/gRPC API contracts with consistent naming and error handling",
        "trigger": "phase:design",
    },
    "protobuf": {
        "path": "{TEAM_PATH}/skills/protobuf/SKILL.md",
        "description": "Protocol Buffer schema design and code generation",
        "trigger": "phase:design",
    },
    # Task-type skills
    "team-flow-design": {
        "path": "{TEAM_PATH}/skills/team-flow-design/SKILL.md",
        "description": "TechLead: spec-driven design, architecture, API contracts",
        "trigger": "phase:design",
    },
    "team-flow-build": {
        "path": "{TEAM_PATH}/skills/team-flow-build/SKILL.md",
        "description": "Dev: implementation, bugfix, testing",
        "trigger": "phase:implement",
    },
    "team-flow-git": {
        "path": "{TEAM_PATH}/skills/team-flow-git/SKILL.md",
        "description": "Dev: git operations (branch, commit, push, PR)",
        "trigger": "manual",
    },
    "team-flow-check-impl": {
        "path": "{TEAM_PATH}/skills/team-flow-check-impl/SKILL.md",
        "description": "Dev/QA: verify implementation matches specs",
        "trigger": "phase:verify",
    },
    "team-flow-review": {
        "path": "{TEAM_PATH}/skills/team-flow-review/SKILL.md",
        "description": "QA: PR review, spec review, QA verification",
        "trigger": "phase:review",
    },
    "bugfix-skill": {
        "path": "{TEAM_PATH}/skills/bugfix-skill/SKILL.md",
        "description": "Dev: systematic bugfix with R-iteration",
        "trigger": "phase:fix",
    },
    "analysis-skill": {
        "path": "{TEAM_PATH}/skills/analysis-skill/SKILL.md",
        "description": "Analyst: codebase analysis and reporting",
        "trigger": "phase:analyze",
    },
    "hotfix-skill": {
        "path": "{TEAM_PATH}/skills/hotfix-skill/SKILL.md",
        "description": "Dev: emergency hotfix with minimal scope",
        "trigger": "phase:fix",
    },
    "deploy-skill": {
        "path": "{TEAM_PATH}/skills/deploy-skill/SKILL.md",
        "description": "DevOps: deployment and rollback operations",
        "trigger": "phase:deploy",
    },
    "release-skill": {
        "path": "{TEAM_PATH}/skills/release-skill/SKILL.md",
        "description": "DevOps/PM: release integration, verification, acceptance",
        "trigger": "phase:deploy",
    },
}


def enrich_flow(filepath):
    with open(filepath, "r", encoding="utf-8") as f:
        flow = json.load(f)

    changed = False

    # 1. Add prompt_directives to role definitions
    for role in flow.get("components", {}).get("roles", []):
        role_id = role.get("id", "")
        if role_id in DIRECTIVES_MAP and "prompt_directives" not in role:
            role["prompt_directives"] = DIRECTIVES_MAP[role_id]
            changed = True

    # 2. Add content_rules to doc specs
    for node in flow.get("nodes", []):
        for doc in node.get("docs", []):
            name = doc.get("name", "")
            if name in DOCS_RULES and "content_rules" not in doc:
                doc["content_rules"] = DOCS_RULES[name]
                changed = True

    # 3. Add path/description/trigger to skill definitions
    for skill in flow.get("components", {}).get("skills", []):
        sid = skill.get("id", "")
        if sid in SKILL_ENRICHMENTS:
            for k, v in SKILL_ENRICHMENTS[sid].items():
                if k not in skill:
                    skill[k] = v
                    changed = True

    if changed:
        with open(filepath, "w", encoding="utf-8") as f:
            json.dump(flow, f, indent=2, ensure_ascii=False)
        return True
    return False


def main():
    updated = []
    skipped = []
    for fname in sorted(os.listdir(FLOWS_DIR)):
        if not fname.endswith(".json"):
            continue
        fpath = os.path.join(FLOWS_DIR, fname)
        if enrich_flow(fpath):
            updated.append(fname)
        else:
            skipped.append(fname)

    print(f"Updated: {len(updated)} files: {updated}")
    print(f"Skipped (no changes): {len(skipped)} files: {skipped}")


if __name__ == "__main__":
    main()
