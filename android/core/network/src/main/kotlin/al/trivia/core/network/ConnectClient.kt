package al.trivia.core.network

import com.connectrpc.ProtocolClientConfig
import com.connectrpc.extensions.GoogleJavaLiteProtobufStrategy
import com.connectrpc.impl.ProtocolClient
import com.connectrpc.okhttp.ConnectOkHttpClient
import com.connectrpc.protocols.NetworkProtocol
import kotlinx.coroutines.Dispatchers
import okhttp3.OkHttpClient

/** Connect client for our API; Keystore signing in the interceptor runs off the main thread (research 17). */
fun connectClient(baseUrl: String, okHttp: OkHttpClient, proofs: DpopProofs): ProtocolClient = ProtocolClient(
    httpClient = ConnectOkHttpClient(okHttp),
    config = ProtocolClientConfig(
        host = baseUrl,
        serializationStrategy = GoogleJavaLiteProtobufStrategy(),
        networkProtocol = NetworkProtocol.CONNECT,
        interceptors = listOf({ _: ProtocolClientConfig -> DpopInterceptor(proofs) }),
        ioCoroutineContext = Dispatchers.IO,
    ),
)
