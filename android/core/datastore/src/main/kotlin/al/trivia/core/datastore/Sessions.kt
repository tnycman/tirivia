package al.trivia.core.datastore

import kotlinx.coroutines.flow.Flow

/** The persisted session; [SessionStore] is the encrypted implementation. */
interface Sessions {
    val session: Flow<StoredSession>

    suspend fun save(session: StoredSession)

    suspend fun clear()
}
