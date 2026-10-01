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

	"github.com/aiveto/veto-demo/harbor"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "harbor: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	desk := harbor.New()
	servers := []*http.Server{
		{Addr: harbor.OrdersAddr, Handler: desk.Handler("orders")},
		{Addr: harbor.CustomersAddr, Handler: desk.Handler("customers")},
		{Addr: harbor.InvoicesAddr, Handler: desk.Handler("invoices")},
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
	fmt.Fprintln(os.Stderr, "harbor stopped")
	return nil
}

func waitReady() error {
	deadline := time.Now().Add(3 * time.Second)
	urls := []string{
		"http://" + harbor.OrdersAddr + "/healthz",
		"http://" + harbor.CustomersAddr + "/healthz",
		"http://" + harbor.InvoicesAddr + "/healthz",
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
	approvals := filepath.Join(dir, ".harbor", "approvals")
	tokens := filepath.Join(dir, ".harbor", "tokens")
	return fmt.Sprintf(`Harbor is listening.

  orders     http://%s
  customers  http://%s
  invoices   http://%s

Order 10482 is Mara Ellison's wool coat. customerId is cus_mara. invoiceId is inv_2291.
warehouseId is wh_sfo_1 and is not a join.

Claude or ChatGPT, same three tools:

  %s serve --config %s --stdio

Env for that process:

  HARBOR_TOKEN=%s
  VETO_APPROVAL_NONCE_DIR=%s
  VETO_TOKEN_DIR=%s

The instruction to paste is prompts/claude.md.
Another terminal: make show.

`, harbor.OrdersAddr, harbor.CustomersAddr, harbor.InvoicesAddr, veto, cfg, harbor.Token, approvals, tokens)
}
