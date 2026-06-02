package condition

import (
	"testing"
)

func TestEval_EmptyExpression(t *testing.T) {
	if !Eval("", nil) {
		t.Error("empty expression should return true")
	}
	if !Eval("null", nil) {
		t.Error("null expression should return true")
	}
	if !Eval("  ", nil) {
		t.Error("whitespace-only expression should return true")
	}
}

func TestEval_SimpleEquality(t *testing.T) {
	ctx := map[string]string{"status": "task_created"}
	if !Eval("status=task_created", ctx) {
		t.Error("status=task_created should match")
	}
	if Eval("status=no_task", ctx) {
		t.Error("status=no_task should NOT match when status=task_created")
	}
}

func TestEval_MissingKey(t *testing.T) {
	ctx := map[string]string{}
	if Eval("missing=anything", ctx) {
		t.Error("missing key should return false")
	}
}

func TestEval_OR(t *testing.T) {
	ctx := map[string]string{"status": "task_created"}
	if !Eval("status=task_created OR status=redirect", ctx) {
		t.Error("OR should match when first branch matches")
	}
	if !Eval("status=no_task OR status=task_created", ctx) {
		t.Error("OR should match when second branch matches")
	}
	if Eval("status=no_task OR status=redirect", ctx) {
		t.Error("OR should NOT match when neither branch matches")
	}
}

func TestEval_NOT(t *testing.T) {
	if !Eval("!false", map[string]string{"false": ""}) {
		t.Error("!false should be true when key is empty")
	}
	if !Eval("!missing", map[string]string{}) {
		t.Error("!missing should be true when key absent")
	}
	if Eval("!exists", map[string]string{"exists": "yes"}) {
		t.Error("!exists should be false when key has value")
	}
}

func TestEval_SimpleBoolean(t *testing.T) {
	if !Eval("gate.passed", map[string]string{"gate.passed": "true"}) {
		t.Error("gate.passed=true should be truthy")
	}
	if Eval("gate.passed", map[string]string{"gate.passed": ""}) {
		t.Error("gate.passed='' should be falsy")
	}
	if Eval("gate.passed", map[string]string{"gate.passed": "false"}) {
		t.Error("gate.passed=false should be falsy")
	}
	if Eval("missing", map[string]string{}) {
		t.Error("missing key should be falsy")
	}
}

func TestEval_NegatedEquality(t *testing.T) {
	ctx := map[string]string{"status": "task_created"}
	if !Eval("!status=no_task", ctx) {
		t.Error("!status=no_task should be true when status=task_created")
	}
	if Eval("!status=task_created", ctx) {
		t.Error("!status=task_created should be false when status=task_created")
	}
}

func TestEval_RealWorldScenarios(t *testing.T) {
	tests := []struct {
		name     string
		expr     string
		ctx      map[string]string
		expected bool
	}{
		{
			name:     "sta0 with task",
			expr:     "status=task_created OR status=redirect",
			ctx:      map[string]string{"status": "task_created"},
			expected: true,
		},
		{
			name:     "sta0 no task",
			expr:     "status=no_task",
			ctx:      map[string]string{"status": "no_task"},
			expected: true,
		},
		{
			name:     "sta0 no task should not match task_created edge",
			expr:     "status=task_created OR status=redirect",
			ctx:      map[string]string{"status": "no_task"},
			expected: false,
		},
		{
			name:     "gate passed",
			expr:     "gate.passed",
			ctx:      map[string]string{"gate.passed": "true"},
			expected: true,
		},
		{
			name:     "gate not passed",
			expr:     "gate.passed",
			ctx:      map[string]string{"gate.passed": ""},
			expected: false,
		},
		{
			name:     "phase check",
			expr:     "phase=review",
			ctx:      map[string]string{"phase": "review"},
			expected: true,
		},
		{
			name:     "phase mismatch",
			expr:     "phase=review",
			ctx:      map[string]string{"phase": "implement"},
			expected: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := Eval(tc.expr, tc.ctx)
			if got != tc.expected {
				t.Errorf("Eval(%q, %v) = %v, want %v", tc.expr, tc.ctx, got, tc.expected)
			}
		})
	}
}