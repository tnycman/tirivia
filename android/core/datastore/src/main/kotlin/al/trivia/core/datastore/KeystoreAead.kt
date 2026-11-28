package al.trivia.core.datastore

import com.google.crypto.tink.Aead
import com.google.crypto.tink.integration.android.AndroidKeystore

/** Tink AndroidKeystore AES-256-GCM; the key is created on first use, off the main thread (ID-9, research 17). */
internal class KeystoreAead(private val alias: String) : Aead {
    private val aead by lazy {
        if (!AndroidKeystore.hasKey(alias)) AndroidKeystore.generateNewAes256GcmKey(alias)
        AndroidKeystore.getAead(alias)
    }

    override fun encrypt(plaintext: ByteArray, associatedData: ByteArray): ByteArray = aead.encrypt(plaintext, associatedData)

    override fun decrypt(ciphertext: ByteArray, associatedData: ByteArray): ByteArray = aead.decrypt(ciphertext, associatedData)
}
