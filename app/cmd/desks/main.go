package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/aiveto/veto-demo/app/desks"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "desks: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	desk := desks.New()
	servers := []*http.Server{
		{Addr: desks.OrdersAddr, Handler: desk.Handler("orders")},
		{Addr: desks.CustomersAddr, Handler: desk.Handler("customers")},
		{Addr: desks.InvoicesAddr, Handler: desk.Handler("invoices")},
	}
	errc := make(chan error, len(servers))
	for _, srv := range servers {
		go func() {
			if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
				errc <- err
			}
		}()
	}
	if err := waitReady(); err != nil {
		return err
	}
	fmt.Print(banner())

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	select {
	case <-ctx.Done():
	case err := <-errc:
		return err
	}
	shut, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	for _, srv := range servers {
		_ = srv.Shutdown(shut)
	}
	fmt.Fprintln(os.Stderr, "desks stopped")
	return nil
}

func waitReady() error {
	deadline := time.Now().Add(3 * time.Second)
	urls := []string{
		"http://" + desks.OrdersAddr + "/healthz",
		"http://" + desks.CustomersAddr + "/healthz",
		"http://" + desks.InvoicesAddr + "/healthz",
	}
	for time.Now().Before(deadline) {
		ready := true
		for _, u := range urls {
			resp, err := http.Get(u)
			if err != nil {
				ready = false
				break
			}
			resp.Body.Close()
			if resp.StatusCode != http.StatusOK {
				ready = false
				break
			}
		}
		if ready {
			return nil
		}
		time.Sleep(30 * time.Millisecond)
	}
	return errors.New("desks did not listen")
}

func banner() string {
	repo := repoRoot()
	veto := filepath.Join(repo, "bin", "veto")
	cfg := filepath.Join(repo, "app", "veto.yaml")
	approvals := filepath.Join(repo, ".demo", "approvals")
	tokens := filepath.Join(repo, ".demo", "tokens")
	return fmt.Sprintf(`Harbor's APIs are listening. Veto is in front of them.

  orders     http://%s
  customers  http://%s
  invoices   http://%s

Order 10482 is Mara Ellison's wool coat. customerId is cus_mara. invoiceId is inv_2291.
warehouseId is wh_sfo_1 and stays on the order.

Claude, Cursor, and ChatGPT use the same server:

  %s serve --config %s --stdio

Env for that process:

  DEMO_TOKEN=%s
  VETO_APPROVAL_NONCE_DIR=%s
  VETO_TOKEN_DIR=%s

Paste app/prompts/claude.md as the instruction.
Relationships: app/relations.yaml. Semantics: app/semantics.yaml.
Another terminal: make show. The CLI: make cli.

`, desks.OrdersAddr, desks.CustomersAddr, desks.InvoicesAddr, veto, cfg, desks.Token, approvals, tokens)
}

func repoRoot() string {
	dir, err := os.Getwd()
	if err != nil {
		return "."
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "app", "veto.yaml")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return dir
		}
		dir = parent
	}
}
