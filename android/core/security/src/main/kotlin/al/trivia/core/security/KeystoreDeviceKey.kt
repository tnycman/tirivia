package al.trivia.core.security

import android.os.Build
import android.security.keystore.KeyGenParameterSpec
import android.security.keystore.KeyProperties
import android.security.keystore.StrongBoxUnavailableException
import java.security.KeyPair
import java.security.KeyPairGenerator
import java.security.KeyStore
import java.security.PrivateKey
import java.security.Signature
import java.security.interfaces.ECPublicKey
import java.security.spec.ECGenParameterSpec

/** DPoP device key in Android Keystore, StrongBox-backed where available (stack/07-android rule 10, research 16). */
class KeystoreDeviceKey(private val alias: String = ALIAS) : DeviceKey {
    private val keyStore by lazy { KeyStore.getInstance(PROVIDER).apply { load(null) } }
    private val stored by lazy { existing() ?: create() }

    override val publicKey: ECPublicKey
        get() = stored.public as ECPublicKey

    override fun sign(data: ByteArray): ByteArray {
        val signer = Signature.getInstance(SIGNATURE)
        signer.initSign(stored.private)
        signer.update(data)
        return EcdsaJose.derToJose(signer.sign())
    }

    private fun existing(): KeyPair? {
        val privateKey = keyStore.getKey(alias, null) as PrivateKey? ?: return null
        return KeyPair(keyStore.getCertificate(alias).publicKey, privateKey)
    }

    private fun create(): KeyPair {
        if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.P) {
            try {
                generate(spec().setIsStrongBoxBacked(true).build())
            } catch (unavailable: StrongBoxUnavailableException) {
                generate(spec().build())
            }
        } else {
            generate(spec().build())
        }
        return checkNotNull(existing()) { "key store did not keep the device key" }
    }

    private fun spec() = KeyGenParameterSpec.Builder(alias, KeyProperties.PURPOSE_SIGN)
        .setAlgorithmParameterSpec(ECGenParameterSpec(CURVE))
        .setDigests(KeyProperties.DIGEST_SHA256)

    private fun generate(spec: KeyGenParameterSpec) {
        KeyPairGenerator.getInstance(KeyProperties.KEY_ALGORITHM_EC, PROVIDER).run {
            initialize(spec)
            generateKeyPair()
        }
    }

    private companion object {
        const val ALIAS = "dpop_device_key"
        const val PROVIDER = "AndroidKeyStore"
        const val CURVE = "secp256r1"
        const val SIGNATURE = "SHA256withECDSA"
    }
}
