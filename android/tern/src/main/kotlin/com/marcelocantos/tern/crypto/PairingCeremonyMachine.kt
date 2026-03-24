// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

// Auto-generated from protocol definition. Do not edit.
// Source of truth: protocol/*.yaml

package com.marcelocantos.tern.crypto

enum class MessageType(val value: String) {
    PairBegin("pair_begin"),
    TokenResponse("token_response"),
    PairHello("pair_hello"),
    PairHelloAck("pair_hello_ack"),
    PairConfirm("pair_confirm"),
    WaitingForCode("waiting_for_code"),
    CodeSubmit("code_submit"),
    PairComplete("pair_complete"),
    PairStatus("pair_status"),
    AuthRequest("auth_request"),
    AuthOk("auth_ok");
}

enum class ServerState(val value: String) {
    Idle("Idle"),
    GenerateToken("GenerateToken"),
    RegisterRelay("RegisterRelay"),
    WaitingForClient("WaitingForClient"),
    DeriveSecret("DeriveSecret"),
    SendAck("SendAck"),
    WaitingForCode("WaitingForCode"),
    ValidateCode("ValidateCode"),
    StorePaired("StorePaired"),
    Paired("Paired"),
    AuthCheck("AuthCheck"),
    SessionActive("SessionActive");
}

class ServerMachine {
    var state: ServerState = ServerState.Idle
        private set

    /** Process a received message. Returns the new state, or null if rejected. */
    fun handleMessage(msg: MessageType, guard: (String) -> Boolean = { true }): ServerState? {
        val newState = when {
            state == ServerState.Idle && msg == MessageType.PairBegin ->
                ServerState.GenerateToken
            state == ServerState.WaitingForClient && msg == MessageType.PairHello && guard("token_valid") ->
                ServerState.DeriveSecret
            state == ServerState.WaitingForClient && msg == MessageType.PairHello && guard("token_invalid") ->
                ServerState.Idle
            state == ServerState.WaitingForCode && msg == MessageType.CodeSubmit ->
                ServerState.ValidateCode
            state == ServerState.Paired && msg == MessageType.AuthRequest ->
                ServerState.AuthCheck
            else -> null
        }
        if (newState != null) state = newState
        return newState
    }

    /** Attempt an internal transition. Returns the new state, or null if none available. */
    fun step(guard: (String) -> Boolean = { true }): ServerState? {
        val newState = when {
            state == ServerState.GenerateToken ->
                ServerState.RegisterRelay
            state == ServerState.RegisterRelay ->
                ServerState.WaitingForClient
            state == ServerState.DeriveSecret ->
                ServerState.SendAck
            state == ServerState.SendAck ->
                ServerState.WaitingForCode
            state == ServerState.ValidateCode && guard("code_correct") ->
                ServerState.StorePaired
            state == ServerState.ValidateCode && guard("code_wrong") ->
                ServerState.Idle
            state == ServerState.StorePaired ->
                ServerState.Paired
            state == ServerState.AuthCheck && guard("device_known") ->
                ServerState.SessionActive
            state == ServerState.AuthCheck && guard("device_unknown") ->
                ServerState.Idle
            state == ServerState.SessionActive ->
                ServerState.Paired
            else -> null
        }
        if (newState != null) state = newState
        return newState
    }
}

enum class IosState(val value: String) {
    Idle("Idle"),
    ScanQR("ScanQR"),
    ConnectRelay("ConnectRelay"),
    GenKeyPair("GenKeyPair"),
    WaitAck("WaitAck"),
    E2EReady("E2EReady"),
    ShowCode("ShowCode"),
    WaitPairComplete("WaitPairComplete"),
    Paired("Paired"),
    Reconnect("Reconnect"),
    SendAuth("SendAuth"),
    SessionActive("SessionActive");
}

class IosMachine {
    var state: IosState = IosState.Idle
        private set

    /** Process a received message. Returns the new state, or null if rejected. */
    fun handleMessage(msg: MessageType, guard: (String) -> Boolean = { true }): IosState? {
        val newState = when {
            state == IosState.WaitAck && msg == MessageType.PairHelloAck ->
                IosState.E2EReady
            state == IosState.E2EReady && msg == MessageType.PairConfirm ->
                IosState.ShowCode
            state == IosState.WaitPairComplete && msg == MessageType.PairComplete ->
                IosState.Paired
            state == IosState.SendAuth && msg == MessageType.AuthOk ->
                IosState.SessionActive
            else -> null
        }
        if (newState != null) state = newState
        return newState
    }

    /** Attempt an internal transition. Returns the new state, or null if none available. */
    fun step(guard: (String) -> Boolean = { true }): IosState? {
        val newState = when {
            state == IosState.Idle ->
                IosState.ScanQR
            state == IosState.ScanQR ->
                IosState.ConnectRelay
            state == IosState.ConnectRelay ->
                IosState.GenKeyPair
            state == IosState.GenKeyPair ->
                IosState.WaitAck
            state == IosState.ShowCode ->
                IosState.WaitPairComplete
            state == IosState.Paired ->
                IosState.Reconnect
            state == IosState.Reconnect ->
                IosState.SendAuth
            state == IosState.SessionActive ->
                IosState.Paired
            else -> null
        }
        if (newState != null) state = newState
        return newState
    }
}

enum class CliState(val value: String) {
    Idle("Idle"),
    GetKey("GetKey"),
    BeginPair("BeginPair"),
    ShowQR("ShowQR"),
    PromptCode("PromptCode"),
    SubmitCode("SubmitCode"),
    Done("Done");
}

class CliMachine {
    var state: CliState = CliState.Idle
        private set

    /** Process a received message. Returns the new state, or null if rejected. */
    fun handleMessage(msg: MessageType, guard: (String) -> Boolean = { true }): CliState? {
        val newState = when {
            state == CliState.BeginPair && msg == MessageType.TokenResponse ->
                CliState.ShowQR
            state == CliState.ShowQR && msg == MessageType.WaitingForCode ->
                CliState.PromptCode
            state == CliState.SubmitCode && msg == MessageType.PairStatus ->
                CliState.Done
            else -> null
        }
        if (newState != null) state = newState
        return newState
    }

    /** Attempt an internal transition. Returns the new state, or null if none available. */
    fun step(guard: (String) -> Boolean = { true }): CliState? {
        val newState = when {
            state == CliState.Idle ->
                CliState.GetKey
            state == CliState.GetKey ->
                CliState.BeginPair
            state == CliState.PromptCode ->
                CliState.SubmitCode
            else -> null
        }
        if (newState != null) state = newState
        return newState
    }
}

