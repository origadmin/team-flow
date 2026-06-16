# Team Prompts Index

Each role has a dedicated prompt template. These are optional extension content —
the canonical persona/guidance lives in `assets/orgs/<org>/team.json`.

| Role | File | Alias |
|------|------|-------|
| concierge | concierge.md | 闻先迎 |
| triage | triage.md | 齐活林 |
| tech-lead | tech-lead.md | 高见远 |
| dev | dev.md | 寇豆码 |
| qa | qa.md | 严过关 |
| devops | devops.md | 稳如磐 |
| analysis | analysis.md | 溯源 |
| pm | pm.md | 签定音 |

## Loading Order

1. **Primary**: `team.json` → role.persona + role.guidance (inlined)
2. **Override** (future): `prompts/{role-id}.md` → if present, supplements persona
