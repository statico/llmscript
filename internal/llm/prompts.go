package llm

const (
	generateScriptsPrompt = `You are a shell script expert. Create two shell scripts based on this description:

%s

IMPORTANT: You MUST format your response EXACTLY as follows, with NO additional text or explanations:

#!/bin/bash
# First script content here
# This is the main script that does the work
---
#!/bin/bash
# Second script content here
# This is the test script that verifies the main script

The test script should:
1. Set up any necessary test environment
2. Run the main script with ./script.sh
3. Verify the output and state
4. Exit with 0 for success, non-zero for failure

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
