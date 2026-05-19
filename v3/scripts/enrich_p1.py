#!/usr/bin/env python3
"""Enrich v3 flow JSONs with QG rules, checklist-based gate conditions, and template paths."""

import json
import os
import sys
from pathlib import Path

FLOWS_DIR = Path(__file__).parent.parent / "flows"
TEMPLATES_DIR = Path(__file__).parent.parent.parent / "team" / "v3" / "templates"

# QG-1~4 quality gates as rules
QG_RULES = [
    {
        "id": "qg1-think-before-coding",
        "name": "QG-1: Think Before Coding",
        "description": "Before any action, verify assumptions: what assumptions am I acting on? What's the worst case if wrong? What should I clarify first?",
        "enforcement": "hard",
        "trigger": "pre_action"
    },
    {
        "id": "qg2-simplicity-first",
        "name": "QG-2: Simplicity First",
        "description": "Can this be solved with less code? Am I adding unrequested features? Is this over-engineered?",
        "enforcement": "hard",
        "trigger": "pre_output"
    },
    {
        "id": "qg3-surgical-changes",
        "name": "QG-3: Surgical Changes",
        "description": "Only modify code within task scope. Never optimize or refactor adjacent code. Clean up orphan code from own changes.",
        "enforcement": "hard",
        "trigger": "code_change"
    },
    {
        "id": "qg4-goal-driven",
        "name": "QG-4: Goal-Driven Execution",
        "description": "Define success criteria before executing. Each step must have concrete verification criteria. Never use vague 'completed' descriptions.",
        "enforcement": "hard",
        "trigger": "task_start"
    }
]

# Template path mapping for DocSpec
TEMPLATE_MAP = {
    "SPEC.md": "{TEAM_PATH}/templates/prd-template.md",
    "AC.md": "{TEAM_PATH}/templates/prd-template.md",
    "R0_NAVIGATION_MATRIX.md": "{TEAM_PATH}/templates/r0-navigation-matrix-template.md",
    "R1_DATA_MODEL.md": "{TEAM_PATH}/templates/architecture-template.md",
    "R2_STATE_MACHINE.md": "{TEAM_PATH}/templates/architecture-template.md",
    "R3_API_CONTRACT.md": "{TEAM_PATH}/templates/api-issue-template.md",
    "RCA.md": "{TEAM_PATH}/templates/bug-test-template.md",
    "TEST_CASE.md": "{TEAM_PATH}/templates/bug-test-template.md",
    "SCOPE.md": "{TEAM_PATH}/templates/scope-template.md",
    "TEST_COVERAGE.md": "{TEAM_PATH}/templates/feature-test-template.md",
    "UI_VERIFICATION.md": "{TEAM_PATH}/templates/ui-verification-template.md",
    "TRIAGE.md": None,
    "PLAN.md": None,
    "REPORT.md": None,
    "IMPL.md": None,
    "HANDOFF.md": "{TEAM_PATH}/templates/closed-loop-verification-template.md",
    "INDEX.md": "{TEAM_PATH}/templates/bug-index-template.md",
    "README.md": None,
}

# Flow-specific gate condition enrichment
GATE_ENRICHMENTS = {
    "feature-flow": {
        "gate-completion": [
            {"type": "custom", "required": True, "check": "Output Guard Summary all passed"},
            {"type": "custom", "required": True, "check": "AC Compliance Matrix no failures"},
            {"type": "custom", "required": True, "check": "R0_NAVIGATION_MATRIX.md exists and all entries implemented"},
            {"type": "custom", "required": True, "check": "SPEC.md and AC.md exist and non-empty"},
            {"type": "custom", "required": True, "check": "R1_DATA_MODEL.md exists and non-empty"},
            {"type": "custom", "required": True, "check": "R3_API_CONTRACT.md exists and non-empty"},
            {"type": "custom", "required": True, "check": "Frontend-backend API paths 100% match"},
            {"type": "custom", "required": True, "check": "Frontend-backend param names 100% match"},
            {"type": "custom", "required": True, "check": "Frontend-backend response types 100% match"},
            {"type": "custom", "required": True, "check": "All Proto APIs have handler registrations"},
            {"type": "custom", "required": True, "check": "No Chinese comments in code"},
            {"type": "custom", "required": True, "check": "SCOPE.md generated"},
            {"type": "custom", "required": True, "check": "TEST_COVERAGE.md exists"},
            {"type": "custom", "required": True, "check": "User confirmation received"},
        ]
    },
    "bugfix-flow": {
        "gate-quality": [
            {"type": "custom", "required": True, "check": "RCA.md contains: phenomenon, root_cause, impact, prevention"},
            {"type": "custom", "required": True, "check": "TEST_CASE.md contains: reproduce steps, expected result, verification result"},
            {"type": "custom", "required": True, "check": "Bug reproduce test passed (show test output)"},
            {"type": "custom", "required": True, "check": "No Chinese comments in code"},
            {"type": "custom", "required": True, "check": "Regression tests added for the fix"},
            {"type": "custom", "required": True, "check": "Data flow tracing included in RCA (for API/auth/state/interaction bugs)"},
            {"type": "custom", "required": True, "check": "Real scenario verification passed (not just mock tests)"},
            {"type": "custom", "required": True, "check": "Full chain verified (not just breakpoint fix)"},
        ],
        "gate-completion": [
            {"type": "custom", "required": True, "check": "Output Guard Summary all passed"},
            {"type": "custom", "required": True, "check": "AC Compliance Matrix no failures"},
            {"type": "custom", "required": True, "check": "SCOPE.md generated"},
            {"type": "custom", "required": True, "check": "API changes: frontend-backend interface/path/param/response match checked"},
            {"type": "custom", "required": True, "check": "User confirmation received"},
        ]
    },
    "change-flow": {
        "gate-completion": [
            {"type": "custom", "required": True, "check": "Output Guard Summary all passed"},
            {"type": "custom", "required": True, "check": "SCOPE.md generated"},
            {"type": "custom", "required": True, "check": "User confirmation received"},
        ]
    },
    "analysis-flow": {
        "gate-completion": [
            {"type": "custom", "required": True, "check": "Output Guard Summary all passed"},
            {"type": "custom", "required": True, "check": "Analysis document exists and non-empty"},
            {"type": "custom", "required": True, "check": "User confirmation received"},
        ]
    },
    "hotfix-flow": {
        "gate-completion": [
            {"type": "custom", "required": True, "check": "Output Guard Summary all passed"},
            {"type": "custom", "required": True, "check": "RCA.md exists"},
            {"type": "custom", "required": True, "check": "SCOPE.md generated"},
            {"type": "custom", "required": True, "check": "User confirmation received"},
        ]
    },
    "dev-flow": {
        "gate-completion": [
            {"type": "custom", "required": True, "check": "Output Guard Summary all passed"},
            {"type": "custom", "required": True, "check": "SCOPE.md generated"},
            {"type": "custom", "required": True, "check": "User confirmation received"},
        ]
    },
    "release-flow": {
        "gate-acceptance": [
            {"type": "custom", "required": True, "check": "All acceptance criteria 100% met"},
            {"type": "custom", "required": True, "check": "Business loop confirmed"},
            {"type": "custom", "required": True, "check": "PM sign-off received"},
        ],
        "gate-deploy": [
            {"type": "custom", "required": True, "check": "CI pipeline all jobs passed"},
            {"type": "custom", "required": True, "check": "Docker image built successfully"},
            {"type": "custom", "required": True, "check": "Smoke tests passed"},
            {"type": "custom", "required": True, "check": "Monitoring metrics normal"},
            {"type": "custom", "required": True, "check": "CHANGELOG updated"},
            {"type": "custom", "required": True, "check": "Version number updated (SemVer)"},
        ]
    },
    "batch-flow": {
        "gate-completion": [
            {"type": "custom", "required": True, "check": "All sub-tasks completed or handled"},
            {"type": "custom", "required": True, "check": "PLAN.md final status updated"},
            {"type": "custom", "required": True, "check": "User confirmation received"},
        ]
    },
}


def enrich_flow(filepath):
    """Add QG rules, gate conditions, and template paths to a flow JSON."""
    with open(filepath, 'r', encoding='utf-8') as f:
        flow = json.load(f)

    flow_name = flow.get("metadata", {}).get("name", filepath.stem)
    modified = False

    # 1. Add QG rules to components.rules (avoid duplicates)
    existing_rule_ids = set()
    rules = flow.get("components", {}).get("rules", [])
    for r in rules:
        if isinstance(r, dict) and "id" in r:
            existing_rule_ids.add(r["id"])

    for qg in QG_RULES:
        if qg["id"] not in existing_rule_ids:
            rules.append(qg)
            existing_rule_ids.add(qg["id"])
            modified = True

    if rules:
        flow.setdefault("components", {})["rules"] = rules

    # 2. Enrich gate conditions (in config.conditions)
    enrichments = GATE_ENRICHMENTS.get(flow_name, {})
    for node in flow.get("nodes", []):
        node_id = node.get("id", "")
        if node_id in enrichments:
            config = node.get("config", {})
            existing_conditions = config.get("conditions", [])
            existing_checks = set()
            for c in existing_conditions:
                if isinstance(c, dict) and "check" in c:
                    existing_checks.add(c["check"])

            for new_cond in enrichments[node_id]:
                if new_cond.get("check") not in existing_checks:
                    existing_conditions.append(new_cond)
                    existing_checks.add(new_cond.get("check", ""))
                    modified = True

            if modified:
                config["conditions"] = existing_conditions
                node["config"] = config

    # 3. Fill DocSpec template paths
    for node in flow.get("nodes", []):
        docs = node.get("docs", [])
        for doc in docs:
            if isinstance(doc, dict) and "name" in doc and not doc.get("template"):
                template = TEMPLATE_MAP.get(doc["name"])
                if template:
                    doc["template"] = template
                    modified = True

    if modified:
        with open(filepath, 'w', encoding='utf-8') as f:
            json.dump(flow, f, indent=2, ensure_ascii=False)
        print(f"  UPDATED: {filepath.name}")
    else:
        print(f"  SKIP: {filepath.name} (no changes)")

    return modified


def main():
    updated = 0
    for fp in sorted(FLOWS_DIR.glob("*.json")):
        if fp.name == "test-simple-flow.json":
            continue
        if enrich_flow(fp):
            updated += 1
    print(f"\nTotal: {updated} flow(s) updated")


if __name__ == "__main__":
    main()
