package al.trivia.core.datastore

import androidx.datastore.core.CorruptionException
import androidx.datastore.tink.AeadSerializer
import com.google.crypto.tink.Aead
import kotlinx.coroutines.runBlocking
import org.junit.Assert.assertEquals
import org.junit.Assert.assertFalse
import org.junit.Assert.assertThrows
import org.junit.Test
import java.io.ByteArrayOutputStream
import java.security.SecureRandom
import javax.crypto.Cipher
import javax.crypto.KeyGenerator
import javax.crypto.spec.GCMParameterSpec

/** On the JVM a JDK AES-GCM Aead stands in for Tink's AndroidKeystore one, which needs a device. */
class SessionSerializerTest {
    private val session = StoredSession("user-1", "refresh-token-value")
    private val associatedData = SessionStore.FILE_NAME.encodeToByteArray()
    private val encrypted = AeadSerializer(JdkAesGcm(), SessionSerializer, associatedData)

    @Test
    fun roundTripsAndReadsEmptyAsSignedOut() = runBlocking {
        assertEquals(session, SessionSerializer.readFrom(write(SessionSerializer, session).inputStream()))
        assertEquals(StoredSession.SIGNED_OUT, SessionSerializer.readFrom(ByteArray(0).inputStream()))
        assertFalse(StoredSession.SIGNED_OUT.signedIn)
    }

    @Test
    fun rejectsUnknownVersionTruncationAndTrailingBytes() {
        val good = runBlocking { write(SessionSerializer, session) }
        listOf(
            byteArrayOf(9) + good.copyOfRange(1, good.size),
            good.copyOf(good.size - 3),
            good + byteArrayOf(0),
        ).forEach { bad ->
            assertThrows(CorruptionException::class.java) { runBlocking { SessionSerializer.readFrom(bad.inputStream()) } }
        }
    }

    @Test
    fun encryptedFileRoundTripsAndHidesTheToken() = runBlocking {
        val bytes = write(encrypted, session)
        assertFalse(String(bytes, Charsets.ISO_8859_1).contains(session.refreshToken))
        assertEquals(session, encrypted.readFrom(bytes.inputStream()))
    }

    @Test
    fun tamperedOrSwappedCiphertextIsCorruption() {
        val bytes = runBlocking { write(encrypted, session) }
        val tampered = bytes.copyOf().also { it[it.size - 1] = (it[it.size - 1].toInt() xor 1).toByte() }
        assertThrows(CorruptionException::class.java) { runBlocking { encrypted.readFrom(tampered.inputStream()) } }
        val otherFile = AeadSerializer(JdkAesGcm(), SessionSerializer, "other.bin".encodeToByteArray())
        assertThrows(CorruptionException::class.java) { runBlocking { otherFile.readFrom(bytes.inputStream()) } }
    }

    private suspend fun write(serializer: androidx.datastore.core.Serializer<StoredSession>, value: StoredSession): ByteArray =
        ByteArrayOutputStream().also { serializer.writeTo(value, it) }.toByteArray()

    private class JdkAesGcm : Aead {
        companion object {
            private val key = KeyGenerator.getInstance("AES").apply { init(256) }.generateKey()
        }

        override fun encrypt(plaintext: ByteArray, associatedData: ByteArray): ByteArray {
            val iv = ByteArray(12).also { SecureRandom().nextBytes(it) }
            val cipher = Cipher.getInstance("AES/GCM/NoPadding")
            cipher.init(Cipher.ENCRYPT_MODE, key, GCMParameterSpec(128, iv))
            cipher.updateAAD(associatedData)
            return iv + cipher.doFinal(plaintext)
        }

        override fun decrypt(ciphertext: ByteArray, associatedData: ByteArray): ByteArray {
            val cipher = Cipher.getInstance("AES/GCM/NoPadding")
            cipher.init(Cipher.DECRYPT_MODE, key, GCMParameterSpec(128, ciphertext, 0, 12))
            cipher.updateAAD(associatedData)
            return cipher.doFinal(ciphertext, 12, ciphertext.size - 12)
        }
    }
}
