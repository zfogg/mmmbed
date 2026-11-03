// Example of generating a Go source file containing a small asset.
package main

import "fmt"

//go:generate go run ../../cmd/mmmbed -l go --go-package main --omit-parent --variable Greeting greeting.txt greeting_generated.go

func main() { fmt.Print(string(Greeting)) }
