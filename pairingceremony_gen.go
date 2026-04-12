// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

// Code generated from protocol/*.yaml. DO NOT EDIT.

package pigeon

import (
	"github.com/marcelocantos/pigeon/protocol"
)

type (
	State      = protocol.State
	MsgType    = protocol.MsgType
	GuardID    = protocol.GuardID
	ActionID   = protocol.ActionID
	EventID    = protocol.EventID
	CmdID      = protocol.CmdID
	Protocol   = protocol.Protocol
	Actor      = protocol.Actor
	Transition = protocol.Transition
	Send       = protocol.Send
	Message    = protocol.Message
	VarDef     = protocol.VarDef
	VarUpdate  = protocol.VarUpdate
	GuardDef   = protocol.GuardDef
	Operator   = protocol.Operator
	AdvAction  = protocol.AdvAction
	Property   = protocol.Property
)

var (
	Recv     = protocol.Recv
	Internal = protocol.Internal
	Invariant = protocol.Invariant
	Liveness  = protocol.Liveness
)

// PairingCeremonyProtocol server states.
const (
	PairingCeremonyProtocolServerIdle State = "Idle"
	PairingCeremonyProtocolServerGenerateToken State = "GenerateToken"
	PairingCeremonyProtocolServerRegisterRelay State = "RegisterRelay"
	PairingCeremonyProtocolServerWaitingForClient State = "WaitingForClient"
	PairingCeremonyProtocolServerDeriveSecret State = "DeriveSecret"
	PairingCeremonyProtocolServerSendAck State = "SendAck"
	PairingCeremonyProtocolServerWaitingForCode State = "WaitingForCode"
	PairingCeremonyProtocolServerValidateCode State = "ValidateCode"
	PairingCeremonyProtocolServerStorePaired State = "StorePaired"
	PairingCeremonyProtocolServerPaired State = "Paired"
	PairingCeremonyProtocolServerAuthCheck State = "AuthCheck"
	PairingCeremonyProtocolServerSessionActive State = "SessionActive"
)

// PairingCeremonyProtocol ios states.
const (
	PairingCeremonyProtocolAppIdle State = "Idle"
	PairingCeremonyProtocolAppScanQR State = "ScanQR"
	PairingCeremonyProtocolAppConnectRelay State = "ConnectRelay"
	PairingCeremonyProtocolAppGenKeyPair State = "GenKeyPair"
	PairingCeremonyProtocolAppWaitAck State = "WaitAck"
	PairingCeremonyProtocolAppE2EReady State = "E2EReady"
	PairingCeremonyProtocolAppShowCode State = "ShowCode"
	PairingCeremonyProtocolAppWaitPairComplete State = "WaitPairComplete"
	PairingCeremonyProtocolAppPaired State = "Paired"
	PairingCeremonyProtocolAppReconnect State = "Reconnect"
	PairingCeremonyProtocolAppSendAuth State = "SendAuth"
	PairingCeremonyProtocolAppSessionActive State = "SessionActive"
)

// PairingCeremonyProtocol cli states.
const (
	PairingCeremonyProtocolCLIIdle State = "Idle"
	PairingCeremonyProtocolCLIGetKey State = "GetKey"
	PairingCeremonyProtocolCLIBeginPair State = "BeginPair"
	PairingCeremonyProtocolCLIShowQR State = "ShowQR"
	PairingCeremonyProtocolCLIPromptCode State = "PromptCode"
	PairingCeremonyProtocolCLISubmitCode State = "SubmitCode"
	PairingCeremonyProtocolCLIDone State = "Done"
)

// PairingCeremonyProtocol message types.
const (
	PairingCeremonyProtocolMsgPairBegin MsgType = "pair_begin"
	PairingCeremonyProtocolMsgTokenResponse MsgType = "token_response"
	PairingCeremonyProtocolMsgPairHello MsgType = "pair_hello"
	PairingCeremonyProtocolMsgPairHelloAck MsgType = "pair_hello_ack"
	PairingCeremonyProtocolMsgPairConfirm MsgType = "pair_confirm"
	PairingCeremonyProtocolMsgWaitingForCode MsgType = "waiting_for_code"
	PairingCeremonyProtocolMsgCodeSubmit MsgType = "code_submit"
	PairingCeremonyProtocolMsgPairComplete MsgType = "pair_complete"
	PairingCeremonyProtocolMsgPairStatus MsgType = "pair_status"
	PairingCeremonyProtocolMsgAuthRequest MsgType = "auth_request"
	PairingCeremonyProtocolMsgAuthOk MsgType = "auth_ok"
)

// PairingCeremonyProtocol guards.
const (
	PairingCeremonyProtocolGuardTokenValid GuardID = "token_valid"
	PairingCeremonyProtocolGuardTokenInvalid GuardID = "token_invalid"
	PairingCeremonyProtocolGuardCodeCorrect GuardID = "code_correct"
	PairingCeremonyProtocolGuardCodeWrong GuardID = "code_wrong"
	PairingCeremonyProtocolGuardDeviceKnown GuardID = "device_known"
	PairingCeremonyProtocolGuardDeviceUnknown GuardID = "device_unknown"
	PairingCeremonyProtocolGuardNonceFresh GuardID = "nonce_fresh"
)

// PairingCeremonyProtocol actions.
const (
	PairingCeremonyProtocolActionDeriveSecret ActionID = "derive_secret"
	PairingCeremonyProtocolActionGenerateToken ActionID = "generate_token"
	PairingCeremonyProtocolActionRegisterRelay ActionID = "register_relay"
	PairingCeremonyProtocolActionSendPairHello ActionID = "send_pair_hello"
	PairingCeremonyProtocolActionStoreDevice ActionID = "store_device"
	PairingCeremonyProtocolActionStoreSecret ActionID = "store_secret"
	PairingCeremonyProtocolActionVerifyDevice ActionID = "verify_device"
)

// PairingCeremonyProtocol events.
const (
	PairingCeremonyProtocolEventECDHComplete EventID = "ECDH complete"
	PairingCeremonyProtocolEventQRParsed EventID = "QR parsed"
	PairingCeremonyProtocolEventAppLaunch EventID = "app launch"
	PairingCeremonyProtocolEventCheckCode EventID = "check code"
	PairingCeremonyProtocolEventCliInit EventID = "cli --init"
	PairingCeremonyProtocolEventCodeDisplayed EventID = "code displayed"
	PairingCeremonyProtocolEventDisconnect EventID = "disconnect"
	PairingCeremonyProtocolEventFinalise EventID = "finalise"
	PairingCeremonyProtocolEventKeyPairGenerated EventID = "key pair generated"
	PairingCeremonyProtocolEventKeyStored EventID = "key stored"
	PairingCeremonyProtocolEventRecvAuthOk EventID = "recv_auth_ok"
	PairingCeremonyProtocolEventRecvAuthRequest EventID = "recv_auth_request"
	PairingCeremonyProtocolEventRecvCodeSubmit EventID = "recv_code_submit"
	PairingCeremonyProtocolEventRecvPairBegin EventID = "recv_pair_begin"
	PairingCeremonyProtocolEventRecvPairComplete EventID = "recv_pair_complete"
	PairingCeremonyProtocolEventRecvPairConfirm EventID = "recv_pair_confirm"
	PairingCeremonyProtocolEventRecvPairHello EventID = "recv_pair_hello"
	PairingCeremonyProtocolEventRecvPairHelloAck EventID = "recv_pair_hello_ack"
	PairingCeremonyProtocolEventRecvPairStatus EventID = "recv_pair_status"
	PairingCeremonyProtocolEventRecvTokenResponse EventID = "recv_token_response"
	PairingCeremonyProtocolEventRecvWaitingForCode EventID = "recv_waiting_for_code"
	PairingCeremonyProtocolEventRelayConnected EventID = "relay connected"
	PairingCeremonyProtocolEventRelayRegistered EventID = "relay registered"
	PairingCeremonyProtocolEventSignalCodeDisplay EventID = "signal code display"
	PairingCeremonyProtocolEventTokenCreated EventID = "token created"
	PairingCeremonyProtocolEventUserEntersCode EventID = "user enters code"
	PairingCeremonyProtocolEventUserScansQR EventID = "user scans QR"
	PairingCeremonyProtocolEventVerify EventID = "verify"
)

func PairingCeremonyProtocol() *Protocol {
	return &Protocol{
		Name: "PairingCeremony",
		Actors: []Actor{
			{Name: "server", Initial: "Idle", Transitions: []Transition{
				{From: "Idle", To: "GenerateToken", On: Recv("pair_begin"), Do: "generate_token", Updates: []VarUpdate{{Var: "current_token", Expr: "\"tok_1\""}, {Var: "active_tokens", Expr: "active_tokens \\union {\"tok_1\"}"}, }},
				{From: "GenerateToken", To: "RegisterRelay", On: Internal("token created"), Do: "register_relay"},
				{From: "RegisterRelay", To: "WaitingForClient", On: Internal("relay registered"), Sends: []Send{{To: "cli", Msg: "token_response", Fields: map[string]string{"instance_id": "\"inst_1\"", "token": "current_token", }}, }},
				{From: "WaitingForClient", To: "DeriveSecret", On: Recv("pair_hello"), Guard: "token_valid", Do: "derive_secret", Updates: []VarUpdate{{Var: "received_client_pub", Expr: "recv_msg.pubkey"}, {Var: "server_ecdh_pub", Expr: "\"server_pub\""}, {Var: "server_shared_key", Expr: "DeriveKey(\"server_pub\", recv_msg.pubkey)"}, {Var: "server_code", Expr: "DeriveCode(\"server_pub\", recv_msg.pubkey)"}, }},
				{From: "WaitingForClient", To: "Idle", On: Recv("pair_hello"), Guard: "token_invalid"},
				{From: "DeriveSecret", To: "SendAck", On: Internal("ECDH complete"), Sends: []Send{{To: "ios", Msg: "pair_hello_ack", Fields: map[string]string{"pubkey": "server_ecdh_pub", }}, }},
				{From: "SendAck", To: "WaitingForCode", On: Internal("signal code display"), Sends: []Send{{To: "ios", Msg: "pair_confirm"}, {To: "cli", Msg: "waiting_for_code"}, }},
				{From: "WaitingForCode", To: "ValidateCode", On: Recv("code_submit"), Updates: []VarUpdate{{Var: "received_code", Expr: "recv_msg.code"}, }},
				{From: "ValidateCode", To: "StorePaired", On: Internal("check code"), Guard: "code_correct"},
				{From: "ValidateCode", To: "Idle", On: Internal("check code"), Guard: "code_wrong", Updates: []VarUpdate{{Var: "code_attempts", Expr: "code_attempts + 1"}, }},
				{From: "StorePaired", To: "Paired", On: Internal("finalise"), Do: "store_device", Sends: []Send{{To: "ios", Msg: "pair_complete", Fields: map[string]string{"key": "server_shared_key", "secret": "\"dev_secret_1\"", }}, {To: "cli", Msg: "pair_status", Fields: map[string]string{"status": "\"paired\"", }}, }, Updates: []VarUpdate{{Var: "device_secret", Expr: "\"dev_secret_1\""}, {Var: "paired_devices", Expr: "paired_devices \\union {\"device_1\"}"}, {Var: "active_tokens", Expr: "active_tokens \\ {current_token}"}, {Var: "used_tokens", Expr: "used_tokens \\union {current_token}"}, }},
				{From: "Paired", To: "AuthCheck", On: Recv("auth_request"), Updates: []VarUpdate{{Var: "received_device_id", Expr: "recv_msg.device_id"}, {Var: "received_auth_nonce", Expr: "recv_msg.nonce"}, }},
				{From: "AuthCheck", To: "SessionActive", On: Internal("verify"), Guard: "device_known", Do: "verify_device", Sends: []Send{{To: "ios", Msg: "auth_ok"}, }, Updates: []VarUpdate{{Var: "auth_nonces_used", Expr: "auth_nonces_used \\union {received_auth_nonce}"}, }},
				{From: "AuthCheck", To: "Idle", On: Internal("verify"), Guard: "device_unknown"},
				{From: "SessionActive", To: "Paired", On: Internal("disconnect")},
			}},
			{Name: "ios", Initial: "Idle", Transitions: []Transition{
				{From: "Idle", To: "ScanQR", On: Internal("user scans QR")},
				{From: "ScanQR", To: "ConnectRelay", On: Internal("QR parsed")},
				{From: "ConnectRelay", To: "GenKeyPair", On: Internal("relay connected")},
				{From: "GenKeyPair", To: "WaitAck", On: Internal("key pair generated"), Do: "send_pair_hello", Sends: []Send{{To: "server", Msg: "pair_hello", Fields: map[string]string{"pubkey": "\"client_pub\"", "token": "current_token", }}, }},
				{From: "WaitAck", To: "E2EReady", On: Recv("pair_hello_ack"), Do: "derive_secret", Updates: []VarUpdate{{Var: "received_server_pub", Expr: "recv_msg.pubkey"}, {Var: "client_shared_key", Expr: "DeriveKey(\"client_pub\", recv_msg.pubkey)"}, }},
				{From: "E2EReady", To: "ShowCode", On: Recv("pair_confirm"), Updates: []VarUpdate{{Var: "ios_code", Expr: "DeriveCode(received_server_pub, \"client_pub\")"}, }},
				{From: "ShowCode", To: "WaitPairComplete", On: Internal("code displayed")},
				{From: "WaitPairComplete", To: "Paired", On: Recv("pair_complete"), Do: "store_secret"},
				{From: "Paired", To: "Reconnect", On: Internal("app launch")},
				{From: "Reconnect", To: "SendAuth", On: Internal("relay connected"), Sends: []Send{{To: "server", Msg: "auth_request", Fields: map[string]string{"device_id": "\"device_1\"", "key": "client_shared_key", "nonce": "\"nonce_1\"", "secret": "device_secret", }}, }},
				{From: "SendAuth", To: "SessionActive", On: Recv("auth_ok")},
				{From: "SessionActive", To: "Paired", On: Internal("disconnect")},
			}},
			{Name: "cli", Initial: "Idle", Transitions: []Transition{
				{From: "Idle", To: "GetKey", On: Internal("cli --init")},
				{From: "GetKey", To: "BeginPair", On: Internal("key stored"), Sends: []Send{{To: "server", Msg: "pair_begin"}, }},
				{From: "BeginPair", To: "ShowQR", On: Recv("token_response")},
				{From: "ShowQR", To: "PromptCode", On: Recv("waiting_for_code")},
				{From: "PromptCode", To: "SubmitCode", On: Internal("user enters code"), Sends: []Send{{To: "server", Msg: "code_submit", Fields: map[string]string{"code": "ios_code", }}, }},
				{From: "SubmitCode", To: "Done", On: Recv("pair_status")},
			}},
		},
		Messages: []Message{
			{Type: "pair_begin", From: "cli", To: "server", Desc: "POST /api/pair/begin"},
			{Type: "token_response", From: "server", To: "cli", Desc: "{instance_id, pairing_token}"},
			{Type: "pair_hello", From: "ios", To: "server", Desc: "ECDH pubkey + pairing token"},
			{Type: "pair_hello_ack", From: "server", To: "ios", Desc: "ECDH pubkey"},
			{Type: "pair_confirm", From: "server", To: "ios", Desc: "signal to compute and display code"},
			{Type: "waiting_for_code", From: "server", To: "cli", Desc: "prompt for code entry"},
			{Type: "code_submit", From: "cli", To: "server", Desc: "POST /api/pair/confirm"},
			{Type: "pair_complete", From: "server", To: "ios", Desc: "encrypted device secret"},
			{Type: "pair_status", From: "server", To: "cli", Desc: "status: paired"},
			{Type: "auth_request", From: "ios", To: "server", Desc: "encrypted auth with nonce"},
			{Type: "auth_ok", From: "server", To: "ios", Desc: "session established"},
		},
		Vars: []VarDef{
			{Name: "current_token", Initial: "\"none\"", Desc: "pairing token currently in play"},
			{Name: "active_tokens", Initial: "{}", Desc: "set of valid (non-revoked) tokens"},
			{Name: "used_tokens", Initial: "{}", Desc: "set of revoked tokens"},
			{Name: "server_ecdh_pub", Initial: "\"none\"", Desc: "server ECDH public key"},
			{Name: "received_client_pub", Initial: "\"none\"", Desc: "pubkey server received in pair_hello (may be adversary's)"},
			{Name: "received_server_pub", Initial: "\"none\"", Desc: "pubkey ios received in pair_hello_ack (may be adversary's)"},
			{Name: "server_shared_key", Initial: "<<\"none\">>", Desc: "ECDH key derived by server (tuple to match DeriveKey output type)"},
			{Name: "client_shared_key", Initial: "<<\"none\">>", Desc: "ECDH key derived by ios (tuple to match DeriveKey output type)"},
			{Name: "server_code", Initial: "<<\"none\">>", Desc: "code computed by server from its view of the pubkeys (tuple to match DeriveCode output type)"},
			{Name: "ios_code", Initial: "<<\"none\">>", Desc: "code computed by ios from its view of the pubkeys (tuple to match DeriveCode output type)"},
			{Name: "received_code", Initial: "<<\"none\">>", Desc: "code received in code_submit (tuple to match DeriveCode output type)"},
			{Name: "code_attempts", Initial: "0", Desc: "failed code submission attempts"},
			{Name: "device_secret", Initial: "\"none\"", Desc: "persistent device secret"},
			{Name: "paired_devices", Initial: "{}", Desc: "device IDs that completed pairing"},
			{Name: "received_device_id", Initial: "\"none\"", Desc: "device_id from auth_request"},
			{Name: "auth_nonces_used", Initial: "{}", Desc: "set of consumed auth nonces"},
			{Name: "received_auth_nonce", Initial: "\"none\"", Desc: "nonce from auth_request"},
			{Name: "adversary_keys", Initial: "{}", Desc: "encryption keys the adversary knows"},
			{Name: "adv_ecdh_pub", Initial: "\"adv_pub\"", Desc: "adversary's ECDH public key"},
			{Name: "adv_saved_client_pub", Initial: "\"none\"", Desc: "real client pubkey saved during MitM"},
			{Name: "adv_saved_server_pub", Initial: "\"none\"", Desc: "real server pubkey saved during MitM"},
			{Name: "recv_msg", Initial: "[type |-> \"none\"]", Desc: "last received message (staging)"},
		},
		Guards: []GuardDef{
			{ID: "token_valid", Expr: "recv_msg.token \\in active_tokens"},
			{ID: "token_invalid", Expr: "recv_msg.token \\notin active_tokens"},
			{ID: "code_correct", Expr: "received_code = server_code"},
			{ID: "code_wrong", Expr: "received_code /= server_code"},
			{ID: "device_known", Expr: "received_device_id \\in paired_devices"},
			{ID: "device_unknown", Expr: "received_device_id \\notin paired_devices"},
			{ID: "nonce_fresh", Expr: "received_auth_nonce \\notin auth_nonces_used"},
		},
		Operators: []Operator{
			{Name: "KeyRank", Params: "k", Expr: "CASE k = \"adv_pub\" -> 0 [] k = \"client_pub\" -> 1 [] k = \"server_pub\" -> 2 [] OTHER -> 3", Desc: "Assign numeric rank to pubkey names for deterministic ordering"},
			{Name: "DeriveKey", Params: "a, b", Expr: "IF KeyRank(a) <= KeyRank(b) THEN <<\"ecdh\", a, b>> ELSE <<\"ecdh\", b, a>>", Desc: "Symbolic ECDH: deterministic key from two public keys (order-independent)"},
			{Name: "DeriveCode", Params: "a, b", Expr: "IF KeyRank(a) <= KeyRank(b) THEN <<\"code\", a, b>> ELSE <<\"code\", b, a>>", Desc: "Key-bound confirmation code: deterministic from both pubkeys (order-independent)"},
		},
		AdvActions: []AdvAction{
			{Name: "QR_shoulder_surf", Desc: "observe QR code content (token + instance_id)", Code: "      await current_token /= \"none\";\n      adversary_knowledge := adversary_knowledge \\union {[type |-> \"qr_token\", token |-> current_token]};"},
			{Name: "MitM_pair_hello", Desc: "intercept pair_hello and substitute adversary ECDH pubkey", Code: "      await Len(chan_ios_server) > 0 /\\ Head(chan_ios_server).type = MSG_pair_hello;\n      adv_saved_client_pub := Head(chan_ios_server).pubkey;\n      chan_ios_server := <<[type |-> MSG_pair_hello, token |-> Head(chan_ios_server).token, pubkey |-> adv_ecdh_pub]>> \\o Tail(chan_ios_server);"},
			{Name: "MitM_pair_hello_ack", Desc: "intercept pair_hello_ack and substitute adversary ECDH pubkey, derive both shared secrets", Code: "      await Len(chan_server_ios) > 0 /\\ Head(chan_server_ios).type = MSG_pair_hello_ack;\n      adv_saved_server_pub := Head(chan_server_ios).pubkey;\n      adversary_keys := adversary_keys \\union {DeriveKey(adv_ecdh_pub, adv_saved_server_pub), DeriveKey(adv_ecdh_pub, adv_saved_client_pub)};\n      chan_server_ios := <<[type |-> MSG_pair_hello_ack, pubkey |-> adv_ecdh_pub]>> \\o Tail(chan_server_ios);"},
			{Name: "MitM_reencrypt_secret", Desc: "decrypt pair_complete with MitM key, learn device secret", Code: "      await Len(chan_server_ios) > 0 /\\ Head(chan_server_ios).type = MSG_pair_complete /\\ Head(chan_server_ios).key \\in adversary_keys;\n      with msg = Head(chan_server_ios) do\n        adversary_knowledge := adversary_knowledge \\union {[type |-> \"plaintext_secret\", secret |-> msg.secret]};\n        chan_server_ios := <<[type |-> MSG_pair_complete, key |-> DeriveKey(adv_ecdh_pub, adv_saved_client_pub), secret |-> msg.secret]>> \\o Tail(chan_server_ios);\n      end with;"},
			{Name: "concurrent_pair", Desc: "race a forged pair_hello using shoulder-surfed token", Code: "      await \\E m \\in adversary_knowledge : m = [type |-> \"qr_token\", token |-> current_token];\n      await Len(chan_ios_server) < 3;\n      chan_ios_server := Append(chan_ios_server, [type |-> MSG_pair_hello, token |-> current_token, pubkey |-> adv_ecdh_pub]);"},
			{Name: "token_bruteforce", Desc: "send pair_hello with fabricated token", Code: "      await Len(chan_ios_server) < 3;\n      chan_ios_server := Append(chan_ios_server, [type |-> MSG_pair_hello, token |-> \"fake_token\", pubkey |-> adv_ecdh_pub]);"},
			{Name: "code_guess", Desc: "submit fabricated confirmation code via CLI channel", Code: "      await Len(chan_cli_server) < 3;\n      chan_cli_server := Append(chan_cli_server, [type |-> MSG_code_submit, code |-> <<\"guess\", \"000000\">>]);"},
			{Name: "session_replay", Desc: "replay captured auth_request with stale nonce", Code: "      await Len(chan_ios_server) < 3;\n      await \\E m \\in adversary_knowledge : m.type = MSG_auth_request;\n      with msg \\in {m \\in adversary_knowledge : m.type = MSG_auth_request} do\n        chan_ios_server := Append(chan_ios_server, msg);\n      end with;"},
		},
		Properties: []Property{
			{Name: "NoTokenReuse", Kind: Invariant, Expr: "used_tokens \\intersect active_tokens = {}", Desc: "A revoked pairing token is never accepted again"},
			{Name: "MitMDetectedByCodeMismatch", Kind: Invariant, Expr: "(server_shared_key \\in adversary_keys /\\ server_code /= <<\"none\">> /\\ ios_code /= <<\"none\">>) => server_code /= ios_code", Desc: "If the current session's shared key is compromised and both sides computed codes, the codes differ"},
			{Name: "MitMPrevented", Kind: Invariant, Expr: "server_shared_key \\in adversary_keys => server_state \\notin {server_StorePaired, server_Paired, server_AuthCheck, server_SessionActive}", Desc: "If the current session's key is compromised, pairing never completes"},
			{Name: "AuthRequiresCompletedPairing", Kind: Invariant, Expr: "server_state = server_SessionActive => received_device_id \\in paired_devices", Desc: "A session is only active for a device that completed pairing"},
			{Name: "NoNonceReuse", Kind: Invariant, Expr: "server_state = server_SessionActive => received_auth_nonce \\notin (auth_nonces_used \\ {received_auth_nonce})", Desc: "Each auth nonce is accepted at most once"},
			{Name: "WrongCodeDoesNotPair", Kind: Invariant, Expr: "(server_state = server_StorePaired \\/ server_state = server_Paired) => received_code = server_code \\/ received_code = <<\"none\">>", Desc: "Pairing only completes with the correct confirmation code"},
			{Name: "DeviceSecretSecrecy", Kind: Invariant, Expr: "\\A m \\in adversary_knowledge : \"type\" \\in DOMAIN m => m.type /= \"plaintext_secret\"", Desc: "Adversary never learns the device secret in plaintext"},
			{Name: "HonestPairingCompletes", Kind: Liveness, Expr: "cli_state = cli_Done /\\ ios_state = ios_Paired", Desc: "If all actors cooperate honestly (no MitM), pairing eventually completes"},
		},
		ChannelBound: 3,
		OneShot: true,
	}
}

// PairingCeremonyProtocolServerMachine is the generated state machine for the server actor.
type PairingCeremonyProtocolServerMachine struct {
	State State
	CurrentToken string // pairing token currently in play
	ActiveTokens string // set of valid (non-revoked) tokens
	UsedTokens string // set of revoked tokens
	ServerEcdhPub string // server ECDH public key
	ReceivedClientPub string // pubkey server received in pair_hello (may be adversary's)
	ServerSharedKey string // ECDH key derived by server (tuple to match DeriveKey output type)
	ServerCode string // code computed by server from its view of the pubkeys (tuple to match DeriveCode output type)
	ReceivedCode string // code received in code_submit (tuple to match DeriveCode output type)
	CodeAttempts int // failed code submission attempts
	DeviceSecret string // persistent device secret
	PairedDevices string // device IDs that completed pairing
	ReceivedDeviceId string // device_id from auth_request
	AuthNoncesUsed string // set of consumed auth nonces
	ReceivedAuthNonce string // nonce from auth_request

	Guards  map[GuardID]func() bool
	Actions map[ActionID]func() error
	OnChange func(varName string)
}

func NewPairingCeremonyProtocolServerMachine() *PairingCeremonyProtocolServerMachine {
	return &PairingCeremonyProtocolServerMachine{
		State: PairingCeremonyProtocolServerIdle,
		CurrentToken: "none",
		ActiveTokens: "",
		UsedTokens: "",
		ServerEcdhPub: "none",
		ReceivedClientPub: "none",
		ServerSharedKey: "",
		ServerCode: "",
		ReceivedCode: "",
		CodeAttempts: 0,
		DeviceSecret: "none",
		PairedDevices: "",
		ReceivedDeviceId: "none",
		AuthNoncesUsed: "",
		ReceivedAuthNonce: "none",
		Guards:  make(map[GuardID]func() bool),
		Actions: make(map[ActionID]func() error),
	}
}

func (m *PairingCeremonyProtocolServerMachine) HandleMessage(msg MsgType) (bool, error) {
	switch {
	case m.State == PairingCeremonyProtocolServerIdle && msg == PairingCeremonyProtocolMsgPairBegin:
		if fn := m.Actions[PairingCeremonyProtocolActionGenerateToken]; fn != nil {
			if err := fn(); err != nil { return false, err }
		}
		m.CurrentToken = "tok_1"
		if m.OnChange != nil { m.OnChange("current_token") }
		// active_tokens: active_tokens \union {"tok_1"} (set by action)
		m.State = PairingCeremonyProtocolServerGenerateToken
		return true, nil
	case m.State == PairingCeremonyProtocolServerWaitingForClient && msg == PairingCeremonyProtocolMsgPairHello && m.Guards[PairingCeremonyProtocolGuardTokenValid] != nil && m.Guards[PairingCeremonyProtocolGuardTokenValid]():
		if fn := m.Actions[PairingCeremonyProtocolActionDeriveSecret]; fn != nil {
			if err := fn(); err != nil { return false, err }
		}
		// received_client_pub: recv_msg.pubkey (set by action)
		m.ServerEcdhPub = "server_pub"
		if m.OnChange != nil { m.OnChange("server_ecdh_pub") }
		// server_shared_key: DeriveKey("server_pub", recv_msg.pubkey) (set by action)
		// server_code: DeriveCode("server_pub", recv_msg.pubkey) (set by action)
		m.State = PairingCeremonyProtocolServerDeriveSecret
		return true, nil
	case m.State == PairingCeremonyProtocolServerWaitingForClient && msg == PairingCeremonyProtocolMsgPairHello && m.Guards[PairingCeremonyProtocolGuardTokenInvalid] != nil && m.Guards[PairingCeremonyProtocolGuardTokenInvalid]():
		m.State = PairingCeremonyProtocolServerIdle
		return true, nil
	case m.State == PairingCeremonyProtocolServerWaitingForCode && msg == PairingCeremonyProtocolMsgCodeSubmit:
		// received_code: recv_msg.code (set by action)
		m.State = PairingCeremonyProtocolServerValidateCode
		return true, nil
	case m.State == PairingCeremonyProtocolServerPaired && msg == PairingCeremonyProtocolMsgAuthRequest:
		// received_device_id: recv_msg.device_id (set by action)
		// received_auth_nonce: recv_msg.nonce (set by action)
		m.State = PairingCeremonyProtocolServerAuthCheck
		return true, nil
	}
	return false, nil
}

func (m *PairingCeremonyProtocolServerMachine) Step(event EventID) (bool, error) {
	switch {
	case m.State == PairingCeremonyProtocolServerGenerateToken && event == PairingCeremonyProtocolEventTokenCreated:
		if fn := m.Actions[PairingCeremonyProtocolActionRegisterRelay]; fn != nil {
			if err := fn(); err != nil { return false, err }
		}
		m.State = PairingCeremonyProtocolServerRegisterRelay
		return true, nil
	case m.State == PairingCeremonyProtocolServerRegisterRelay && event == PairingCeremonyProtocolEventRelayRegistered:
		m.State = PairingCeremonyProtocolServerWaitingForClient
		return true, nil
	case m.State == PairingCeremonyProtocolServerDeriveSecret && event == PairingCeremonyProtocolEventECDHComplete:
		m.State = PairingCeremonyProtocolServerSendAck
		return true, nil
	case m.State == PairingCeremonyProtocolServerSendAck && event == PairingCeremonyProtocolEventSignalCodeDisplay:
		m.State = PairingCeremonyProtocolServerWaitingForCode
		return true, nil
	case m.State == PairingCeremonyProtocolServerValidateCode && event == PairingCeremonyProtocolEventCheckCode && m.Guards[PairingCeremonyProtocolGuardCodeCorrect] != nil && m.Guards[PairingCeremonyProtocolGuardCodeCorrect]():
		m.State = PairingCeremonyProtocolServerStorePaired
		return true, nil
	case m.State == PairingCeremonyProtocolServerValidateCode && event == PairingCeremonyProtocolEventCheckCode && m.Guards[PairingCeremonyProtocolGuardCodeWrong] != nil && m.Guards[PairingCeremonyProtocolGuardCodeWrong]():
		m.CodeAttempts = m.CodeAttempts + 1
		if m.OnChange != nil { m.OnChange("code_attempts") }
		m.State = PairingCeremonyProtocolServerIdle
		return true, nil
	case m.State == PairingCeremonyProtocolServerStorePaired && event == PairingCeremonyProtocolEventFinalise:
		if fn := m.Actions[PairingCeremonyProtocolActionStoreDevice]; fn != nil {
			if err := fn(); err != nil { return false, err }
		}
		m.DeviceSecret = "dev_secret_1"
		if m.OnChange != nil { m.OnChange("device_secret") }
		// paired_devices: paired_devices \union {"device_1"} (set by action)
		// active_tokens: active_tokens \ {current_token} (set by action)
		// used_tokens: used_tokens \union {current_token} (set by action)
		m.State = PairingCeremonyProtocolServerPaired
		return true, nil
	case m.State == PairingCeremonyProtocolServerAuthCheck && event == PairingCeremonyProtocolEventVerify && m.Guards[PairingCeremonyProtocolGuardDeviceKnown] != nil && m.Guards[PairingCeremonyProtocolGuardDeviceKnown]():
		if fn := m.Actions[PairingCeremonyProtocolActionVerifyDevice]; fn != nil {
			if err := fn(); err != nil { return false, err }
		}
		// auth_nonces_used: auth_nonces_used \union {received_auth_nonce} (set by action)
		m.State = PairingCeremonyProtocolServerSessionActive
		return true, nil
	case m.State == PairingCeremonyProtocolServerAuthCheck && event == PairingCeremonyProtocolEventVerify && m.Guards[PairingCeremonyProtocolGuardDeviceUnknown] != nil && m.Guards[PairingCeremonyProtocolGuardDeviceUnknown]():
		m.State = PairingCeremonyProtocolServerIdle
		return true, nil
	case m.State == PairingCeremonyProtocolServerSessionActive && event == PairingCeremonyProtocolEventDisconnect:
		m.State = PairingCeremonyProtocolServerPaired
		return true, nil
	}
	return false, nil
}

func (m *PairingCeremonyProtocolServerMachine) HandleEvent(ev EventID) ([]CmdID, error) {
	switch {
	case m.State == PairingCeremonyProtocolServerIdle && ev == PairingCeremonyProtocolEventRecvPairBegin:
		if fn := m.Actions[PairingCeremonyProtocolActionGenerateToken]; fn != nil {
			if err := fn(); err != nil { return nil, err }
		}
		m.CurrentToken = "tok_1"
		if m.OnChange != nil { m.OnChange("current_token") }
		// active_tokens: active_tokens \union {"tok_1"} (set by action)
		m.State = PairingCeremonyProtocolServerGenerateToken
		return nil, nil
	case m.State == PairingCeremonyProtocolServerGenerateToken && ev == PairingCeremonyProtocolEventTokenCreated:
		if fn := m.Actions[PairingCeremonyProtocolActionRegisterRelay]; fn != nil {
			if err := fn(); err != nil { return nil, err }
		}
		m.State = PairingCeremonyProtocolServerRegisterRelay
		return nil, nil
	case m.State == PairingCeremonyProtocolServerRegisterRelay && ev == PairingCeremonyProtocolEventRelayRegistered:
		m.State = PairingCeremonyProtocolServerWaitingForClient
		return nil, nil
	case m.State == PairingCeremonyProtocolServerWaitingForClient && ev == PairingCeremonyProtocolEventRecvPairHello && m.Guards[PairingCeremonyProtocolGuardTokenValid] != nil && m.Guards[PairingCeremonyProtocolGuardTokenValid]():
		if fn := m.Actions[PairingCeremonyProtocolActionDeriveSecret]; fn != nil {
			if err := fn(); err != nil { return nil, err }
		}
		// received_client_pub: recv_msg.pubkey (set by action)
		m.ServerEcdhPub = "server_pub"
		if m.OnChange != nil { m.OnChange("server_ecdh_pub") }
		// server_shared_key: DeriveKey("server_pub", recv_msg.pubkey) (set by action)
		// server_code: DeriveCode("server_pub", recv_msg.pubkey) (set by action)
		m.State = PairingCeremonyProtocolServerDeriveSecret
		return nil, nil
	case m.State == PairingCeremonyProtocolServerWaitingForClient && ev == PairingCeremonyProtocolEventRecvPairHello && m.Guards[PairingCeremonyProtocolGuardTokenInvalid] != nil && m.Guards[PairingCeremonyProtocolGuardTokenInvalid]():
		m.State = PairingCeremonyProtocolServerIdle
		return nil, nil
	case m.State == PairingCeremonyProtocolServerDeriveSecret && ev == PairingCeremonyProtocolEventECDHComplete:
		m.State = PairingCeremonyProtocolServerSendAck
		return nil, nil
	case m.State == PairingCeremonyProtocolServerSendAck && ev == PairingCeremonyProtocolEventSignalCodeDisplay:
		m.State = PairingCeremonyProtocolServerWaitingForCode
		return nil, nil
	case m.State == PairingCeremonyProtocolServerWaitingForCode && ev == PairingCeremonyProtocolEventRecvCodeSubmit:
		// received_code: recv_msg.code (set by action)
		m.State = PairingCeremonyProtocolServerValidateCode
		return nil, nil
	case m.State == PairingCeremonyProtocolServerValidateCode && ev == PairingCeremonyProtocolEventCheckCode && m.Guards[PairingCeremonyProtocolGuardCodeCorrect] != nil && m.Guards[PairingCeremonyProtocolGuardCodeCorrect]():
		m.State = PairingCeremonyProtocolServerStorePaired
		return nil, nil
	case m.State == PairingCeremonyProtocolServerValidateCode && ev == PairingCeremonyProtocolEventCheckCode && m.Guards[PairingCeremonyProtocolGuardCodeWrong] != nil && m.Guards[PairingCeremonyProtocolGuardCodeWrong]():
		m.CodeAttempts = m.CodeAttempts + 1
		if m.OnChange != nil { m.OnChange("code_attempts") }
		m.State = PairingCeremonyProtocolServerIdle
		return nil, nil
	case m.State == PairingCeremonyProtocolServerStorePaired && ev == PairingCeremonyProtocolEventFinalise:
		if fn := m.Actions[PairingCeremonyProtocolActionStoreDevice]; fn != nil {
			if err := fn(); err != nil { return nil, err }
		}
		m.DeviceSecret = "dev_secret_1"
		if m.OnChange != nil { m.OnChange("device_secret") }
		// paired_devices: paired_devices \union {"device_1"} (set by action)
		// active_tokens: active_tokens \ {current_token} (set by action)
		// used_tokens: used_tokens \union {current_token} (set by action)
		m.State = PairingCeremonyProtocolServerPaired
		return nil, nil
	case m.State == PairingCeremonyProtocolServerPaired && ev == PairingCeremonyProtocolEventRecvAuthRequest:
		// received_device_id: recv_msg.device_id (set by action)
		// received_auth_nonce: recv_msg.nonce (set by action)
		m.State = PairingCeremonyProtocolServerAuthCheck
		return nil, nil
	case m.State == PairingCeremonyProtocolServerAuthCheck && ev == PairingCeremonyProtocolEventVerify && m.Guards[PairingCeremonyProtocolGuardDeviceKnown] != nil && m.Guards[PairingCeremonyProtocolGuardDeviceKnown]():
		if fn := m.Actions[PairingCeremonyProtocolActionVerifyDevice]; fn != nil {
			if err := fn(); err != nil { return nil, err }
		}
		// auth_nonces_used: auth_nonces_used \union {received_auth_nonce} (set by action)
		m.State = PairingCeremonyProtocolServerSessionActive
		return nil, nil
	case m.State == PairingCeremonyProtocolServerAuthCheck && ev == PairingCeremonyProtocolEventVerify && m.Guards[PairingCeremonyProtocolGuardDeviceUnknown] != nil && m.Guards[PairingCeremonyProtocolGuardDeviceUnknown]():
		m.State = PairingCeremonyProtocolServerIdle
		return nil, nil
	case m.State == PairingCeremonyProtocolServerSessionActive && ev == PairingCeremonyProtocolEventDisconnect:
		m.State = PairingCeremonyProtocolServerPaired
		return nil, nil
	}
	return nil, nil
}

// PairingCeremonyProtocolAppMachine is the generated state machine for the ios actor.
type PairingCeremonyProtocolAppMachine struct {
	State State
	ReceivedServerPub string // pubkey ios received in pair_hello_ack (may be adversary's)
	ClientSharedKey string // ECDH key derived by ios (tuple to match DeriveKey output type)
	IosCode string // code computed by ios from its view of the pubkeys (tuple to match DeriveCode output type)

	Guards  map[GuardID]func() bool
	Actions map[ActionID]func() error
	OnChange func(varName string)
}

func NewPairingCeremonyProtocolAppMachine() *PairingCeremonyProtocolAppMachine {
	return &PairingCeremonyProtocolAppMachine{
		State: PairingCeremonyProtocolAppIdle,
		ReceivedServerPub: "none",
		ClientSharedKey: "",
		IosCode: "",
		Guards:  make(map[GuardID]func() bool),
		Actions: make(map[ActionID]func() error),
	}
}

func (m *PairingCeremonyProtocolAppMachine) HandleMessage(msg MsgType) (bool, error) {
	switch {
	case m.State == PairingCeremonyProtocolAppWaitAck && msg == PairingCeremonyProtocolMsgPairHelloAck:
		if fn := m.Actions[PairingCeremonyProtocolActionDeriveSecret]; fn != nil {
			if err := fn(); err != nil { return false, err }
		}
		// received_server_pub: recv_msg.pubkey (set by action)
		// client_shared_key: DeriveKey("client_pub", recv_msg.pubkey) (set by action)
		m.State = PairingCeremonyProtocolAppE2EReady
		return true, nil
	case m.State == PairingCeremonyProtocolAppE2EReady && msg == PairingCeremonyProtocolMsgPairConfirm:
		// ios_code: DeriveCode(received_server_pub, "client_pub") (set by action)
		m.State = PairingCeremonyProtocolAppShowCode
		return true, nil
	case m.State == PairingCeremonyProtocolAppWaitPairComplete && msg == PairingCeremonyProtocolMsgPairComplete:
		if fn := m.Actions[PairingCeremonyProtocolActionStoreSecret]; fn != nil {
			if err := fn(); err != nil { return false, err }
		}
		m.State = PairingCeremonyProtocolAppPaired
		return true, nil
	case m.State == PairingCeremonyProtocolAppSendAuth && msg == PairingCeremonyProtocolMsgAuthOk:
		m.State = PairingCeremonyProtocolAppSessionActive
		return true, nil
	}
	return false, nil
}

func (m *PairingCeremonyProtocolAppMachine) Step(event EventID) (bool, error) {
	switch {
	case m.State == PairingCeremonyProtocolAppIdle && event == PairingCeremonyProtocolEventUserScansQR:
		m.State = PairingCeremonyProtocolAppScanQR
		return true, nil
	case m.State == PairingCeremonyProtocolAppScanQR && event == PairingCeremonyProtocolEventQRParsed:
		m.State = PairingCeremonyProtocolAppConnectRelay
		return true, nil
	case m.State == PairingCeremonyProtocolAppConnectRelay && event == PairingCeremonyProtocolEventRelayConnected:
		m.State = PairingCeremonyProtocolAppGenKeyPair
		return true, nil
	case m.State == PairingCeremonyProtocolAppGenKeyPair && event == PairingCeremonyProtocolEventKeyPairGenerated:
		if fn := m.Actions[PairingCeremonyProtocolActionSendPairHello]; fn != nil {
			if err := fn(); err != nil { return false, err }
		}
		m.State = PairingCeremonyProtocolAppWaitAck
		return true, nil
	case m.State == PairingCeremonyProtocolAppShowCode && event == PairingCeremonyProtocolEventCodeDisplayed:
		m.State = PairingCeremonyProtocolAppWaitPairComplete
		return true, nil
	case m.State == PairingCeremonyProtocolAppPaired && event == PairingCeremonyProtocolEventAppLaunch:
		m.State = PairingCeremonyProtocolAppReconnect
		return true, nil
	case m.State == PairingCeremonyProtocolAppReconnect && event == PairingCeremonyProtocolEventRelayConnected:
		m.State = PairingCeremonyProtocolAppSendAuth
		return true, nil
	case m.State == PairingCeremonyProtocolAppSessionActive && event == PairingCeremonyProtocolEventDisconnect:
		m.State = PairingCeremonyProtocolAppPaired
		return true, nil
	}
	return false, nil
}

func (m *PairingCeremonyProtocolAppMachine) HandleEvent(ev EventID) ([]CmdID, error) {
	switch {
	case m.State == PairingCeremonyProtocolAppIdle && ev == PairingCeremonyProtocolEventUserScansQR:
		m.State = PairingCeremonyProtocolAppScanQR
		return nil, nil
	case m.State == PairingCeremonyProtocolAppScanQR && ev == PairingCeremonyProtocolEventQRParsed:
		m.State = PairingCeremonyProtocolAppConnectRelay
		return nil, nil
	case m.State == PairingCeremonyProtocolAppConnectRelay && ev == PairingCeremonyProtocolEventRelayConnected:
		m.State = PairingCeremonyProtocolAppGenKeyPair
		return nil, nil
	case m.State == PairingCeremonyProtocolAppGenKeyPair && ev == PairingCeremonyProtocolEventKeyPairGenerated:
		if fn := m.Actions[PairingCeremonyProtocolActionSendPairHello]; fn != nil {
			if err := fn(); err != nil { return nil, err }
		}
		m.State = PairingCeremonyProtocolAppWaitAck
		return nil, nil
	case m.State == PairingCeremonyProtocolAppWaitAck && ev == PairingCeremonyProtocolEventRecvPairHelloAck:
		if fn := m.Actions[PairingCeremonyProtocolActionDeriveSecret]; fn != nil {
			if err := fn(); err != nil { return nil, err }
		}
		// received_server_pub: recv_msg.pubkey (set by action)
		// client_shared_key: DeriveKey("client_pub", recv_msg.pubkey) (set by action)
		m.State = PairingCeremonyProtocolAppE2EReady
		return nil, nil
	case m.State == PairingCeremonyProtocolAppE2EReady && ev == PairingCeremonyProtocolEventRecvPairConfirm:
		// ios_code: DeriveCode(received_server_pub, "client_pub") (set by action)
		m.State = PairingCeremonyProtocolAppShowCode
		return nil, nil
	case m.State == PairingCeremonyProtocolAppShowCode && ev == PairingCeremonyProtocolEventCodeDisplayed:
		m.State = PairingCeremonyProtocolAppWaitPairComplete
		return nil, nil
	case m.State == PairingCeremonyProtocolAppWaitPairComplete && ev == PairingCeremonyProtocolEventRecvPairComplete:
		if fn := m.Actions[PairingCeremonyProtocolActionStoreSecret]; fn != nil {
			if err := fn(); err != nil { return nil, err }
		}
		m.State = PairingCeremonyProtocolAppPaired
		return nil, nil
	case m.State == PairingCeremonyProtocolAppPaired && ev == PairingCeremonyProtocolEventAppLaunch:
		m.State = PairingCeremonyProtocolAppReconnect
		return nil, nil
	case m.State == PairingCeremonyProtocolAppReconnect && ev == PairingCeremonyProtocolEventRelayConnected:
		m.State = PairingCeremonyProtocolAppSendAuth
		return nil, nil
	case m.State == PairingCeremonyProtocolAppSendAuth && ev == PairingCeremonyProtocolEventRecvAuthOk:
		m.State = PairingCeremonyProtocolAppSessionActive
		return nil, nil
	case m.State == PairingCeremonyProtocolAppSessionActive && ev == PairingCeremonyProtocolEventDisconnect:
		m.State = PairingCeremonyProtocolAppPaired
		return nil, nil
	}
	return nil, nil
}

// PairingCeremonyProtocolCLIMachine is the generated state machine for the cli actor.
type PairingCeremonyProtocolCLIMachine struct {
	State State

	Guards  map[GuardID]func() bool
	Actions map[ActionID]func() error
	OnChange func(varName string)
}

func NewPairingCeremonyProtocolCLIMachine() *PairingCeremonyProtocolCLIMachine {
	return &PairingCeremonyProtocolCLIMachine{
		State: PairingCeremonyProtocolCLIIdle,
		Guards:  make(map[GuardID]func() bool),
		Actions: make(map[ActionID]func() error),
	}
}

func (m *PairingCeremonyProtocolCLIMachine) HandleMessage(msg MsgType) (bool, error) {
	switch {
	case m.State == PairingCeremonyProtocolCLIBeginPair && msg == PairingCeremonyProtocolMsgTokenResponse:
		m.State = PairingCeremonyProtocolCLIShowQR
		return true, nil
	case m.State == PairingCeremonyProtocolCLIShowQR && msg == PairingCeremonyProtocolMsgWaitingForCode:
		m.State = PairingCeremonyProtocolCLIPromptCode
		return true, nil
	case m.State == PairingCeremonyProtocolCLISubmitCode && msg == PairingCeremonyProtocolMsgPairStatus:
		m.State = PairingCeremonyProtocolCLIDone
		return true, nil
	}
	return false, nil
}

func (m *PairingCeremonyProtocolCLIMachine) Step(event EventID) (bool, error) {
	switch {
	case m.State == PairingCeremonyProtocolCLIIdle && event == PairingCeremonyProtocolEventCliInit:
		m.State = PairingCeremonyProtocolCLIGetKey
		return true, nil
	case m.State == PairingCeremonyProtocolCLIGetKey && event == PairingCeremonyProtocolEventKeyStored:
		m.State = PairingCeremonyProtocolCLIBeginPair
		return true, nil
	case m.State == PairingCeremonyProtocolCLIPromptCode && event == PairingCeremonyProtocolEventUserEntersCode:
		m.State = PairingCeremonyProtocolCLISubmitCode
		return true, nil
	}
	return false, nil
}

func (m *PairingCeremonyProtocolCLIMachine) HandleEvent(ev EventID) ([]CmdID, error) {
	switch {
	case m.State == PairingCeremonyProtocolCLIIdle && ev == PairingCeremonyProtocolEventCliInit:
		m.State = PairingCeremonyProtocolCLIGetKey
		return nil, nil
	case m.State == PairingCeremonyProtocolCLIGetKey && ev == PairingCeremonyProtocolEventKeyStored:
		m.State = PairingCeremonyProtocolCLIBeginPair
		return nil, nil
	case m.State == PairingCeremonyProtocolCLIBeginPair && ev == PairingCeremonyProtocolEventRecvTokenResponse:
		m.State = PairingCeremonyProtocolCLIShowQR
		return nil, nil
	case m.State == PairingCeremonyProtocolCLIShowQR && ev == PairingCeremonyProtocolEventRecvWaitingForCode:
		m.State = PairingCeremonyProtocolCLIPromptCode
		return nil, nil
	case m.State == PairingCeremonyProtocolCLIPromptCode && ev == PairingCeremonyProtocolEventUserEntersCode:
		m.State = PairingCeremonyProtocolCLISubmitCode
		return nil, nil
	case m.State == PairingCeremonyProtocolCLISubmitCode && ev == PairingCeremonyProtocolEventRecvPairStatus:
		m.State = PairingCeremonyProtocolCLIDone
		return nil, nil
	}
	return nil, nil
}

