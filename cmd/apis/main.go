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

	"github.com/aiveto/veto-demo/apis"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "desks: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	desk := apis.New()
	servers := []*http.Server{
		{Addr: apis.OrdersAddr, Handler: desk.Handler("orders")},
		{Addr: apis.CustomersAddr, Handler: desk.Handler("customers")},
		{Addr: apis.InvoicesAddr, Handler: desk.Handler("invoices")},
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
		"http://" + apis.OrdersAddr + "/healthz",
		"http://" + apis.CustomersAddr + "/healthz",
		"http://" + apis.InvoicesAddr + "/healthz",
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
	dir, err := os.Getwd()
	if err != nil {
		dir = "."
	}
	veto := filepath.Join(dir, "bin", "veto")
	cfg := filepath.Join(dir, "veto.yaml")
	approvals := filepath.Join(dir, ".demo", "approvals")
	tokens := filepath.Join(dir, ".demo", "tokens")
	return fmt.Sprintf(`The desks are listening. Veto is the catalog.

  orders     http://%s
  customers  http://%s
  invoices   http://%s

Order 10482 is Mara Ellison's wool coat. customerId is cus_mara. invoiceId is inv_2291.
warehouseId is wh_sfo_1 and is not a join.

Claude or ChatGPT, same three tools:

  %s serve --config %s --stdio

Env for that process:

  DEMO_TOKEN=%s
  VETO_APPROVAL_NONCE_DIR=%s
  VETO_TOKEN_DIR=%s

The instruction to paste is prompts/claude.md.
Another terminal: make show.

`, apis.OrdersAddr, apis.CustomersAddr, apis.InvoicesAddr, veto, cfg, apis.Token, approvals, tokens)
}
