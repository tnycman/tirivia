package al.trivia.core.datastore

/** What survives a restart: the refresh token. The access token (10 min, ID-4) stays in memory. */
data class StoredSession(val userId: String, val refreshToken: String) {
    val signedIn: Boolean get() = refreshToken.isNotEmpty()

    companion object {
        val SIGNED_OUT = StoredSession("", "")
    }
}
