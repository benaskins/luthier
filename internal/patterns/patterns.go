// Package patterns holds the embedded system prompt used for the analysis call.
package patterns

import _ "embed"

//go:embed system_prompt.txt
var SystemPrompt string
