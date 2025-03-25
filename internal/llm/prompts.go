package llm

const (
	generateScriptsPrompt = `You are a shell script expert. Create two shell scripts based on this description:

%s

The first script should be the main script that does the work. The second script should be a test script that verifies the main script works correctly.

Format your response as:
<main script>
---
<test script>

The test script should:
1. Set up any necessary test environment
2. Run the main script
3. Verify the output and state
4. Exit with success/failure status

Platform info: %s`

	fixScriptsPrompt = `The following scripts failed their tests:

Main script:
%s

Test script:
%s

Error:
%s

Please fix both scripts to make the tests pass. Format your response as:
<fixed main script>
---
<fixed test script>

Platform info: %s`
)
