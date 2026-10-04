package demo

import (
	"bytes"
	"encoding/json"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/aiveto/veto-demo/app/desks"
)

func TestMain(m *testing.M) {
	if err := waitReady(); err != nil {
		desk := desks.New()
		for _, srv := range []*http.Server{
			{Addr: desks.OrdersAddr, Handler: desk.Handler("orders")},
			{Addr: desks.CustomersAddr, Handler: desk.Handler("customers")},
			{Addr: desks.InvoicesAddr, Handler: desk.Handler("invoices")},
		} {
			go func() { _ = srv.ListenAndServe() }()
		}
		if err := waitReady(); err != nil {
			os.Exit(m.Run())
		}
	}
	os.Exit(m.Run())
}

func TestCLIJSONAndMCPHoldTheSameClaims(t *testing.T) {
	env, err := Load()
	if err != nil {
		t.Skip(err.Error())
	}
	if err := waitReady(); err != nil {
		t.Skip(err.Error())
	}

	session, err := Connect(env)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = session.Close() })

	surfaces := []struct {
		name string
		a    Adapter
	}{
		{name: "cli", a: CLI{Env: env}},
		{name: "json", a: jsonSkill{Env: env}},
		{name: "mcp", a: MCP{Env: env, Session: session}},
	}

	for _, s := range surfaces {
		t.Run(s.name, func(t *testing.T) {
			if err := s.a.Discover(); err != nil {
				t.Fatal(err)
			}

			t.Run("semantics", func(t *testing.T) {
				for _, q := range []string{"retire order", "scrap order"} {
					hits, err := s.a.Hits(q)
					if err != nil {
						t.Fatal(err)
					}
					if len(hits) == 0 || hits[0].ID != "orders.delete" || !hits[0].Confirmation {
						t.Fatalf("search %q ranked %+v", q, hits)
					}
				}
			})

			t.Run("relations", func(t *testing.T) {
				text, err := s.a.Describe("orders.get")
				if err != nil {
					t.Fatal(err)
				}
				if !strings.Contains(text, "customers.get") || !strings.Contains(text, "invoices.get") {
					t.Fatalf("describe missed joins\n%s", text)
				}
				if strings.Contains(text, "warehouseId identifies") || strings.Contains(text, "openapi:") {
					t.Fatalf("describe leaked warehouse join or spec\n%s", text)
				}
				hits, err := s.a.Hits("orders.get")
				if err != nil {
					t.Fatal(err)
				}
				get := hitByID(hits, "orders.get")
				if get.ID == "" || !contains(get.Related, "customers.get") || !contains(get.Related, "invoices.get") {
					t.Fatalf("related %v", get.Related)
				}
				if !hasCall(get.RelatedCalls, "customers.get") || !hasCall(get.RelatedCalls, "invoices.get") {
					t.Fatalf("related_calls %v", get.RelatedCalls)
				}
				n := 0
				for _, id := range get.Related {
					if id == "customers.get" {
						n++
					}
				}
				if n != 1 {
					t.Fatalf("customers.get listed %d times", n)
				}
			})

			t.Run("read", func(t *testing.T) {
				out, err := s.a.Invoke("orders.get", map[string]any{"id": "10482"}, "")
				if err != nil {
					t.Fatal(err)
				}
				if out.Status != "ok" || !out.HTTP || !out.Sent || !strings.Contains(out.Body, "cus_mara") {
					t.Fatalf("orders.get %+v", out)
				}
			})

			t.Run("invalid_query", func(t *testing.T) {
				before := mustLog()
				out, err := s.a.Invoke("orders.list", map[string]any{"status": "banana"}, "")
				if err != nil && out.Code == "" {
					t.Fatal(err)
				}
				if out.Code != "invalid_param" || out.HTTP || out.Sent || strings.Contains(out.Error, "banana") {
					t.Fatalf("status=banana %+v", out)
				}
				if len(mustLog().Calls) != len(before.Calls) {
					t.Fatal("invalid query reached HTTP")
				}
			})

			t.Run("invalid_limit", func(t *testing.T) {
				out, err := s.a.Invoke("orders.list", map[string]any{"limit": "nope"}, "")
				if err != nil && out.Code == "" {
					t.Fatal(err)
				}
				if out.Code != "invalid_param" || out.HTTP {
					t.Fatalf("limit=nope %+v", out)
				}
			})

			t.Run("preview_integer", func(t *testing.T) {
				const id = "9007199254740993"
				got := invokePreview(t, s.a, env, "orders.create", map[string]any{
					"body": map[string]any{
						"customerId": "cus_mara",
						"lines": []map[string]any{{
							"sku": "X", "name": "x", "quantity": 1, "unitAmount": json.Number(id),
						}},
					},
				})
				if !strings.Contains(got, id) || strings.Contains(got, "9007199254740992") {
					t.Fatalf("preview lost integer digits\n%s", got)
				}
			})

			t.Run("hold_and_pending", func(t *testing.T) {
				before := mustLog().Deletes
				held, err := s.a.Invoke("orders.delete", map[string]any{"id": "10502"}, "")
				if err != nil {
					t.Fatal(err)
				}
				if held.Status != "confirmation_required" || held.ApprovalID == "" || held.HTTP || held.Sent {
					t.Fatalf("hold %+v", held)
				}
				replay, err := s.a.Invoke("orders.delete", map[string]any{"id": "10502"}, held.ApprovalID)
				if err != nil && replay.Code == "" {
					t.Fatal(err)
				}
				if replay.Code != "pending_approval" || replay.HTTP || mustLog().Deletes != before {
					t.Fatalf("pending as approval %+v", replay)
				}
			})
		})
	}

	t.Run("parity_search_describe", func(t *testing.T) {
		cliHits, err := CLI{Env: env}.Hits("retire order 10482")
		if err != nil {
			t.Fatal(err)
		}
		jsonHits, err := jsonSkill{Env: env}.Hits("retire order 10482")
		if err != nil {
			t.Fatal(err)
		}
		mcpHits, err := MCP{Env: env, Session: session}.Hits("retire order 10482")
		if err != nil {
			t.Fatal(err)
		}
		if cliHits[0].ID != jsonHits[0].ID || cliHits[0].ID != mcpHits[0].ID {
			t.Fatalf("search ids cli=%s json=%s mcp=%s", cliHits[0].ID, jsonHits[0].ID, mcpHits[0].ID)
		}
		if !sameStrings(cliHits[0].Related, jsonHits[0].Related) || !sameStrings(cliHits[0].Related, mcpHits[0].Related) {
			t.Fatalf("related cli=%v json=%v mcp=%v", cliHits[0].Related, jsonHits[0].Related, mcpHits[0].Related)
		}
		if !sameStrings(cliHits[0].RelatedCalls, jsonHits[0].RelatedCalls) || !sameStrings(cliHits[0].RelatedCalls, mcpHits[0].RelatedCalls) {
			t.Fatalf("related_calls cli=%v json=%v mcp=%v", cliHits[0].RelatedCalls, jsonHits[0].RelatedCalls, mcpHits[0].RelatedCalls)
		}
		cliDesc, err := CLI{Env: env}.Describe("orders.get")
		if err != nil {
			t.Fatal(err)
		}
		jsonDesc, err := jsonSkill{Env: env}.Describe("orders.get")
		if err != nil {
			t.Fatal(err)
		}
		mcpDesc, err := MCP{Env: env, Session: session}.Describe("orders.get")
		if err != nil {
			t.Fatal(err)
		}
		if !jsonEqual(cliDesc, jsonDesc) || !jsonEqual(cliDesc, mcpDesc) {
			t.Fatalf("describe drifted\ncli %s\njson %s\nmcp %s", cliDesc, jsonDesc, mcpDesc)
		}
	})

	t.Run("approve_consume_once_across_surfaces", func(t *testing.T) {
		held, err := CLI{Env: env}.Invoke("orders.delete", map[string]any{"id": "10502"}, "")
		if err != nil {
			t.Fatal(err)
		}
		if held.Status != "confirmation_required" {
			t.Fatalf("hold %+v", held)
		}
		approved, errText, err := env.command("approve", "--config", env.Config(), held.ApprovalID)
		if err != nil {
			t.Fatalf("approve: %v\n%s", err, errText)
		}
		approved = strings.TrimSpace(approved)
		if approved == "" || approved == held.ApprovalID {
			t.Fatalf("approve returned %q", approved)
		}
		sent, err := jsonSkill{Env: env}.Invoke("orders.delete", map[string]any{"id": "10502"}, approved)
		if err != nil {
			t.Fatal(err)
		}
		if sent.Status != "ok" || !sent.HTTP {
			t.Fatalf("approved delete via json %+v", sent)
		}
		again, err := MCP{Env: env, Session: session}.Invoke("orders.delete", map[string]any{"id": "10502"}, approved)
		if err != nil && again.Error == "" {
			t.Fatal(err)
		}
		if again.Status == "ok" || !strings.Contains(again.Error, "invalid approval") {
			t.Fatalf("spent approval via mcp %+v", again)
		}
	})

	t.Run("json_v1_v2", func(t *testing.T) {
		for _, ver := range []string{"v1", "v2"} {
			cfg := writeJSONConfig(t, env, ver)
			out, errText, err := env.command("invoke", "--config", cfg, `{"operation_id":"orders.get","params":{"id":"10482"}}`)
			if err != nil {
				t.Fatalf("json %s: %v\n%s", ver, err, errText)
			}
			var res InvokeOut
			if err := json.Unmarshal([]byte(out), &res); err != nil {
				t.Fatal(err)
			}
			if res.Status != "ok" || !strings.Contains(res.Body, "cus_mara") {
				t.Fatalf("json %s %+v", ver, res)
			}
		}
	})

	t.Run("same_op_relation_refused", func(t *testing.T) {
		dir := t.TempDir()
		rel := filepath.Join(dir, "relations.yaml")
		if err := os.WriteFile(rel, []byte("relations:\n  - schema: OrderCreate\n    field: customerId\n    to: orders.create\n"), 0o600); err != nil {
			t.Fatal(err)
		}
		cfg := filepath.Join(dir, "veto.yaml")
		body := "contracts:\n  - " + filepath.Join(env.Root, "app/contracts/orders.yaml") + "\nrelations_file: " + rel + "\n"
		if err := os.WriteFile(cfg, []byte(body), 0o600); err != nil {
			t.Fatal(err)
		}
		_, errText, err := env.command("doctor", "--config", cfg)
		if err == nil || !strings.Contains(errText, "same operation") {
			t.Fatalf("same-op relation: err=%v stderr=%s", err, errText)
		}
	})

	t.Run("isolation", func(t *testing.T) {
		ada, errText, err := env.command("invoke", "--config", env.Config(), "--caller", "ada", `{"operation_id":"orders.delete","params":{"id":"10482"}}`)
		if err != nil && !strings.Contains(ada, "confirmation_required") {
			t.Fatalf("ada: %v\n%s%s", err, ada, errText)
		}
		var held InvokeOut
		if err := json.Unmarshal([]byte(ada), &held); err != nil {
			t.Fatal(err)
		}
		if held.Status != "confirmation_required" || held.ApprovalID == "" {
			t.Fatalf("ada hold %+v", held)
		}
		approved, errText, err := env.command("approve", "--config", env.Config(), held.ApprovalID)
		if err != nil {
			t.Fatalf("approve ada: %v\n%s", err, errText)
		}
		payload := `{"operation_id":"orders.delete","params":{"id":"10482"},"approval_id":"` + strings.TrimSpace(approved) + `"}`
		grace, errText, err := env.command("invoke", "--config", env.Config(), "--caller", "grace", payload)
		if err == nil && strings.Contains(grace, `"status":"ok"`) {
			t.Fatalf("grace used ada approval\n%s", grace)
		}
		if !strings.Contains(grace+errText, "invalid approval") && !strings.Contains(grace, "invalid approval") {
			t.Fatalf("grace should be invalid approval\n%s%s", grace, errText)
		}
	})

	t.Run("pack_omits_spec", func(t *testing.T) {
		out, errText, err := env.command("pack", "--config", env.Config(), "--message", "who placed order 10482")
		if err != nil {
			t.Fatalf("pack: %v\n%s%s", err, out, errText)
		}
		if strings.Contains(out, "openapi:") || strings.Contains(out, "paths:") {
			t.Fatalf("pack included the spec\n%s", out)
		}
		if !strings.Contains(out, "customers.get") {
			t.Fatalf("pack missed the join\n%s", out)
		}
	})

	t.Run("perf", func(t *testing.T) {
		jsonP50 := timeJSONSearch(t, env, 40)
		cliP50 := timeCLISearch(t, env, 16)
		t.Logf("json search mean=%s cli search mean=%s", jsonP50, cliP50)
		if jsonP50 > 80*time.Millisecond {
			t.Fatalf("json search mean %s", jsonP50)
		}
		if cliP50 > 250*time.Millisecond {
			t.Fatalf("cli search mean %s", cliP50)
		}
	})
}

type jsonSkill struct{ Env Env }

func (j jsonSkill) Name() string { return "json" }

func (j jsonSkill) Discover() error {
	root, errText, err := j.Env.command("--help-json")
	if err != nil {
		return fail("veto --help-json", root, errText, err)
	}
	if !strings.Contains(root, "capabilities_search") || !strings.Contains(root, "capabilities_invoke") {
		return fail("veto --help-json", root, errText, err)
	}
	return nil
}

func (j jsonSkill) Hits(query string) ([]Hit, error) {
	text, err := j.call(map[string]any{"search": map[string]any{"query": query}})
	if err != nil {
		return nil, err
	}
	return decodeHits(text)
}

func (j jsonSkill) Describe(operationID string) (string, error) {
	return j.call(map[string]any{"describe": map[string]any{"operation_id": operationID}})
}

func (j jsonSkill) Invoke(operationID string, params map[string]any, approval string) (InvokeOut, error) {
	args := map[string]any{"operation_id": operationID, "params": params}
	if approval != "" {
		args["approval_id"] = approval
	}
	text, err := j.call(map[string]any{"invoke": args})
	if err != nil {
		if out, decErr := decodeInvoke(text); decErr == nil {
			return out, nil
		}
		return InvokeOut{}, err
	}
	return decodeInvoke(text)
}

func (j jsonSkill) call(line map[string]any) (string, error) {
	raw, err := json.Marshal(line)
	if err != nil {
		return "", err
	}
	cmd := exec.Command(j.Env.Veto, "serve", "--json", "--config", j.Env.Config())
	cmd.Dir = j.Env.Root
	cmd.Env = os.Environ()
	cmd.Stdin = bytes.NewReader(append(raw, '\n'))
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err = cmd.Run()
	out := strings.TrimSpace(stdout.String())
	if err != nil && out == "" {
		return "", fail("serve --json", out, stderr.String(), err)
	}
	return out, nil
}

func invokePreview(t *testing.T, a Adapter, env Env, op string, params map[string]any) string {
	t.Helper()
	switch v := a.(type) {
	case jsonSkill:
		text, err := v.call(map[string]any{"invoke": map[string]any{"operation_id": op, "params": params, "preview": true}})
		if err != nil {
			t.Fatal(err)
		}
		return text
	case MCP:
		text, err := v.call("capabilities_invoke", map[string]any{"operation_id": op, "params": params, "preview": true})
		if err != nil {
			t.Fatal(err)
		}
		return text
	default:
		raw, err := json.Marshal(map[string]any{"operation_id": op, "params": params, "preview": true})
		if err != nil {
			t.Fatal(err)
		}
		out, errText, err := env.command("invoke", "--config", env.Config(), string(raw))
		if err != nil {
			t.Fatalf("preview: %v\n%s", err, errText)
		}
		return out
	}
}

func writeJSONConfig(t *testing.T, env Env, ver string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "veto.yaml")
	body := "json: " + ver + "\ncontracts:\n  - " + filepath.Join(env.Root, "app/contracts/orders.yaml") + "\n  - " + filepath.Join(env.Root, "app/contracts/customers.yaml") + "\n  - " + filepath.Join(env.Root, "app/contracts/invoices.yaml") + "\nrelations_file: " + filepath.Join(env.Root, "app/relations.yaml") + "\nsemantics: file\nsemantics_file: " + filepath.Join(env.Root, "app/semantics.yaml") + "\nauth:\n  bearerAuth: DEMO_TOKEN\n"
	if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

func timeJSONSearch(t *testing.T, env Env, n int) time.Duration {
	t.Helper()
	var lines bytes.Buffer
	for i := 0; i < n; i++ {
		lines.WriteString(`{"search":{"query":"open orders"}}` + "\n")
	}
	cmd := exec.Command(env.Veto, "serve", "--json", "--config", env.Config())
	cmd.Dir = env.Root
	cmd.Env = os.Environ()
	cmd.Stdin = &lines
	start := time.Now()
	out, err := cmd.CombinedOutput()
	elapsed := time.Since(start)
	if err != nil {
		t.Fatalf("json search: %v\n%s", err, out)
	}
	return elapsed / time.Duration(n)
}

func timeCLISearch(t *testing.T, env Env, n int) time.Duration {
	t.Helper()
	var total time.Duration
	for i := 0; i < n; i++ {
		start := time.Now()
		_, errText, err := env.command("search", "--config", env.Config(), "open", "orders")
		if err != nil {
			t.Fatal(errText, err)
		}
		total += time.Since(start)
	}
	return total / time.Duration(n)
}

func sameStrings(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func jsonEqual(a, b string) bool {
	var x, y any
	if json.Unmarshal([]byte(a), &x) != nil || json.Unmarshal([]byte(b), &y) != nil {
		return false
	}
	left, err := json.Marshal(x)
	if err != nil {
		return false
	}
	right, err := json.Marshal(y)
	if err != nil {
		return false
	}
	return string(left) == string(right)
}
