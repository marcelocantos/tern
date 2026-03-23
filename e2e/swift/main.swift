// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

// Standalone E2E test for the Swift TernRelay client.
//
// Usage:
//   swift build && .build/debug/tern-e2e-swift                          # local relay (auto-started)
//   TERN_RELAY_HOST=tern.fly.dev TERN_TOKEN=<tok> swift run tern-e2e-swift  # live relay
//
// Exits 0 if all tests pass, 1 on first failure.

#if canImport(Network)

import Foundation
import Network
import TernCrypto
import TernRelay

// MARK: - Configuration

let relayHost = ProcessInfo.processInfo.environment["TERN_RELAY_HOST"] ?? "127.0.0.1"
let relayPort = UInt16(ProcessInfo.processInfo.environment["TERN_RELAY_PORT"] ?? "4433")!
let relayToken = ProcessInfo.processInfo.environment["TERN_TOKEN"]
let isLocal = relayHost == "127.0.0.1" || relayHost == "localhost"

// MARK: - Test runner

var passed = 0
var failed = 0

func test(_ name: String, _ body: () async throws -> Void) async {
    do {
        try await body()
        passed += 1
        print("[PASS] \(name)")
    } catch {
        failed += 1
        print("[FAIL] \(name): \(error)")
    }
}

func insecureOptions() -> NWProtocolQUIC.Options {
    let opts = NWProtocolQUIC.Options(alpn: ["tern"])
    sec_protocol_options_set_verify_block(opts.securityProtocolOptions, { _, _, c in
        c(true)
    }, DispatchQueue.main)
    return opts
}

func quicOptions() -> NWProtocolQUIC.Options? {
    // Local server uses self-signed cert — bypass verification.
    // Live server has Let's Encrypt cert but the QUIC port uses the
    // same certmagic config, so we still need insecure for now because
    // the QUIC ALPN "tern" may not match what certmagic expects.
    return insecureOptions()
}

// MARK: - Local relay server management

class LocalRelay {
    let process: Process
    let quicPort: UInt16

    static func start() throws -> LocalRelay {
        let repoRoot = URL(fileURLWithPath: #filePath)
            .deletingLastPathComponent()  // swift/
            .deletingLastPathComponent()  // e2e/
            .deletingLastPathComponent()  // repo root

        // Build the relay binary.
        let build = Process()
        build.executableURL = URL(fileURLWithPath: "/usr/bin/env")
        build.arguments = ["go", "build", "-o", "/tmp/tern-e2e-server", "./cmd/tern"]
        build.currentDirectoryURL = repoRoot
        build.standardOutput = FileHandle.nullDevice
        build.standardError = FileHandle.nullDevice
        try build.run()
        build.waitUntilExit()
        guard build.terminationStatus == 0 else {
            throw NSError(domain: "LocalRelay", code: 1,
                          userInfo: [NSLocalizedDescriptionKey: "go build failed with status \(build.terminationStatus)"])
        }

        // Find a free UDP port.
        let port = findFreePort()

        // Start the server.
        let proc = Process()
        proc.executableURL = URL(fileURLWithPath: "/tmp/tern-e2e-server")
        proc.arguments = ["--quic-port", String(port)]
        proc.standardOutput = FileHandle.nullDevice

        // Capture stderr to detect "tern starting".
        let pipe = Pipe()
        proc.standardError = pipe

        try proc.run()

        // Wait for "tern starting" or timeout after 10 seconds.
        let deadline = Date().addingTimeInterval(10)
        var started = false
        let readQueue = DispatchQueue(label: "relay-stderr")
        pipe.fileHandleForReading.readabilityHandler = { handle in
            let data = handle.availableData
            if let str = String(data: data, encoding: .utf8), str.contains("tern starting") {
                started = true
            }
        }
        while !started && Date() < deadline {
            Thread.sleep(forTimeInterval: 0.1)
        }
        pipe.fileHandleForReading.readabilityHandler = nil

        if !started {
            proc.terminate()
            throw NSError(domain: "LocalRelay", code: 2,
                          userInfo: [NSLocalizedDescriptionKey: "server did not start within 10 seconds"])
        }

        return LocalRelay(process: proc, quicPort: port)
    }

    private init(process: Process, quicPort: UInt16) {
        self.process = process
        self.quicPort = quicPort
    }

    func stop() {
        process.terminate()
        process.waitUntilExit()
    }

    private static func findFreePort() -> UInt16 {
        let sock = socket(AF_INET, SOCK_DGRAM, 0)
        defer { close(sock) }
        var addr = sockaddr_in()
        addr.sin_family = sa_family_t(AF_INET)
        addr.sin_port = 0
        addr.sin_addr.s_addr = INADDR_ANY.bigEndian
        var addrLen = socklen_t(MemoryLayout<sockaddr_in>.size)
        withUnsafePointer(to: &addr) {
            $0.withMemoryRebound(to: sockaddr.self, capacity: 1) {
                _ = bind(sock, $0, addrLen)
            }
        }
        withUnsafeMutablePointer(to: &addr) {
            $0.withMemoryRebound(to: sockaddr.self, capacity: 1) {
                _ = getsockname(sock, $0, &addrLen)
            }
        }
        return UInt16(bigEndian: addr.sin_port)
    }
}

// MARK: - Tests

@main struct E2ETests {
    static func main() async {
        let actualHost: String
        let actualPort: UInt16
        var localRelay: LocalRelay? = nil

        if isLocal {
            print("Starting local relay server...")
            do {
                localRelay = try LocalRelay.start()
                actualHost = "127.0.0.1"
                actualPort = localRelay!.quicPort
                print("Local relay started on port \(actualPort)")
            } catch {
                print("[FATAL] Failed to start local relay: \(error)")
                exit(1)
            }
        } else {
            actualHost = relayHost
            actualPort = relayPort
            print("Using relay at \(actualHost):\(actualPort)")
        }

        defer { localRelay?.stop() }

        // --- Test: Register ---
        await test("register") {
            let conn = try await TernConn.register(
                host: actualHost, port: actualPort,
                token: relayToken, quicOptions: quicOptions())
            defer { conn.close() }
            guard !conn.instanceID.isEmpty else {
                throw err("empty instance ID")
            }
        }

        // --- Test: Stream round-trip ---
        await test("stream round-trip") {
            let backend = try await TernConn.register(
                host: actualHost, port: actualPort,
                token: relayToken, quicOptions: quicOptions())
            defer { backend.close() }

            let client = try await TernConn.connect(
                host: actualHost, port: actualPort,
                instanceID: backend.instanceID, quicOptions: quicOptions())
            defer { client.close() }

            // Client → backend
            try await client.send(Data("hello from swift".utf8))
            let msg = try await backend.recv()
            guard String(decoding: msg, as: UTF8.self) == "hello from swift" else {
                throw err("got '\(String(decoding: msg, as: UTF8.self))'")
            }

            // Backend → client
            try await backend.send(Data("reply from swift".utf8))
            let reply = try await client.recv()
            guard String(decoding: reply, as: UTF8.self) == "reply from swift" else {
                throw err("got '\(String(decoding: reply, as: UTF8.self))'")
            }
        }

        // --- Test: Multiple messages ---
        await test("10 messages in order") {
            let backend = try await TernConn.register(
                host: actualHost, port: actualPort,
                token: relayToken, quicOptions: quicOptions())
            defer { backend.close() }

            let client = try await TernConn.connect(
                host: actualHost, port: actualPort,
                instanceID: backend.instanceID, quicOptions: quicOptions())
            defer { client.close() }

            for i in 0..<10 {
                try await client.send(Data("msg-\(i)".utf8))
            }
            for i in 0..<10 {
                let data = try await backend.recv()
                let text = String(decoding: data, as: UTF8.self)
                guard text == "msg-\(i)" else {
                    throw err("msg \(i): got '\(text)'")
                }
            }
        }

        // --- Test: Encrypted stream ---
        await test("encrypted stream round-trip") {
            let backend = try await TernConn.register(
                host: actualHost, port: actualPort,
                token: relayToken, quicOptions: quicOptions())
            defer { backend.close() }

            let client = try await TernConn.connect(
                host: actualHost, port: actualPort,
                instanceID: backend.instanceID, quicOptions: quicOptions())
            defer { client.close() }

            // Exchange public keys through relay.
            let bKP = E2EKeyPair()
            let cKP = E2EKeyPair()

            try await client.send(cKP.publicKeyData)
            try await backend.send(bKP.publicKeyData)

            let cPubData = try await backend.recv()
            let bPubData = try await client.recv()

            // Derive session keys.
            let bSendKey = try bKP.deriveSessionKey(peerPublicKey: cPubData, info: Data("b2c".utf8))
            let bRecvKey = try bKP.deriveSessionKey(peerPublicKey: cPubData, info: Data("c2b".utf8))
            let cSendKey = try cKP.deriveSessionKey(peerPublicKey: bPubData, info: Data("c2b".utf8))
            let cRecvKey = try cKP.deriveSessionKey(peerPublicKey: bPubData, info: Data("b2c".utf8))

            let bCh = E2EChannel(sendKey: bSendKey, recvKey: bRecvKey)
            let cCh = E2EChannel(sendKey: cSendKey, recvKey: cRecvKey)

            // Send encrypted message client → backend.
            let plaintext = Data("encrypted hello".utf8)
            let ciphertext = try cCh.encrypt(plaintext)
            try await client.send(ciphertext)

            let received = try await backend.recv()
            let decrypted = try bCh.decrypt(received)
            guard decrypted == plaintext else {
                throw err("decrypted mismatch")
            }

            // Backend → client.
            let reply = Data("encrypted reply".utf8)
            let replyCt = try bCh.encrypt(reply)
            try await backend.send(replyCt)

            let replyRecv = try await client.recv()
            let replyPt = try cCh.decrypt(replyRecv)
            guard replyPt == reply else {
                throw err("reply decrypted mismatch")
            }
        }

        // --- Test: Confirmation codes ---
        await test("confirmation codes match") {
            let bKP = E2EKeyPair()
            let cKP = E2EKeyPair()

            let bCode = deriveConfirmationCode(bKP.publicKeyData, cKP.publicKeyData)
            let cCode = deriveConfirmationCode(cKP.publicKeyData, bKP.publicKeyData)

            guard bCode == cCode else {
                throw err("codes differ: \(bCode) vs \(cCode)")
            }
            guard bCode.count == 6 else {
                throw err("code not 6 digits: \(bCode)")
            }
        }

        // --- Test: Cross-platform confirmation code vector ---
        await test("confirmation code cross-platform vector") {
            let keyA = Data(repeating: 0x01, count: 32)
            let keyB = Data(repeating: 0x02, count: 32)
            let code = deriveConfirmationCode(keyA, keyB)
            guard code == "629624" else {
                throw err("got '\(code)', want '629624'")
            }
        }

        // --- Summary ---
        print("\n=== \(passed) passed, \(failed) failed ===")
        exit(failed > 0 ? 1 : 0)
    }
}

func err(_ msg: String) -> NSError {
    NSError(domain: "E2E", code: 0, userInfo: [NSLocalizedDescriptionKey: msg])
}

#else
@main struct E2ETests {
    static func main() {
        print("Network.framework not available — skipping E2E tests")
        exit(0)
    }
}
#endif
