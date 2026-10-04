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
	session, err := demo.Connect(env)
	if err != nil {
		fmt.Fprintf(os.Stderr, "\n%s\n", err)
		os.Exit(1)
	}
	defer session.Close()
	if err := demo.Play(env, demo.MCP{Env: env, Session: session}); err != nil {
		fmt.Fprintf(os.Stderr, "\n%s\n", err)
		os.Exit(1)
	}
}
