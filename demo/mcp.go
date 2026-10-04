package demo

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// MCP is capabilities_search, capabilities_describe, and capabilities_invoke.
type MCP struct {
	Env     Env
	Session *mcp.ClientSession
}

func Connect(env Env) (*mcp.ClientSession, error) {
	cmd := exec.Command(env.Veto, "serve", "--config", env.Config(), "--stdio")
	cmd.Dir = env.Root
	cmd.Env = os.Environ()
	client := mcp.NewClient(&mcp.Implementation{Name: "veto-demo", Version: "0.1.0"}, nil)
	session, err := client.Connect(context.Background(), &mcp.CommandTransport{Command: cmd}, nil)
	if err != nil {
		return nil, fmt.Errorf("mcp connect: %w", err)
	}
	return session, nil
}

func (m MCP) Name() string { return "MCP" }

func (m MCP) Discover() error {
	fmt.Print("\n2  The three capabilities\n")
	listed, err := m.Session.ListTools(context.Background(), nil)
	if err != nil {
		return err
	}
	var names []string
	for _, tool := range listed.Tools {
		names = append(names, tool.Name)
	}
	got := strings.Join(names, ", ")
	if len(names) != 3 || !strings.Contains(got, "capabilities_search") || !strings.Contains(got, "capabilities_describe") || !strings.Contains(got, "capabilities_invoke") {
		return fmt.Errorf("mcp tools are %s", got)
	}
	var search, invoke string
	for _, tool := range listed.Tools {
		switch tool.Name {
		case "capabilities_search":
			search = tool.Description
		case "capabilities_invoke":
			invoke = tool.Description
		}
	}
	if !strings.Contains(search, "related_calls") {
		return fmt.Errorf("search description missed related_calls\n%s", search)
	}
	if !strings.Contains(invoke, "pending_approval") || !strings.Contains(invoke, "idempotency_key") {
		return fmt.Errorf("invoke description missed pending_approval or idempotency_key\n%s", invoke)
	}
	fmt.Printf("    %s\n", got)
	fmt.Print("    Harbor publishes more than twenty operations. The agent sees these three.\n")
	return nil
}

func (m MCP) Hits(query string) ([]Hit, error) {
	text, err := m.call("capabilities_search", map[string]any{"query": query})
	if err != nil {
		return nil, err
	}
	return decodeHits(text)
}

func (m MCP) Describe(operationID string) (string, error) {
	return m.call("capabilities_describe", map[string]any{"operation_id": operationID})
}

func (m MCP) Invoke(operationID string, params map[string]any, approval string) (InvokeOut, error) {
	args := map[string]any{"operation_id": operationID, "params": params}
	if approval != "" {
		args["approval_id"] = approval
	}
	text, err := m.call("capabilities_invoke", args)
	if err != nil {
		return InvokeOut{}, err
	}
	return decodeInvoke(text)
}

func (m MCP) call(name string, args map[string]any) (string, error) {
	res, err := m.Session.CallTool(context.Background(), &mcp.CallToolParams{Name: name, Arguments: args})
	if err != nil {
		return "", err
	}
	if res == nil || len(res.Content) == 0 {
		return "", fmt.Errorf("%s returned no content", name)
	}
	text, ok := res.Content[0].(*mcp.TextContent)
	if !ok {
		return "", fmt.Errorf("%s content is not text", name)
	}
	return text.Text, nil
}
