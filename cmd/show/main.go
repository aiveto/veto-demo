// Command show walks the catalog the way a person, a CLI, and Claude would.
package main

import (
	"bytes"
	"context"
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

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type (
	show struct {
		root    string
		veto    string
		pending string
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

	invokeOut struct {
		Status      string `json:"status"`
		ApprovalID  string `json:"approval_id"`
		OperationID string `json:"operation_id"`
		HTTPStatus  int    `json:"http_status"`
		Body        string `json:"body"`
		Error       string `json:"error"`
	}
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "\n%s\n", err)
		os.Exit(1)
	}
}

func run() error {
	root, err := findRoot()
	if err != nil {
		return err
	}
	prepareEnv(root)
	s := &show{root: root, veto: filepath.Join(root, "bin", "veto")}
	if _, err := os.Stat(s.veto); err != nil {
		return errors.New("bin/veto is missing. Run make demo from the veto-demo directory")
	}
	if err := waitReady(); err != nil {
		return err
	}

	fmt.Print("veto\nThree desks. One catalog. The model proposes. Veto decides.\n")

	if err := s.catalog(); err != nil {
		return err
	}
	session, err := s.connect()
	if err != nil {
		return err
	}
	defer session.Close()
	if err := s.tools(session); err != nil {
		return err
	}
	if err := s.semantics(session); err != nil {
		return err
	}
	if err := s.pack(); err != nil {
		return err
	}
	if err := s.relation(session); err != nil {
		return err
	}
	if err := s.allowedWrite(session); err != nil {
		return err
	}
	if err := s.agentStops(); err != nil {
		return err
	}
	if err := s.wireStops(session); err != nil {
		return err
	}
	if err := s.personApproves(session); err != nil {
		return err
	}
	if err := s.cases(); err != nil {
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

func (s *show) catalog() error {
	fmt.Print("\n1  Catalog\n")
	out, errText, err := s.command("validate", "--config", "veto.yaml")
	if err != nil {
		return fail("validate", out, errText, err)
	}
	if !strings.Contains(out, "operations)") || !strings.Contains(out, "customers.get") || !strings.Contains(out, "invoices.get") {
		return fmt.Errorf("validate did not show the catalog and the joins\n%s", out)
	}
	fmt.Print(indent(out))
	doc, errText, err := s.command("doctor", "--config", "veto.yaml")
	if err != nil {
		return fail("doctor", doc, errText, err)
	}
	doc += errText
	if !strings.Contains(doc, "orders.delete") {
		return fmt.Errorf("doctor did not name orders.delete\n%s", doc)
	}
	fmt.Print("    Doctor loaded the desks. orders.delete is waiting on a person.\n")
	return nil
}

func (s *show) tools(session *mcp.ClientSession) error {
	fmt.Print("\n2  Three tools\n")
	listed, err := session.ListTools(context.Background(), nil)
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
	fmt.Printf("    %s\n", got)
	fmt.Print("    The desks publish more than twenty operations. The agent sees these three.\n")
	return nil
}

func (s *show) semantics(session *mcp.ClientSession) error {
	fmt.Print("\n3  Semantics\n")
	for _, query := range []string{"retire order", "scrap order"} {
		id, err := firstHit(session, query)
		if err != nil {
			return err
		}
		if id != "orders.delete" {
			return fmt.Errorf("search %q ranked %s first", query, id)
		}
		fmt.Printf("    %q  →  orders.delete\n", query)
	}
	text, err := call(session, "capabilities_describe", map[string]any{"operation_id": "orders.delete"})
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

func (s *show) pack() error {
	fmt.Print("\n4  Context pack\n")
	out, errText, err := s.command("pack", "--config", "veto.yaml", "--message", "who placed order 10482")
	if err != nil {
		return fail("pack", out, errText, err)
	}
	if strings.Contains(out, "openapi:") {
		return errors.New("the pack embedded the raw spec")
	}
	fmt.Print("    The pack for \"who placed order 10482\" has no raw spec in it.\n")
	return nil
}

func (s *show) relation(session *mcp.ClientSession) error {
	fmt.Print("\n5  Relations\n")
	before := mustLog()
	text, err := call(session, "capabilities_describe", map[string]any{"operation_id": "orders.get"})
	if err != nil {
		return err
	}
	if !strings.Contains(text, "customers.get") || !strings.Contains(text, "invoices.get") {
		return fmt.Errorf("describe did not name both joins\n%s", text)
	}
	if strings.Contains(text, "warehouse") && strings.Contains(text, "identifies") && strings.Contains(text, "warehouseId identifies") {
		return errors.New("warehouseId was treated as a join")
	}
	fmt.Print("    orders.get  →  customers.get   by Order.customerId\n")
	fmt.Print("    orders.get  →  invoices.get    by Order.invoiceId\n")
	order, err := invoke(session, "orders.get", map[string]any{"id": "10482"}, "")
	if err != nil {
		return err
	}
	if order.Status != "ok" || !strings.Contains(order.Body, "cus_mara") || !strings.Contains(order.Body, "inv_2291") || !strings.Contains(order.Body, "wh_sfo_1") {
		return fmt.Errorf("orders.get body %s", order.Body)
	}
	customer, err := invoke(session, "customers.get", map[string]any{"id": "cus_mara"}, "")
	if err != nil {
		return err
	}
	if customer.Status != "ok" || !strings.Contains(customer.Body, "Mara Ellison") {
		return fmt.Errorf("customers.get %s %s", customer.Status, customer.Body)
	}
	invoice, err := invoke(session, "invoices.get", map[string]any{"id": "inv_2291"}, "")
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

func (s *show) allowedWrite(session *mcp.ClientSession) error {
	fmt.Print("\n6  Cancel reaches the desk\n")
	preview, err := call(session, "capabilities_invoke", map[string]any{
		"operation_id": "orders.cancel",
		"preview":      true,
		"params": map[string]any{
			"id":   "10490",
			"body": map[string]any{"reason": "buyer asked before it shipped"},
		},
	})
	if err != nil {
		return err
	}
	if strings.Contains(preview, `"approval_required":true`) || strings.Contains(preview, `"errors"`) {
		return fmt.Errorf("cancel looked like it needed a person\n%s", preview)
	}
	before := mustLog()
	cancelled, err := invoke(session, "orders.cancel", map[string]any{
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

func (s *show) agentStops() error {
	fmt.Print("\n7  The agent stops\n")
	before := mustLog()
	out, errText, err := s.command("preview", "--config", "veto.yaml", "--operation", "orders.delete", "--param", "id=10482")
	if err != nil {
		return fail("preview delete", out, errText, err)
	}
	if !strings.Contains(out, `"approval_required": true`) && !strings.Contains(out, `"approval_required":true`) {
		return fmt.Errorf("preview did not hold the delete\n%s", out)
	}
	trace, errText, err := s.command("replay", "--config", "veto.yaml", "--message", "Delete order 10482")
	if err != nil {
		return fail("replay", trace, errText, err)
	}
	if !strings.Contains(trace, "orders.delete") || !strings.Contains(trace, "confirmation_required") {
		return fmt.Errorf("replay did not stop on confirmation\n%s", trace)
	}
	if mustLog().Deletes != before.Deletes {
		return errors.New("the scripted agent reached the desk")
	}
	fmt.Print("    \"Delete order 10482\" selected orders.delete.\n")
	fmt.Print("    The trace says confirmation_required. The desk saw no DELETE.\n")
	return nil
}

func (s *show) wireStops(session *mcp.ClientSession) error {
	fmt.Print("\n8  MCP stops the same way\n")
	before := mustLog()
	held, err := invoke(session, "orders.delete", map[string]any{"id": "10482"}, "")
	if err != nil {
		return err
	}
	if held.Status != "confirmation_required" || held.ApprovalID == "" {
		return fmt.Errorf("invoke status %s id %q", held.Status, held.ApprovalID)
	}
	if mustLog().Deletes != before.Deletes {
		return errors.New("the pending id sent HTTP")
	}
	fmt.Printf("    capabilities_invoke orders.delete  →  confirmation_required\n")
	fmt.Printf("    pending %s\n", held.ApprovalID)
	s.pending = held.ApprovalID
	return nil
}

func (s *show) personApproves(session *mcp.ClientSession) error {
	fmt.Print("\n9  A person says yes\n")
	if s.pending == "" {
		return errors.New("no pending id")
	}
	approved, errText, err := s.command("approve", s.pending)
	if err != nil {
		return fail("approve", approved, errText, err)
	}
	approved = strings.TrimSpace(approved)
	if approved == "" || approved == s.pending {
		return fmt.Errorf("approve returned %q", approved)
	}
	fmt.Printf("    veto approve  →  %s\n", approved)
	sent, err := invoke(session, "orders.delete", map[string]any{"id": "10482"}, approved)
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
	again, err := invoke(session, "orders.delete", map[string]any{"id": "10482"}, approved)
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

func (s *show) cases() error {
	fmt.Print("\n10  Eval\n")
	before := mustLog().Deletes
	out, errText, err := s.command("eval", "--config", "veto.yaml", "--case", "cases")
	if err != nil {
		return fail("eval", out, errText, err)
	}
	if mustLog().Deletes != before {
		return errors.New("eval reached the desk")
	}
	fmt.Print("    veto eval  delete still requires confirmation, and the pack still omits the spec.\n")
	return nil
}

func (s *show) command(args ...string) (string, string, error) {
	cmd := exec.Command(s.veto, args...)
	cmd.Dir = s.root
	cmd.Env = os.Environ()
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	return stdout.String(), stderr.String(), err
}

func (s *show) connect() (*mcp.ClientSession, error) {
	cmd := exec.Command(s.veto, "serve", "--config", filepath.Join(s.root, "veto.yaml"), "--stdio")
	cmd.Dir = s.root
	cmd.Env = os.Environ()
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	client := mcp.NewClient(&mcp.Implementation{Name: "veto-demo", Version: "0.1.0"}, nil)
	session, err := client.Connect(context.Background(), &mcp.CommandTransport{Command: cmd}, nil)
	if err != nil {
		return nil, fmt.Errorf("mcp connect: %w\n%s", err, stderr.String())
	}
	return session, nil
}

func firstHit(session *mcp.ClientSession, query string) (string, error) {
	text, err := call(session, "capabilities_search", map[string]any{"query": query})
	if err != nil {
		return "", err
	}
	var hits []struct {
		Operation struct {
			ID string `json:"ID"`
		} `json:"Operation"`
	}
	if err := json.Unmarshal([]byte(text), &hits); err != nil {
		return "", fmt.Errorf("search %q: %w\n%s", query, err, text)
	}
	if len(hits) == 0 || hits[0].Operation.ID == "" {
		return "", fmt.Errorf("search %q returned nothing\n%s", query, text)
	}
	return hits[0].Operation.ID, nil
}

func invoke(session *mcp.ClientSession, operation string, params map[string]any, approval string) (invokeOut, error) {
	args := map[string]any{"operation_id": operation, "params": params}
	if approval != "" {
		args["approval_id"] = approval
	}
	text, err := call(session, "capabilities_invoke", args)
	if err != nil {
		return invokeOut{}, err
	}
	var out invokeOut
	if err := json.Unmarshal([]byte(text), &out); err != nil {
		return invokeOut{}, fmt.Errorf("invoke %s: %w\n%s", operation, err, text)
	}
	return out, nil
}

func call(session *mcp.ClientSession, name string, args map[string]any) (string, error) {
	res, err := session.CallTool(context.Background(), &mcp.CallToolParams{Name: name, Arguments: args})
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
		if _, err := os.Stat(filepath.Join(dir, "veto.yaml")); err == nil {
			if _, err := os.Stat(filepath.Join(dir, "contracts", "orders.yaml")); err == nil {
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
	return errors.New("the desks are not listening. Run make up in another terminal, or make demo")
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
