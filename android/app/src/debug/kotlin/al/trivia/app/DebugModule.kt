package al.trivia.app

import al.trivia.core.attestation.IntegrityTokens
import dagger.Module
import dagger.Provides
import dagger.hilt.InstallIn
import dagger.hilt.components.SingletonComponent

/**
 * Debug builds only: the emulator reaches the workstation's 127.0.0.1 at 10.0.2.2 (Android emulator
 * networking docs), where the backend runs with AUTH_DEV_MODE=1. Release has no such bindings.
 */
@Module
@InstallIn(SingletonComponent::class)
object DebugModule {
    @Provides
    fun apiConfig(): ApiConfig = ApiConfig("http://10.0.2.2:8081")

    @Provides
    fun integrityTokens(): IntegrityTokens = DevIntegrityTokens
}

/** AUTH_DEV_MODE accepts any non-empty token (backend/internal/auth/attest/dev.go). */
private object DevIntegrityTokens : IntegrityTokens {
    override suspend fun token(challenge: ByteArray) = "dev-mode"
}
