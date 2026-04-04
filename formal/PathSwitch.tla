---- MODULE PathSwitch ----
\* Path-switching protocol with resource lifecycle management.
\* All resource bindings (dispatcher, monitor, signal, bridge) are
\* state variables. Transitions that change paths must update them.
\* The executor diffs state and rebinds — no independent logic.

EXTENDS Integers, Sequences, FiniteSets, TLC

\* --- States ---

CONSTANTS backend_RelayConnected, backend_LANOffered,
          backend_LANActive, backend_LANDegraded, backend_RelayBackoff

CONSTANTS client_RelayConnected, client_LANConnecting,
          client_LANVerifying, client_LANActive, client_RelayFallback

CONSTANTS relay_Idle, relay_BackendRegistered, relay_Bridged

VARIABLES
    backend_state, client_state, relay_state,
    chan_b2c, chan_c2b,
    \* Protocol state
    challenge, offer_challenge,
    ping_failures, backoff_level,
    \* Resource lifecycle — per actor where applicable
    b_active_path,       \* backend: "relay" or "lan"
    c_active_path,       \* client: "relay" or "lan"
    b_dispatcher_path,   \* backend: "relay" or "lan"
    c_dispatcher_path,   \* client: "relay" or "lan"
    monitor_target,      \* backend only: "lan" or "none"
    lan_signal,          \* "pending" or "ready"
    relay_bridge         \* "active" or "idle"

vars == <<backend_state, client_state, relay_state,
          chan_b2c, chan_c2b,
          challenge, offer_challenge, ping_failures, backoff_level,
          b_active_path, c_active_path,
          b_dispatcher_path, c_dispatcher_path,
          monitor_target, lan_signal, relay_bridge>>

CONSTANTS MaxChanLen, MaxPingFailures, MaxBackoffLevel

Init ==
    /\ backend_state = backend_RelayConnected
    /\ client_state = client_RelayConnected
    /\ relay_state = relay_Bridged
    /\ chan_b2c = <<>>
    /\ chan_c2b = <<>>
    /\ challenge = "c1"
    /\ offer_challenge = "none"
    /\ ping_failures = 0
    /\ backoff_level = 0
    /\ b_active_path = "relay"
    /\ c_active_path = "relay"
    /\ b_dispatcher_path = "relay"
    /\ c_dispatcher_path = "relay"
    /\ monitor_target = "none"
    /\ lan_signal = "pending"
    /\ relay_bridge = "active"

CanSend(ch) == Len(ch) < MaxChanLen
Min(a, b) == IF a < b THEN a ELSE b

\* Unchanged helper for resource vars that don't change in a transition.
ResourcesUnchanged == UNCHANGED <<b_active_path, c_active_path,
                                   b_dispatcher_path, c_dispatcher_path,
                                   monitor_target, lan_signal, relay_bridge>>

\* ================================================================
\* Backend
\* ================================================================

BackendOfferLAN ==
    /\ backend_state = backend_RelayConnected
    /\ CanSend(chan_b2c)
    /\ chan_b2c' = Append(chan_b2c, [type |-> "lan_offer", challenge |-> challenge])
    /\ backend_state' = backend_LANOffered
    /\ UNCHANGED <<client_state, relay_state, chan_c2b,
                   challenge, offer_challenge, ping_failures, backoff_level>>
    /\ ResourcesUnchanged

BackendVerifyOK ==
    /\ backend_state = backend_LANOffered
    /\ Len(chan_c2b) > 0
    /\ Head(chan_c2b).type = "lan_verify"
    /\ Head(chan_c2b).challenge = challenge
    /\ chan_c2b' = Tail(chan_c2b)
    /\ CanSend(chan_b2c)
    /\ chan_b2c' = Append(chan_b2c, [type |-> "lan_confirm"])
    /\ backend_state' = backend_LANActive
    /\ ping_failures' = 0
    /\ backoff_level' = 0
    \* Resource updates: backend switches to LAN.
    /\ b_active_path' = "lan"
    /\ b_dispatcher_path' = "lan"
    /\ monitor_target' = "lan"
    /\ lan_signal' = "ready"
    /\ UNCHANGED <<client_state, relay_state, challenge, offer_challenge,
                   c_active_path, c_dispatcher_path, relay_bridge>>

BackendVerifyFail ==
    /\ backend_state = backend_LANOffered
    /\ Len(chan_c2b) > 0
    /\ Head(chan_c2b).type = "lan_verify"
    /\ Head(chan_c2b).challenge /= challenge
    /\ chan_c2b' = Tail(chan_c2b)
    /\ backend_state' = backend_RelayConnected
    /\ UNCHANGED <<client_state, relay_state, chan_b2c,
                   challenge, offer_challenge, ping_failures, backoff_level>>
    /\ ResourcesUnchanged

BackendOfferTimeout ==
    /\ backend_state = backend_LANOffered
    /\ backend_state' = backend_RelayBackoff
    /\ backoff_level' = Min(backoff_level + 1, MaxBackoffLevel)
    /\ UNCHANGED <<client_state, relay_state, chan_b2c, chan_c2b,
                   challenge, offer_challenge, ping_failures>>
    /\ ResourcesUnchanged

BackendPing ==
    /\ backend_state = backend_LANActive
    /\ CanSend(chan_b2c)
    /\ chan_b2c' = Append(chan_b2c, [type |-> "ping"])
    /\ UNCHANGED <<backend_state, client_state, relay_state, chan_c2b,
                   challenge, offer_challenge, ping_failures, backoff_level>>
    /\ ResourcesUnchanged

BackendPingTimeout ==
    /\ backend_state = backend_LANActive
    /\ ping_failures' = 1
    /\ backend_state' = backend_LANDegraded
    /\ UNCHANGED <<client_state, relay_state, chan_b2c, chan_c2b,
                   challenge, offer_challenge, backoff_level>>
    /\ ResourcesUnchanged

BackendRecoverPong ==
    /\ backend_state = backend_LANDegraded
    /\ Len(chan_c2b) > 0
    /\ Head(chan_c2b).type = "pong"
    /\ chan_c2b' = Tail(chan_c2b)
    /\ backend_state' = backend_LANActive
    /\ ping_failures' = 0
    /\ UNCHANGED <<client_state, relay_state, chan_b2c,
                   challenge, offer_challenge, backoff_level>>
    /\ ResourcesUnchanged

BackendDegradedTimeoutContinue ==
    /\ backend_state = backend_LANDegraded
    /\ ping_failures + 1 < MaxPingFailures
    /\ ping_failures' = ping_failures + 1
    /\ UNCHANGED <<backend_state, client_state, relay_state,
                   chan_b2c, chan_c2b,
                   challenge, offer_challenge, backoff_level>>
    /\ ResourcesUnchanged

BackendDegradedFallback ==
    /\ backend_state = backend_LANDegraded
    /\ ping_failures + 1 >= MaxPingFailures
    /\ backend_state' = backend_RelayBackoff
    /\ ping_failures' = 0
    /\ backoff_level' = Min(backoff_level + 1, MaxBackoffLevel)
    /\ chan_b2c' = <<>>
    /\ chan_c2b' = <<>>
    \* Resource updates: backend back to relay.
    /\ b_active_path' = "relay"
    /\ b_dispatcher_path' = "relay"
    /\ monitor_target' = "none"
    /\ lan_signal' = "pending"
    /\ UNCHANGED <<client_state, relay_state, challenge, offer_challenge,
                   c_active_path, c_dispatcher_path, relay_bridge>>

BackendBackoffExpired ==
    /\ backend_state = backend_RelayBackoff
    /\ CanSend(chan_b2c)
    /\ chan_b2c' = Append(chan_b2c, [type |-> "lan_offer", challenge |-> challenge])
    /\ backend_state' = backend_LANOffered
    /\ UNCHANGED <<client_state, relay_state, chan_c2b,
                   challenge, offer_challenge, ping_failures, backoff_level>>
    /\ ResourcesUnchanged

\* ================================================================
\* Client
\* ================================================================

ClientRecvOffer ==
    /\ client_state = client_RelayConnected
    /\ Len(chan_b2c) > 0
    /\ Head(chan_b2c).type = "lan_offer"
    /\ offer_challenge' = Head(chan_b2c).challenge
    /\ chan_b2c' = Tail(chan_b2c)
    /\ client_state' = client_LANConnecting
    /\ UNCHANGED <<backend_state, relay_state, chan_c2b,
                   challenge, ping_failures, backoff_level>>
    /\ ResourcesUnchanged

ClientDialOK ==
    /\ client_state = client_LANConnecting
    /\ CanSend(chan_c2b)
    /\ chan_c2b' = Append(chan_c2b, [type |-> "lan_verify", challenge |-> offer_challenge])
    /\ client_state' = client_LANVerifying
    /\ UNCHANGED <<backend_state, relay_state, chan_b2c,
                   challenge, offer_challenge, ping_failures, backoff_level>>
    /\ ResourcesUnchanged

ClientDialFail ==
    /\ client_state = client_LANConnecting
    /\ client_state' = client_RelayConnected
    /\ UNCHANGED <<backend_state, relay_state, chan_b2c, chan_c2b,
                   challenge, offer_challenge, ping_failures, backoff_level>>
    /\ ResourcesUnchanged

ClientConfirm ==
    /\ client_state = client_LANVerifying
    /\ Len(chan_b2c) > 0
    /\ Head(chan_b2c).type = "lan_confirm"
    /\ chan_b2c' = Tail(chan_b2c)
    /\ client_state' = client_LANActive
    \* Resource updates: client switches to LAN.
    /\ c_active_path' = "lan"
    /\ c_dispatcher_path' = "lan"
    /\ lan_signal' = "ready"
    /\ UNCHANGED <<backend_state, relay_state, chan_c2b,
                   challenge, offer_challenge, ping_failures, backoff_level,
                   b_active_path, b_dispatcher_path, monitor_target, relay_bridge>>

ClientVerifyTimeout ==
    /\ client_state = client_LANVerifying
    /\ client_state' = client_RelayConnected
    /\ chan_b2c' = <<>>
    \* Resource: ensure client dispatcher is on relay.
    /\ c_dispatcher_path' = "relay"
    /\ UNCHANGED <<backend_state, relay_state, chan_c2b,
                   challenge, offer_challenge, ping_failures, backoff_level,
                   b_active_path, c_active_path, b_dispatcher_path,
                   monitor_target, lan_signal, relay_bridge>>

ClientPong ==
    /\ client_state = client_LANActive
    /\ Len(chan_b2c) > 0
    /\ Head(chan_b2c).type = "ping"
    /\ chan_b2c' = Tail(chan_b2c)
    /\ CanSend(chan_c2b)
    /\ chan_c2b' = Append(chan_c2b, [type |-> "pong"])
    /\ UNCHANGED <<backend_state, client_state, relay_state,
                   challenge, offer_challenge, ping_failures, backoff_level>>
    /\ ResourcesUnchanged

ClientLANError ==
    /\ client_state = client_LANActive
    /\ client_state' = client_RelayFallback
    /\ chan_b2c' = <<>>
    /\ chan_c2b' = <<>>
    \* Resource updates: client back to relay.
    /\ c_active_path' = "relay"
    /\ c_dispatcher_path' = "relay"
    /\ lan_signal' = "pending"
    /\ UNCHANGED <<backend_state, relay_state,
                   challenge, offer_challenge, ping_failures, backoff_level,
                   b_active_path, b_dispatcher_path, monitor_target, relay_bridge>>

ClientRelayOK ==
    /\ client_state = client_RelayFallback
    /\ client_state' = client_RelayConnected
    /\ UNCHANGED <<backend_state, relay_state, chan_b2c, chan_c2b,
                   challenge, offer_challenge, ping_failures, backoff_level>>
    /\ ResourcesUnchanged

ClientNewOfferOnLAN ==
    /\ client_state = client_LANActive
    /\ Len(chan_b2c) > 0
    /\ Head(chan_b2c).type = "lan_offer"
    /\ offer_challenge' = Head(chan_b2c).challenge
    /\ chan_b2c' = Tail(chan_b2c)
    /\ client_state' = client_LANConnecting
    /\ UNCHANGED <<backend_state, relay_state, chan_c2b,
                   challenge, ping_failures, backoff_level>>
    /\ ResourcesUnchanged

\* ================================================================
\* Next / Spec
\* ================================================================

Next ==
    \/ BackendOfferLAN     \/ BackendVerifyOK    \/ BackendVerifyFail
    \/ BackendOfferTimeout \/ BackendPing        \/ BackendPingTimeout
    \/ BackendRecoverPong  \/ BackendDegradedTimeoutContinue
    \/ BackendDegradedFallback \/ BackendBackoffExpired
    \/ ClientRecvOffer     \/ ClientDialOK       \/ ClientDialFail
    \/ ClientConfirm       \/ ClientVerifyTimeout
    \/ ClientPong          \/ ClientLANError
    \/ ClientRelayOK       \/ ClientNewOfferOnLAN

Spec == Init /\ [][Next]_vars /\ WF_vars(Next)

\* ================================================================
\* Invariants
\* ================================================================

\* --- Protocol invariants ---
PathConsistency ==
    /\ b_active_path \in {"relay", "lan"}
    /\ c_active_path \in {"relay", "lan"}
RelayStaysConnected == relay_state \in {relay_BackendRegistered, relay_Bridged}
LANRequiresVerification ==
    (backend_state = backend_LANActive /\ client_state = client_LANActive)
        => offer_challenge = challenge
ChannelsBounded == Len(chan_b2c) <= MaxChanLen /\ Len(chan_c2b) <= MaxChanLen
PingFailuresBounded == ping_failures <= MaxPingFailures
BackoffBounded == backoff_level <= MaxBackoffLevel
BackoffResetsOnSuccess == backend_state = backend_LANActive => backoff_level = 0
FallbackEntersBackoff == backend_state = backend_RelayBackoff => backoff_level >= 1

\* --- Resource lifecycle invariants (per actor) ---
DispatcherAlwaysBound ==
    /\ b_dispatcher_path \in {"relay", "lan"}
    /\ c_dispatcher_path \in {"relay", "lan"}
BackendDispatcherMatchesActive ==
    backend_state = backend_LANActive => b_dispatcher_path = "lan"
ClientDispatcherMatchesActive ==
    client_state = client_LANActive => c_dispatcher_path = "lan"
BackendDispatcherRelayOnFallback ==
    backend_state = backend_RelayBackoff => b_dispatcher_path = "relay"
ClientDispatcherRelayOnFallback ==
    client_state = client_RelayFallback => c_dispatcher_path = "relay"
MonitorOnlyWhenLAN ==
    monitor_target = "lan" => backend_state \in {backend_LANActive, backend_LANDegraded}
MonitorOffOnFallback ==
    backend_state = backend_RelayBackoff => monitor_target = "none"
LANSignalReady ==
    (backend_state = backend_LANActive /\ lan_signal = "ready") => b_active_path = "lan"
LANSignalPendingOnFallback ==
    backend_state = backend_RelayBackoff => lan_signal = "pending"

====
