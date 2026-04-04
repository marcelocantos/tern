---- MODULE PathSwitch ----
\* Path-switching protocol with exponential backoff.
\* Verifies transport correctness and backoff properties.

EXTENDS Integers, Sequences, FiniteSets, TLC

\* --- States ---

CONSTANTS backend_RelayConnected, backend_LANOffered,
          backend_LANActive, backend_LANDegraded, backend_RelayBackoff

CONSTANTS client_RelayConnected, client_LANConnecting,
          client_LANVerifying, client_LANActive, client_RelayFallback

CONSTANTS relay_Idle, relay_BackendRegistered, relay_Bridged

\* --- Variables ---

VARIABLES
    backend_state,
    client_state,
    relay_state,
    chan_b2c,          \* backend -> client
    chan_c2b,          \* client -> backend
    backend_path,      \* "relay" or "lan"
    client_path,       \* "relay" or "lan"
    challenge,         \* backend's challenge value
    offer_challenge,   \* client's received challenge
    ping_failures,     \* consecutive failed pings
    backoff_level      \* exponential backoff level (0 = immediate)

vars == <<backend_state, client_state, relay_state,
          chan_b2c, chan_c2b,
          backend_path, client_path,
          challenge, offer_challenge,
          ping_failures, backoff_level>>

CONSTANTS MaxChanLen, MaxPingFailures, MaxBackoffLevel

\* --- Init ---

Init ==
    /\ backend_state = backend_RelayConnected
    /\ client_state = client_RelayConnected
    /\ relay_state = relay_Bridged
    /\ chan_b2c = <<>>
    /\ chan_c2b = <<>>
    /\ backend_path = "relay"
    /\ client_path = "relay"
    /\ challenge = "c1"
    /\ offer_challenge = "none"
    /\ ping_failures = 0
    /\ backoff_level = 0

CanSend(ch) == Len(ch) < MaxChanLen
Min(a, b) == IF a < b THEN a ELSE b

\* ================================================================
\* Backend transitions
\* ================================================================

\* First LAN advertisement (no backoff — immediate).
BackendOfferLAN ==
    /\ backend_state = backend_RelayConnected
    /\ CanSend(chan_b2c)
    /\ chan_b2c' = Append(chan_b2c, [type |-> "lan_offer", challenge |-> challenge])
    /\ backend_state' = backend_LANOffered
    /\ UNCHANGED <<client_state, relay_state, chan_c2b,
                   backend_path, client_path, challenge, offer_challenge,
                   ping_failures, backoff_level>>

\* LAN verified — challenge matches. Reset backoff.
BackendVerifyOK ==
    /\ backend_state = backend_LANOffered
    /\ Len(chan_c2b) > 0
    /\ Head(chan_c2b).type = "lan_verify"
    /\ Head(chan_c2b).challenge = challenge
    /\ chan_c2b' = Tail(chan_c2b)
    /\ CanSend(chan_b2c)
    /\ chan_b2c' = Append(chan_b2c, [type |-> "lan_confirm"])
    /\ backend_state' = backend_LANActive
    /\ backend_path' = "lan"
    /\ ping_failures' = 0
    /\ backoff_level' = 0
    /\ UNCHANGED <<client_state, relay_state, client_path,
                   challenge, offer_challenge>>

\* LAN verify — challenge mismatch.
BackendVerifyFail ==
    /\ backend_state = backend_LANOffered
    /\ Len(chan_c2b) > 0
    /\ Head(chan_c2b).type = "lan_verify"
    /\ Head(chan_c2b).challenge /= challenge
    /\ chan_c2b' = Tail(chan_c2b)
    /\ backend_state' = backend_RelayConnected
    /\ UNCHANGED <<client_state, relay_state, chan_b2c,
                   backend_path, client_path, challenge, offer_challenge,
                   ping_failures, backoff_level>>

\* Offer times out — increase backoff.
BackendOfferTimeout ==
    /\ backend_state = backend_LANOffered
    /\ backend_state' = backend_RelayBackoff
    /\ backoff_level' = Min(backoff_level + 1, MaxBackoffLevel)
    /\ UNCHANGED <<client_state, relay_state, chan_b2c, chan_c2b,
                   backend_path, client_path, challenge, offer_challenge,
                   ping_failures>>

\* Ping on healthy LAN.
BackendPing ==
    /\ backend_state = backend_LANActive
    /\ CanSend(chan_b2c)
    /\ chan_b2c' = Append(chan_b2c, [type |-> "ping"])
    /\ UNCHANGED <<backend_state, client_state, relay_state, chan_c2b,
                   backend_path, client_path, challenge, offer_challenge,
                   ping_failures, backoff_level>>

\* Ping timeout — first failure.
BackendPingTimeout ==
    /\ backend_state = backend_LANActive
    /\ ping_failures' = 1
    /\ backend_state' = backend_LANDegraded
    /\ UNCHANGED <<client_state, relay_state, chan_b2c, chan_c2b,
                   backend_path, client_path, challenge, offer_challenge,
                   backoff_level>>

\* Pong received while degraded — recover.
BackendRecoverPong ==
    /\ backend_state = backend_LANDegraded
    /\ Len(chan_c2b) > 0
    /\ Head(chan_c2b).type = "pong"
    /\ chan_c2b' = Tail(chan_c2b)
    /\ backend_state' = backend_LANActive
    /\ ping_failures' = 0
    /\ UNCHANGED <<client_state, relay_state, chan_b2c,
                   backend_path, client_path, challenge, offer_challenge,
                   backoff_level>>

\* Degraded timeout — under max failures, keep trying.
BackendDegradedTimeoutContinue ==
    /\ backend_state = backend_LANDegraded
    /\ ping_failures + 1 < MaxPingFailures
    /\ ping_failures' = ping_failures + 1
    /\ UNCHANGED <<backend_state, client_state, relay_state,
                   chan_b2c, chan_c2b,
                   backend_path, client_path, challenge, offer_challenge,
                   backoff_level>>

\* Degraded timeout — max failures reached, fall back with backoff.
BackendDegradedFallback ==
    /\ backend_state = backend_LANDegraded
    /\ ping_failures + 1 >= MaxPingFailures
    /\ backend_state' = backend_RelayBackoff
    /\ backend_path' = "relay"
    /\ ping_failures' = 0
    /\ backoff_level' = 1
    /\ chan_b2c' = <<>>
    /\ chan_c2b' = <<>>
    /\ UNCHANGED <<client_state, relay_state,
                   client_path, challenge, offer_challenge>>

\* Backoff expired — re-advertise LAN.
BackendBackoffExpired ==
    /\ backend_state = backend_RelayBackoff
    /\ CanSend(chan_b2c)
    /\ chan_b2c' = Append(chan_b2c, [type |-> "lan_offer", challenge |-> challenge])
    /\ backend_state' = backend_LANOffered
    /\ UNCHANGED <<client_state, relay_state, chan_c2b,
                   backend_path, client_path, challenge, offer_challenge,
                   ping_failures, backoff_level>>

\* ================================================================
\* Client transitions
\* ================================================================

\* Receive LAN offer.
ClientRecvOffer ==
    /\ client_state = client_RelayConnected
    /\ Len(chan_b2c) > 0
    /\ Head(chan_b2c).type = "lan_offer"
    /\ offer_challenge' = Head(chan_b2c).challenge
    /\ chan_b2c' = Tail(chan_b2c)
    /\ client_state' = client_LANConnecting
    /\ UNCHANGED <<backend_state, relay_state, chan_c2b,
                   backend_path, client_path, challenge,
                   ping_failures, backoff_level>>

\* Dial OK — send verify.
ClientDialOK ==
    /\ client_state = client_LANConnecting
    /\ CanSend(chan_c2b)
    /\ chan_c2b' = Append(chan_c2b, [type |-> "lan_verify", challenge |-> offer_challenge])
    /\ client_state' = client_LANVerifying
    /\ UNCHANGED <<backend_state, relay_state, chan_b2c,
                   backend_path, client_path, challenge, offer_challenge,
                   ping_failures, backoff_level>>

\* Dial fails.
ClientDialFail ==
    /\ client_state = client_LANConnecting
    /\ client_state' = client_RelayConnected
    /\ UNCHANGED <<backend_state, relay_state, chan_b2c, chan_c2b,
                   backend_path, client_path, challenge, offer_challenge,
                   ping_failures, backoff_level>>

\* Receive LAN confirm.
ClientConfirm ==
    /\ client_state = client_LANVerifying
    /\ Len(chan_b2c) > 0
    /\ Head(chan_b2c).type = "lan_confirm"
    /\ chan_b2c' = Tail(chan_b2c)
    /\ client_state' = client_LANActive
    /\ client_path' = "lan"
    /\ UNCHANGED <<backend_state, relay_state, chan_c2b,
                   backend_path, challenge, offer_challenge,
                   ping_failures, backoff_level>>

\* Verify timeout — drain stale messages.
ClientVerifyTimeout ==
    /\ client_state = client_LANVerifying
    /\ client_state' = client_RelayConnected
    /\ chan_b2c' = <<>>
    /\ UNCHANGED <<backend_state, relay_state, chan_c2b,
                   backend_path, client_path, challenge, offer_challenge,
                   ping_failures, backoff_level>>

\* Receive ping — respond with pong.
ClientPong ==
    /\ client_state = client_LANActive
    /\ Len(chan_b2c) > 0
    /\ Head(chan_b2c).type = "ping"
    /\ chan_b2c' = Tail(chan_b2c)
    /\ CanSend(chan_c2b)
    /\ chan_c2b' = Append(chan_c2b, [type |-> "pong"])
    /\ UNCHANGED <<backend_state, client_state, relay_state,
                   backend_path, client_path, challenge, offer_challenge,
                   ping_failures, backoff_level>>

\* LAN error — fall back to relay, drain channels.
ClientLANError ==
    /\ client_state = client_LANActive
    /\ client_state' = client_RelayFallback
    /\ client_path' = "relay"
    /\ chan_b2c' = <<>>
    /\ chan_c2b' = <<>>
    /\ UNCHANGED <<backend_state, relay_state,
                   backend_path, challenge, offer_challenge,
                   ping_failures, backoff_level>>

\* Relay fallback completes.
ClientRelayOK ==
    /\ client_state = client_RelayFallback
    /\ client_state' = client_RelayConnected
    /\ UNCHANGED <<backend_state, relay_state, chan_b2c, chan_c2b,
                   backend_path, client_path, challenge, offer_challenge,
                   ping_failures, backoff_level>>

\* New LAN offer while already on LAN (address change).
ClientNewOfferOnLAN ==
    /\ client_state = client_LANActive
    /\ Len(chan_b2c) > 0
    /\ Head(chan_b2c).type = "lan_offer"
    /\ offer_challenge' = Head(chan_b2c).challenge
    /\ chan_b2c' = Tail(chan_b2c)
    /\ client_state' = client_LANConnecting
    /\ UNCHANGED <<backend_state, relay_state, chan_c2b,
                   backend_path, client_path, challenge,
                   ping_failures, backoff_level>>

\* ================================================================
\* Next state
\* ================================================================

Next ==
    \/ BackendOfferLAN
    \/ BackendVerifyOK
    \/ BackendVerifyFail
    \/ BackendOfferTimeout
    \/ BackendPing
    \/ BackendPingTimeout
    \/ BackendRecoverPong
    \/ BackendDegradedTimeoutContinue
    \/ BackendDegradedFallback
    \/ BackendBackoffExpired
    \/ ClientRecvOffer
    \/ ClientDialOK
    \/ ClientDialFail
    \/ ClientConfirm
    \/ ClientVerifyTimeout
    \/ ClientPong
    \/ ClientLANError
    \/ ClientRelayOK
    \/ ClientNewOfferOnLAN

Spec == Init /\ [][Next]_vars /\ WF_vars(Next)

\* ================================================================
\* Invariants
\* ================================================================

\* Both paths are always valid values.
PathConsistency ==
    /\ backend_path \in {"relay", "lan"}
    /\ client_path \in {"relay", "lan"}

\* Relay stays connected (session established at init).
RelayStaysConnected ==
    relay_state \in {relay_BackendRegistered, relay_Bridged}

\* LAN only active after successful verification.
LANRequiresVerification ==
    (backend_state = backend_LANActive /\ client_state = client_LANActive)
        => offer_challenge = challenge

\* Channels never overflow.
ChannelsBounded ==
    /\ Len(chan_b2c) <= MaxChanLen
    /\ Len(chan_c2b) <= MaxChanLen

\* Ping failures stay bounded.
PingFailuresBounded ==
    ping_failures <= MaxPingFailures

\* Backoff level never exceeds the cap.
BackoffBounded ==
    backoff_level <= MaxBackoffLevel

\* Successful LAN establishment always resets backoff.
BackoffResetsOnSuccess ==
    backend_state = backend_LANActive => backoff_level = 0

\* Fallback always enters backoff (never gets stuck).
FallbackEntersBackoff ==
    (backend_state = backend_RelayBackoff) => backoff_level >= 1

====
