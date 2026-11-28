package al.trivia.app

import al.trivia.core.attestation.IntegrityTokens
import al.trivia.core.data.SignInRepository
import al.trivia.core.datastore.SessionStore
import al.trivia.core.datastore.Sessions
import al.trivia.core.network.DpopProofs
import al.trivia.core.network.connectClient
import al.trivia.core.security.DeviceKey
import al.trivia.core.security.KeystoreDeviceKey
import al.trivia.core.v1.AuthServiceClient
import al.trivia.core.v1.AuthServiceClientInterface
import android.content.Context
import com.connectrpc.impl.ProtocolClient
import dagger.Module
import dagger.Provides
import dagger.hilt.InstallIn
import dagger.hilt.android.qualifiers.ApplicationContext
import dagger.hilt.components.SingletonComponent
import okhttp3.OkHttpClient
import javax.inject.Singleton

/** Where the API lives; provided per build type (src/debug only until production exists). */
data class ApiConfig(val baseUrl: String)

@Module
@InstallIn(SingletonComponent::class)
object AppModule {
    @Provides
    @Singleton
    fun deviceKey(): DeviceKey = KeystoreDeviceKey()

    @Provides
    @Singleton
    fun okHttp(): OkHttpClient = OkHttpClient()

    @Provides
    @Singleton
    fun protocolClient(config: ApiConfig, okHttp: OkHttpClient, key: DeviceKey): ProtocolClient =
        connectClient(config.baseUrl, okHttp, DpopProofs(key))

    @Provides
    @Singleton
    fun authService(client: ProtocolClient): AuthServiceClientInterface = AuthServiceClient(client)

    @Provides
    fun signIn(auth: AuthServiceClientInterface, integrity: IntegrityTokens): SignInRepository =
        SignInRepository(auth, integrity)

    @Provides
    @Singleton
    fun sessions(@ApplicationContext context: Context): Sessions = SessionStore(context)
}
