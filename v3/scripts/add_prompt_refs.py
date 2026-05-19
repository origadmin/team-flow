#!/usr/bin/env python3
"""Add prompt_source, standards_source to role definitions and prompts refs to node components across all v3 flow JSONs."""
import json
import sys
from pathlib import Path

PROMPT_MAP = {
    "triage": {
        "prompt_source": "{TEAM_PATH}/prompts/triage.md",
        "standards_source": "{TEAM_PATH}/workflows/roles/triage-standards.md",
    },
    "tech-lead": {
        "prompt_source": "{TEAM_PATH}/prompts/tech-lead.md",
        "standards_source": "{TEAM_PATH}/workflows/roles/architecture-standards.md",
    },
    "dev": {
        "prompt_source": "{TEAM_PATH}/prompts/dev.md",
        "standards_source": "{TEAM_PATH}/workflows/roles/development-standards.md",
    },
    "dev-backend": {
        "prompt_source": "{TEAM_PATH}/prompts/dev-backend.md",
        "standards_source": "{TEAM_PATH}/workflows/roles/development-standards.md",
    },
    "dev-frontend": {
        "prompt_source": "{TEAM_PATH}/prompts/dev-frontend.md",
        "standards_source": "{TEAM_PATH}/workflows/roles/development-standards.md",
    },
    "qa": {
        "prompt_source": "{TEAM_PATH}/prompts/qa-engineer.md",
        "standards_source": "{TEAM_PATH}/workflows/roles/test-standards.md",
    },
    "bugfix": {
        "prompt_source": "{TEAM_PATH}/prompts/bugfix.md",
        "standards_source": "{TEAM_PATH}/workflows/roles/bugfix-standards.md",
    },
    "framework-architect": {
        "prompt_source": "{TEAM_PATH}/prompts/framework-architect.md",
        "standards_source": "{TEAM_PATH}/workflows/roles/architecture-standards.md",
    },
    "devops": {
        "prompt_source": "{TEAM_PATH}/prompts/devops.md",
        "standards_source": "{TEAM_PATH}/workflows/roles/devops-standards.md",
    },
    "pm": {
        "prompt_source": "{TEAM_PATH}/prompts/pm.md",
        "standards_source": "{TEAM_PATH}/workflows/roles/requirements-standards.md",
    },
    "analysis": {
        "prompt_source": "{TEAM_PATH}/prompts/analysis.md",
        "standards_source": "{TEAM_PATH}/workflows/roles/analysis-standards.md",
    },
    "ui-designer": {
        "prompt_source": "{TEAM_PATH}/prompts/ui-designer.md",
        "standards_source": "{TEAM_PATH}/workflows/roles/ui-standards.md",
    },
}

PROMPT_PATH_MAP = {
    "triage": "{TEAM_PATH}/prompts/triage.md",
    "tech-lead": "{TEAM_PATH}/prompts/tech-lead.md",
    "dev": "{TEAM_PATH}/prompts/dev.md",
    "dev-backend": "{TEAM_PATH}/prompts/dev-backend.md",
    "dev-frontend": "{TEAM_PATH}/prompts/dev-frontend.md",
    "qa": "{TEAM_PATH}/prompts/qa-engineer.md",
    "bugfix": "{TEAM_PATH}/prompts/bugfix.md",
    "framework-architect": "{TEAM_PATH}/prompts/framework-architect.md",
    "devops": "{TEAM_PATH}/prompts/devops.md",
    "pm": "{TEAM_PATH}/prompts/pm.md",
    "analysis": "{TEAM_PATH}/prompts/analysis.md",
    "ui-designer": "{TEAM_PATH}/prompts/ui-designer.md",
}

FLOW_DIR = Path(r"D:\workspace\project\golang\origadmin\framework\projects\team-flow\v3\flows")

def update_role_definitions(flow):
    """Add prompt_source and standards_source to role definitions in components."""
    if not flow.get("components") or not flow["components"].get("roles"):
        return 0
    count = 0
    for role in flow["components"]["roles"]:
        rid = role.get("id", "")
        if rid in PROMPT_MAP and "prompt_source" not in role:
            role.update(PROMPT_MAP[rid])
            count += 1
    return count

def add_prompts_to_nodes(flow):
    """Add prompts refs to node components based on the role ref."""
    if not flow.get("nodes"):
        return 0
    count = 0
    for node in flow["nodes"]:
        comps = node.get("components")
        if not comps or not comps.get("roles"):
            continue
        # Skip if already has prompts
        if comps.get("prompts"):
            continue
        # Only add prompts to phase/start/branch nodes
        if node.get("type") not in ("phase", "start", "branch"):
            continue
        prompts = []
        for role_ref in comps["roles"]:
            rid = role_ref.get("ref", "")
            if rid in PROMPT_PATH_MAP:
                prompts.append({
                    "ref": rid,
                    "source": "file",
                    "path": PROMPT_PATH_MAP[rid]
                })
        if prompts:
            comps["prompts"] = prompts
            count += 1
    return count

def main():
    changed = []
    for fp in sorted(FLOW_DIR.glob("*.json")):
        if fp.name == "test-simple-flow.json":
            continue
        with open(fp, "r", encoding="utf-8") as f:
            flow = json.load(f)
        rc = update_role_definitions(flow)
        nc = add_prompts_to_nodes(flow)
        if rc > 0 or nc > 0:
            with open(fp, "w", encoding="utf-8") as f:
                json.dump(flow, f, indent=2, ensure_ascii=False)
            changed.append(f"{fp.name}: {rc} roles updated, {nc} nodes got prompts")
    for c in changed:
        print(c)
    if not changed:
        print("No changes needed")

if __name__ == "__main__":
    main()
