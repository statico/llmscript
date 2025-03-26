package llm

const (
	// FeatureScriptPrompt is used to generate the main feature script
	FeatureScriptPrompt = `You are an expert shell script developer with deep knowledge of Unix/Linux systems, shell scripting best practices, and error handling.
Your task is to create robust, maintainable shell scripts that work reliably across different environments.

Create a shell script that accomplishes the following task:

<description>
%s
</description>

<target_platform>
Target platform Information:
%s
</target_platform>

<requirements>
- Use standard shell commands (sh/bash) with POSIX compliance where possible
- Only use argument and environment variables if the description requires it
- Follow shell scripting best practices
- Ensure cross-platform compatibility
- Use proper exit codes for different scenarios
- Keep the script short, concise, and simple
</requirements>

<output_format>
Output your response in the following format:

<script>
#!/usr/bin/env bash
# Your shell script content here
</script>

You *MUST NOT* include any other text, explanations, or markdown formatting.
</output_format>`

	// TestScriptPrompt is used to generate the test script
	TestScriptPrompt = `You are an expert in testing shell scripts with extensive experience in test automation and quality assurance.
Your goal is to create a test script that verifies the functionality of the main script, ./script.sh, including edge cases and error conditions.

Create a test script for the following script:

<script>
%s
</script>

<description>
%s
</description>

<target_platform>
Target platform Information:
%s
</target_platform>

<requirements>
- The script you're testing is ./script.sh
- Create a test script that runs one or more test cases to make sure that ./script.sh works as expected
- Each test case should:
   - Set up the test environment
   - Run the main script with test inputs
   - Verify the output matches expectations
   - Clean up after the test only if necessary
- Use standard shell commands (sh/bash) with POSIX compliance where possible
- The test script should not need any arguments to run
- Return exit code 0 if all tests pass, or 1 if any test fails
- Set appropriate timeouts for any long-running tests
- Handle environment variables and cleanup
- Ensure platform compatibility
- Keep the script short, concise, and simple
</requirements>

<output_format>
Output your response in the following format:

<script>
#!/usr/bin/env bash
# Your test script content here
</script>

You *MUST NOT* include any other text, explanations, or markdown formatting.
</output_format>`

	// FixScriptPrompt is used to fix a script based on test failures
	FixScriptPrompt = `You are an expert shell script developer specializing in debugging and fixing shell scripts.
Your expertise includes error handling, cross-platform compatibility, and shell scripting best practices.

Fix the following script based on the test failures:

<script>
%s
</script>

<test_failures>
%s
</test_failures>

<target_platform>
Target platform Information:
%s
</target_platform>

<requirements>
- Fix all test failures while maintaining existing functionality
- Improve error handling and validation
- Follow shell scripting best practices
- Ensure cross-platform compatibility
- Keep the script short, concise, and simple
</requirements>

<output_format>
Output your response in the following format:

<script>
#!/usr/bin/env bash
# Your fixed shell script content here
</script>

Do not include any other text, explanations, or markdown formatting. Only output the script between the markers.
</output_format>`
)
