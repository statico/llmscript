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
- Implement comprehensive error handling with descriptive messages
- Use clear, descriptive variable names following shell naming conventions
- Add concise, meaningful comments for complex logic
- Follow shell scripting best practices and security guidelines
- Ensure cross-platform compatibility
- Include input validation where appropriate
- Use proper exit codes for different scenarios
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
Your goal is to create a comprehensive test script that verifies the functionality of the main script, including edge cases and error conditions.

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
- Create a test script that runs multiple test cases
- Each test case should:
   - Set up the test environment
   - Run the main script with test inputs
   - Verify the output matches expectations
   - Clean up after the test
- Include tests for:
   - Success scenarios
   - Error cases
   - Edge cases
   - Boundary conditions
- Use clear test case names and descriptions
- Implement proper error handling and reporting
- Set appropriate timeouts for long-running tests
- Handle environment variables and cleanup
- Ensure platform compatibility
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
- Enhance code readability and maintainability
- Follow shell scripting best practices
- Ensure cross-platform compatibility
- Add appropriate logging for debugging
- Implement proper cleanup in error cases
- Optimize performance where possible
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
