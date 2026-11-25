import XCTest
@testable import CoreSecurity

/// Runs protocol/testvectors/shuffle_vectors.json (copied into Vectors/ by CI) against the iOS
/// OptionShuffle. The same file is run by Go (reference) and Kotlin; all must agree.
/// The cue-verification runner is added with the iOS CueVerifier in Phase 2a.
final class ShuffleVectorTests: XCTestCase {
    private struct Suite: Decodable {
        let format_version: Int
        let vectors: [Vector]
    }

    private struct Vector: Decodable {
        let name: String
        let seed_hex: String
        let question_id: String
        let n: Int
        let order: [Int]
    }

    private func hexData(_ hex: String) -> Data {
        var data = Data(capacity: hex.count / 2)
        var index = hex.startIndex
        while index < hex.endIndex {
            let next = hex.index(index, offsetBy: 2)
            data.append(UInt8(hex[index..<next], radix: 16)!)
            index = next
        }
        return data
    }

    func testSharedShuffleVectors() throws {
        let url = try XCTUnwrap(Bundle.module.url(forResource: "shuffle_vectors", withExtension: "json", subdirectory: "Vectors"))
        let suite = try JSONDecoder().decode(Suite.self, from: Data(contentsOf: url))
        XCTAssertEqual(suite.format_version, 1)
        XCTAssertGreaterThanOrEqual(suite.vectors.count, 16, "runner must not silently skip vectors")
        for v in suite.vectors {
            let got = OptionShuffle.order(seed: hexData(v.seed_hex), questionID: v.question_id, count: v.n)
            XCTAssertEqual(got, v.order, v.name)
        }
    }
}
