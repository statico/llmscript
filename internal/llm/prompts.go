package llm

const (
	generateScriptsPrompt = `You are a shell script expert. Create two shell scripts based on this description:

<description>
%s
</description>

<scripts>
# Feature script content here
# This is the feature script that implements the functionality

%s

# Test script content here
# This is the test script that verifies the feature script

%s
</scripts>

The test script should:
1. Be executable
2. Run the feature script with ./script.sh
3. Exit with code 0 if all tests pass, non-zero if any fail`

	fixScriptsPrompt = `The following scripts failed their tests:

Feature script:
%s

Test script:
%s

Error:
%s

Please fix both scripts to make the tests pass. Format your response EXACTLY as follows:

# Fixed feature script content here
%s

# Fixed test script content here
%s`
)
