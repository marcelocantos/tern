// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

package com.marcelocantos.tern.relay

import java.io.InputStream
import java.io.OutputStream
import java.nio.ByteBuffer

/**
 * A [QuicTransport] that delegates to the tern-bridge Go binary via
 * stdin/stdout. Avoids Java QUIC library issues (kwik builder bug)
 * by using the proven Go QUIC client.
 *
 * The bridge binary speaks length-prefixed framing on stdin/stdout,
 * identical to the relay protocol. Messages written to [outputStream]
 * are forwarded to the relay via QUIC; messages from the relay arrive
 * on [inputStream].
 */
class BridgeQuicTransport private constructor(
    private val process: Process,
) : QuicTransport {

    override val inputStream: InputStream = process.inputStream
    override val outputStream: OutputStream = process.outputStream

    override fun sendDatagram(data: ByteArray) {
        // Datagrams not supported through the bridge (stdin/stdout is stream-only).
        throw UnsupportedOperationException("datagrams not supported via bridge")
    }

    override fun receiveDatagram(): ByteArray {
        throw UnsupportedOperationException("datagrams not supported via bridge")
    }

    override fun close() {
        try { outputStream.close() } catch (_: Exception) {}
        process.destroyForcibly()
    }

    companion object {
        /**
         * Build the tern-bridge binary and start it in register or connect mode.
         *
         * @param repoRoot path to the tern repository root (for `go build`)
         * @param relayUrl the relay URL (e.g. "https://tern.fly.dev:4433")
         * @param mode "register" or "connect"
         * @param token bearer token (for register mode)
         * @param instanceID instance ID (for connect mode)
         * @return a connected [BridgeQuicTransport]
         */
        fun start(
            repoRoot: String,
            relayUrl: String,
            mode: String,
            token: String? = null,
            instanceID: String? = null,
        ): BridgeQuicTransport {
            // Build the bridge binary.
            val bridgeBin = "/tmp/tern-bridge-kotlin"
            val build = ProcessBuilder("go", "build", "-o", bridgeBin, "./cmd/tern-bridge")
                .directory(java.io.File(repoRoot))
                .redirectErrorStream(true)
                .start()
            val buildExit = build.waitFor()
            require(buildExit == 0) { "go build failed: ${build.inputStream.readAllBytes().decodeToString()}" }

            // Start the bridge.
            val args = mutableListOf(bridgeBin, mode, relayUrl)
            if (mode == "register" && token != null) args.add(token)
            if (mode == "connect" && instanceID != null) args.add(instanceID)

            val proc = ProcessBuilder(args)
                .redirectError(ProcessBuilder.Redirect.INHERIT)
                .start()

            return BridgeQuicTransport(proc)
        }
    }
}
