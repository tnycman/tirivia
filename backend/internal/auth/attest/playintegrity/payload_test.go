package playintegrity

import (
	"encoding/json"
	"slices"
	"testing"

	pi "google.golang.org/api/playintegrity/v1"
)

// Field names follow the verdicts page example; timestampMillis is a JSON string there.
const samplePayload = `{
  "requestDetails": {"requestPackageName": "al.tirivia.app", "requestHash": "h", "timestampMillis": "1800000000000"},
  "appIntegrity": {"appRecognitionVerdict": "PLAY_RECOGNIZED", "packageName": "al.tirivia.app",
    "certificateSha256Digest": ["cert-a"], "versionCode": "42"},
  "deviceIntegrity": {"deviceRecognitionVerdict": ["MEETS_DEVICE_INTEGRITY", "MEETS_STRONG_INTEGRITY"],
    "deviceAttributes": {"sdkVersion": 34}, "recentDeviceActivity": {"deviceActivityLevel": "LEVEL_3"}},
  "accountDetails": {"appLicensingVerdict": "LICENSED"},
  "environmentDetails": {"playProtectVerdict": "MEDIUM_RISK", "appAccessRiskVerdict": {"appsDetected": ["KNOWN_INSTALLED", "UNKNOWN_CAPTURING"]}}
}`

func TestFromPayload(t *testing.T) {
	var p pi.TokenPayloadExternal
	if err := json.Unmarshal([]byte(samplePayload), &p); err != nil {
		t.Fatal(err)
	}
	v := fromPayload(&p)
	if !v.HasRequest || v.TimestampMillis != 1_800_000_000_000 || v.RequestHash != "h" || v.SDKVersion != 34 ||
		v.ActivityLevel != "LEVEL_3" || v.Licensing != Licensed || v.PlayProtect != "MEDIUM_RISK" || len(v.AppsDetected) != 2 {
		t.Fatalf("mapping: %+v", v)
	}
	if !PrizeEligible(v, expect) {
		t.Fatal("sample must be prize eligible")
	}
	s := Signals(v)
	for _, want := range []string{"pi.strong_integrity=present", "pi.sdk=34", "pi.activity=LEVEL_3", "pi.play_protect=MEDIUM_RISK", "pi.apps=UNKNOWN_CAPTURING"} {
		if !slices.Contains(s, want) {
			t.Errorf("missing signal %s in %v", want, s)
		}
	}
}
