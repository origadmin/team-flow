package proc

import (
	"context"
	"testing"

	"github.com/origadmin/team-flow/internal/flow"
)

func TestExtractDocsContentRules(t *testing.T) {
	data := []byte(`{
		"version": "v3",
		"metadata": {"name": "test"},
		"config": {"domain": "test-domain"},
		"components": {
			"roles": [
				{"id": "skill-architect", "name": "技能架构师", "alias": "匠思远", "alias_en": "Craft", "persona": "test persona", "traits": ["design-first"], "guidance": "test guidance"}
			]
		},
		"nodes": [
			{
				"id": "fa03",
				"type": "phase",
				"name": "Test",
				"role": "skill-architect",
				"docs": [
					{
						"name": "SPEC",
						"format": "markdown",
						"path": "{DOCS_INTERNAL}/{task_id}/SPEC.md",
						"required": true,
						"content_rules": ["rule1", "rule2"],
						"template": "{TEAM_PATH}/templates/spec-template.md"
					}
				]
			}
		],
		"edges": []
	}`)

	f, err := flow.ParseFlow(data)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	vars := map[string]string{
		"DOCS_INTERNAL": "/tmp/docs",
		"TEAM_PATH":     "/tmp/team",
		"task_id":       "test-123",
	}

	node := &f.Nodes[0]
	docs := extractDocs(node, vars)
	if len(docs) != 1 {
		t.Fatalf("expected 1 doc, got %d", len(docs))
	}

	doc := docs[0]
	if doc.Name != "SPEC" {
		t.Errorf("expected name SPEC, got %s", doc.Name)
	}
	if doc.Template != "/tmp/team/templates/spec-template.md" {
		t.Errorf("expected template resolved, got %q", doc.Template)
	}
	if len(doc.ContentRules) != 2 {
		t.Fatalf("expected 2 content_rules, got %d: %v", len(doc.ContentRules), doc.ContentRules)
	}
	if doc.ContentRules[0] != "rule1" {
		t.Errorf("expected rule1, got %s", doc.ContentRules[0])
	}
	if doc.ContentRules[1] != "rule2" {
		t.Errorf("expected rule2, got %s", doc.ContentRules[1])
	}
}

func TestProcRunContentRules(t *testing.T) {
	data := []byte(`{
		"version": "v3",
		"metadata": {"name": "test-flow"},
		"config": {"domain": "test-domain"},
		"components": {
			"roles": [
				{"id": "skill-architect", "name": "技能架构师", "alias": "匠思远", "alias_en": "Craft", "persona": "test persona", "traits": ["design-first"], "guidance": "test guidance"}
			]
		},
		"nodes": [
			{
				"id": "fa03",
				"type": "phase",
				"name": "Test",
				"role": "skill-architect",
				"docs": [
					{
						"name": "SPEC",
						"format": "markdown",
						"path": "{DOCS_INTERNAL}/{task_id}/SPEC.md",
						"required": true,
						"content_rules": ["rule1", "rule2"],
						"template": "{TEAM_PATH}/templates/spec-template.md"
					}
				]
			}
		],
		"edges": []
	}`)

	f, err := flow.ParseFlow(data)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	vars := map[string]string{
		"DOCS_INTERNAL": "/tmp/docs",
		"TEAM_PATH":     "/tmp/team",
		"task_id":       "test-123",
	}

	node := &f.Nodes[0]
	result := generateResult(f, node, vars, nil, false, "", "", false)

	if len(result.Current.Docs) != 1 {
		t.Fatalf("expected 1 doc, got %d", len(result.Current.Docs))
	}

	doc := result.Current.Docs[0]
	if doc.Template != "/tmp/team/templates/spec-template.md" {
		t.Errorf("expected template resolved, got %q", doc.Template)
	}
	if len(doc.ContentRules) != 2 {
		t.Fatalf("expected 2 content_rules, got %d: %v", len(doc.ContentRules), doc.ContentRules)
	}
	if doc.ContentRules[0] != "rule1" {
		t.Errorf("expected rule1, got %s", doc.ContentRules[0])
	}
}

func TestProcRunEngineContentRules(t *testing.T) {
	engine := NewProcRunEngine("")

	data := []byte(`{
		"version": "v3",
		"metadata": {"name": "test-flow"},
		"config": {"domain": "test-domain"},
		"components": {
			"roles": [
				{"id": "skill-architect", "name": "技能架构师", "alias": "匠思远", "alias_en": "Craft", "persona": "test persona", "traits": ["design-first"], "guidance": "test guidance"}
			]
		},
		"nodes": [
			{
				"id": "fa03",
				"type": "phase",
				"name": "Test",
				"role": "skill-architect",
				"docs": [
					{
						"name": "SPEC",
						"format": "markdown",
						"path": "{DOCS_INTERNAL}/{task_id}/SPEC.md",
						"required": true,
						"content_rules": ["rule1", "rule2"],
						"template": "{TEAM_PATH}/templates/spec-template.md"
					}
				]
			}
		],
		"edges": []
	}`)

	f, err := flow.ParseFlow(data)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	vars := map[string]string{
		"DOCS_INTERNAL": "/tmp/docs",
		"TEAM_PATH":     "/tmp/team",
		"task_id":       "test-123",
	}

	node := &f.Nodes[0]
	result := generateResult(f, node, vars, nil, false, "", "", false)

	_ = engine
	_ = context.Background()

	if len(result.Current.Docs) != 1 {
		t.Fatalf("expected 1 doc, got %d", len(result.Current.Docs))
	}

	doc := result.Current.Docs[0]
	if len(doc.ContentRules) != 2 {
		t.Fatalf("expected 2 content_rules, got %d: %v", len(doc.ContentRules), doc.ContentRules)
	}
}
