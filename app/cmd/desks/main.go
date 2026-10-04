package main

import (
	"context"
	"encoding/json"
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
	quiet := false
	for _, a := range os.Args[1:] {
		if a == "--quiet" {
			quiet = true
		}
	}
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
	if !quiet {
		fmt.Print(banner())
	}

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
	body, err := json.MarshalIndent(mcpConfig{
		MCPServers: map[string]mcpServer{
			"veto": {
				Command: filepath.Join(repo, "bin", "veto"),
				Args:    []string{"serve", "--config", filepath.Join(repo, "app", "veto.yaml"), "--stdio"},
				Env: map[string]string{
					"DEMO_TOKEN":              desks.Token,
					"VETO_APPROVAL_NONCE_DIR": filepath.Join(repo, ".demo", "approvals"),
					"VETO_TOKEN_DIR":          filepath.Join(repo, ".demo", "tokens"),
				},
			},
		},
	}, "", "  ")
	if err != nil {
		return err.Error() + "\n"
	}
	return fmt.Sprintf(`Harbor's APIs are listening.

  orders     http://%s
  customers  http://%s
  invoices   http://%s

Paste this into Claude, Cursor, or ChatGPT. Paste app/prompts/claude.md as the instruction.

veto serve --json serves JSON lines for skills and scripts. app/prompts/skill.md is that instruction. make cli plays search, describe, and invoke.

%s
`, desks.OrdersAddr, desks.CustomersAddr, desks.InvoicesAddr, body)
}

type (
	mcpConfig struct {
		MCPServers map[string]mcpServer `json:"mcpServers"`
	}
	mcpServer struct {
		Command string            `json:"command"`
		Args    []string          `json:"args"`
		Env     map[string]string `json:"env"`
	}
)

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
