package al.trivia.core.network

import com.connectrpc.Interceptor
import com.connectrpc.StreamFunction
import com.connectrpc.UnaryFunction
import com.connectrpc.http.clone

/** Adds a DPoP proof for the final request URL to every unary call (research 17). */
class DpopInterceptor(private val proofs: DpopProofs) : Interceptor {
    override fun unaryFunction() = UnaryFunction(
        requestFunction = { request ->
            val proof = proofs.create(request.httpMethod.string, request.url.toString())
            request.clone(headers = request.headers + (HEADER to listOf(proof)))
        },
    )

    override fun streamFunction() = StreamFunction()

    private companion object {
        const val HEADER = "DPoP"
    }
}
