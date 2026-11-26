package playintegrity

import (
	"slices"
	"strconv"
)

// Signals lists risk facts for the risk score (AC-7); they never decide eligibility on their own.
func Signals(v Verdict) []string {
	var s []string
	if !slices.Contains(v.DeviceLabels, MeetsDeviceIntegrity) {
		s = append(s, "pi.device_integrity=missing")
	}
	if slices.Contains(v.DeviceLabels, MeetsStrongIntegrity) {
		s = append(s, "pi.strong_integrity=present")
	}
	if v.SDKVersion > 0 {
		s = append(s, "pi.sdk="+strconv.FormatInt(v.SDKVersion, 10))
	}
	if v.ActivityLevel != "" && v.ActivityLevel != "LEVEL_1" && v.ActivityLevel != "UNEVALUATED" {
		s = append(s, "pi.activity="+v.ActivityLevel)
	}
	if v.PlayProtect != "" && v.PlayProtect != PlayProtectNoIssues {
		s = append(s, "pi.play_protect="+v.PlayProtect)
	}
	for _, a := range v.AppsDetected {
		s = append(s, "pi.apps="+a)
	}
	if v.Testing {
		s = append(s, "pi.testing_response=true")
	}
	return s
}
