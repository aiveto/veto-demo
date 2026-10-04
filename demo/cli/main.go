package main

import (
	"fmt"
	"os"

	"github.com/aiveto/veto-demo/demo"
)

func main() {
	env, err := demo.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "\n%s\n", err)
		os.Exit(1)
	}
	if err := demo.Play(env, demo.CLI{Env: env}); err != nil {
		fmt.Fprintf(os.Stderr, "\n%s\n", err)
		os.Exit(1)
	}
}
