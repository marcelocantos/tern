// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

package com.marcelocantos.tern.crypto

import javax.crypto.Mac
import javax.crypto.spec.SecretKeySpec

/**
 * HKDF-SHA256 key derivation (RFC 5869).
 * Matches the Go and Swift implementations: empty salt (treated as 32 zero bytes).
 */
internal object Hkdf {
    private const val HASH_LEN = 32 // SHA-256 output length

    /**
     * Derive a key from input keying material.
     *
     * @param ikm input keying material (e.g. ECDH shared secret)
     * @param info context/application-specific info
     * @param length output key length in bytes (default 32)
     * @return derived key
     */
    fun derive(ikm: ByteArray, info: ByteArray, length: Int = 32): ByteArray {
        // Extract: PRK = HMAC-SHA256(salt, IKM)
        val salt = ByteArray(HASH_LEN) // empty salt → zero-filled
        val prk = hmacSha256(salt, ikm)

        // Expand: output = HMAC-SHA256(PRK, info || 0x01) truncated to length
        require(length <= HASH_LEN) { "requested length $length exceeds single HMAC output" }
        val expanded = hmacSha256(prk, info + byteArrayOf(0x01))
        return expanded.copyOf(length)
    }

    private fun hmacSha256(key: ByteArray, data: ByteArray): ByteArray {
        val mac = Mac.getInstance("HmacSHA256")
        mac.init(SecretKeySpec(key, "HmacSHA256"))
        return mac.doFinal(data)
    }
}
