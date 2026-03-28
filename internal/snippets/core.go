package snippets

// CoreSnippets returns snippets for the core axon modules:
// axon, axon-talk, axon-loop, axon-tool.
func CoreSnippets() []Snippet {
	return []Snippet{
		axonSnippet(),
		axonTalkSnippet(),
		axonToolSnippet(),
		axonLoopSnippet(),
	}
}

func axonSnippet() Snippet {
	return Snippet{
		Module: "axon",
		Imports: []Import{
			{Path: "log/slog"},
			{Path: "net/http"},
			{Path: "os"},
			{Path: "github.com/benaskins/axon"},
		},
		Requires: []string{"github.com/benaskins/axon"},
		Setup: `	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})`,
		Helpers: `func serve(port string, mux *http.ServeMux) {
	slog.Info("serving", "port", port)
	axon.ListenAndServe(port, axon.StandardMiddleware(mux))
}`,
	}
}

func axonTalkSnippet() Snippet {
	return Snippet{
		Module: "axon-talk",
		Imports: []Import{
			{Path: "fmt"},
			{Path: "log/slog"},
			{Path: "os"},
			{Path: "strings"},
			{Path: "github.com/benaskins/axon-loop", Alias: "loop"},
			{Path: "github.com/benaskins/axon-talk/anthropic"},
		},
		Requires: []string{
			"github.com/benaskins/axon-talk",
			"github.com/benaskins/axon-loop",
		},
		Setup: `	client := selectLLMClient()
	if client == nil {
		fmt.Fprintln(os.Stderr, "no LLM provider available — set ANTHROPIC_API_KEY or CLOUDFLARE_AI_GATEWAY_TOKEN")
		os.Exit(1)
	}`,
		Helpers: `func selectLLMClient() loop.LLMClient {
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
}`,
	}
}

func axonToolSnippet() Snippet {
	return Snippet{
		Module: "axon-tool",
		Imports: []Import{
			{Path: "github.com/benaskins/axon-tool", Alias: "tool"},
		},
		Requires: []string{"github.com/benaskins/axon-tool"},
		Setup:    `	allTools := map[string]tool.ToolDef{}`,
	}
}

func axonLoopSnippet() Snippet {
	return Snippet{
		Module: "axon-loop",
		Imports: []Import{
			{Path: "github.com/benaskins/axon-loop", Alias: "loop"},
		},
		Requires: []string{"github.com/benaskins/axon-loop"},
		Deps:     []string{"axon-talk", "axon-tool"},
		Setup: `	_ = loop.RunConfig{
		Client: client,
		Tools:  allTools,
	}`,
	}
}
