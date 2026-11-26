# App Attest test vectors

| File | Source | Notes |
|---|---|---|
| `apple_sample_attestation.b64` | Apple, [Attestation object validation guide](https://developer.apple.com/documentation/devicecheck/attestation-object-validation-guide), fetched 2026-11-25 | Production. Team `1234567890`, bundle `com.example.myapp`, key ID `zgSY9YSD+7TaDXssY6WlOPVS1K3Lmk+pFhlcSWE+ZV0=`. The leaf is valid 2026-04-20 to 2026-04-23. The sample's nonce uses the raw challenge `example_server_challenge` as clientDataHash, despite the guide's prose |
| `veehaitch_ios14_4.json` | [veehaitch/devicecheck-appattest](https://github.com/veehaitch/devicecheck-appattest), `src/test/resources/ios-14.4.yaml`. Copyright Vincent Haupert, Apache License 2.0 | A real iOS 14.4 development device: one attestation plus one assertion (counter 1). It has no extensions |

Both chains verify against the pinned Apple App Attestation Root CA, which proves the bytes are genuine.
