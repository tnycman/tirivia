package al.trivia.core.datastore

import android.content.Context
import androidx.datastore.core.DataStore
import androidx.datastore.core.DataStoreFactory
import androidx.datastore.core.handlers.ReplaceFileCorruptionHandler
import androidx.datastore.dataStoreFile
import androidx.datastore.tink.AeadSerializer
import kotlinx.coroutines.flow.Flow

/**
 * The encrypted session file (ID-9). Create one instance per process (DataStore rule).
 * A file that cannot be decrypted, such as after the Keystore key is lost, resets to signed out.
 */
class SessionStore(context: Context) {
    private val store: DataStore<StoredSession> = DataStoreFactory.create(
        serializer = AeadSerializer(KeystoreAead(KEY_ALIAS), SessionSerializer, FILE_NAME.encodeToByteArray()),
        corruptionHandler = ReplaceFileCorruptionHandler { StoredSession.SIGNED_OUT },
        produceFile = { context.dataStoreFile(FILE_NAME) },
    )

    val session: Flow<StoredSession> = store.data

    suspend fun save(session: StoredSession) {
        store.updateData { session }
    }

    suspend fun clear() = save(StoredSession.SIGNED_OUT)

    companion object {
        /** Excluded from backup and device transfer in app/src/main/res/xml/data_extraction_rules.xml. */
        const val FILE_NAME = "session.bin"
        private const val KEY_ALIAS = "session_store_key"
    }
}
