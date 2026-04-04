---- MODULE Session_Pairing ----
\* Auto-generated from protocol definition. Do not edit.
\* Phase: Pairing

EXTENDS Integers, Sequences, FiniteSets, TLC

\* States for backend
backend_Idle == "backend_Idle"
backend_GenerateToken == "backend_GenerateToken"
backend_RegisterRelay == "backend_RegisterRelay"
backend_WaitingForClient == "backend_WaitingForClient"
backend_DeriveSecret == "backend_DeriveSecret"
backend_SendAck == "backend_SendAck"
backend_WaitingForCode == "backend_WaitingForCode"
backend_ValidateCode == "backend_ValidateCode"
backend_StorePaired == "backend_StorePaired"
backend_Paired == "backend_Paired"
backend_AuthCheck == "backend_AuthCheck"
backend_SessionActive == "backend_SessionActive"
backend_RelayConnected == "backend_RelayConnected"

\* States for client
client_Idle == "client_Idle"
client_ScanQR == "client_ScanQR"
client_ConnectRelay == "client_ConnectRelay"
client_GenKeyPair == "client_GenKeyPair"
client_WaitAck == "client_WaitAck"
client_E2EReady == "client_E2EReady"
client_ShowCode == "client_ShowCode"
client_WaitPairComplete == "client_WaitPairComplete"
client_Paired == "client_Paired"
client_Reconnect == "client_Reconnect"
client_SendAuth == "client_SendAuth"
client_SessionActive == "client_SessionActive"
client_RelayConnected == "client_RelayConnected"

\* States for relay
relay_Idle == "relay_Idle"
relay_BackendRegistered == "relay_BackendRegistered"

\* Message types
MSG_pair_hello == "pair_hello" \* client -> backend (ECDH pubkey + pairing token)
MSG_pair_hello_ack == "pair_hello_ack" \* backend -> client (ECDH pubkey)
MSG_pair_confirm == "pair_confirm" \* backend -> client (signal to compute and display code)
MSG_pair_complete == "pair_complete" \* backend -> client (encrypted device secret)
MSG_auth_request == "auth_request" \* client -> backend (encrypted auth with nonce)
MSG_auth_ok == "auth_ok" \* backend -> client (session established)
MSG_lan_offer == "lan_offer" \* backend -> client (LAN address + challenge (sent via relay))
MSG_lan_verify == "lan_verify" \* client -> backend (challenge response + instance ID (sent via LAN))
MSG_lan_confirm == "lan_confirm" \* backend -> client (LAN verified, path is live (sent via LAN))
MSG_path_ping == "path_ping" \* backend -> client (health check on active direct path)
MSG_path_pong == "path_pong" \* client -> backend (health check response)

\* Helper operators
\* deterministic ordering for ECDH
KeyRank(k) == CASE k = "adv_pub" -> 0 [] k = "client_pub" -> 1 [] k = "backend_pub" -> 2 [] OTHER -> 3
\* symbolic ECDH
DeriveKey(a, b) == IF KeyRank(a) <= KeyRank(b) THEN <<"ecdh", a, b>> ELSE <<"ecdh", b, a>>
\* confirmation code from pubkeys
DeriveCode(a, b) == IF KeyRank(a) <= KeyRank(b) THEN <<"code", a, b>> ELSE <<"code", b, a>>
\* minimum of two values
Min(a, b) == IF a < b THEN a ELSE b

(*--algorithm Session_Pairing

variables
    backend_state = backend_Idle,
    client_state = client_Idle,
    relay_state = relay_Idle,
    chan_client_backend = <<>>,
    chan_backend_client = <<>>,
    adversary_knowledge = {},
    \* pairing token currently in play
    current_token = "none",
    \* set of valid (non-revoked) tokens
    active_tokens = {},
    \* set of revoked tokens
    used_tokens = {},
    \* backend ECDH public key
    backend_ecdh_pub = "none",
    \* pubkey backend received in pair_hello
    received_client_pub = "none",
    \* pubkey client received in pair_hello_ack
    received_backend_pub = "none",
    \* ECDH key derived by backend
    backend_shared_key = <<"none">>,
    \* ECDH key derived by client
    client_shared_key = <<"none">>,
    \* code computed by backend
    backend_code = <<"none">>,
    \* code computed by client
    client_code = <<"none">>,
    \* code entered via CLI
    received_code = <<"none">>,
    \* staging for CLI code input
    cli_entered_code = <<"none">>,
    \* failed code submission attempts
    code_attempts = 0,
    \* persistent device secret
    device_secret = "none",
    \* device IDs that completed pairing
    paired_devices = {},
    \* device_id from auth_request
    received_device_id = "none",
    \* set of consumed auth nonces
    auth_nonces_used = {},
    \* nonce from auth_request
    received_auth_nonce = "none",
    \* whether QR code has been shown (local backend state)
    qr_displayed = FALSE,
    \* last received message (staging)
    recv_msg = [type |-> "none"],
    \* encryption keys the adversary knows
    adversary_keys = {},
    \* adversary's ECDH public key
    adv_ecdh_pub = "adv_pub",
    \* real client pubkey saved during MitM
    adv_saved_client_pub = "none",
    \* real backend pubkey saved during MitM
    adv_saved_server_pub = "none";

fair process backend = 1
begin
  backend_loop:
  while TRUE do
    either
      \* Idle -> GenerateToken (cli_init_pair)
      await backend_state = backend_Idle;
      current_token := "tok_1";
      active_tokens := active_tokens \union {"tok_1"};
      backend_state := backend_GenerateToken;
    or
      \* GenerateToken -> RegisterRelay (token_created)
      await backend_state = backend_GenerateToken;
      backend_state := backend_RegisterRelay;
    or
      \* RegisterRelay -> WaitingForClient (relay_registered)
      await backend_state = backend_RegisterRelay;
      qr_displayed := TRUE;
      backend_state := backend_WaitingForClient;
    or
      \* WaitingForClient -> DeriveSecret on pair_hello
      await backend_state = backend_WaitingForClient /\ Len(chan_client_backend) > 0 /\ Head(chan_client_backend).type = MSG_pair_hello /\ (Head(chan_client_backend).token \in active_tokens);
      recv_msg := Head(chan_client_backend);
      chan_client_backend := Tail(chan_client_backend);
      received_client_pub := recv_msg.pubkey;
      backend_ecdh_pub := "backend_pub";
      backend_shared_key := DeriveKey("backend_pub", recv_msg.pubkey);
      backend_code := DeriveCode("backend_pub", recv_msg.pubkey);
      backend_state := backend_DeriveSecret;
    or
      \* WaitingForClient -> Idle on pair_hello
      await backend_state = backend_WaitingForClient /\ Len(chan_client_backend) > 0 /\ Head(chan_client_backend).type = MSG_pair_hello /\ (Head(chan_client_backend).token \notin active_tokens);
      recv_msg := Head(chan_client_backend);
      chan_client_backend := Tail(chan_client_backend);
      backend_state := backend_Idle;
    or
      \* DeriveSecret -> SendAck (ecdh_complete)
      await backend_state = backend_DeriveSecret;
      chan_backend_client := Append(chan_backend_client, [type |-> MSG_pair_hello_ack, pubkey |-> backend_ecdh_pub]);
      backend_state := backend_SendAck;
    or
      \* SendAck -> WaitingForCode (signal_code_display)
      await backend_state = backend_SendAck;
      chan_backend_client := Append(chan_backend_client, [type |-> MSG_pair_confirm]);
      backend_state := backend_WaitingForCode;
    or
      \* WaitingForCode -> ValidateCode (cli_code_entered)
      await backend_state = backend_WaitingForCode;
      received_code := cli_entered_code;
      backend_state := backend_ValidateCode;
    or
      \* ValidateCode -> StorePaired (check_code)
      await backend_state = backend_ValidateCode /\ (received_code = backend_code);
      backend_state := backend_StorePaired;
    or
      \* ValidateCode -> Idle (check_code)
      await backend_state = backend_ValidateCode /\ (received_code /= backend_code);
      code_attempts := code_attempts + 1;
      backend_state := backend_Idle;
    or
      \* StorePaired -> Paired (finalise)
      await backend_state = backend_StorePaired;
      chan_backend_client := Append(chan_backend_client, [type |-> MSG_pair_complete, key |-> backend_shared_key, secret |-> "dev_secret_1"]);
      device_secret := "dev_secret_1";
      paired_devices := paired_devices \union {"device_1"};
      active_tokens := active_tokens \ {current_token};
      used_tokens := used_tokens \union {current_token};
      backend_state := backend_Paired;
    or
      \* Paired -> AuthCheck on auth_request
      await backend_state = backend_Paired /\ Len(chan_client_backend) > 0 /\ Head(chan_client_backend).type = MSG_auth_request;
      recv_msg := Head(chan_client_backend);
      chan_client_backend := Tail(chan_client_backend);
      received_device_id := recv_msg.device_id;
      received_auth_nonce := recv_msg.nonce;
      backend_state := backend_AuthCheck;
    or
      \* AuthCheck -> SessionActive (verify)
      await backend_state = backend_AuthCheck /\ (received_device_id \in paired_devices);
      chan_backend_client := Append(chan_backend_client, [type |-> MSG_auth_ok]);
      auth_nonces_used := auth_nonces_used \union {received_auth_nonce};
      backend_state := backend_SessionActive;
    or
      \* AuthCheck -> Idle (verify)
      await backend_state = backend_AuthCheck /\ (received_device_id \notin paired_devices);
      backend_state := backend_Idle;
    or
      \* SessionActive -> RelayConnected (session_established)
      await backend_state = backend_SessionActive;
      backend_state := backend_RelayConnected;
    or
      \* RelayConnected -> Paired (disconnect)
      await backend_state = backend_RelayConnected;
      backend_state := backend_Paired;
    end either;
  end while;
end process;

fair process client = 2
begin
  client_loop:
  while TRUE do
    either
      \* Idle -> ScanQR (user_scans_qr)
      await client_state = client_Idle;
      client_state := client_ScanQR;
    or
      \* ScanQR -> ConnectRelay (qr_parsed)
      await client_state = client_ScanQR;
      client_state := client_ConnectRelay;
    or
      \* ConnectRelay -> GenKeyPair (relay_connected)
      await client_state = client_ConnectRelay;
      client_state := client_GenKeyPair;
    or
      \* GenKeyPair -> WaitAck (key_pair_generated)
      await client_state = client_GenKeyPair;
      chan_client_backend := Append(chan_client_backend, [type |-> MSG_pair_hello, pubkey |-> "client_pub", token |-> current_token]);
      client_state := client_WaitAck;
    or
      \* WaitAck -> E2EReady on pair_hello_ack
      await client_state = client_WaitAck /\ Len(chan_backend_client) > 0 /\ Head(chan_backend_client).type = MSG_pair_hello_ack;
      recv_msg := Head(chan_backend_client);
      chan_backend_client := Tail(chan_backend_client);
      received_backend_pub := recv_msg.pubkey;
      client_shared_key := DeriveKey("client_pub", recv_msg.pubkey);
      client_state := client_E2EReady;
    or
      \* E2EReady -> ShowCode on pair_confirm
      await client_state = client_E2EReady /\ Len(chan_backend_client) > 0 /\ Head(chan_backend_client).type = MSG_pair_confirm;
      recv_msg := Head(chan_backend_client);
      chan_backend_client := Tail(chan_backend_client);
      client_code := DeriveCode(received_backend_pub, "client_pub");
      client_state := client_ShowCode;
    or
      \* ShowCode -> WaitPairComplete (code_displayed)
      await client_state = client_ShowCode;
      client_state := client_WaitPairComplete;
    or
      \* WaitPairComplete -> Paired on pair_complete
      await client_state = client_WaitPairComplete /\ Len(chan_backend_client) > 0 /\ Head(chan_backend_client).type = MSG_pair_complete;
      recv_msg := Head(chan_backend_client);
      chan_backend_client := Tail(chan_backend_client);
      client_state := client_Paired;
    or
      \* Paired -> Reconnect (app_launch)
      await client_state = client_Paired;
      client_state := client_Reconnect;
    or
      \* Reconnect -> SendAuth (relay_connected)
      await client_state = client_Reconnect;
      chan_client_backend := Append(chan_client_backend, [type |-> MSG_auth_request, device_id |-> "device_1", key |-> client_shared_key, nonce |-> "nonce_1", secret |-> device_secret]);
      client_state := client_SendAuth;
    or
      \* SendAuth -> SessionActive on auth_ok
      await client_state = client_SendAuth /\ Len(chan_backend_client) > 0 /\ Head(chan_backend_client).type = MSG_auth_ok;
      recv_msg := Head(chan_backend_client);
      chan_backend_client := Tail(chan_backend_client);
      client_state := client_SessionActive;
    or
      \* SessionActive -> RelayConnected (session_established)
      await client_state = client_SessionActive;
      client_state := client_RelayConnected;
    or
      \* RelayConnected -> Paired (disconnect)
      await client_state = client_RelayConnected;
      client_state := client_Paired;
    end either;
  end while;
end process;

fair process relay = 3
begin
  relay_loop:
  while TRUE do
    either
      \* Idle -> BackendRegistered (backend_register)
      await relay_state = relay_Idle;
      relay_state := relay_BackendRegistered;
    or
      \* BackendRegistered -> Idle (backend_disconnect)
      await relay_state = relay_BackendRegistered;
      relay_state := relay_Idle;
    end either;
  end while;
end process;

\* Dolev-Yao adversary: controls the network.
fair process Adversary = 4
begin
  adv_loop:
  while TRUE do
    await backend_state \notin {backend_RelayConnected, backend_LANOffered, backend_LANActive, backend_LANDegraded, backend_RelayBackoff};
    either
      skip \* no-op: honest relay
    or
      \* Eavesdrop on client -> backend
      await Len(chan_client_backend) > 0;
      adversary_knowledge := adversary_knowledge \union {Head(chan_client_backend)};
    or
      \* Drop from client -> backend
      await Len(chan_client_backend) > 0;
      chan_client_backend := Tail(chan_client_backend);
    or
      \* Replay into client -> backend
      await adversary_knowledge /= {} /\ Len(chan_client_backend) < 3;
      with msg \in adversary_knowledge do
        chan_client_backend := Append(chan_client_backend, msg);
      end with;
    or
      \* Eavesdrop on backend -> client
      await Len(chan_backend_client) > 0;
      adversary_knowledge := adversary_knowledge \union {Head(chan_backend_client)};
    or
      \* Drop from backend -> client
      await Len(chan_backend_client) > 0;
      chan_backend_client := Tail(chan_backend_client);
    or
      \* Replay into backend -> client
      await adversary_knowledge /= {} /\ Len(chan_backend_client) < 3;
      with msg \in adversary_knowledge do
        chan_backend_client := Append(chan_backend_client, msg);
      end with;
    or
      \* QR_shoulder_surf: observe QR code content
      await current_token /= "none";
      adversary_knowledge := adversary_knowledge \union {[type |-> "qr_token", token |-> current_token]};
    or
      \* MitM_pair_hello: intercept pair_hello and substitute adversary pubkey
      await Len(chan_client_backend) > 0 /\ Head(chan_client_backend).type = MSG_pair_hello;
      adv_saved_client_pub := Head(chan_client_backend).pubkey;
      chan_client_backend := <<[type |-> MSG_pair_hello, token |-> Head(chan_client_backend).token, pubkey |-> adv_ecdh_pub]>> \o Tail(chan_client_backend);
    or
      \* MitM_pair_hello_ack: intercept pair_hello_ack and substitute adversary pubkey
      await Len(chan_backend_client) > 0 /\ Head(chan_backend_client).type = MSG_pair_hello_ack;
      adv_saved_server_pub := Head(chan_backend_client).pubkey;
      adversary_keys := adversary_keys \union {DeriveKey(adv_ecdh_pub, adv_saved_server_pub), DeriveKey(adv_ecdh_pub, adv_saved_client_pub)};
      chan_backend_client := <<[type |-> MSG_pair_hello_ack, pubkey |-> adv_ecdh_pub]>> \o Tail(chan_backend_client);
    or
      \* MitM_reencrypt_secret: decrypt pair_complete with MitM key
      await Len(chan_backend_client) > 0 /\ Head(chan_backend_client).type = MSG_pair_complete /\ Head(chan_backend_client).key \in adversary_keys;
      with msg = Head(chan_backend_client) do
        adversary_knowledge := adversary_knowledge \union {[type |-> "plaintext_secret", secret |-> msg.secret]};
        chan_backend_client := <<[type |-> MSG_pair_complete, key |-> DeriveKey(adv_ecdh_pub, adv_saved_client_pub), secret |-> msg.secret]>> \o Tail(chan_backend_client);
      end with;
    or
      \* concurrent_pair: race a forged pair_hello using shoulder-surfed token
      await \E m \in adversary_knowledge : m = [type |-> "qr_token", token |-> current_token];
      await Len(chan_client_backend) < 3;
      chan_client_backend := Append(chan_client_backend, [type |-> MSG_pair_hello, token |-> current_token, pubkey |-> adv_ecdh_pub]);
    or
      \* token_bruteforce: send pair_hello with fabricated token
      await Len(chan_client_backend) < 3;
      chan_client_backend := Append(chan_client_backend, [type |-> MSG_pair_hello, token |-> "fake_token", pubkey |-> adv_ecdh_pub]);
    or
      \* code_guess: submit fabricated confirmation code
      await backend_state = backend_WaitingForCode;
      cli_entered_code := <<"guess", "000000">>;
    or
      \* session_replay: replay captured auth_request with stale nonce
      await Len(chan_client_backend) < 3;
      await \E m \in adversary_knowledge : m.type = MSG_auth_request;
      with msg \in {m \in adversary_knowledge : m.type = MSG_auth_request} do
        chan_client_backend := Append(chan_client_backend, msg);
      end with;
    end either;
  end while;
end process;

end algorithm; *)
\* BEGIN TRANSLATION
\* END TRANSLATION

\* Verification properties
\* A revoked pairing token is never accepted again
NoTokenReuse == used_tokens \intersect active_tokens = {}
\* MitM produces mismatched codes
MitMDetectedByCodeMismatch == (backend_shared_key \in adversary_keys /\ backend_code /= <<"none">> /\ client_code /= <<"none">>) => backend_code /= client_code
\* Compromised key prevents pairing completion
MitMPrevented == backend_shared_key \in adversary_keys => backend_state \notin {backend_StorePaired, backend_Paired, backend_AuthCheck, backend_SessionActive}
\* Session requires completed pairing
AuthRequiresCompletedPairing == backend_state = backend_SessionActive => received_device_id \in paired_devices
\* Each auth nonce accepted at most once
NoNonceReuse == backend_state = backend_SessionActive => received_auth_nonce \notin (auth_nonces_used \ {received_auth_nonce})
\* Adversary never learns device secret
DeviceSecretSecrecy == \A m \in adversary_knowledge : "type" \in DOMAIN m => m.type /= "plaintext_secret"
\* After fallback, backend eventually re-advertises LAN
FallbackLeadsToReadvertise == (backend_state = backend_RelayBackoff) ~> (backend_state = backend_LANOffered)
\* Degraded state eventually resolves (recovery or fallback)
DegradedLeadsToResolutionOrFallback == (backend_state = backend_LANDegraded) ~> (backend_state \in {backend_LANActive, backend_RelayBackoff})

====
