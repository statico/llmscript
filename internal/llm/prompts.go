package llm

const (
	// FeatureScriptPrompt is used to generate the main feature script
	FeatureScriptPrompt = `You are an expert shell script developer with deep knowledge of Unix/Linux systems, shell scripting best practices, and error handling.
Your task is to create robust, maintainable shell scripts that work reliably across different environments.

Create a shell script that accomplishes the following task:

<description>
%s
</description>

<platform>
Platform Information:
%s
</platform>

<requirements>
1. Use standard shell commands (sh/bash) with POSIX compliance where possible
2. Implement comprehensive error handling with descriptive messages
3. Use clear, descriptive variable names following shell naming conventions
4. Add concise, meaningful comments for complex logic
5. Follow shell scripting best practices and security guidelines
6. Ensure cross-platform compatibility
7. Include input validation where appropriate
8. Use proper exit codes for different scenarios
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

<platform>
Platform Information:
%s
</platform>

<requirements>
1. Create a test script that runs multiple test cases
2. Each test case should:
   - Set up the test environment
   - Run the main script with test inputs
   - Verify the output matches expectations
   - Clean up after the test
3. Include tests for:
   - Success scenarios
   - Error cases
   - Edge cases
   - Boundary conditions
4. Use clear test case names and descriptions
5. Implement proper error handling and reporting
6. Set appropriate timeouts for long-running tests
7. Handle environment variables and cleanup
8. Ensure platform compatibility
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

<platform>
Platform Information:
%s
</platform>

<requirements>
1. Fix all test failures while maintaining existing functionality
2. Improve error handling and validation
3. Enhance code readability and maintainability
4. Follow shell scripting best practices
5. Ensure cross-platform compatibility
6. Add appropriate logging for debugging
7. Implement proper cleanup in error cases
8. Optimize performance where possible
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
