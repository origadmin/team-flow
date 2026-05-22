package flow

import (
	"testing"
)

func TestDocSpecContentRules(t *testing.T) {
	data := []byte(`{
		"version": "v3",
		"metadata": {"name": "test"},
		"nodes": [
			{
				"id": "fa03",
				"type": "phase",
				"name": "Test",
				"docs": [
					{
						"name": "SPEC",
						"format": "markdown",
						"path": "/tmp/SPEC.md",
						"required": true,
						"content_rules": ["rule1", "rule2"],
						"template": "/tmp/template.md"
					}
				]
			}
		],
		"edges": []
	}`)

	f, err := ParseFlow(data)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	if len(f.Nodes) != 1 {
		t.Fatalf("expected 1 node, got %d", len(f.Nodes))
	}

	node := f.Nodes[0]
	if len(node.Docs) != 1 {
		t.Fatalf("expected 1 doc, got %d", len(node.Docs))
	}

	doc := node.Docs[0]
	if doc.Name != "SPEC" {
		t.Errorf("expected name SPEC, got %s", doc.Name)
	}
	if doc.Template != "/tmp/template.md" {
		t.Errorf("expected template /tmp/template.md, got %q", doc.Template)
	}
	if len(doc.ContentRules) != 2 {
		t.Fatalf("expected 2 content_rules, got %d", len(doc.ContentRules))
	}
	if doc.ContentRules[0] != "rule1" {
		t.Errorf("expected rule1, got %s", doc.ContentRules[0])
	}
	if doc.ContentRules[1] != "rule2" {
		t.Errorf("expected rule2, got %s", doc.ContentRules[1])
	}
}
