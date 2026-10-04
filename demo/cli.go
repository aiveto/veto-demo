package demo

import (
	"encoding/json"
	"fmt"
	"strings"
)

// CLI is veto search, describe, and invoke. --json serves the same three as JSON lines.
type CLI struct {
	Env Env
}

func (c CLI) Name() string { return "the CLI" }

func (c CLI) Discover() error {
	fmt.Print("\n2  The three capabilities\n")
	root, errText, err := c.Env.command("--help-json")
	if err != nil {
		return fail("veto --help-json", root, errText, err)
	}
	if !strings.Contains(root, "capabilities_search") || !strings.Contains(root, `"search"`) {
		return fmt.Errorf("veto --help-json did not list the capabilities\n%s", root)
	}
	spec, errText, err := c.Env.command("search", "--help-json")
	if err != nil {
		return fail("veto search --help-json", spec, errText, err)
	}
	var doc struct {
		Name  string `json:"name"`
		Input []struct {
			Name string `json:"name"`
		} `json:"input"`
	}
	if err := json.Unmarshal([]byte(spec), &doc); err != nil {
		return fmt.Errorf("search --help-json: %w\n%s", err, spec)
	}
	if doc.Name != "capabilities_search" {
		return fmt.Errorf("search --help-json name %q", doc.Name)
	}
	var fields []string
	for _, f := range doc.Input {
		fields = append(fields, f.Name)
	}
	if !contains(fields, "query") {
		return fmt.Errorf("search --help-json input is %v", fields)
	}
	if contains(fields, "config") {
		return fmt.Errorf("catalog flags leaked into input\n%s", spec)
	}
	invoke, errText, err := c.Env.command("invoke", "--help-json")
	if err != nil {
		return fail("veto invoke --help-json", invoke, errText, err)
	}
	if !strings.Contains(invoke, "idempotency_key") || !strings.Contains(invoke, "pending_approval") {
		return fmt.Errorf("invoke --help-json missed idempotency_key or pending_approval\n%s", invoke)
	}
	fmt.Print("    capabilities_search, capabilities_describe, capabilities_invoke\n")
	fmt.Print("    veto --help-json lists them. input is the tool arguments.\n")
	return nil
}

func (c CLI) Hits(query string) ([]Hit, error) {
	args := []string{"search", "--config", c.Env.Config()}
	args = append(args, strings.Fields(query)...)
	out, errText, err := c.Env.command(args...)
	if err != nil {
		return nil, fail("search "+query, out, errText, err)
	}
	return decodeHits(out)
}

func (c CLI) Describe(operationID string) (string, error) {
	out, errText, err := c.Env.command("describe", "--config", c.Env.Config(), operationID)
	if err != nil {
		return "", fail("describe "+operationID, out, errText, err)
	}
	return out, nil
}

func (c CLI) Invoke(operationID string, params map[string]any, approval string) (InvokeOut, error) {
	payload := map[string]any{"operation_id": operationID, "params": params}
	if approval != "" {
		payload["approval_id"] = approval
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		return InvokeOut{}, err
	}
	out, errText, err := c.Env.command("invoke", "--config", c.Env.Config(), string(raw))
	if err != nil && out == "" {
		return InvokeOut{}, fail("invoke "+operationID, out, errText, err)
	}
	got, decErr := decodeInvoke(out)
	if decErr != nil {
		if err != nil {
			return InvokeOut{}, fail("invoke "+operationID, out, errText, err)
		}
		return InvokeOut{}, decErr
	}
	return got, nil
}
