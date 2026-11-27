import CryptoKit
import Foundation

/// Per-player option order (SPEC §6.3), pinned by protocol/testvectors.
public enum OptionShuffle {
    public static let maxOptions = 33

    /// Returns a permutation of 0..<n: display position p shows option `order[p]`.
    public static func order(seed: Data, questionID: String, count n: Int) -> [Int] {
        precondition(n >= 1 && n <= maxOptions, "option count out of range")
        let mac = HMAC<SHA256>.authenticationCode(
            for: Data(questionID.utf8),
            using: SymmetricKey(data: seed)
        )
        let k = Array(mac)
        var idx = Array(0..<n)
        var i = n - 1
        while i >= 1 {
            let j = Int(k[n - 1 - i]) % (i + 1)
            idx.swapAt(i, j)
            i -= 1
        }
        return idx
    }
}
