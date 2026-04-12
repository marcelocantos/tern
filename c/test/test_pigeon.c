// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

#include "pigeon.h"
#include <assert.h>
#include <stdio.h>
#include <string.h>
#include <sodium.h>

static int tests_run = 0;
static int tests_passed = 0;

#define TEST(name) \
    do { \
        tests_run++; \
        printf("  %-50s", name); \
    } while (0)

#define PASS() \
    do { \
        tests_passed++; \
        printf("OK\n"); \
    } while (0)

#define FAIL(msg) \
    do { \
        printf("FAIL: %s\n", msg); \
    } while (0)

// --- Keypair generation ---

static void test_keypair(void)
{
    TEST("keypair generation");
    pigeon_keypair kp;
    int ret = pigeon_generate_keypair(&kp);
    if (ret != 0) { FAIL("generate returned error"); return; }

    // Public key should not be all zeroes.
    uint8_t zero[32] = {0};
    if (memcmp(kp.public_key, zero, 32) == 0) { FAIL("public key is zero"); return; }
    if (memcmp(kp.private_key, zero, 32) == 0) { FAIL("private key is zero"); return; }

    // Two keypairs should differ.
    pigeon_keypair kp2;
    pigeon_generate_keypair(&kp2);
    if (memcmp(kp.private_key, kp2.private_key, 32) == 0) { FAIL("duplicate keys"); return; }

    PASS();
}

// --- Session key derivation (ECDH + HKDF) ---

static void test_session_key_derivation(void)
{
    TEST("session key derivation (ECDH + HKDF)");
    pigeon_keypair alice, bob;
    pigeon_generate_keypair(&alice);
    pigeon_generate_keypair(&bob);

    const uint8_t info[] = "test-session";
    uint8_t alice_key[32], bob_key[32];

    int ret1 = pigeon_derive_session_key(alice.private_key, bob.public_key,
                                          info, sizeof(info) - 1, alice_key);
    int ret2 = pigeon_derive_session_key(bob.private_key, alice.public_key,
                                          info, sizeof(info) - 1, bob_key);
    if (ret1 != 0 || ret2 != 0) { FAIL("derivation error"); return; }

    // Both sides should derive the same key.
    if (memcmp(alice_key, bob_key, 32) != 0) { FAIL("keys don't match"); return; }

    // Different info should produce different keys.
    uint8_t other_key[32];
    const uint8_t info2[] = "other-session";
    pigeon_derive_session_key(alice.private_key, bob.public_key,
                              info2, sizeof(info2) - 1, other_key);
    if (memcmp(alice_key, other_key, 32) == 0) { FAIL("different info same key"); return; }

    PASS();
}

// --- Confirmation code ---

static void test_confirmation_code(void)
{
    TEST("confirmation code (order-independent)");
    pigeon_keypair alice, bob;
    pigeon_generate_keypair(&alice);
    pigeon_generate_keypair(&bob);

    char code_ab[7], code_ba[7];
    int ret1 = pigeon_derive_confirmation_code(alice.public_key, bob.public_key, code_ab);
    int ret2 = pigeon_derive_confirmation_code(bob.public_key, alice.public_key, code_ba);
    if (ret1 != 0 || ret2 != 0) { FAIL("derivation error"); return; }

    // Same code regardless of order.
    if (strcmp(code_ab, code_ba) != 0) { FAIL("codes differ"); return; }

    // 6 digits, null-terminated.
    if (strlen(code_ab) != 6) { FAIL("wrong length"); return; }
    for (int i = 0; i < 6; i++) {
        if (code_ab[i] < '0' || code_ab[i] > '9') { FAIL("non-digit"); return; }
    }

    PASS();
}

// --- Channel encrypt/decrypt round-trip ---

static void test_channel_roundtrip(void)
{
    TEST("channel encrypt/decrypt round-trip");
    pigeon_keypair alice, bob;
    pigeon_generate_keypair(&alice);
    pigeon_generate_keypair(&bob);

    // Derive send/recv keys for alice→bob direction.
    uint8_t a2b[32], b2a[32];
    const uint8_t info_a2b[] = "alice-to-bob";
    const uint8_t info_b2a[] = "bob-to-alice";
    pigeon_derive_session_key(alice.private_key, bob.public_key, info_a2b, sizeof(info_a2b) - 1, a2b);
    pigeon_derive_session_key(alice.private_key, bob.public_key, info_b2a, sizeof(info_b2a) - 1, b2a);

    pigeon_channel alice_ch, bob_ch;
    pigeon_channel_init(&alice_ch, a2b, b2a, PIGEON_MODE_STRICT);
    pigeon_channel_init(&bob_ch, b2a, a2b, PIGEON_MODE_STRICT);

    const char *msg = "hello from pigeon";
    uint8_t ciphertext[256], plaintext[256];

    int ct_len = pigeon_channel_encrypt(&alice_ch, (const uint8_t *)msg, strlen(msg),
                                         ciphertext, sizeof(ciphertext));
    if (ct_len < 0) { FAIL("encrypt failed"); return; }

    // Ciphertext should be longer than plaintext (8-byte seq + 16-byte tag).
    if ((size_t)ct_len != 8 + strlen(msg) + 16) { FAIL("wrong ciphertext length"); return; }

    int pt_len = pigeon_channel_decrypt(&bob_ch, ciphertext, (size_t)ct_len,
                                         plaintext, sizeof(plaintext));
    if (pt_len < 0) { FAIL("decrypt failed"); return; }
    if ((size_t)pt_len != strlen(msg)) { FAIL("wrong plaintext length"); return; }
    if (memcmp(plaintext, msg, (size_t)pt_len) != 0) { FAIL("plaintext mismatch"); return; }

    PASS();
}

// --- Symmetric channel ---

static void test_symmetric_channel(void)
{
    TEST("symmetric channel (client/server)");
    uint8_t master[32];
    randombytes_buf(master, 32);

    pigeon_channel client_ch, server_ch;
    int ret1 = pigeon_channel_init_symmetric(&client_ch, master, false);
    int ret2 = pigeon_channel_init_symmetric(&server_ch, master, true);
    if (ret1 != 0 || ret2 != 0) { FAIL("init error"); return; }

    // Client → server.
    const char *msg = "client says hi";
    uint8_t ct[256], pt[256];
    int ct_len = pigeon_channel_encrypt(&client_ch, (const uint8_t *)msg, strlen(msg), ct, sizeof(ct));
    int pt_len = pigeon_channel_decrypt(&server_ch, ct, (size_t)ct_len, pt, sizeof(pt));
    if (pt_len < 0 || (size_t)pt_len != strlen(msg) || memcmp(pt, msg, (size_t)pt_len) != 0) {
        FAIL("client→server failed"); return;
    }

    // Server → client.
    const char *reply = "server replies";
    ct_len = pigeon_channel_encrypt(&server_ch, (const uint8_t *)reply, strlen(reply), ct, sizeof(ct));
    pt_len = pigeon_channel_decrypt(&client_ch, ct, (size_t)ct_len, pt, sizeof(pt));
    if (pt_len < 0 || (size_t)pt_len != strlen(reply) || memcmp(pt, reply, (size_t)pt_len) != 0) {
        FAIL("server→client failed"); return;
    }

    PASS();
}

// --- Sequence number enforcement ---

static void test_sequence_strict(void)
{
    TEST("strict mode rejects out-of-order");
    uint8_t key[32];
    randombytes_buf(key, 32);

    pigeon_channel send_ch, recv_ch;
    pigeon_channel_init(&send_ch, key, key, PIGEON_MODE_STRICT);
    pigeon_channel_init(&recv_ch, key, key, PIGEON_MODE_STRICT);

    const char *msg = "seq test";
    uint8_t ct1[256], ct2[256], pt[256];

    pigeon_channel_encrypt(&send_ch, (const uint8_t *)msg, strlen(msg), ct1, sizeof(ct1));
    int ct2_len = pigeon_channel_encrypt(&send_ch, (const uint8_t *)msg, strlen(msg), ct2, sizeof(ct2));

    // Decrypt ct2 first (seq=1) — should fail because recv expects seq=0.
    int ret = pigeon_channel_decrypt(&recv_ch, ct2, (size_t)ct2_len, pt, sizeof(pt));
    if (ret >= 0) { FAIL("should have rejected out-of-order"); return; }

    PASS();
}

// --- Wire framing ---

static void test_framing(void)
{
    TEST("wire framing (4-byte BE length prefix)");
    const char *payload = "test payload";
    size_t len = strlen(payload);
    uint8_t buf[256];

    int frame_len = pigeon_frame_message((const uint8_t *)payload, len, buf, sizeof(buf));
    if (frame_len != (int)(4 + len)) { FAIL("wrong frame length"); return; }

    uint32_t decoded_len = pigeon_read_frame_length(buf);
    if (decoded_len != len) { FAIL("decoded length mismatch"); return; }
    if (memcmp(buf + 4, payload, len) != 0) { FAIL("payload mismatch"); return; }

    PASS();
}

// --- State machine init ---

static void test_state_machine_init(void)
{
    TEST("pairing machine init (ios actor)");
    pigeon_ios_machine m;
    pigeon_ios_machine_init(&m);
    if (m.state != PIGEON_APP_IDLE) { FAIL("wrong initial state"); return; }

    PASS();
}

// --- pigeon_ctx init ---

static void test_ctx_init(void)
{
    TEST("pigeon_ctx init");
    // Use a smaller buffer size for the test to avoid stack overflow.
    // The real pigeon_ctx with PIGEON_MAX_MSG=1MiB is too large for the stack.
    // Just verify that init works with a NULL transport.
    static pigeon_ctx ctx;
    pigeon_init(&ctx, NULL);
    if (ctx.pairing.state != PIGEON_APP_IDLE) { FAIL("wrong pairing state"); return; }
    if (ctx.stream_channel.send_seq != 0) { FAIL("send_seq not zero"); return; }

    PASS();
}

int main(void)
{
    if (sodium_init() < 0) {
        fprintf(stderr, "sodium_init failed\n");
        return 1;
    }

    printf("pigeon C library tests\n\n");

    test_keypair();
    test_session_key_derivation();
    test_confirmation_code();
    test_channel_roundtrip();
    test_symmetric_channel();
    test_sequence_strict();
    test_framing();
    test_state_machine_init();
    test_ctx_init();

    printf("\n%d/%d tests passed\n", tests_passed, tests_run);
    return tests_passed == tests_run ? 0 : 1;
}
