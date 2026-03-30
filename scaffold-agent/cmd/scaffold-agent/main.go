package main

import (
	"fmt"
	"log/slog"
	"os"
	"strings"

	loop "github.com/benaskins/axon-loop"
	"github.com/benaskins/axon-talk/anthropic"
	tool "github.com/benaskins/axon-tool"
)

func main() {
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

	// TODO: wire scaffold-agent business logic here
	fmt.Fprintln(os.Stderr, "scaffold-agent:", "ready")
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
