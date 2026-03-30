package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strings"

	"github.com/benaskins/axon"
	fact "github.com/benaskins/axon-fact"
	loop "github.com/benaskins/axon-loop"
	memo "github.com/benaskins/axon-memo"
	"github.com/benaskins/axon-talk/anthropic"
	task "github.com/benaskins/axon-task"
	tool "github.com/benaskins/axon-tool"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
	  port = "8080"
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
	  w.WriteHeader(http.StatusOK)
	})

	client := selectLLMClient()
	if client == nil {
	  fmt.Fprintln(os.Stderr, "no LLM provider available — set ANTHROPIC_API_KEY or CLOUDFLARE_AI_GATEWAY_TOKEN")
	  os.Exit(1)
	}

	allTools := map[string]tool.ToolDef{}

	_ = loop.RunConfig{
	  Client: client,
	  Tools:  allTools,
	}

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
	  fmt.Fprintln(os.Stderr, "DATABASE_URL must be set")
	  os.Exit(1)
	}
	db := axon.MustOpenDB(dsn, "app")
	events := fact.NewPostgresStore(db,
	  // TODO: register domain projectors here
	)
	if err := events.Replay(context.Background()); err != nil {
	  slog.Error("replay events", "error", err)
	  os.Exit(1)
	}

	taskStore := task.NewPostgresStore(db)
	_ = task.NewExecutor("agent", ".", "sonnet", taskStore)
	// TODO: register workers with executor.RegisterWorker(name, worker)

	_ = memo.NewConversationClient(envOrDefault("CHAT_SERVICE_URL", "http://localhost:8080"))
	// TODO: wire memo extractor, retriever, and consolidator

	authURL := os.Getenv("AUTH_URL")
	if authURL == "" {
	  authURL = "http://localhost:9000"
	}
	authClient := axon.NewAuthClientPlain(authURL)
	defer authClient.Close()
	_ = axon.RequireAuth(authClient)

	// TODO: wire coding-agent business logic here
	fmt.Fprintln(os.Stderr, "coding-agent:", "ready")
}

func serve(port string, mux *http.ServeMux) {
  slog.Info("serving", "port", port)
  axon.ListenAndServe(port, axon.StandardMiddleware(mux))
}

func selectLLMClient() loop.LLMClient {
  apiKey := os.Getenv("ANTHROPIC_API_KEY")
  accountID := os.Getenv("CLOUDFLARE_ACCOUNT_ID")
  gwToken := os.Getenv("CLOUDFLARE_AI_GATEWAY_TOKEN")

  if accountID != "" && gwToken != "" {
    gateway := envOrDefault("CLOUDFLARE_GATEWAY", "axon-gate")
    baseURL := "https://gateway.ai.cloudflare.com/v1/" + strings.TrimSpace(accountID) + "/" + gateway + "/anthropic"
    slog.Info("using Anthropic via Cloudflare AI Gateway", "gateway", gateway)
    return anthropic.NewClient(baseURL, apiKey, anthropic.WithGatewayToken(gwToken))
  }

  if apiKey != "" {
    slog.Info("using Anthropic API directly")
    return anthropic.NewClient("https://api.anthropic.com", apiKey)
  }

  return nil
}

func envOrDefault(key, fallback string) string {
  if v := os.Getenv(key); v != "" {
    return v
  }
  return fallback
}
