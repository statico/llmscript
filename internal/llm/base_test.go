package llm

import (
	"context"
	"strings"
	"testing"
)

// fakeGenerator records the prompts it receives and returns canned responses.
type fakeGenerator struct {
	responses []string
	calls     int
	prompts   []string
}

func (f *fakeGenerator) name() string { return "fake" }

func (f *fakeGenerator) generate(_ context.Context, prompt string) (string, error) {
	f.prompts = append(f.prompts, prompt)
	resp := f.responses[f.calls%len(f.responses)]
	f.calls++
	return resp, nil
}

func TestScriptProvider_GenerateScripts(t *testing.T) {
	gen := &fakeGenerator{responses: []string{
		"<script>\n#!/usr/bin/env bash\necho main\n</script>",
		"<script>\n#!/usr/bin/env bash\necho test\n</script>",
	}}
	p := &scriptProvider{gen: gen, extraPrompt: "Use ANSI colors"}

	pair, err := p.GenerateScripts(context.Background(), "print main")
	if err != nil {
		t.Fatalf("GenerateScripts: %v", err)
	}

	if !strings.Contains(pair.MainScript, "echo main") {
		t.Errorf("main script not extracted from <script> tags: %q", pair.MainScript)
	}
	if strings.Contains(pair.MainScript, "<script>") {
		t.Errorf("main script still contains tags: %q", pair.MainScript)
	}
	if !strings.Contains(pair.TestScript, "echo test") {
		t.Errorf("test script not extracted: %q", pair.TestScript)
	}

	if gen.calls != 2 {
		t.Fatalf("expected 2 generate calls, got %d", gen.calls)
	}

	// The first prompt should carry the description and inject the additional
	// instructions before the output-format section.
	mainPrompt := gen.prompts[0]
	if !strings.Contains(mainPrompt, "print main") {
		t.Errorf("main prompt missing description")
	}
	if !strings.Contains(mainPrompt, "<additional_instructions>") ||
		!strings.Contains(mainPrompt, "Use ANSI colors") {
		t.Errorf("main prompt missing additional instructions block")
	}
	if idx, out := strings.Index(mainPrompt, "<additional_instructions>"),
		strings.Index(mainPrompt, "<output_format>"); idx == -1 || out == -1 || idx > out {
		t.Errorf("additional instructions should appear before output_format")
	}
}

func TestScriptProvider_NoExtraPrompt(t *testing.T) {
	gen := &fakeGenerator{responses: []string{"echo hi"}}
	p := &scriptProvider{gen: gen}

	if _, err := p.GenerateScripts(context.Background(), "desc"); err != nil {
		t.Fatalf("GenerateScripts: %v", err)
	}
	if strings.Contains(gen.prompts[0], "<additional_instructions>") {
		t.Errorf("empty extra prompt should not inject an instructions block")
	}
}

func TestScriptProvider_FixScripts(t *testing.T) {
	gen := &fakeGenerator{responses: []string{"<script>fixed</script>"}}
	p := &scriptProvider{gen: gen}

	in := ScriptPair{MainScript: "broken", TestScript: "the-test"}
	out, err := p.FixScripts(context.Background(), in, "it failed")
	if err != nil {
		t.Fatalf("FixScripts: %v", err)
	}
	if out.MainScript != "fixed" {
		t.Errorf("expected fixed main script, got %q", out.MainScript)
	}
	if out.TestScript != "the-test" {
		t.Errorf("FixScripts should preserve the test script, got %q", out.TestScript)
	}
	if !strings.Contains(gen.prompts[0], "it failed") {
		t.Errorf("fix prompt should include the failure text")
	}
}

func TestNewProvider(t *testing.T) {
	tests := []struct {
		name     string
		cfg      Config
		wantErr  bool
		wantName string
	}{
		{name: "default is ollama", cfg: Config{}, wantName: "Ollama"},
		{name: "ollama explicit", cfg: Config{Provider: "ollama"}, wantName: "Ollama"},
		{name: "claude needs key", cfg: Config{Provider: "claude"}, wantErr: true},
		{name: "claude ok", cfg: Config{Provider: "claude", Claude: ClaudeConfig{APIKey: "k"}}, wantName: "Claude"},
		{name: "openai needs key", cfg: Config{Provider: "openai"}, wantErr: true},
		{name: "openai ok", cfg: Config{Provider: "openai", OpenAI: OpenAIConfig{APIKey: "k"}}, wantName: "OpenAI"},
		{name: "openrouter needs key", cfg: Config{Provider: "openrouter"}, wantErr: true},
		{name: "openrouter ok", cfg: Config{Provider: "openrouter", OpenRouter: OpenRouterConfig{APIKey: "k"}}, wantName: "OpenRouter"},
		{name: "gemini needs key", cfg: Config{Provider: "gemini"}, wantErr: true},
		{name: "unknown provider", cfg: Config{Provider: "bogus"}, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, err := NewProvider(tt.cfg)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if p.Name() != tt.wantName {
				t.Errorf("expected name %q, got %q", tt.wantName, p.Name())
			}
		})
	}
}
