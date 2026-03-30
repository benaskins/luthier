package review

import (
	"context"
	"testing"

	talk "github.com/benaskins/axon-talk"

	"github.com/benaskins/maestro/internal/plan"
)

// mockClient implements loop.LLMClient (talk.LLMClient) for testing.
type mockClient struct {
	response string
	err      error
}

func (m *mockClient) Chat(_ context.Context, _ *talk.Request, fn func(talk.Response) error) error {
	if m.err != nil {
		return m.err
	}
	return fn(talk.Response{Content: m.response, Done: true})
}

func TestReview_PassesGoodDiff(t *testing.T) {
	step := plan.Step{
		Number:      1,
		Title:       "add greeting function",
		Description: "Create a Hello function that returns a greeting string.",
	}

	diff := `diff --git a/hello.go b/hello.go
new file mode 100644
--- /dev/null
+++ b/hello.go
@@ -0,0 +1,7 @@
+package main
+
+// Hello returns a greeting string.
+func Hello(name string) string {
+	return "Hello, " + name + "!"
+}`

	client := &mockClient{
		response: `{"passed": true, "reason": "The diff adds a Hello function that returns a greeting string, satisfying the step requirements."}`,
	}

	r := New(Config{Client: client})
	result, err := r.Review(context.Background(), diff, step)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Passed {
		t.Errorf("expected Passed=true, got false; reason: %s", result.Reason)
	}
	if result.Reason == "" {
		t.Error("expected non-empty reason")
	}
}

func TestReview_FailsBadDiff(t *testing.T) {
	step := plan.Step{
		Number:      2,
		Title:       "implement user authentication",
		Description: "Add JWT-based authentication middleware that validates tokens on protected routes.",
	}

	diff := `diff --git a/config.go b/config.go
--- a/config.go
+++ b/config.go
@@ -1,3 +1,4 @@
 package main

+const appName = "myapp"`

	client := &mockClient{
		response: `{"passed": false, "reason": "The diff only adds a constant and does not implement JWT authentication middleware as required."}`,
	}

	r := New(Config{Client: client})
	result, err := r.Review(context.Background(), diff, step)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Passed {
		t.Errorf("expected Passed=false, got true; reason: %s", result.Reason)
	}
	if result.Reason == "" {
		t.Error("expected non-empty reason")
	}
}

func TestReview_FallbackKeywordDetection_Pass(t *testing.T) {
	// When the LLM returns non-JSON, parseResult falls back to keyword detection.
	result, err := parseResult("The implementation looks correct and should pass review.")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Passed {
		t.Errorf("expected Passed=true from keyword fallback")
	}
}

func TestReview_FallbackKeywordDetection_Fail(t *testing.T) {
	result, err := parseResult("This diff fails to implement the required changes.")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Passed {
		t.Errorf("expected Passed=false from keyword fallback")
	}
}

func TestReview_LLMError(t *testing.T) {
	step := plan.Step{Number: 1, Title: "test step"}
	client := &mockClient{err: errLLMUnavailable}

	r := New(Config{Client: client})
	_, err := r.Review(context.Background(), "diff content", step)
	if err == nil {
		t.Error("expected error when LLM client fails")
	}
}

var errLLMUnavailable = &mockError{"LLM unavailable"}

type mockError struct{ msg string }

func (e *mockError) Error() string { return e.msg }

func TestNew_DefaultModel(t *testing.T) {
	r := New(Config{Client: &mockClient{}})
	if r.model != defaultModel {
		t.Errorf("expected default model %q, got %q", defaultModel, r.model)
	}
}

func TestNew_CustomModel(t *testing.T) {
	r := New(Config{Client: &mockClient{}, Model: "mistral"})
	if r.model != "mistral" {
		t.Errorf("expected model %q, got %q", "mistral", r.model)
	}
}

func TestParseResult_JSONWithExtraText(t *testing.T) {
	content := `Here is my assessment:

{"passed": true, "reason": "All required changes are present."}

Let me know if you need more details.`

	result, err := parseResult(content)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Passed {
		t.Errorf("expected Passed=true")
	}
	if result.Reason != "All required changes are present." {
		t.Errorf("unexpected reason: %q", result.Reason)
	}
}
