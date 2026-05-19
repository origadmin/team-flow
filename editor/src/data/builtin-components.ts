import type { ComponentSource } from '../types/flow';

interface BuiltinComponent {
  ref: string;
  source: ComponentSource;
  label: string;
  domain?: string;
}

interface BuiltinTool extends BuiltinComponent {
  commands?: string[];
}

export const BUILTIN_ROLES: BuiltinComponent[] = [
  { ref: 'triage', source: 'builtin', label: 'Triage', domain: 'development' },
  { ref: 'tech-lead', source: 'builtin', label: 'TechLead', domain: 'development' },
  { ref: 'dev', source: 'builtin', label: 'Dev', domain: 'development' },
  { ref: 'qa', source: 'builtin', label: 'QA', domain: 'development' },
  { ref: 'analysis', source: 'builtin', label: 'Analysis', domain: 'development' },
  { ref: 'game-designer', source: 'builtin', label: 'GameDesigner', domain: 'game-design' },
  { ref: 'game-dev', source: 'builtin', label: 'GameDev', domain: 'game-design' },
  { ref: 'level-designer', source: 'builtin', label: 'LevelDesigner', domain: 'game-design' },
  { ref: 'artist', source: 'builtin', label: 'Artist', domain: 'game-design' },
  { ref: 'writer', source: 'builtin', label: 'Writer', domain: 'writing' },
  { ref: 'editor', source: 'builtin', label: 'Editor', domain: 'writing' },
  { ref: 'pm', source: 'builtin', label: 'PM', domain: 'development' },
  { ref: 'designer', source: 'builtin', label: 'Designer', domain: 'development' },
];

export const BUILTIN_RULES: BuiltinComponent[] = [
  { ref: 'dispatch-guard', source: 'builtin', label: 'Dispatch Guard', domain: 'development' },
  { ref: 'requirements-standards', source: 'builtin', label: 'Requirements Standards', domain: 'development' },
  { ref: 'output-guard', source: 'framework', label: 'Output Guard' },
  { ref: 'architecture-standards', source: 'builtin', label: 'Architecture Standards', domain: 'development' },
  { ref: 'development-standards', source: 'builtin', label: 'Development Standards', domain: 'development' },
  { ref: 'regression-guard', source: 'builtin', label: 'Regression Guard', domain: 'development' },
  { ref: 'loop-guard', source: 'builtin', label: 'Loop Guard', domain: 'development' },
  { ref: 'test-standards', source: 'builtin', label: 'Test Standards', domain: 'development' },
  { ref: 'review-standards', source: 'builtin', label: 'Review Standards' },
  { ref: 'game-concept-standards', source: 'builtin', label: 'Game Concept Standards', domain: 'game-design' },
  { ref: 'game-design-standards', source: 'builtin', label: 'Game Design Standards', domain: 'game-design' },
  { ref: 'prototype-standards', source: 'builtin', label: 'Prototype Standards', domain: 'game-design' },
  { ref: 'art-standards', source: 'builtin', label: 'Art Standards', domain: 'game-design' },
  { ref: 'level-design-standards', source: 'builtin', label: 'Level Design Standards', domain: 'game-design' },
  { ref: 'narrative-standards', source: 'builtin', label: 'Narrative Standards', domain: 'writing' },
  { ref: 'character-consistency', source: 'builtin', label: 'Character Consistency', domain: 'writing' },
  { ref: 'editing-standards', source: 'builtin', label: 'Editing Standards', domain: 'writing' },
];

export const BUILTIN_TOOLS: BuiltinTool[] = [
  { ref: 'task', source: 'builtin', label: 'Task', commands: ['show', 'update', 'append'] },
  { ref: 'search', source: 'builtin', label: 'Search' },
  { ref: 'file-write', source: 'builtin', label: 'File Write' },
  { ref: 'file-ops', source: 'builtin', label: 'File Ops' },
];

export const BUILTIN_SKILLS: BuiltinComponent[] = [
  { ref: 'api-design', source: 'builtin', label: 'API Design', domain: 'development' },
  { ref: 'protobuf', source: 'builtin', label: 'Protobuf', domain: 'development' },
];

export const DOMAIN_OPTIONS = [
  'development',
  'game-design',
  'writing',
  'operations',
  'analytics',
  'design',
  'management',
];

export function getRolesByDomain(domain: string): BuiltinComponent[] {
  return BUILTIN_ROLES.filter((r) => !r.domain || r.domain === domain);
}

export function getRulesByDomain(domain: string): BuiltinComponent[] {
  return BUILTIN_RULES.filter((r) => !r.domain || r.domain === domain);
}

export function getSkillsByDomain(domain: string): BuiltinComponent[] {
  return BUILTIN_SKILLS.filter((s) => !s.domain || s.domain === domain);
}
