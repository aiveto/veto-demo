// Package demo plays Harbor's story through one adapter. MCP and the CLI are the two adapters.
package demo

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

type (
	Env struct {
		Root string
		Veto string
	}

	Adapter interface {
		Name() string
		Discover() error
		Hits(query string) ([]Hit, error)
		Describe(operationID string) (string, error)
		Invoke(operationID string, params map[string]any, approval string) (InvokeOut, error)
	}

	Hit struct {
		ID           string   `json:"id"`
		Call         string   `json:"call"`
		Related      []string `json:"related"`
		RelatedCalls []string `json:"related_calls"`
		Confirmation bool     `json:"confirmation"`
	}

	InvokeOut struct {
		Status      string `json:"status"`
		ApprovalID  string `json:"approval_id"`
		OperationID string `json:"operation_id"`
		HTTPStatus  int    `json:"http_status"`
		Body        string `json:"body"`
		Error       string `json:"error"`
		Code        string `json:"code"`
		Why         string `json:"why"`
		HTTP        bool   `json:"http"`
		Sent        bool   `json:"sent"`
	}

	hitLog struct {
		Deletes int `json:"deletes"`
		Calls   []struct {
			Service string `json:"service"`
			Method  string `json:"method"`
			Path    string `json:"path"`
			Status  int    `json:"status"`
		} `json:"calls"`
	}
)

func Load() (Env, error) {
	root, err := findRoot()
	if err != nil {
		return Env{}, err
	}
	prepareEnv(root)
	veto := filepath.Join(root, "bin", "veto")
	if _, err := os.Stat(veto); err != nil {
		return Env{}, errors.New("bin/veto is missing. Run make from the veto-demo directory")
	}
	return Env{Root: root, Veto: veto}, nil
}

func Play(env Env, a Adapter) error {
	if err := waitReady(); err != nil {
		return err
	}

	fmt.Printf("veto\nHarbor's orders, customers, and billing are up. This walk is %s.\nThe model proposes. Veto decides.\n", a.Name())

	if err := catalog(env); err != nil {
		return err
	}
	if err := a.Discover(); err != nil {
		return err
	}
	if err := semantics(a); err != nil {
		return err
	}
	if err := relations(a); err != nil {
		return err
	}
	if err := allowedWrite(a); err != nil {
		return err
	}
	pending, err := invokeStops(a)
	if err != nil {
		return err
	}
	if err := personApproves(env, a, pending); err != nil {
		return err
	}
	if err := cases(env); err != nil {
		return err
	}

	fmt.Print(`
Mara's order was read, then her customer, then her invoice.
The warehouse id stayed a field.
Cancel reached the desk. Delete did not, until a person approved it.
The approved id ran once.
`)
	return nil
}

func catalog(env Env) error {
	fmt.Print("\n1  Catalog\n")
	out, errText, err := env.command("validate", "--config", env.Config())
	if err != nil {
		return fail("validate", out, errText, err)
	}
	if !strings.Contains(out, "operations)") || !strings.Contains(out, "customers.get") || !strings.Contains(out, "invoices.get") {
		return fmt.Errorf("validate did not show the catalog and the joins\n%s", out)
	}
	fmt.Print(indent(out))
	doc, errText, err := env.command("doctor", "--config", env.Config())
	if err != nil {
		return fail("doctor", doc, errText, err)
	}
	if !strings.Contains(doc+errText, "orders.delete") {
		return fmt.Errorf("doctor did not name orders.delete\n%s%s", doc, errText)
	}
	fmt.Print("    Doctor loaded Harbor's APIs. orders.delete is waiting on a person.\n")
	return nil
}

func semantics(a Adapter) error {
	fmt.Print("\n3  Semantics\n")
	for _, query := range []string{"retire order", "scrap order"} {
		hits, err := a.Hits(query)
		if err != nil {
			return err
		}
		if len(hits) == 0 || hits[0].ID != "orders.delete" {
			return fmt.Errorf("search %q ranked %v first", query, hits)
		}
		if !hits[0].Confirmation {
			return fmt.Errorf("search %q missed confirmation on orders.delete", query)
		}
		fmt.Printf("    %q  →  orders.delete\n", query)
	}
	text, err := a.Describe("orders.delete")
	if err != nil {
		return err
	}
	if !strings.Contains(text, "waits for a person") {
		return fmt.Errorf("describe missed the semantics sentence\n%s", text)
	}
	if strings.Contains(text, "openapi:") {
		return errors.New("describe included the raw spec")
	}
	fmt.Print("    retire is the built-in synonym. scrap is the overlay. The sentence says the call waits.\n")
	return nil
}

func relations(a Adapter) error {
	fmt.Print("\n4  Relations\n")
	before := mustLog()
	text, err := a.Describe("orders.get")
	if err != nil {
		return err
	}
	if !strings.Contains(text, "customers.get") || !strings.Contains(text, "invoices.get") {
		return fmt.Errorf("describe did not name both joins\n%s", text)
	}
	if strings.Contains(text, "warehouseId identifies") {
		return errors.New("warehouseId was treated as a join")
	}
	hits, err := a.Hits("orders.get")
	if err != nil {
		return err
	}
	get := hitByID(hits, "orders.get")
	if get.ID == "" {
		return fmt.Errorf("search missed orders.get\n%+v", hits)
	}
	if !contains(get.Related, "customers.get") || !contains(get.Related, "invoices.get") {
		return fmt.Errorf("search related is %v", get.Related)
	}
	if !hasCall(get.RelatedCalls, "customers.get") || !hasCall(get.RelatedCalls, "invoices.get") {
		return fmt.Errorf("search related_calls is %v", get.RelatedCalls)
	}
	fmt.Print("    orders.get  →  customers.get   by Order.customerId\n")
	fmt.Print("    orders.get  →  invoices.get    by Order.invoiceId\n")
	fmt.Print("    search related_calls is the next invoke line.\n")
	order, err := a.Invoke("orders.get", map[string]any{"id": "10482"}, "")
	if err != nil {
		return err
	}
	if order.Status != "ok" || !strings.Contains(order.Body, "cus_mara") || !strings.Contains(order.Body, "inv_2291") || !strings.Contains(order.Body, "wh_sfo_1") {
		return fmt.Errorf("orders.get body %s", order.Body)
	}
	customer, err := a.Invoke("customers.get", map[string]any{"id": "cus_mara"}, "")
	if err != nil {
		return err
	}
	if customer.Status != "ok" || !strings.Contains(customer.Body, "Mara Ellison") {
		return fmt.Errorf("customers.get %s %s", customer.Status, customer.Body)
	}
	invoice, err := a.Invoke("invoices.get", map[string]any{"id": "inv_2291"}, "")
	if err != nil {
		return err
	}
	if invoice.Status != "ok" || !strings.Contains(invoice.Body, "paid") {
		return fmt.Errorf("invoices.get %s %s", invoice.Status, invoice.Body)
	}
	after := mustLog()
	if after.Deletes != before.Deletes {
		return errors.New("a read deleted an order")
	}
	if !saw(after, "orders", "GET", "/orders/10482") || !saw(after, "customers", "GET", "/customers/cus_mara") || !saw(after, "invoices", "GET", "/invoices/inv_2291") {
		return fmt.Errorf("the joins did not reach the desks\n%+v", after.Calls)
	}
	fmt.Print("    10482  Mara Ellison  wool coat and two cedar trays  $556.00  invoice paid\n")
	fmt.Print("    warehouse wh_sfo_1 was on the order. Nothing called it.\n")
	return nil
}

func allowedWrite(a Adapter) error {
	fmt.Print("\n5  Cancel reaches the desk\n")
	before := mustLog()
	cancelled, err := a.Invoke("orders.cancel", map[string]any{
		"id":   "10490",
		"body": map[string]any{"reason": "buyer asked before it shipped"},
	}, "")
	if err != nil {
		return err
	}
	if cancelled.Status != "ok" {
		return fmt.Errorf("cancel status %s %s", cancelled.Status, cancelled.Error)
	}
	after := mustLog()
	if after.Deletes != before.Deletes || !saw(after, "orders", "POST", "/orders/10490/cancel") {
		return fmt.Errorf("cancel did not reach the desk once\n%+v", after.Calls)
	}
	fmt.Print("    orders.cancel 10490 went to the desk. No approval. Jonas Adler's linen throw is cancelled.\n")
	return nil
}

func invokeStops(a Adapter) (string, error) {
	fmt.Print("\n6  Invoke stops\n")
	before := mustLog()
	held, err := a.Invoke("orders.delete", map[string]any{"id": "10482"}, "")
	if err != nil {
		return "", err
	}
	if held.Status != "confirmation_required" || held.ApprovalID == "" {
		return "", fmt.Errorf("invoke status %s id %q", held.Status, held.ApprovalID)
	}
	if held.HTTP || held.Sent || mustLog().Deletes != before.Deletes {
		return "", errors.New("the pending id reached the transport")
	}
	replay, err := a.Invoke("orders.delete", map[string]any{"id": "10482"}, held.ApprovalID)
	if err != nil && replay.Code == "" {
		return "", err
	}
	if replay.Code != "pending_approval" || replay.HTTP || replay.Sent || mustLog().Deletes != before.Deletes {
		return "", fmt.Errorf("pending id as approval: status %s code %s http %t sent %t", replay.Status, replay.Code, replay.HTTP, replay.Sent)
	}
	fmt.Printf("    %s orders.delete  →  confirmation_required\n", a.Name())
	fmt.Printf("    pending %s\n", held.ApprovalID)
	fmt.Print("    that id as approval_id is pending_approval. No HTTP.\n")
	return held.ApprovalID, nil
}

func personApproves(env Env, a Adapter, pending string) error {
	fmt.Print("\n7  A person says yes\n")
	if pending == "" {
		return errors.New("no pending id")
	}
	approved, errText, err := env.command("approve", "--config", env.Config(), pending)
	if err != nil {
		return fail("approve", approved, errText, err)
	}
	approved = strings.TrimSpace(approved)
	if approved == "" || approved == pending {
		return fmt.Errorf("approve returned %q", approved)
	}
	fmt.Printf("    veto approve  →  %s\n", approved)
	sent, err := a.Invoke("orders.delete", map[string]any{"id": "10482"}, approved)
	if err != nil {
		return err
	}
	if sent.Status != "ok" || sent.HTTPStatus != http.StatusNoContent {
		return fmt.Errorf("approved delete status %s http %d %s", sent.Status, sent.HTTPStatus, sent.Error)
	}
	log := mustLog()
	if log.Deletes != 1 {
		return fmt.Errorf("desk recorded %d deletes", log.Deletes)
	}
	again, err := a.Invoke("orders.delete", map[string]any{"id": "10482"}, approved)
	if err != nil {
		return err
	}
	if again.Status == "ok" || !strings.Contains(again.Error, "invalid approval") {
		return fmt.Errorf("the approved id ran twice: %s %s", again.Status, again.Error)
	}
	if mustLog().Deletes != 1 {
		return errors.New("the spent approval deleted again")
	}
	if status, _ := getStatus("http://127.0.0.1:18410/orders/10482"); status != http.StatusNotFound {
		return fmt.Errorf("order 10482 still answers %d", status)
	}
	fmt.Print("    One DELETE /orders/10482. The same id does not run again. The order is gone.\n")
	return nil
}

func cases(env Env) error {
	fmt.Print("\n8  Eval\n")
	before := mustLog().Deletes
	out, errText, err := env.command("eval", "--config", env.Config(), "--case", filepath.Join(env.Root, "app", "cases"))
	if err != nil {
		return fail("eval", out, errText, err)
	}
	if mustLog().Deletes != before {
		return errors.New("eval reached the desk")
	}
	fmt.Print("    veto eval  delete still requires confirmation, and the pack still omits the spec.\n")
	return nil
}

func (e Env) Config() string {
	return filepath.Join(e.Root, "app", "veto.yaml")
}

func (e Env) command(args ...string) (string, string, error) {
	cmd := exec.Command(e.Veto, args...)
	cmd.Dir = e.Root
	cmd.Env = os.Environ()
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	return stdout.String(), stderr.String(), err
}

func decodeHits(raw string) ([]Hit, error) {
	var hits []Hit
	if err := json.Unmarshal([]byte(raw), &hits); err != nil {
		return nil, fmt.Errorf("search: %w\n%s", err, raw)
	}
	if len(hits) == 0 || hits[0].ID == "" {
		return nil, fmt.Errorf("search returned nothing\n%s", raw)
	}
	return hits, nil
}

func hitByID(hits []Hit, id string) Hit {
	for _, h := range hits {
		if h.ID == id {
			return h
		}
	}
	return Hit{}
}

func hasCall(lines []string, id string) bool {
	for _, line := range lines {
		if strings.HasPrefix(line, id+" ") || line == id {
			return true
		}
	}
	return false
}

func contains(list []string, want string) bool {
	for _, s := range list {
		if s == want {
			return true
		}
	}
	return false
}

func decodeInvoke(raw string) (InvokeOut, error) {
	var out InvokeOut
	if err := json.Unmarshal([]byte(raw), &out); err != nil {
		return InvokeOut{}, fmt.Errorf("invoke: %w\n%s", err, raw)
	}
	return out, nil
}

func prepareEnv(root string) {
	if os.Getenv("DEMO_TOKEN") == "" {
		_ = os.Setenv("DEMO_TOKEN", "demo-token")
	}
	if os.Getenv("VETO_APPROVAL_NONCE_DIR") == "" {
		_ = os.Setenv("VETO_APPROVAL_NONCE_DIR", filepath.Join(root, ".demo", "approvals"))
	}
	if os.Getenv("VETO_TOKEN_DIR") == "" {
		_ = os.Setenv("VETO_TOKEN_DIR", filepath.Join(root, ".demo", "tokens"))
	}
	_ = os.MkdirAll(os.Getenv("VETO_APPROVAL_NONCE_DIR"), 0o700)
	_ = os.MkdirAll(os.Getenv("VETO_TOKEN_DIR"), 0o700)
	_ = os.Unsetenv("VETO_APPROVAL_SECRET")
}

func findRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "app", "veto.yaml")); err == nil {
			if _, err := os.Stat(filepath.Join(dir, "app", "contracts", "orders.yaml")); err == nil {
				return dir, nil
			}
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", errors.New("run this from the veto-demo directory")
		}
		dir = parent
	}
}

func waitReady() error {
	deadline := time.Now().Add(8 * time.Second)
	urls := []string{
		"http://127.0.0.1:18410/healthz",
		"http://127.0.0.1:18411/healthz",
		"http://127.0.0.1:18412/healthz",
	}
	for time.Now().Before(deadline) {
		ready := true
		for _, u := range urls {
			status, err := getStatus(u)
			if err != nil || status != http.StatusOK {
				ready = false
				break
			}
		}
		if ready {
			return nil
		}
		time.Sleep(50 * time.Millisecond)
	}
	return errors.New("the APIs are not listening. Run make mcp in another terminal, or make cli")
}

func mustLog() hitLog {
	resp, err := http.Get("http://127.0.0.1:18410/_demo/log")
	if err != nil {
		return hitLog{}
	}
	defer resp.Body.Close()
	var view hitLog
	_ = json.NewDecoder(resp.Body).Decode(&view)
	return view
}

func saw(view hitLog, service, method, path string) bool {
	for _, c := range view.Calls {
		if c.Service == service && c.Method == method && c.Path == path && c.Status < 400 {
			return true
		}
	}
	return false
}

func getStatus(url string) (int, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return 0, err
	}
	req.Header.Set("Authorization", "Bearer "+os.Getenv("DEMO_TOKEN"))
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return 0, err
	}
	io.Copy(io.Discard, resp.Body)
	resp.Body.Close()
	return resp.StatusCode, nil
}

func indent(s string) string {
	s = strings.TrimRight(s, "\n")
	return "    " + strings.ReplaceAll(s, "\n", "\n    ") + "\n"
}

func fail(name, stdout, stderr string, err error) error {
	return fmt.Errorf("%s: %w\n%s%s", name, err, stdout, stderr)
}
