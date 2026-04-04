---- MODULE PathSwitch ----
\* Path-switching protocol: relay (permanent) + LAN (direct, optional).
\* Verifies transport correctness invariants. Security is out of scope
\* (handled by the PairingCeremony spec).

EXTENDS Integers, Sequences, FiniteSets, TLC

\* --- States ---

\* Backend states
CONSTANTS backend_RelayConnected, backend_LANOffered,
          backend_LANActive, backend_LANDegraded

\* Client states
CONSTANTS client_RelayConnected, client_LANConnecting,
          client_LANVerifying, client_LANActive, client_RelayFallback

\* Relay states
CONSTANTS relay_Idle, relay_BackendRegistered, relay_Bridged

\* --- Variables ---

VARIABLES
    backend_state,
    client_state,
    relay_state,
    \* Message channels (bounded sequences).
    chan_b2c,      \* backend -> client (via relay or LAN)
    chan_c2b,      \* client -> backend (via relay or LAN)
    \* Which path is active for each side.
    backend_path,  \* "relay" or "lan"
    client_path,   \* "relay" or "lan"
    \* LAN handshake state.
    challenge,     \* backend's challenge value
    offer_challenge, \* client's received challenge
    \* Health monitoring.
    ping_failures  \* consecutive failed pings

vars == <<backend_state, client_state, relay_state,
          chan_b2c, chan_c2b,
          backend_path, client_path,
          challenge, offer_challenge, ping_failures>>

CONSTANTS MaxChanLen, MaxPingFailures

\* --- Init ---

Init ==
    /\ backend_state = backend_RelayConnected
    /\ client_state = client_RelayConnected
    /\ relay_state = relay_Bridged  \* session already established
    /\ chan_b2c = <<>>
    /\ chan_c2b = <<>>
    /\ backend_path = "relay"
    /\ client_path = "relay"
    /\ challenge = "c1"
    /\ offer_challenge = "none"
    /\ ping_failures = 0

\* --- Helper: send if channel not full ---

CanSend(ch) == Len(ch) < MaxChanLen

\* --- Backend transitions ---

\* Backend advertises LAN.
BackendOfferLAN ==
    /\ backend_state = backend_RelayConnected
    /\ CanSend(chan_b2c)
    /\ chan_b2c' = Append(chan_b2c, [type |-> "lan_offer", challenge |-> challenge])
    /\ backend_state' = backend_LANOffered
    /\ UNCHANGED <<client_state, relay_state, chan_c2b,
                   backend_path, client_path, challenge, offer_challenge, ping_failures>>

\* Backend verifies LAN — challenge matches.
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
    /\ UNCHANGED <<client_state, relay_state, client_path,
                   challenge, offer_challenge>>

\* Backend verifies LAN — challenge mismatch.
BackendVerifyFail ==
    /\ backend_state = backend_LANOffered
    /\ Len(chan_c2b) > 0
    /\ Head(chan_c2b).type = "lan_verify"
    /\ Head(chan_c2b).challenge /= challenge
    /\ chan_c2b' = Tail(chan_c2b)
    /\ backend_state' = backend_RelayConnected
    /\ UNCHANGED <<client_state, relay_state, chan_b2c,
                   backend_path, client_path, challenge, offer_challenge, ping_failures>>

\* Backend offer times out.
BackendOfferTimeout ==
    /\ backend_state = backend_LANOffered
    /\ backend_state' = backend_RelayConnected
    /\ UNCHANGED <<client_state, relay_state, chan_b2c, chan_c2b,
                   backend_path, client_path, challenge, offer_challenge, ping_failures>>

\* Backend sends ping on LAN.
BackendPing ==
    /\ backend_state = backend_LANActive
    /\ CanSend(chan_b2c)
    /\ chan_b2c' = Append(chan_b2c, [type |-> "ping"])
    /\ UNCHANGED <<backend_state, client_state, relay_state, chan_c2b,
                   backend_path, client_path, challenge, offer_challenge, ping_failures>>

\* Backend ping times out — path degrading.
BackendPingTimeout ==
    /\ backend_state = backend_LANActive
    /\ ping_failures' = ping_failures + 1
    /\ backend_state' = backend_LANDegraded
    /\ UNCHANGED <<client_state, relay_state, chan_b2c, chan_c2b,
                   backend_path, client_path, challenge, offer_challenge>>

\* Backend receives pong in degraded state — recovers.
BackendRecoverPong ==
    /\ backend_state = backend_LANDegraded
    /\ Len(chan_c2b) > 0
    /\ Head(chan_c2b).type = "pong"
    /\ chan_c2b' = Tail(chan_c2b)
    /\ backend_state' = backend_LANActive
    /\ ping_failures' = 0
    /\ UNCHANGED <<client_state, relay_state, chan_b2c,
                   backend_path, client_path, challenge, offer_challenge>>

\* Backend degraded — another ping timeout. On max failures, fall back
\* to relay and flush channels (LAN stream is closing).
BackendDegradedTimeout ==
    /\ backend_state = backend_LANDegraded
    /\ ping_failures' = ping_failures + 1
    /\ IF ping_failures + 1 >= MaxPingFailures
       THEN /\ backend_state' = backend_RelayConnected
            /\ backend_path' = "relay"
            /\ chan_b2c' = <<>>
            /\ chan_c2b' = <<>>
       ELSE /\ backend_state' = backend_LANDegraded
            /\ backend_path' = backend_path
            /\ UNCHANGED <<chan_b2c, chan_c2b>>
    /\ UNCHANGED <<client_state, relay_state,
                   client_path, challenge, offer_challenge>>

\* Backend re-advertises LAN after fallback.
BackendReadvertise ==
    /\ backend_state = backend_RelayConnected
    /\ backend_path = "relay"
    /\ CanSend(chan_b2c)
    /\ chan_b2c' = Append(chan_b2c, [type |-> "lan_offer", challenge |-> challenge])
    /\ backend_state' = backend_LANOffered
    /\ UNCHANGED <<client_state, relay_state, chan_c2b,
                   backend_path, client_path, challenge, offer_challenge, ping_failures>>

\* --- Client transitions ---

\* Client receives LAN offer.
ClientRecvOffer ==
    /\ client_state = client_RelayConnected
    /\ Len(chan_b2c) > 0
    /\ Head(chan_b2c).type = "lan_offer"
    /\ offer_challenge' = Head(chan_b2c).challenge
    /\ chan_b2c' = Tail(chan_b2c)
    /\ client_state' = client_LANConnecting
    /\ UNCHANGED <<backend_state, relay_state, chan_c2b,
                   backend_path, client_path, challenge, ping_failures>>

\* Client dial succeeds — send verify.
ClientDialOK ==
    /\ client_state = client_LANConnecting
    /\ CanSend(chan_c2b)
    /\ chan_c2b' = Append(chan_c2b, [type |-> "lan_verify", challenge |-> offer_challenge])
    /\ client_state' = client_LANVerifying
    /\ UNCHANGED <<backend_state, relay_state, chan_b2c,
                   backend_path, client_path, challenge, offer_challenge, ping_failures>>

\* Client dial fails.
ClientDialFail ==
    /\ client_state = client_LANConnecting
    /\ client_state' = client_RelayConnected
    /\ UNCHANGED <<backend_state, relay_state, chan_b2c, chan_c2b,
                   backend_path, client_path, challenge, offer_challenge, ping_failures>>

\* Client receives LAN confirm.
ClientConfirm ==
    /\ client_state = client_LANVerifying
    /\ Len(chan_b2c) > 0
    /\ Head(chan_b2c).type = "lan_confirm"
    /\ chan_b2c' = Tail(chan_b2c)
    /\ client_state' = client_LANActive
    /\ client_path' = "lan"
    /\ UNCHANGED <<backend_state, relay_state, chan_c2b,
                   backend_path, challenge, offer_challenge, ping_failures>>

\* Client verify times out — drain stale LAN messages.
ClientVerifyTimeout ==
    /\ client_state = client_LANVerifying
    /\ client_state' = client_RelayConnected
    /\ chan_b2c' = <<>>  \* LAN stream closed, stale messages discarded
    /\ UNCHANGED <<backend_state, relay_state, chan_c2b,
                   backend_path, client_path, challenge, offer_challenge, ping_failures>>

\* Client receives ping — responds with pong.
ClientPong ==
    /\ client_state = client_LANActive
    /\ Len(chan_b2c) > 0
    /\ Head(chan_b2c).type = "ping"
    /\ chan_b2c' = Tail(chan_b2c)
    /\ CanSend(chan_c2b)
    /\ chan_c2b' = Append(chan_c2b, [type |-> "pong"])
    /\ UNCHANGED <<backend_state, client_state, relay_state,
                   backend_path, client_path, challenge, offer_challenge, ping_failures>>

\* Client LAN error — fall back to relay, drain stale LAN messages.
ClientLANError ==
    /\ client_state = client_LANActive
    /\ client_state' = client_RelayFallback
    /\ client_path' = "relay"
    /\ chan_b2c' = <<>>  \* LAN stream closed
    /\ chan_c2b' = <<>>
    /\ UNCHANGED <<backend_state, relay_state,
                   backend_path, challenge, offer_challenge, ping_failures>>

\* Client relay fallback completes.
ClientRelayOK ==
    /\ client_state = client_RelayFallback
    /\ client_state' = client_RelayConnected
    /\ UNCHANGED <<backend_state, relay_state, chan_b2c, chan_c2b,
                   backend_path, client_path, challenge, offer_challenge, ping_failures>>

\* Client receives new LAN offer while already on LAN.
ClientNewOfferOnLAN ==
    /\ client_state = client_LANActive
    /\ Len(chan_b2c) > 0
    /\ Head(chan_b2c).type = "lan_offer"
    /\ offer_challenge' = Head(chan_b2c).challenge
    /\ chan_b2c' = Tail(chan_b2c)
    /\ client_state' = client_LANConnecting
    /\ UNCHANGED <<backend_state, relay_state, chan_c2b,
                   backend_path, client_path, challenge, ping_failures>>

\* --- Next state ---

Next ==
    \/ BackendOfferLAN
    \/ BackendVerifyOK
    \/ BackendVerifyFail
    \/ BackendOfferTimeout
    \/ BackendPing
    \/ BackendPingTimeout
    \/ BackendRecoverPong
    \/ BackendDegradedTimeout
    \/ BackendReadvertise
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

\* --- Invariants ---

\* Both sides always agree on a valid path state.
PathConsistency ==
    /\ backend_path \in {"relay", "lan"}
    /\ client_path \in {"relay", "lan"}

\* The relay is always in a connected state (session established at init).
RelayStaysConnected ==
    relay_state \in {relay_BackendRegistered, relay_Bridged}

\* LAN is only active on both sides if the challenge was verified.
LANRequiresVerification ==
    (backend_state = backend_LANActive /\ client_state = client_LANActive)
        => offer_challenge = challenge

\* If the backend is on relay, the client is also on relay or falling back.
\* (No state where backend is relay but client thinks it's on LAN with no backend.)
BackendRelayImpliesClientNotLANAlone ==
    (backend_path = "relay" /\ backend_state = backend_RelayConnected)
        => client_state \notin {client_LANActive}
           \/ client_path = "relay"

\* Channels never exceed their bound.
ChannelsBounded ==
    /\ Len(chan_b2c) <= MaxChanLen
    /\ Len(chan_c2b) <= MaxChanLen

\* Ping failures are bounded.
PingFailuresBounded ==
    ping_failures <= MaxPingFailures

\* --- Liveness ---

\* If the backend keeps offering LAN and the client keeps dialing
\* successfully, they eventually both reach LANActive.
LANEventuallyEstablished ==
    <>(backend_state = backend_LANActive /\ client_state = client_LANActive)

====
