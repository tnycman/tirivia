# 17. Android sign-in client: Connect setup, DPoP on requests, session storage

PLAN Phase 1, slice 1c. Builds on research 16 (generated code, device key, proofs). Sources read 2026-11-28; quotes verbatim unless marked.

## What the backend expects (our code, `main` 7f0a18a)

| Step | RPC | DPoP | Source |
|---|---|---|---|
| 1 | `GetAttestationChallenge` | no | `backend/internal/auth/api/attest.go` |
| 2 | `Attest` | yes; the ticket is bound to the proof's key | `attest.go`, `attest/service.go` |
| 3 | `StartOtp` (phone, ticket, channel, locale) | no | `api/otp.go`; phone must be canonical E.164 and mobile (`otp/phone.go`, libphonenumber), else `NUMBER_NOT_ALLOWED` |
| 4 | `VerifyOtp` | yes; same key as `Attest` | `api/otp.go` |
| later | `RefreshSession`, `SignOut` | yes | `api/session.go` |

- Proof checked against header `DPoP`, `htm` `POST`, `htu` = `PublicBaseURL` + procedure, no nonce or `ath` (`api/handler.go:62`).
- `AUTH_DEV_MODE=1`: any non-empty Play Integrity token passes (`attest/dev.go`). Codes go to the server log only. Default base `http://localhost:8081` (`cmd/core/authdev.go`).

## connect-kotlin 0.10.0 (source at tag v0.10.0)

| Source | Text |
|---|---|
| [ProtocolClientConfig.kt](https://github.com/connectrpc/connect-kotlin/blob/v0.10.0/library/src/main/kotlin/com/connectrpc/ProtocolClientConfig.kt) | `class ProtocolClientConfig @JvmOverloads constructor(val host: String, val serializationStrategy: SerializationStrategy, networkProtocol: NetworkProtocol = NetworkProtocol.CONNECT, … interceptors: List<(ProtocolClientConfig) -> Interceptor> = emptyList(), … val ioCoroutineContext: CoroutineContext? = null, val timeoutOracle: TimeoutOracle = { 10.toDuration(DurationUnit.SECONDS) }, …)` |
| [ConnectOkHttpClient.kt](https://github.com/connectrpc/connect-kotlin/blob/v0.10.0/okhttp/src/main/kotlin/com/connectrpc/okhttp/ConnectOkHttpClient.kt) | `class ConnectOkHttpClient @JvmOverloads constructor(private val unaryClient: OkHttpClient = OkHttpClient(), …)` |
| [GoogleJavaLiteProtobufStrategy.kt](https://github.com/connectrpc/connect-kotlin/blob/v0.10.0/extensions/google-javalite/src/main/kotlin/com/connectrpc/extensions/GoogleJavaLiteProtobufStrategy.kt) | `class GoogleJavaLiteProtobufStrategy(private val registry: ExtensionRegistryLite = ExtensionRegistryLite.getEmptyRegistry())` |
| [Interceptor.kt](https://github.com/connectrpc/connect-kotlin/blob/v0.10.0/library/src/main/kotlin/com/connectrpc/Interceptor.kt) | `class UnaryFunction(val requestFunction: (UnaryHTTPRequest) -> UnaryHTTPRequest = { it }, …)`: not `suspend` |
| [ProtocolClient.kt](https://github.com/connectrpc/connect-kotlin/blob/v0.10.0/library/src/main/kotlin/com/connectrpc/impl/ProtocolClient.kt) | `UnaryHTTPRequest(url = urlFromMethodSpec(methodSpec), …)` then `config.createInterceptorChain()`; `urlFromMethodSpec` appends `methodSpec.path` to the host; with `ioCoroutineContext` set the call runs in `withContext(config.ioCoroutineContext)` |
| [Interceptors doc](https://connectrpc.com/docs/kotlin/interceptors/) | "Request types are immutable: to mutate one, use the `clone()` extension function on `UnaryHTTPRequest`"; "The client will invoke each interceptor in FIFO order on the request path" |
| [ProtocolClientInterface.kt](https://github.com/connectrpc/connect-kotlin/blob/v0.10.0/library/src/main/kotlin/com/connectrpc/ProtocolClientInterface.kt) | `typealias Headers = Map<String, List<String>>` |
| [ResponseMessage.kt](https://github.com/connectrpc/connect-kotlin/blob/v0.10.0/library/src/main/kotlin/com/connectrpc/ResponseMessage.kt) | `Success(val message)`, `Failure(val cause: ConnectException)`; `fold(onSuccess, onFailure)`, `getOrThrow()` |
| [Code.kt](https://github.com/connectrpc/connect-kotlin/blob/v0.10.0/library/src/main/kotlin/com/connectrpc/Code.kt) | `UNAUTHENTICATED`, `RESOURCE_EXHAUSTED`, `PERMISSION_DENIED`, `UNAVAILABLE`, …; internal is `INTERNAL_ERROR` |

Settled: a DPoP interceptor signs in `requestFunction` from `request.url` (the final POST URL; the Connect protocol interceptor runs after it and keeps the URL unless GET is enabled, which is off by default). Keystore signing is synchronous, so the client sets `ioCoroutineContext` to an IO dispatcher. The Keystore page: "apps should avoid using `AndroidKeyStore` on their main thread".

UNVERIFIED: how ktor's `appendPathSegments` encodes the path (proven only by the first end-to-end call, which the server rejects if `htu` differs).

## Session storage

| Source | Text |
|---|---|
| [Security release notes](https://developer.android.com/jetpack/androidx/releases/security), 1.1.0-beta01 | "Deprecated all APIs in favour of existing platform APIs and direct use of Android Keystore." |
| [Cryptography](https://developer.android.com/privacy-and-security/cryptography) | "All APIs in the `security-crypto` Jetpack library were deprecated in the stable release of version `1.1.0`."; "AES in either CBC or GCM mode with 256-bit keys (such as `AES/GCM/NoPadding`)" |
| Security best practices (developer.android.com) | "use the Android Keystore, and encrypt stored keys using a robust tool such as Tink Java." |
| [DataStore releases](https://developer.android.com/jetpack/androidx/releases/datastore), 1.3.0 | "Added the new `androidx.datastore:datastore-tink` artifact, which provides encryption support using the Tink library. It introduces `AeadSerializer`"; alpha07: keyset via `AndroidKeysetManager … withMasterKeyUri("android-keystore://master_key")`; `associatedData`: "Specify a unique name to prevent a ciphertext swapping attack." |
| [KeyGenParameterSpec.Builder](https://developer.android.com/reference/android/security/keystore/KeyGenParameterSpec.Builder) `setRandomizedEncryptionRequired` | "caller-provided IVs are rejected when encrypting"; "let the `Cipher` generate a random IV … The IV can then be queried via Cipher.getIV()." |
| [DataStore guide](https://developer.android.com/topic/libraries/architecture/datastore) | "By default, these files are included in Android Auto Backup and device-to-device (D2D) transfers." |
| [Auto Backup](https://developer.android.com/identity/data/autobackup) | "To back up user credentials and authentication tokens, don't store them in shared preferences or a file."; API 31+ `dataExtractionRules`, API ≤ 30 `fullBackupContent`; "If there are no rules for a particular backup mode … that mode is fully enabled" |
| [Restore Credentials](https://developer.android.com/identity/sign-in/restore-credentials) | "Exclude raw authentication tokens, hardware-bound Android Keystore ciphertext … so stale session state isn't copied onto the new device." |
| [Keystore](https://developer.android.com/privacy-and-security/keystore) "Require user authentication for key use" | "generally useful only if your requirements are that a compromise of your application process after key generation … can't bypass the requirement" |

Settled:
1. `security-crypto` is not used (deprecated).
2. Session files and any keyset file are excluded from cloud backup and device transfer, with both rule sets (API 31+ and ≤ 30).
3. The token key does not require user authentication (the doc scopes that to a different threat; the step-up key is separate).

Owner decision 2026-11-28 (register ID-9): `datastore-tink`, refined to Tink `AndroidKeystore` (below).

### Settled by the 1.3.0 / 1.24.0 source jars (Google Maven, Maven Central)

| Source | Text |
|---|---|
| `datastore-tink-android-1.3.0.module` | depends on `com.google.crypto.tink:tink-android` `1.20.0` (the Android variant; the JVM variant uses `tink`) |
| `datastore-tink-android-1.3.0-sources.jar` `AeadSerializer.kt` | `public class AeadSerializer<T>(private val aead: Aead, private val wrappedSerializer: Serializer<T>, private val associatedData: ByteArray = byteArrayOf())`. Decryption failure: `throw CorruptionException("Decryption failed", e)`. An empty file is passed to the wrapped serializer as empty bytes |
| `tink-android-1.24.0-sources.jar` `AndroidKeysetManager.java` | "We don't recommend using this class. Instead, directly use {@link AndroidKeystore}."; "When Android Keystore is disabled or otherwise unavailable, keysets will be stored in cleartext." |
| same, `AndroidKeystore.java` | `generateNewAes256GcmKey(alias)`: `setKeySize(256)`, `BLOCK_MODE_GCM`, `ENCRYPTION_PADDING_NONE`; `getAead(alias)`: "All instances of this class use a 12-byte IV and a 16-byte tag"; the IV comes from `cipher.getIV()`; "Warning: This API is not thread-safe." |
| `datastore-core-android-1.3.0-sources.jar` | `DataStoreFactory.create(serializer, corruptionHandler: ReplaceFileCorruptionHandler<T>? = null, …, produceFile)`; `Serializer.readFrom` "@throws … CorruptionException if the data … is corrupted" |
| `datastore-android-1.3.0-sources.jar` | `Context.dataStoreFile(fileName)` = `File(applicationContext.filesDir, "datastore/$fileName")` |
| [Auto Backup](https://developer.android.com/identity/data/autobackup) syntax reference | `<data-extraction-rules><cloud-backup>…</cloud-backup><device-transfer>…</device-transfer></data-extraction-rules>`; "On devices from some device manufacturers, specifying android:allowBackup="false" disables cloud-based backup and restore … but doesn't disable device-to-device transfers" |

This resolves the earlier UNVERIFIED items:
- **Tag and IV lengths, and key size:** set by Tink's `AndroidKeystore`.
- **Element name:** `<device-transfer>`, per the syntax reference.
- **Library versions:** `datastore-tink` 1.3.0 with `tink-android` 1.24.0, the latest release on Maven Central.
- **When the key is lost or the file is corrupt:** decryption fails with `CorruptionException`, and the replace handler resets to signed-out. End-to-end check: the owner runs the app in Android Studio's emulator on the workstation against the dev backend.

UNVERIFIED:
- AndroidKeyStore GCM tag and IV lengths.
- The default AES key size.
- The `<device-transfer>` vs `<device-to-device>` element name. The two official pages differ, so the data-extraction-rules reference must be read before the XML is written.
- The `datastore-tink` and Tink dependency versions.

## Sign-in screens and wiring (slice 1c-2)

Checked in the published source jars and docs on 2026-11-28:

| Source | Text |
|---|---|
| `hilt-lifecycle-viewmodel-compose-1.4.0-sources.jar` | `package androidx.hilt.lifecycle.viewmodel.compose`; `public inline fun <reified VM : ViewModel> hiltViewModel(` |
| `lifecycle-runtime-compose-android-2.11.1-sources.jar` | `package androidx.lifecycle.compose`; `fun <T> StateFlow<T>.collectAsStateWithLifecycle(` |
| `material3-android-1.5.0-sources.jar` (Compose BOM 2026.11.01) | `public fun OutlinedTextField(value: String, onValueChange: (String) -> Unit, …, keyboardOptions: KeyboardOptions = KeyboardOptions.Default, …, singleLine: Boolean = false, …)`; `RadioButton(selected, onClick: (() -> Unit)?, …)`; `Button`, `TextButton`, `Surface`, `Text(…, style: TextStyle …)` |
| `foundation-android-1.13.0` / `ui-text-android-1.13.0` / `ui-android-1.13.0` sources | `Modifier.selectable(selected, enabled, role, onClick)`; `KeyboardOptions(…, keyboardType: KeyboardType …)`; `KeyboardType.Phone`, `KeyboardType.Number`; `stringResource(@StringRes id: Int, vararg formatArgs: Any)`; `LocalConfiguration`; `Role.RadioButton` |
| [Emulator networking](https://developer.android.com/studio/run/emulator-networking-address) | "10.0.2.2 Special alias to your host loopback interface (127.0.0.1 on your development machine)" |
| [Network security config](https://developer.android.com/privacy-and-security/security-config) | `<domain-config cleartextTrafficPermitted="true"><domain includeSubdomains="true">insecure.example.com</domain></domain-config>`; the default for API 28+ is `<base-config cleartextTrafficPermitted="false">` |
| [Build variants](https://developer.android.com/build/build-variants) | Debug-only Kotlin goes in "src/debug/kotlin/ or src/debug/java/"; "Gradle gives files in the build type and product flavor source sets a higher priority than the same files in main/." |
| [ConnectOkHttpClient.kt v0.10.0](https://github.com/connectrpc/connect-kotlin/blob/v0.10.0/okhttp/src/main/kotlin/com/connectrpc/okhttp/ConnectOkHttpClient.kt) | network `IOException` → `cause = ConnectException(code, …)`, so sign-in sees network failures as Connect codes |

Settled:
- **Debug-only wiring.** `src/debug` holds the API address (`http://10.0.2.2:8081`), the dev integrity stand-in and the cleartext rule for `10.0.2.2` only. `main` allows HTTPS only.
- **Release builds stop at compile time.** Hilt cannot satisfy `ApiConfig` and `IntegrityTokens` in release, so a release build fails until the production API and Play Integrity exist. Dev attestation therefore can never ship.
- **The dev backend must use the address the app signs.** The server checks `htu` against `PUBLIC_BASE_URL`, so the dev backend runs with `PUBLIC_BASE_URL=http://10.0.2.2:8081`.
- **No navigation library yet.** The two steps are state in one ViewModel. The app owns navigation once more screens exist.

Open:
- The Albanian strings need the owner's review.
- The ID-5 self-declared 18+ step needs a backend field. It is not in `VerifyOtp` or any implemented RPC.
