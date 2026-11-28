# 16. Android sign-in: API code generation and the DPoP device key

PLAN Phase 1, Android. The server side is built (`AuthService`): challenge, then `Attest` (DPoP proof), then `StartOtp`, then `VerifyOtp` (DPoP proof) returns a session. Proofs must be ES256 with `typ: dpop+jwt` (`backend/internal/auth/dpop/types.go`). The Android client has no generated API code yet, and `protocol/buf.gen.yaml` lists the Kotlin plugins unpinned. Sources read 2026-11-28 (verbatim unless marked).

## Generating the Kotlin API code

| Source | Text |
|---|---|
| [Protobuf cross-version guarantee](https://protobuf.dev/support/cross-version-runtime-guarantee/) | "gencode should never be paired with a runtime that predates the `protoc` and plugin that was used to generate those bindings." Our runtime is `protobuf-kotlin-lite` 4.36.2 |
| [Maven: protoc](https://repo1.maven.org/maven2/com/google/protobuf/protoc/maven-metadata.xml) | 4.36.2 present and the highest stable 4.x (4.37.0 is RC only) |
| [Maven: protoc-gen-connect-kotlin](https://repo1.maven.org/maven2/com/connectrpc/protoc-gen-connect-kotlin/maven-metadata.xml) | release 0.10.0 (matches our connect-kotlin pin) |
| [protobuf-gradle-plugin](https://github.com/google/protobuf-gradle-plugin) | "The latest version is `0.10.1`"; 0.10.0 notes: "Make Protobuf Gradle Plugin compatible with Android Gradle Plugin 9"; "lite code generation is built into protoc's "java" output"; Android uses `android.sourceSets { proto }` |
| [connect-kotlin getting started](https://connectrpc.com/docs/kotlin/getting-started/) | `ProtocolClient(httpClient = ConnectOkHttpClient(okHttp), ProtocolClientConfig(host, serializationStrategy = GoogleJavaLiteProtobufStrategy(), networkProtocol = NetworkProtocol.CONNECT))`. Its buf example pins `connectrpc/kotlin:v0.8.0`, behind the 0.10.0 release |
| [connect-kotlin interceptors](https://connectrpc.com/docs/kotlin/interceptors/) | "Each interceptor is instantiated **once per request**"; headers are set with `request.clone(headers = ...)` |
| BSR plugin pages | Rendered client-side; no plugin versions readable (UNVERIFIED) |

UNVERIFIED: Kotlin-DSL wiring of the protobuf Gradle plugin for proto files outside the module, and the connect-kotlin plugin's Maven classifier. The first CI run confirms both.

## The device key and DPoP proofs

| Source | Text |
|---|---|
| [Android Keystore](https://developer.android.com/privacy-and-security/keystore) | `KeyPairGenerator.getInstance(KeyProperties.KEY_ALGORITHM_EC, "AndroidKeyStore")` with `PURPOSE_SIGN`; signs with `Signature.getInstance("SHA256withECDSA")`. StrongBox: "If the StrongBox KeyMint does not support the specified algorithm or key size, the framework will throw a StrongBoxUnavailableException. If this occurs, generate or import the key without calling `setIsStrongBoxBacked(true)`." StrongBox needs API 28+ |
| [Android Signature](https://developer.android.com/reference/java/security/Signature) | lists `SHA256withECDSA` (API 11+); no `...inP1363Format` variant |
| [RFC 3279 §2.2.3](https://www.rfc-editor.org/rfc/rfc3279.html#section-2.2.3) | ECDSA r and s "MUST be ASN.1 encoded": `Ecdsa-Sig-Value ::= SEQUENCE { r INTEGER, s INTEGER }` |
| [RFC 7518 §3.4](https://www.rfc-editor.org/rfc/rfc7518.html#section-3.4) | JWS: R and S "in big-endian order, with each array being be 32 octets long"; "MUST NOT be shortened"; "The JWS Signature value MUST be a 64-octet sequence." |
| [RFC 9449 §4.2](https://www.rfc-editor.org/rfc/rfc9449.html#section-4.2) | header `typ` `dpop+jwt`, `alg` (not `none`, not MAC), `jwk` that "MUST NOT contain a private key"; claims `jti` (≥ 96 bits random or UUIDv4), `htm`, `htu` "without query and fragment parts", `iat`; `ath` with an access token; `nonce` when the server sends `DPoP-Nonce` |
| [RFC 7638 §3.2](https://www.rfc-editor.org/rfc/rfc7638.html#section-3.2) | EC thumbprint members in order: `crv`, `kty`, `x`, `y` |
| RFC 7518 §6.2.1 | EC JWK `kty`/`crv`/`x`/`y`, each coordinate the full curve size (summary only: UNVERIFIED verbatim) |

## Settled by the sources

1. The Kotlin code is generated with protoc 4.36.2, which equals the runtime, so the guarantee holds exactly.
2. Android's `SHA256withECDSA` gives the ASN.1 form, and a JWS needs the 64-byte R‖S form. The proof code converts DER to R‖S, left-padding each value to 32 bytes.
3. Keys are generated with StrongBox first, and again without it on `StrongBoxUnavailableException`.
4. Proofs carry `jti`, `htm`, `htu` (no query or fragment), `iat`, plus `ath` and `nonce` when they apply. The JWK holds only `kty`, `crv`, `x` and `y`.

## Code generation: Gradle protobuf plugin (A) or pinned buf remote plugins (B)

| Source | Text |
|---|---|
| [AGP: migrate to built-in Kotlin](https://developer.android.com/build/migrate-to-built-in-kotlin), §4 | "With built-in Kotlin, the only supported option is to add the directories to the `AndroidSourceSet.kotlin` set using the `android.sourceSets{}` DSL"; "Adding Kotlin source directories to AndroidSourceSet.java is also not supported" |
| protobuf-gradle-plugin source (PR #797) | registers all generated output through `variant.sources.java` only (from the research agent's source read). Its AGP 9 test project generates Java lite only, not Kotlin |
| [protobuf-gradle-plugin #814](https://github.com/google/protobuf-gradle-plugin/issues/814) (open) | "Generated class not accessible in andoid studio in version 0.10" |
| [protobuf-gradle-plugin #818](https://github.com/google/protobuf-gradle-plugin/issues/818) (open) | race "with jar-based protoc plugins", which is how protoc-gen-connect-kotlin runs |
| [connect-kotlin examples](https://github.com/connectrpc/connect-kotlin/blob/main/examples/buf.gen.yaml) | generate with **buf** (`protoc_builtin: java`/`kotlin` lite plus connect-kotlin); no example uses the Gradle plugin |
| Now in Android PR #2054 | moved its proto module off Android to a JVM library for AGP 9 |
| [bufbuild/plugins](https://github.com/bufbuild/plugins/tree/main/plugins) | `protocolbuffers/java` and `protocolbuffers/kotlin` **v36.2** (= our runtime 4.36.2) and `connectrpc/kotlin` **v0.10.0** (= our pin) exist |
| [Buf managed mode](https://buf.build/docs/generate/managed-mode/) | "The default `java_package` is the Protobuf package with a `com.` prefix"; override with `java_package_prefix` |
| [buf.gen.yaml v2](https://buf.build/docs/configuration/v2/buf-gen-yaml/) | "'clean', when set to true, deletes the directories ... specified in the `out` field for all plugins" |
| [BSR rate limits](https://buf.build/docs/bsr/rate-limits/) | unauthenticated code generation: 10 requests per hour per IP; with a token: 960 (from the research agent's read) |

Conclusion: **B**. Option A depends on a path AGP 9 calls unsupported (Kotlin through the java source set), is untested for Kotlin by its own AGP 9 test, and has two open bugs, one on exactly the jar-plugin path we need. B pins versions that equal the runtime exactly, uses the same tool as the Go and TS code, and is how connect-kotlin itself generates. Shape:

- A separate template, `protocol/buf.gen.kotlin.yaml`, like `buf.gen.go.yaml`, so `clean` touches only the Kotlin output and the unpinned Swift/TS entries stay out of it.
- `java_package_prefix: al` gives `al.trivia.*.v1`, matching the app ID, instead of the default `com.trivia.*`.
- Output under `android/core/protocol/src/gen/{java,kotlin}`, committed, and added through `sourceSets.main.kotlin` / `java`.
- CI regenerates and fails if the output is stale. It authenticates with a `BUF_TOKEN` secret, because hosted runners share IPs under the 10-per-hour unauthenticated limit.
- Fallback: make `core:protocol` a JVM library, as NiA and connect-kotlin's examples do.

| Source | Text |
|---|---|
| [bufbuild/plugins connectrpc/kotlin v0.10.0](https://github.com/bufbuild/plugins/blob/main/plugins/connectrpc/kotlin/v0.10.0/buf.plugin.yaml) | `deps: - plugin: buf.build/protocolbuffers/kotlin:v36.2`; lite runtime `com.connectrpc:connect-kotlin:0.10.0` |
| [bufbuild/plugins protocolbuffers/kotlin v36.2](https://github.com/bufbuild/plugins/blob/main/plugins/protocolbuffers/kotlin/v36.2/buf.plugin.yaml) | `deps: - plugin: buf.build/protocolbuffers/java:v36.2`; lite runtime `com.google.protobuf:protobuf-kotlin-lite:4.36.2` |
| [connect-kotlin README](https://github.com/connectrpc/connect-kotlin) | `generateCallbackMethods` default `false`; `generateCoroutineMethods` default `true` |
| [BSR authentication](https://buf.build/docs/bsr/authentication/) | "store your token in the provider's secret store and expose it to the job as `BUF_TOKEN`" |
| [AGP: migrate to built-in Kotlin](https://developer.android.com/build/migrate-to-built-in-kotlin) | `android.sourceSets.named("main") { kotlin.directories += "additionalSourceDirectory/kotlin" }` |

Plugins are pinned by version (`buf.build/<owner>/<plugin>:<version>`), not revision. UNVERIFIED until the first CI run: that AGP 9.5.1 compiles the committed sources added this way (`java.directories` is assumed to work like `kotlin.directories`), and whether the connect-kotlin output needs an opt-in flag.
