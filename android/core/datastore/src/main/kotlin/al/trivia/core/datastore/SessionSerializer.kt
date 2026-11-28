package al.trivia.core.datastore

import androidx.datastore.core.CorruptionException
import androidx.datastore.core.Serializer
import java.io.DataInputStream
import java.io.DataOutputStream
import java.io.IOException
import java.io.InputStream
import java.io.OutputStream

/** Plaintext layout inside the encrypted file: version byte, then userId and refreshToken. */
internal object SessionSerializer : Serializer<StoredSession> {
    private const val VERSION = 1

    override val defaultValue = StoredSession.SIGNED_OUT

    override suspend fun readFrom(input: InputStream): StoredSession {
        val bytes = input.readBytes()
        if (bytes.isEmpty()) return defaultValue
        val data = DataInputStream(bytes.inputStream())
        try {
            if (data.readUnsignedByte() != VERSION) throw CorruptionException("unknown session format")
            val session = StoredSession(data.readUTF(), data.readUTF())
            if (data.read() != -1) throw CorruptionException("trailing bytes in session")
            return session
        } catch (e: CorruptionException) {
            throw e
        } catch (e: IOException) {
            throw CorruptionException("unreadable session", e)
        }
    }

    override suspend fun writeTo(t: StoredSession, output: OutputStream) {
        val data = DataOutputStream(output)
        data.writeByte(VERSION)
        data.writeUTF(t.userId)
        data.writeUTF(t.refreshToken)
        data.flush()
    }
}
