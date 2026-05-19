import type { Flow } from '../types/flow';

export const sampleDevFlow: Flow = {
  version: 'v3',
  metadata: {
    name: 'dev-flow',
    description: 'Software development main flow: Triage → Gate → Feature/Bugfix/Hotfix/Analysis/Change → Terminal',
    domain: 'development',
    author: 'team-flow',
    tags: ['development', 'main-flow'],
  },
  config: {
    task_type: 'development',
    auto_dispatch: true,
    parallel_limit: 3,
    timeout_minutes: 480,
  },
  variables: {
    docs_path: '',
    TEAM_PATH: '',
    task_id: '',
    SKILL_PATH: '',
    workspace: '',
    domain: 'development',
  },
  nodes: [
    {
      id: 's3k1',
      type: 'start',
      name: 'Start',
    },
    {
      id: 'tri3',
      type: 'phase',
      name: 'Task Triage',
      description: 'Classify incoming task by type and dispatch to the appropriate branch',
      config: { role: 'Triage' },
      components: {
        roles: [{ ref: 'triage', source: 'builtin' }],
        rules: [{ ref: 'dispatch-guard', source: 'builtin' }],
        tools: [
          { ref: 'task', source: 'builtin', commands: ['show', 'update', 'append'] },
          { ref: 'search', source: 'builtin' },
        ],
      },
      docs: [
        {
          name: 'TRIAGE.md',
          format: 'markdown',
          path: '{docs_path}/{task_id}/TRIAGE.md',
          required: true,
          description: '任务分类记录。必须包含：1.任务类型判断(feature/bug/hotfix/analysis/change) 2.优先级评估 3.指派建议 4.初步范围估算',
        },
      ],
      on_enter: [{ action: 'update_task_phase', phase: 'ready' }],
    },
    {
      id: 'ent7',
      type: 'gate',
      name: 'Entry Gate',
      description: 'Validate task has sufficient context and matches a known task type',
      config: {
        conditions: [
          { type: 'task_exists', required: true, check: 'Verify a beads task exists for this request' },
          { type: 'context_sufficient', required: true, check: 'Task description contains enough detail to proceed', threshold: 'minimal' },
          { type: 'type_identified', required: false, check: 'Task type can be determined from context' },
        ],
      },
    },

    {
      id: 'fa01',
      type: 'phase',
      name: 'Requirements Analysis',
      description: 'Analyze feature requirements and produce specification with acceptance criteria',
      config: { role: 'TechLead' },
      components: {
        roles: [{ ref: 'tech-lead', source: 'builtin' }],
        rules: [
          { ref: 'requirements-standards', source: 'builtin' },
          { ref: 'output-guard', source: 'framework' },
        ],
        tools: [
          { ref: 'task', source: 'builtin', commands: ['show', 'update', 'append'] },
          { ref: 'search', source: 'builtin' },
        ],
        skills: [{ ref: 'api-design', source: 'builtin' }],
        constraints: [{ type: 'file_write', allowed_paths: ['{docs_path}/{task_id}/'] }],
      },
      docs: [
        {
          name: 'SPEC.md',
          format: 'markdown',
          path: '{docs_path}/{task_id}/SPEC.md',
          required: true,
          description: '功能规格说明书。必须包含：1.功能概述 2.用户故事 3.接口定义 4.数据模型 5.非功能需求 6.依赖关系',
        },
        {
          name: 'AC.md',
          format: 'markdown',
          path: '{docs_path}/{task_id}/AC.md',
          required: true,
          description: '验收标准。必须包含：1.功能验收条件(Given-When-Then格式) 2.性能验收标准 3.安全验收标准 4.每个条件必须有明确的通过/失败判定依据',
        },
      ],
      on_enter: [{ action: 'update_task_phase', phase: 'analyze' }],
    },
    {
      id: 'fd02',
      type: 'phase',
      name: 'Technical Design',
      description: 'Design data models, state machines, and API contracts based on requirements',
      config: { role: 'TechLead' },
      components: {
        roles: [{ ref: 'tech-lead', source: 'builtin' }],
        rules: [{ ref: 'architecture-standards', source: 'builtin' }],
        skills: [{ ref: 'protobuf', source: 'builtin' }, { ref: 'api-design', source: 'builtin' }],
        constraints: [{ type: 'file_write', allowed_paths: ['{docs_path}/{task_id}/'] }],
      },
      docs: [
        {
          name: 'R1_DATA_MODEL.md',
          format: 'markdown',
          path: '{docs_path}/{task_id}/R1_DATA_MODEL.md',
          required: true,
          description: '数据模型设计。必须包含：1.实体定义(字段/类型/约束) 2.实体关系图 3.索引设计 4.数据迁移方案',
        },
        {
          name: 'R2_STATE_MACHINE.md',
          format: 'markdown',
          path: '{docs_path}/{task_id}/R2_STATE_MACHINE.md',
          required: true,
          description: '状态机设计。必须包含：1.状态定义 2.转换条件 3.转换动作 4.状态图',
        },
        {
          name: 'R3_API_CONTRACT.md',
          format: 'markdown',
          path: '{docs_path}/{task_id}/R3_API_CONTRACT.md',
          required: true,
          description: 'API契约定义。必须包含：1.接口列表(路径/方法/请求/响应) 2.错误码定义 3.版本策略 4.认证方式',
        },
      ],
      on_enter: [{ action: 'update_task_phase', phase: 'design' }],
    },
    {
      id: 'fi03',
      type: 'phase',
      name: 'Implementation',
      description: 'Implement the feature code and unit tests following the technical design',
      config: { role: 'Dev' },
      components: {
        roles: [{ ref: 'dev', source: 'builtin' }],
        rules: [
          { ref: 'development-standards', source: 'builtin' },
          { ref: 'regression-guard', source: 'builtin' },
          { ref: 'loop-guard', source: 'builtin' },
        ],
        tools: [
          { ref: 'task', source: 'builtin' },
          { ref: 'search', source: 'builtin' },
          { ref: 'file-write', source: 'builtin' },
        ],
        constraints: [
          { type: 'file_write', allowed_paths: ['{workspace}/'] },
          { type: 'tool_restriction', forbidden: ['shell-exec'] },
        ],
      },
      docs: [
        {
          name: 'IMPL.md',
          format: 'markdown',
          path: '{docs_path}/{task_id}/IMPL.md',
          required: true,
          description: '实现记录。必须包含：1.实现方案概述 2.关键代码变更列表 3.新增依赖 4.已知限制',
        },
      ],
      on_enter: [{ action: 'update_task_phase', phase: 'implement' }],
    },

    {
      id: 'bi01',
      type: 'phase',
      name: 'Root Cause Investigation',
      description: 'Investigate the bug, identify root cause, and produce RCA document',
      config: { role: 'Dev' },
      components: {
        roles: [{ ref: 'dev', source: 'builtin' }],
        rules: [
          { ref: 'development-standards', source: 'builtin' },
          { ref: 'output-guard', source: 'framework' },
        ],
        tools: [
          { ref: 'task', source: 'builtin', commands: ['show', 'update', 'append'] },
          { ref: 'search', source: 'builtin' },
        ],
      },
      docs: [
        {
          name: 'RCA.md',
          format: 'markdown',
          path: '{docs_path}/{task_id}/RCA.md',
          required: true,
          description: '根因分析。必须包含：1.Bug来源 2.受影响组件 3.复现步骤 4.严重程度评估',
        },
      ],
      on_enter: [{ action: 'update_task_phase', phase: 'analyze' }],
    },
    {
      id: 'bf02',
      type: 'phase',
      name: 'Bug Fix',
      description: 'Implement the bug fix and add regression tests',
      config: { role: 'Dev' },
      components: {
        roles: [{ ref: 'dev', source: 'builtin' }],
        rules: [
          { ref: 'development-standards', source: 'builtin' },
          { ref: 'regression-guard', source: 'builtin' },
        ],
        tools: [
          { ref: 'task', source: 'builtin' },
          { ref: 'search', source: 'builtin' },
          { ref: 'file-write', source: 'builtin' },
        ],
        constraints: [{ type: 'file_write', allowed_paths: ['{workspace}/'] }],
      },
      docs: [
        {
          name: 'FIX.md',
          format: 'markdown',
          path: '{docs_path}/{task_id}/FIX.md',
          required: true,
          description: '修复记录。必须包含：1.修复方案 2.代码变更列表 3.回归测试新增 4.验证步骤',
        },
      ],
      on_enter: [{ action: 'update_task_phase', phase: 'implement' }],
    },

    {
      id: 'hi01',
      type: 'phase',
      name: 'Emergency Fix',
      description: 'Implement the emergency hotfix with minimal scope',
      config: { role: 'Dev' },
      components: {
        roles: [{ ref: 'dev', source: 'builtin' }],
        rules: [
          { ref: 'development-standards', source: 'builtin' },
          { ref: 'regression-guard', source: 'builtin' },
        ],
        tools: [
          { ref: 'task', source: 'builtin' },
          { ref: 'search', source: 'builtin' },
          { ref: 'file-write', source: 'builtin' },
        ],
        constraints: [
          { type: 'file_write', allowed_paths: ['{workspace}/'] },
          { type: 'timeout', minutes: 60 },
        ],
      },
      docs: [
        {
          name: 'RCA.md',
          format: 'markdown',
          path: '{docs_path}/{task_id}/RCA.md',
          required: true,
          description: '紧急根因分析。必须包含：1.事件时间线 2.根因描述 3.修复说明 4.后续行动',
        },
      ],
      on_enter: [{ action: 'update_task_phase', phase: 'implement' }],
      on_error: { strategy: 'fail', fallback: 'notify' },
    },

    {
      id: 'ai01',
      type: 'phase',
      name: 'Deep Analysis',
      description: 'Perform technical analysis and gather findings',
      config: { role: 'Analysis' },
      components: {
        roles: [{ ref: 'analysis', source: 'builtin' }],
        rules: [{ ref: 'output-guard', source: 'framework' }],
        tools: [
          { ref: 'task', source: 'builtin', commands: ['show', 'update', 'append'] },
          { ref: 'search', source: 'builtin' },
        ],
      },
      docs: [
        {
          name: 'FINDINGS.md',
          format: 'markdown',
          path: '{docs_path}/{task_id}/FINDINGS.md',
          required: true,
          description: '分析发现。必须包含：1.调查结果 2.证据链 3.模式识别 4.技术洞察',
        },
      ],
      on_enter: [{ action: 'update_task_phase', phase: 'analyze' }],
    },
    {
      id: 'ar02',
      type: 'phase',
      name: 'Analysis Report',
      description: 'Compile analysis results into a structured report with conclusions',
      config: { role: 'Analysis' },
      components: {
        roles: [{ ref: 'analysis', source: 'builtin' }],
        rules: [{ ref: 'output-guard', source: 'framework' }],
      },
      docs: [
        {
          name: 'REPORT.md',
          format: 'markdown',
          path: '{docs_path}/{task_id}/REPORT.md',
          required: true,
          description: '分析报告。必须包含：1.结论 2.建议 3.行动项 4.风险评估',
        },
      ],
      on_enter: [{ action: 'update_task_phase', phase: 'review' }],
    },

    {
      id: 'cp01',
      type: 'phase',
      name: 'Change Planning',
      description: 'Plan the change scope, impact analysis, and migration strategy',
      config: { role: 'TechLead' },
      components: {
        roles: [{ ref: 'tech-lead', source: 'builtin' }],
        rules: [
          { ref: 'architecture-standards', source: 'builtin' },
          { ref: 'output-guard', source: 'framework' },
        ],
        tools: [
          { ref: 'task', source: 'builtin', commands: ['show', 'update', 'append'] },
          { ref: 'search', source: 'builtin' },
        ],
        constraints: [{ type: 'file_write', allowed_paths: ['{docs_path}/{task_id}/'] }],
      },
      docs: [
        {
          name: 'SCOPE.md',
          format: 'markdown',
          path: '{docs_path}/{task_id}/SCOPE.md',
          required: true,
          description: '变更范围。必须包含：1.修改/新增/删除内容 2.依赖关系图 3.回滚方案',
        },
        {
          name: 'IMPACT.md',
          format: 'markdown',
          path: '{docs_path}/{task_id}/IMPACT.md',
          required: true,
          description: '影响分析。必须包含：1.受影响组件 2.依赖变更 3.迁移步骤 4.回滚方案',
        },
      ],
      on_enter: [{ action: 'update_task_phase', phase: 'analyze' }],
    },
    {
      id: 'ce02',
      type: 'phase',
      name: 'Change Execution',
      description: 'Execute the planned changes and update tests',
      config: { role: 'Dev' },
      components: {
        roles: [{ ref: 'dev', source: 'builtin' }],
        rules: [
          { ref: 'development-standards', source: 'builtin' },
          { ref: 'regression-guard', source: 'builtin' },
        ],
        tools: [
          { ref: 'task', source: 'builtin' },
          { ref: 'search', source: 'builtin' },
          { ref: 'file-write', source: 'builtin' },
        ],
        constraints: [{ type: 'file_write', allowed_paths: ['{workspace}/'] }],
      },
      docs: [
        {
          name: 'EXEC.md',
          format: 'markdown',
          path: '{docs_path}/{task_id}/EXEC.md',
          required: true,
          description: '变更执行记录。必须包含：1.修改内容 2.受影响文件 3.测试更新 4.验证步骤',
        },
      ],
      on_enter: [{ action: 'update_task_phase', phase: 'implement' }],
    },

    {
      id: 'qua9',
      type: 'gate',
      name: 'Quality Gate',
      description: 'Verify code quality: tests pass, lint clean, no regressions',
      config: {
        conditions: [
          { type: 'tests_pass', threshold: '100%', required: true, check: 'All unit and integration tests pass' },
          { type: 'lint_pass', threshold: '0 errors', required: true, check: 'No lint errors in changed files' },
          { type: 'no_regressions', required: false, check: 'No regressions in existing functionality' },
        ],
        auto_retry: { max_attempts: 2, delay_minutes: 5 },
      },
    },
    {
      id: 'ver4',
      type: 'phase',
      name: 'Integration Verification',
      description: 'Run full integration and E2E tests, verify completeness',
      config: { role: 'QA' },
      components: {
        roles: [{ ref: 'qa', source: 'builtin' }],
        rules: [{ ref: 'test-standards', source: 'builtin' }],
      },
      docs: [
        {
          name: 'TEST_COVERAGE.md',
          format: 'markdown',
          path: '{docs_path}/{task_id}/TEST_COVERAGE.md',
          required: true,
          description: '测试覆盖率报告。必须包含：1.单元测试覆盖率(行/分支) 2.集成测试结果 3.E2E测试结果 4.未覆盖场景说明',
        },
      ],
      on_enter: [{ action: 'update_task_phase', phase: 'verify' }],
    },
    {
      id: 'rev6',
      type: 'phase',
      name: 'Code Review',
      description: 'Review code quality, architecture compliance, and produce completion scope',
      config: { role: 'TechLead' },
      components: {
        roles: [{ ref: 'tech-lead', source: 'builtin' }],
        rules: [{ ref: 'review-standards', source: 'builtin' }],
      },
      docs: [
        {
          name: 'SCOPE.md',
          format: 'markdown',
          path: '{docs_path}/{task_id}/SCOPE.md',
          required: true,
          description: '完成范围文档。必须包含：1.已完成功能清单 2.变更文件列表 3.已知问题 4.后续建议',
        },
      ],
      on_enter: [{ action: 'update_task_phase', phase: 'review' }],
    },

    {
      id: 'suc0',
      type: 'terminal',
      name: 'Completed',
      description: 'Task completed successfully',
      config: { status: 'success', message: 'Task completed successfully' },
    },
    {
      id: 'rej8',
      type: 'terminal',
      name: 'Rejected',
      description: 'Task rejected at entry gate',
      config: { status: 'rejected', message: 'Task does not meet minimum requirements' },
    },
    {
      id: 'fai0',
      type: 'terminal',
      name: 'Failed',
      description: 'Task failed and requires manual escalation',
      config: { status: 'failed', message: 'Task failed - manual intervention required' },
    },
  ],
  edges: [
    { id: 'e-s3k1-tri3', from: 's3k1', to: 'tri3' },
    { id: 'e-tri3-ent7', from: 'tri3', to: 'ent7' },

    { id: 'e-ent7-fa01', from: 'ent7', to: 'fa01', type: 'conditional', conditions: [{ expression: 'task_type=feature' }] },
    { id: 'e-ent7-bi01', from: 'ent7', to: 'bi01', type: 'conditional', conditions: [{ expression: 'task_type=bug' }] },
    { id: 'e-ent7-hi01', from: 'ent7', to: 'hi01', type: 'conditional', conditions: [{ expression: 'task_type=hotfix' }] },
    { id: 'e-ent7-ai01', from: 'ent7', to: 'ai01', type: 'conditional', conditions: [{ expression: 'task_type=analysis' }] },
    { id: 'e-ent7-cp01', from: 'ent7', to: 'cp01', type: 'conditional', conditions: [{ expression: 'task_type=change' }] },
    { id: 'e-ent7-rej8', from: 'ent7', to: 'rej8', type: 'conditional', conditions: [{ expression: '!gate.passed' }] },

    { id: 'e-fa01-fd02', from: 'fa01', to: 'fd02' },
    { id: 'e-fd02-fi03', from: 'fd02', to: 'fi03' },
    { id: 'e-fi03-qua9', from: 'fi03', to: 'qua9' },

    { id: 'e-bi01-bf02', from: 'bi01', to: 'bf02' },
    { id: 'e-bf02-qua9', from: 'bf02', to: 'qua9' },

    { id: 'e-hi01-ver4', from: 'hi01', to: 'ver4' },

    { id: 'e-ai01-ar02', from: 'ai01', to: 'ar02' },
    { id: 'e-ar02-suc0', from: 'ar02', to: 'suc0' },

    { id: 'e-cp01-ce02', from: 'cp01', to: 'ce02' },
    { id: 'e-ce02-qua9', from: 'ce02', to: 'qua9' },

    { id: 'e-qua9-ver4', from: 'qua9', to: 'ver4', type: 'conditional', conditions: [{ expression: 'gate.passed' }] },
    { id: 'e-qua9-fi03', from: 'qua9', to: 'fi03', type: 'conditional', conditions: [{ expression: '!gate.passed AND task_type=feature' }] },
    { id: 'e-qua9-bf02', from: 'qua9', to: 'bf02', type: 'conditional', conditions: [{ expression: '!gate.passed AND task_type=bug' }] },
    { id: 'e-qua9-ce02', from: 'qua9', to: 'ce02', type: 'conditional', conditions: [{ expression: '!gate.passed AND task_type=change' }] },

    { id: 'e-ver4-rev6', from: 'ver4', to: 'rev6' },
    { id: 'e-rev6-suc0', from: 'rev6', to: 'suc0' },
  ],
};

export const sampleGameDesignFlow: Flow = {
  version: 'v3',
  metadata: {
    name: 'game-design-flow',
    description: 'Game design flow: Concept → GDD → Prototype → Gate → Art & Levels → Gate → Polish → Terminal',
    author: 'team-flow',
    tags: ['game', 'design', 'creative'],
    domain: 'game-design',
  },
  config: {
    task_type: 'game-design',
    auto_dispatch: true,
    parallel_limit: 3,
    timeout_minutes: 480,
  },
  variables: {
    docs_path: '',
    TEAM_PATH: '',
    task_id: '',
    SKILL_PATH: '',
    workspace: '',
    domain: 'game-design',
  },
  nodes: [
    {
      id: 's2k9',
      type: 'start',
      name: 'Start',
    },
    {
      id: 'gc01',
      type: 'phase',
      name: 'Game Concept',
      description: 'Define the game concept including genre, target audience, core loop, and creative vision',
      config: { role: 'GameDesigner' },
      components: {
        roles: [{ ref: 'game-designer', source: 'builtin' }],
        rules: [
          { ref: 'output-guard', source: 'framework' },
          { ref: 'game-concept-standards', source: 'builtin' },
        ],
      },
      docs: [
        {
          name: 'CONCEPT.md',
          format: 'markdown',
          path: '{docs_path}/{task_id}/CONCEPT.md',
          required: true,
          description: 'Game concept document: genre, target audience, core gameplay loop, monetization model, creative vision, unique selling points',
        },
      ],
      on_enter: [
        { action: 'update_task_phase', phase: 'analyze' },
      ],
    },
    {
      id: 'gc02',
      type: 'phase',
      name: 'Game Design Document',
      description: 'Write the full GDD covering mechanics, systems, progression, economy, and balance',
      config: { role: 'GameDesigner' },
      components: {
        roles: [{ ref: 'game-designer', source: 'builtin' }],
        rules: [
          { ref: 'output-guard', source: 'framework' },
          { ref: 'game-design-standards', source: 'builtin' },
        ],
      },
      docs: [
        {
          name: 'GDD.md',
          format: 'markdown',
          path: '{docs_path}/{task_id}/GDD.md',
          required: true,
          description: 'Game Design Document: core mechanics, progression systems, economy, balance parameters, player journey, UI/UX flow',
        },
      ],
      on_enter: [
        { action: 'update_task_phase', phase: 'design' },
      ],
    },
    {
      id: 'gp03',
      type: 'phase',
      name: 'Prototype',
      description: 'Build a minimal playable prototype to validate core mechanics and fun factor',
      config: { role: 'GameDev' },
      components: {
        roles: [{ ref: 'game-dev', source: 'builtin' }],
        rules: [
          { ref: 'output-guard', source: 'framework' },
          { ref: 'prototype-standards', source: 'builtin' },
        ],
        tools: [
          { ref: 'task', source: 'builtin' },
          { ref: 'search', source: 'builtin' },
          { ref: 'file-write', source: 'builtin' },
        ],
        constraints: [
          { type: 'file_write', allowed_paths: ['{workspace}/'] },
          { type: 'timeout', minutes: 120 },
        ],
      },
      docs: [
        {
          name: 'PROTOTYPE.md',
          format: 'markdown',
          path: '{docs_path}/{task_id}/PROTOTYPE.md',
          required: true,
          description: 'Prototype report: what was built, core loop validation results, fun factor assessment, known issues',
        },
      ],
      on_enter: [
        { action: 'update_task_phase', phase: 'implement' },
      ],
    },
    {
      id: 'gpt4',
      type: 'gate',
      name: 'Prototype Review',
      description: 'Evaluate if the prototype validates the core gameplay concept',
      config: {
        conditions: [
          { type: 'core_loop_playable', required: true, check: 'Core gameplay loop is functional and playable', expected: 'playable' },
          { type: 'fun_factor', required: true, check: 'Core mechanics are engaging and enjoyable', expected: 'positive' },
          { type: 'technical_feasibility', required: false, check: 'Prototype runs without critical errors', expected: 'stable' },
        ],
        auto_retry: { max_attempts: 2, delay_minutes: 10 },
      },
    },
    {
      id: 'ga05',
      type: 'phase',
      name: 'Art Direction',
      description: 'Define visual style, create art guidelines, and produce key concept art',
      config: { role: 'Artist' },
      components: {
        roles: [{ ref: 'artist', source: 'builtin' }],
        rules: [
          { ref: 'output-guard', source: 'framework' },
          { ref: 'art-standards', source: 'builtin' },
        ],
      },
      docs: [
        {
          name: 'ART_DIRECTION.md',
          format: 'markdown',
          path: '{docs_path}/{task_id}/ART_DIRECTION.md',
          required: true,
          description: 'Art direction document: visual style guide, color palette, asset list, UI/UX wireframes, animation notes',
        },
      ],
      on_enter: [
        { action: 'update_task_phase', phase: 'implement' },
      ],
    },
    {
      id: 'gl06',
      type: 'phase',
      name: 'Level Design',
      description: 'Design game levels, environments, puzzles, and difficulty progression',
      config: { role: 'LevelDesigner' },
      components: {
        roles: [{ ref: 'level-designer', source: 'builtin' }],
        rules: [
          { ref: 'output-guard', source: 'framework' },
          { ref: 'level-design-standards', source: 'builtin' },
        ],
      },
      docs: [
        {
          name: 'LEVELS.md',
          format: 'markdown',
          path: '{docs_path}/{task_id}/LEVELS.md',
          required: true,
          description: 'Level design document: layout descriptions, puzzle mechanics, difficulty curve, environment themes, pacing notes',
        },
      ],
      on_enter: [
        { action: 'update_task_phase', phase: 'implement' },
      ],
    },
    {
      id: 'gcs7',
      type: 'gate',
      name: 'Design Consistency Check',
      description: 'Verify all design elements are coherent and consistent',
      config: {
        conditions: [
          { type: 'worldbuilding_consistency', required: true, check: 'World settings are internally consistent across all documents' },
          { type: 'mechanic_balance', required: true, check: 'Game mechanics are balanced and progression feels fair', expected: 'balanced' },
          { type: 'art_mechanic_coherence', required: false, check: 'Art direction aligns with game mechanics and tone' },
        ],
        auto_retry: { max_attempts: 2, delay_minutes: 10 },
      },
    },
    {
      id: 'gp08',
      type: 'phase',
      name: 'Polish & Integration',
      description: 'Integrate all design elements, polish details, and produce final design package',
      config: { role: 'GameDesigner' },
      components: {
        roles: [{ ref: 'game-designer', source: 'builtin' }],
        rules: [
          { ref: 'output-guard', source: 'framework' },
          { ref: 'review-standards', source: 'builtin' },
        ],
      },
      docs: [
        {
          name: 'DESIGN_PACKAGE.md',
          format: 'markdown',
          path: '{docs_path}/{task_id}/DESIGN_PACKAGE.md',
          required: true,
          description: 'Final design package: integrated GDD with all revisions, art guidelines, level specs, implementation notes',
        },
      ],
      on_enter: [
        { action: 'update_task_phase', phase: 'review' },
      ],
    },
    {
      id: 'gs00',
      type: 'terminal',
      name: 'Design Complete',
      description: 'Game design package completed and ready for development',
      config: { status: 'success', message: 'Game design flow completed successfully' },
    },
    {
      id: 'gr09',
      type: 'terminal',
      name: 'Rework Required',
      description: 'Design requires significant rework from concept stage',
      config: { status: 'rejected', message: 'Core concept or prototype did not pass review - rework needed' },
    },
  ],
  edges: [
    { id: 'e-s2k9-gc01', from: 's2k9', to: 'gc01' },
    { id: 'e-gc01-gc02', from: 'gc01', to: 'gc02' },
    { id: 'e-gc02-gp03', from: 'gc02', to: 'gp03' },
    { id: 'e-gp03-gpt4', from: 'gp03', to: 'gpt4' },
    { id: 'e-gpt4-ga05', from: 'gpt4', to: 'ga05', type: 'conditional', conditions: [{ expression: 'gate.passed' }] },
    { id: 'e-gpt4-gc01', from: 'gpt4', to: 'gc01', type: 'conditional', conditions: [{ expression: '!gate.passed' }] },
    { id: 'e-ga05-gl06', from: 'ga05', to: 'gl06' },
    { id: 'e-gl06-gcs7', from: 'gl06', to: 'gcs7' },
    { id: 'e-gcs7-gp08', from: 'gcs7', to: 'gp08', type: 'conditional', conditions: [{ expression: 'gate.passed' }] },
    { id: 'e-gcs7-ga05', from: 'gcs7', to: 'ga05', type: 'conditional', conditions: [{ expression: '!gate.passed' }] },
    { id: 'e-gp08-gs00', from: 'gp08', to: 'gs00' },
  ],
};

export const sampleNovelFlow: Flow = {
  version: 'v3',
  metadata: {
    name: 'novel-flow',
    description: 'Novel writing flow: Concept → Outline → Characters & World → Gate → Chapter Writing → Gate → Editing → Terminal',
    author: 'team-flow',
    tags: ['novel', 'writing', 'creative'],
    domain: 'writing',
  },
  config: {
    task_type: 'novel-writing',
    auto_dispatch: true,
    parallel_limit: 2,
    timeout_minutes: 600,
  },
  variables: {
    docs_path: '',
    TEAM_PATH: '',
    task_id: '',
    SKILL_PATH: '',
    workspace: '',
    domain: 'writing',
  },
  nodes: [
    {
      id: 's1m7',
      type: 'start',
      name: 'Start',
    },
    {
      id: 'nc01',
      type: 'phase',
      name: 'Novel Concept',
      description: 'Define the novel concept including genre, themes, target audience, and narrative vision',
      config: { role: 'Writer' },
      components: {
        roles: [{ ref: 'writer', source: 'builtin' }],
        rules: [
          { ref: 'output-guard', source: 'framework' },
          { ref: 'narrative-standards', source: 'builtin' },
        ],
      },
      docs: [
        {
          name: 'CONCEPT.md',
          format: 'markdown',
          path: '{docs_path}/{task_id}/CONCEPT.md',
          required: true,
          description: 'Novel concept: genre, themes, target audience, narrative vision, tone, comparable works, unique angle',
        },
      ],
      on_enter: [
        { action: 'update_task_phase', phase: 'analyze' },
      ],
    },
    {
      id: 'nc02',
      type: 'phase',
      name: 'Story Outline',
      description: 'Create the detailed story outline with narrative arc, chapter breakdown, and key plot points',
      config: { role: 'Writer' },
      components: {
        roles: [{ ref: 'writer', source: 'builtin' }],
        rules: [
          { ref: 'output-guard', source: 'framework' },
          { ref: 'narrative-standards', source: 'builtin' },
        ],
      },
      docs: [
        {
          name: 'OUTLINE.md',
          format: 'markdown',
          path: '{docs_path}/{task_id}/OUTLINE.md',
          required: true,
          description: 'Story outline: three-act structure, chapter breakdown, key plot points, turning points, climax, resolution',
        },
      ],
      on_enter: [
        { action: 'update_task_phase', phase: 'design' },
      ],
    },
    {
      id: 'nc03',
      type: 'phase',
      name: 'Character & World Building',
      description: 'Develop character profiles, relationships, world settings, and rules',
      config: { role: 'Writer' },
      components: {
        roles: [{ ref: 'writer', source: 'builtin' }],
        rules: [
          { ref: 'output-guard', source: 'framework' },
          { ref: 'character-consistency', source: 'builtin' },
        ],
      },
      docs: [
        {
          name: 'CHARACTERS.md',
          format: 'markdown',
          path: '{docs_path}/{task_id}/CHARACTERS.md',
          required: true,
          description: 'Character profiles: backstory, motivations, relationships, character arcs, voice, quirks',
        },
        {
          name: 'WORLDBUILDING.md',
          format: 'markdown',
          path: '{docs_path}/{task_id}/WORLDBUILDING.md',
          required: true,
          description: 'World settings: geography, culture, magic/technology system, social structure, history, rules',
        },
      ],
      on_enter: [
        { action: 'update_task_phase', phase: 'design' },
      ],
    },
    {
      id: 'ngo4',
      type: 'gate',
      name: 'Outline Review',
      description: 'Evaluate if the story outline, characters, and world are coherent and complete enough to write',
      config: {
        conditions: [
          { type: 'plot_coherence', required: true, check: 'Story arc is logically coherent with clear cause-effect chain', expected: 'coherent' },
          { type: 'character_depth', required: true, check: 'Main characters have clear motivations, arcs, and distinct voices', expected: 'sufficient' },
          { type: 'world_consistency', required: true, check: 'World rules are internally consistent without contradictions', expected: 'consistent' },
          { type: 'outline_completeness', required: false, check: 'Chapter breakdown covers full story arc from opening to resolution' },
        ],
        auto_retry: { max_attempts: 2, delay_minutes: 5 },
      },
    },
    {
      id: 'nw05',
      type: 'phase',
      name: 'Chapter Writing',
      description: 'Write the full manuscript following the outline, character profiles, and world settings',
      config: { role: 'Writer' },
      components: {
        roles: [{ ref: 'writer', source: 'builtin' }],
        rules: [
          { ref: 'output-guard', source: 'framework' },
          { ref: 'narrative-standards', source: 'builtin' },
          { ref: 'character-consistency', source: 'builtin' },
        ],
        tools: [
          { ref: 'task', source: 'builtin' },
          { ref: 'file-write', source: 'builtin' },
        ],
        constraints: [
          { type: 'file_write', allowed_paths: ['{docs_path}/{task_id}/'] },
        ],
      },
      docs: [
        {
          name: 'MANUSCRIPT.md',
          format: 'markdown',
          path: '{docs_path}/{task_id}/MANUSCRIPT.md',
          required: true,
          description: 'Full manuscript: all chapters with narrative prose, dialogue, scene descriptions, following outline structure',
        },
      ],
      on_enter: [
        { action: 'update_task_phase', phase: 'implement' },
      ],
    },
    {
      id: 'ngl6',
      type: 'gate',
      name: 'Logic & Consistency Check',
      description: 'Verify manuscript for logical consistency, character coherence, and plot holes',
      config: {
        conditions: [
          { type: 'logic_consistency', required: true, check: 'No logical contradictions or plot holes in the narrative', expected: 'consistent' },
          { type: 'character_consistency', required: true, check: 'Characters behave consistently with their established profiles and arcs', expected: 'consistent' },
          { type: 'worldbuilding_consistency', required: true, check: 'World rules are followed consistently throughout the manuscript', expected: 'consistent' },
          { type: 'pacing', required: false, check: 'Story pacing feels natural with appropriate tension and release', expected: 'good' },
        ],
        auto_retry: { max_attempts: 2, delay_minutes: 5 },
      },
    },
    {
      id: 'ne07',
      type: 'phase',
      name: 'Editing & Polish',
      description: 'Edit for prose quality, grammar, style consistency, and produce final manuscript',
      config: { role: 'Editor' },
      components: {
        roles: [{ ref: 'editor', source: 'builtin' }],
        rules: [
          { ref: 'output-guard', source: 'framework' },
          { ref: 'editing-standards', source: 'builtin' },
        ],
      },
      docs: [
        {
          name: 'FINAL_MANUSCRIPT.md',
          format: 'markdown',
          path: '{docs_path}/{task_id}/FINAL_MANUSCRIPT.md',
          required: true,
          description: 'Final edited manuscript: polished prose, corrected grammar, consistent style, ready for publication',
        },
      ],
      on_enter: [
        { action: 'update_task_phase', phase: 'review' },
      ],
    },
    {
      id: 'ns00',
      type: 'terminal',
      name: 'Novel Complete',
      description: 'Novel manuscript completed and ready for publication',
      config: { status: 'success', message: 'Novel writing flow completed successfully' },
    },
    {
      id: 'nr08',
      type: 'terminal',
      name: 'Major Revision Needed',
      description: 'Manuscript requires fundamental rework of story structure',
      config: { status: 'rejected', message: 'Story structure or world building needs fundamental revision' },
    },
  ],
  edges: [
    { id: 'e-s1m7-nc01', from: 's1m7', to: 'nc01' },
    { id: 'e-nc01-nc02', from: 'nc01', to: 'nc02' },
    { id: 'e-nc02-nc03', from: 'nc02', to: 'nc03' },
    { id: 'e-nc03-ngo4', from: 'nc03', to: 'ngo4' },
    { id: 'e-ngo4-nw05', from: 'ngo4', to: 'nw05', type: 'conditional', conditions: [{ expression: 'gate.passed' }] },
    { id: 'e-ngo4-nc02', from: 'ngo4', to: 'nc02', type: 'conditional', conditions: [{ expression: '!gate.passed' }] },
    { id: 'e-nw05-ngl6', from: 'nw05', to: 'ngl6' },
    { id: 'e-ngl6-ne07', from: 'ngl6', to: 'ne07', type: 'conditional', conditions: [{ expression: 'gate.passed' }] },
    { id: 'e-ngl6-nw05', from: 'ngl6', to: 'nw05', type: 'conditional', conditions: [{ expression: '!gate.passed' }] },
    { id: 'e-ne07-ns00', from: 'ne07', to: 'ns00' },
  ],
};
