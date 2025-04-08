package main

import (
	"fmt"

	"github.com/sshturbo/GoTeleMD"
	"github.com/sshturbo/GoTeleMD/pkg/types"
)

func main() {
	// Create a new converter with custom configuration
	converter := GoTeleMD.NewConverter(
		types.WithSafetyLevel(GoTeleMD.SAFETYLEVELSTRICT),
		types.WithTableAlignment(true),
		types.WithMaxMessageLength(4096),
		types.WithDebugLogs(true),
	)

	// Sample Markdown text with various elements
	markdown := `# Header 1
## Header 2
### Header 3

Normal text with **bold** and _italic_ formatting.

* Unordered list item 1
* Unordered list item 2
  * Nested item
* Unordered list item 3

1. Ordered list item 1
2. Ordered list item 2
3. Ordered list item 3

> This is a blockquote
> With multiple lines

` + "```go" + `
package main

func main() {
    fmt.Println("Hello World!")
}
` + "```" + `

| Column 1 | Column 2 | Column 3 |
|:---------|:--------:|----------:|
| Left     | Center   | Right    |
| align    | align    | align    |`

	// Convert the Markdown text
	response, err := converter.Convert(markdown)
	if err != nil {
		panic(err)
	}

	// Print the converted messages
	for _, part := range response.Parts {
		fmt.Printf("Part %d: %s\n", part.Part, part.Content)
	}
}
