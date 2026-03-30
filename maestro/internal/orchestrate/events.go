package orchestrate

// Event type constants for the orchestration audit trail.
const (
	EventStepStarted        = "StepStarted"
	EventAgentInvoked       = "AgentInvoked"
	EventAgentSucceeded     = "AgentSucceeded"
	EventAgentFailed        = "AgentFailed"
	EventVerificationRun    = "VerificationRun"
	EventVerificationPassed = "VerificationPassed"
	EventVerificationFailed = "VerificationFailed"
	EventReviewRun          = "ReviewRun"
	EventReviewPassed       = "ReviewPassed"
	EventReviewFailed       = "ReviewFailed"
	EventReviewErrored      = "ReviewErrored"
	EventCommitSucceeded    = "CommitSucceeded"
	EventCommitSkipped      = "CommitSkipped"
	EventCommitFailed       = "CommitFailed"
	EventRetryAttempt       = "RetryAttempt"
	EventStepCompleted      = "StepCompleted"
	EventStepFailed         = "StepFailed"
)

// StepStartedData is emitted when a step begins its first attempt.
type StepStartedData struct {
	StepNumber int    `json:"step_number"`
	StepTitle  string `json:"step_title"`
}

// AgentInvokedData is emitted before delegating to the coding agent.
type AgentInvokedData struct {
	StepNumber  int  `json:"step_number"`
	Attempt     int  `json:"attempt"`
	HasFeedback bool `json:"has_feedback"`
}

// AgentSucceededData is emitted when the coding agent completes without error.
type AgentSucceededData struct {
	StepNumber int `json:"step_number"`
	Attempt    int `json:"attempt"`
}

// AgentFailedData is emitted when the coding agent returns an error.
type AgentFailedData struct {
	StepNumber int    `json:"step_number"`
	Attempt    int    `json:"attempt"`
	Error      string `json:"error"`
}

// VerificationRunData is emitted before running the verification command.
type VerificationRunData struct {
	StepNumber int    `json:"step_number"`
	Attempt    int    `json:"attempt"`
	Command    string `json:"command"`
}

// VerificationPassedData is emitted when verification exits successfully.
type VerificationPassedData struct {
	StepNumber int `json:"step_number"`
	Attempt    int `json:"attempt"`
}

// VerificationFailedData is emitted when verification exits with an error.
type VerificationFailedData struct {
	StepNumber int `json:"step_number"`
	Attempt    int `json:"attempt"`
}

// ReviewRunData is emitted before invoking the semantic reviewer.
type ReviewRunData struct {
	StepNumber int `json:"step_number"`
	Attempt    int `json:"attempt"`
}

// ReviewPassedData is emitted when the reviewer approves the implementation.
type ReviewPassedData struct {
	StepNumber int    `json:"step_number"`
	Attempt    int    `json:"attempt"`
	Reason     string `json:"reason"`
}

// ReviewFailedData is emitted when the reviewer rejects the implementation.
type ReviewFailedData struct {
	StepNumber int    `json:"step_number"`
	Attempt    int    `json:"attempt"`
	Reason     string `json:"reason"`
}

// ReviewErroredData is emitted when the reviewer returns an error (non-fatal).
type ReviewErroredData struct {
	StepNumber int    `json:"step_number"`
	Attempt    int    `json:"attempt"`
	Error      string `json:"error"`
}

// CommitSucceededData is emitted when a git commit is created successfully.
type CommitSucceededData struct {
	StepNumber int    `json:"step_number"`
	Message    string `json:"message"`
}

// CommitSkippedData is emitted when there are no changes to commit.
type CommitSkippedData struct {
	StepNumber int    `json:"step_number"`
	Message    string `json:"message"`
}

// CommitFailedData is emitted when the git commit operation fails.
type CommitFailedData struct {
	StepNumber int    `json:"step_number"`
	Message    string `json:"message"`
	Error      string `json:"error"`
}

// RetryAttemptData is emitted at the start of each retry (attempt > 1).
type RetryAttemptData struct {
	StepNumber int `json:"step_number"`
	Attempt    int `json:"attempt"`
}

// StepCompletedData is emitted when a step finishes successfully.
type StepCompletedData struct {
	StepNumber int    `json:"step_number"`
	StepTitle  string `json:"step_title"`
}

// StepFailedData is emitted when a step exhausts all retry attempts.
type StepFailedData struct {
	StepNumber int    `json:"step_number"`
	StepTitle  string `json:"step_title"`
	Attempts   int    `json:"attempts"`
	LastError  string `json:"last_error"`
}
