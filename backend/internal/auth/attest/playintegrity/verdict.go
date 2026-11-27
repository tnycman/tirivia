// Package playintegrity verifies Play Integrity standard-request tokens (developer.android.com/google/play/integrity).
package playintegrity

// Verdict values used by the policy (verdicts page, 2025 semantics).
const (
	AppPlayRecognized    = "PLAY_RECOGNIZED"
	Licensed             = "LICENSED"
	MeetsDeviceIntegrity = "MEETS_DEVICE_INTEGRITY"
	MeetsStrongIntegrity = "MEETS_STRONG_INTEGRITY"
	PlayProtectNoIssues  = "NO_ISSUES"
)

// Verdict is the decoded token payload the policy reads.
type Verdict struct {
	HasRequest         bool
	RequestPackageName string
	RequestHash        string
	TimestampMillis    int64
	AppRecognition     string
	PackageName        string
	CertDigests        []string
	Licensing          string
	DeviceLabels       []string
	SDKVersion         int64
	ActivityLevel      string
	PlayProtect        string
	AppsDetected       []string
	Testing            bool
}
