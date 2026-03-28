package snippets

// CapabilitySnippets returns snippets for the capability axon modules:
// axon-fact, axon-task, axon-auth, axon-memo.
func CapabilitySnippets() []Snippet {
	return []Snippet{
		axonFactSnippet(),
		axonTaskSnippet(),
		axonAuthSnippet(),
		axonMemoSnippet(),
	}
}

func axonFactSnippet() Snippet {
	return Snippet{
		Module: "axon-fact",
		Imports: []Import{
			{Path: "context"},
			{Path: "fmt"},
			{Path: "log/slog"},
			{Path: "os"},
			{Path: "github.com/benaskins/axon"},
			{Path: "github.com/benaskins/axon-fact", Alias: "fact"},
		},
		Requires: []string{"github.com/benaskins/axon", "github.com/benaskins/axon-fact"},
		Setup: `	dsn := os.Getenv("DATABASE_URL")
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
	}`,
	}
}

func axonTaskSnippet() Snippet {
	return Snippet{
		Module: "axon-task",
		Imports: []Import{
			{Path: "github.com/benaskins/axon-task", Alias: "task"},
		},
		Requires: []string{"github.com/benaskins/axon-task"},
		Deps:     []string{"axon-fact"},
		Setup: `	taskStore := task.NewPostgresStore(db)
	_ = task.NewExecutor("agent", ".", "sonnet", taskStore)
	// TODO: register workers with executor.RegisterWorker(name, worker)`,
	}
}

func axonAuthSnippet() Snippet {
	return Snippet{
		Module: "axon-auth",
		Imports: []Import{
			{Path: "os"},
			{Path: "github.com/benaskins/axon"},
		},
		Requires: []string{"github.com/benaskins/axon"},
		Setup: `	authURL := os.Getenv("AUTH_URL")
	if authURL == "" {
		authURL = "http://localhost:9000"
	}
	authClient := axon.NewAuthClientPlain(authURL)
	defer authClient.Close()
	_ = axon.RequireAuth(authClient)`,
	}
}

func axonMemoSnippet() Snippet {
	return Snippet{
		Module: "axon-memo",
		Imports: []Import{
			{Path: "github.com/benaskins/axon-memo", Alias: "memo"},
		},
		Requires: []string{"github.com/benaskins/axon-memo"},
		Deps:     []string{"axon"},
		Setup: `	_ = memo.NewConversationClient(envOrDefault("CHAT_SERVICE_URL", "http://localhost:8080"))
	// TODO: wire memo extractor, retriever, and consolidator`,
	}
}
