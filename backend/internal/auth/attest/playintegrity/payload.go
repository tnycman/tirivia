package playintegrity

import pi "google.golang.org/api/playintegrity/v1"

func fromPayload(p *pi.TokenPayloadExternal) Verdict {
	var v Verdict
	if r := p.RequestDetails; r != nil {
		v.HasRequest, v.RequestPackageName, v.RequestHash, v.TimestampMillis = true, r.RequestPackageName, r.RequestHash, r.TimestampMillis
	}
	if a := p.AppIntegrity; a != nil {
		v.AppRecognition, v.PackageName, v.CertDigests = a.AppRecognitionVerdict, a.PackageName, a.CertificateSha256Digest
	}
	if a := p.AccountDetails; a != nil {
		v.Licensing = a.AppLicensingVerdict
	}
	if d := p.DeviceIntegrity; d != nil {
		v.DeviceLabels = d.DeviceRecognitionVerdict
		if d.DeviceAttributes != nil {
			v.SDKVersion = d.DeviceAttributes.SdkVersion
		}
		if d.RecentDeviceActivity != nil {
			v.ActivityLevel = d.RecentDeviceActivity.DeviceActivityLevel
		}
	}
	if e := p.EnvironmentDetails; e != nil {
		v.PlayProtect = e.PlayProtectVerdict
		if e.AppAccessRiskVerdict != nil {
			v.AppsDetected = e.AppAccessRiskVerdict.AppsDetected
		}
	}
	v.Testing = p.TestingDetails != nil && p.TestingDetails.IsTestingResponse
	return v
}
