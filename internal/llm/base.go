package llm

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/statico/llmscript/internal/log"
)

// perRequestTimeout bounds a single LLM completion request. It is generous
// enough for the most capable models yet ensures non-interactive runs (CI,
// cron) can't hang forever on an unresponsive endpoint, since the overall
// pipeline is otherwise only cancellable via Ctrl-C.
const perRequestTimeout = 5 * time.Minute

// ScriptPair represents a feature script and its test script
type ScriptPair struct {
	MainScript string // The feature script that implements the functionality
	TestScript string // The test script that verifies the feature script
}

// generator produces raw text completions from a prompt. Each backend
// (Claude, OpenAI, OpenRouter, Gemini, Ollama) implements this minimal
// interface; the shared script-generation flow lives in scriptProvider.
type generator interface {
	generate(ctx context.Context, prompt string) (string, error)
	name() string
}

// scriptProvider implements Provider for any generator using a single shared
// generate/test/fix flow, so each backend only has to know how to turn a
// prompt into text.
type scriptProvider struct {
	gen         generator
	extraPrompt string
}

// Name returns a human-readable name for the underlying backend.
func (p *scriptProvider) Name() string {
	return p.gen.name()
}

// generate runs a single completion, bounding it with perRequestTimeout so no
// backend can hang indefinitely.
func (p *scriptProvider) generate(ctx context.Context, prompt string) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, perRequestTimeout)
	defer cancel()
	return p.gen.generate(ctx, prompt)
}

// GenerateScripts creates a main script and test script from a natural language description
func (p *scriptProvider) GenerateScripts(ctx context.Context, description string) (ScriptPair, error) {
	log.Info("Generating main script with %s...", p.gen.name())
	mainPrompt := p.formatPrompt(FeatureScriptPrompt, description)
	mainScript, err := p.generate(ctx, mainPrompt)
	if err != nil {
		return ScriptPair{}, fmt.Errorf("failed to generate main script: %w", err)
	}
	mainScript = ExtractScriptContent(mainScript)
	log.Debug("Main script generated:\n%s", mainScript)

	log.Info("Generating test script with %s...", p.gen.name())
	testPrompt := p.formatPrompt(TestScriptPrompt, mainScript, description)
	testScript, err := p.generate(ctx, testPrompt)
	if err != nil {
		return ScriptPair{}, fmt.Errorf("failed to generate test script: %w", err)
	}
	testScript = ExtractScriptContent(testScript)
	log.Debug("Test script generated:\n%s", testScript)

	return ScriptPair{
		MainScript: strings.TrimSpace(mainScript),
		TestScript: strings.TrimSpace(testScript),
	}, nil
}

// FixScripts attempts to fix the main script based on test failures
func (p *scriptProvider) FixScripts(ctx context.Context, scripts ScriptPair, failure string) (ScriptPair, error) {
	mainPrompt := p.formatPrompt(FixScriptPrompt, scripts.MainScript, failure)
	fixedMainScript, err := p.generate(ctx, mainPrompt)
	if err != nil {
		return ScriptPair{}, fmt.Errorf("failed to fix main script: %w", err)
	}
	fixedMainScript = ExtractScriptContent(fixedMainScript)

	return ScriptPair{
		MainScript: strings.TrimSpace(fixedMainScript),
		TestScript: scripts.TestScript,
	}, nil
}

// formatPrompt fills in a prompt template, appending the current platform
// information and, when configured, the user's additional instructions just
// before the output-format section so the model still sees the formatting
// rules last.
func (p *scriptProvider) formatPrompt(template string, args ...interface{}) string {
	args = append(args, GetPlatformInfo())
	prompt := fmt.Sprintf(template, args...)

	if extra := strings.TrimSpace(p.extraPrompt); extra != "" {
		block := "<additional_instructions>\n" + extra + "\n</additional_instructions>\n\n"
		prompt = strings.Replace(prompt, "<output_format>", block+"<output_format>", 1)
	}

	return prompt
}
