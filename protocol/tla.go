// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

package protocol

import (
	"fmt"
	"io"
	"sort"
	"strings"
)

// ExportTLA writes a TLA+ spec for the protocol. If phaseName is empty,
// the full protocol is emitted. If phaseName matches a defined phase,
// only that phase's transitions and variables are emitted — other
// variables are omitted entirely, dramatically reducing the state space.
func (p *Protocol) ExportTLA(w io.Writer) error {
	return p.ExportTLAPhase(w, "")
}

// ExportTLAPhase writes a TLA+ spec for a specific phase, or the full
// protocol if phaseName is empty.
func (p *Protocol) ExportTLAPhase(w io.Writer, phaseName string) error {
	var phase *Phase
	if phaseName != "" {
		for i := range p.Phases {
			if p.Phases[i].Name == phaseName {
				phase = &p.Phases[i]
				break
			}
		}
		if phase == nil {
			return fmt.Errorf("phase %q not found", phaseName)
		}
	}

	var b strings.Builder

	moduleName := sanitiseTLA(p.Name)
	if phase != nil {
		moduleName = sanitiseTLA(p.Name) + "_" + sanitiseTLA(phase.Name)
	}

	b.WriteString("---- MODULE ")
	b.WriteString(moduleName)
	b.WriteString(" ----\n")
	b.WriteString("\\* Auto-generated from protocol definition. Do not edit.\n")
	if phase != nil {
		fmt.Fprintf(&b, "\\* Phase: %s\n", phase.Name)
	}
	b.WriteString("\nEXTENDS Integers, Sequences, FiniteSets, TLC\n\n")

	// Build phase state set for filtering transitions.
	phaseStates := map[State]bool{}
	if phase != nil {
		for _, s := range phase.States {
			phaseStates[s] = true
		}
	}

	// Build phase var set for filtering variables.
	phaseVars := map[string]bool{}
	if phase != nil {
		for _, v := range phase.Vars {
			phaseVars[v] = true
		}
	}

	writeStateConstants(&b, p, phaseStates)
	writeMsgConstants(&b, p)
	writeOperators(&b, p)

	b.WriteString("(*--algorithm ")
	b.WriteString(moduleName)
	b.WriteString("\n\n")

	writeVariables(&b, p, phase, phaseVars)
	writeProcesses(&b, p, phase, phaseStates)

	// Only include adversary if this phase has it enabled (or no phase filter).
	includeAdversary := phase == nil || phase.Adversary
	if includeAdversary {
		writeAdversary(&b, p)
	}

	b.WriteString("end algorithm; *)\n")
	b.WriteString("\\* BEGIN TRANSLATION\n")
	b.WriteString("\\* END TRANSLATION\n\n")

	writeProperties(&b, p, phase, phaseVars)

	b.WriteString("====\n")

	_, err := io.WriteString(w, b.String())
	return err
}

func writeStateConstants(b *strings.Builder, p *Protocol, phaseStates map[State]bool) {
	for _, a := range p.Actors {
		states := collectStates(a)
		b.WriteString("\\* States for ")
		b.WriteString(a.Name)
		b.WriteString("\n")
		for _, s := range states {
			// In phase mode, only emit states in the phase + immediate
			// neighbours (for initial/terminal state references).
			if len(phaseStates) > 0 && !phaseStates[s] && !isNeighbour(a, s, phaseStates) {
				continue
			}
			fmt.Fprintf(b, "%s_%s == \"%s_%s\"\n",
				sanitiseTLA(a.Name), sanitiseTLA(string(s)),
				a.Name, s)
		}
		b.WriteString("\n")
	}
}

// isNeighbour returns true if state s is the source or target of a
// transition where the other end is in phaseStates.
func isNeighbour(a Actor, s State, phaseStates map[State]bool) bool {
	for _, t := range a.Transitions {
		if t.From == s && phaseStates[t.To] {
			return true
		}
		if t.To == s && phaseStates[t.From] {
			return true
		}
	}
	return false
}

func writeMsgConstants(b *strings.Builder, p *Protocol) {
	if len(p.Messages) == 0 {
		return
	}
	b.WriteString("\\* Message types\n")
	for _, m := range p.Messages {
		fmt.Fprintf(b, "MSG_%s == \"%s\" \\* %s -> %s",
			sanitiseTLA(string(m.Type)), m.Type, m.From, m.To)
		if m.Desc != "" {
			fmt.Fprintf(b, " (%s)", m.Desc)
		}
		b.WriteString("\n")
	}
	b.WriteString("\n")
}

func writeOperators(b *strings.Builder, p *Protocol) {
	if len(p.Operators) == 0 {
		return
	}
	b.WriteString("\\* Helper operators\n")
	for _, op := range p.Operators {
		if op.Desc != "" {
			fmt.Fprintf(b, "\\* %s\n", op.Desc)
		}
		if op.Params != "" {
			fmt.Fprintf(b, "%s(%s) == %s\n", sanitiseTLA(op.Name), op.Params, op.Expr)
		} else {
			fmt.Fprintf(b, "%s == %s\n", sanitiseTLA(op.Name), op.Expr)
		}
	}
	b.WriteString("\n")
}

func writeVariables(b *strings.Builder, p *Protocol, phase *Phase, phaseVars map[string]bool) {
	b.WriteString("variables\n")

	for _, a := range p.Actors {
		fmt.Fprintf(b, "    %s_state = %s_%s,\n",
			sanitiseTLA(a.Name),
			sanitiseTLA(a.Name), sanitiseTLA(string(initialForPhase(a, phase))))
	}

	channels := channelPairs(p)
	for _, ch := range channels {
		fmt.Fprintf(b, "    chan_%s_%s = <<>>,\n", ch.from, ch.to)
	}

	if phase == nil || phase.Adversary {
		b.WriteString("    adversary_knowledge = {},\n")
	}

	for _, v := range p.Vars {
		// In phase mode, only include phase-relevant variables.
		// Always include recv_msg — it's infrastructure for message reception.
		if phase != nil && !phaseVars[v.Name] && v.Name != "recv_msg" {
			continue
		}
		if v.Desc != "" {
			fmt.Fprintf(b, "    \\* %s\n", v.Desc)
		}
		fmt.Fprintf(b, "    %s = %s,\n", sanitiseTLA(v.Name), v.Initial)
	}

	// Remove trailing comma, add semicolon.
	s := b.String()
	if idx := strings.LastIndex(s, ",\n"); idx >= 0 {
		b.Reset()
		b.WriteString(s[:idx])
		b.WriteString(";\n\n")
	}
}

// initialForPhase returns the initial state for an actor in a given phase.
// If the phase defines states, use the first phase state that appears as
// a transition target from outside the phase (the entry point). Otherwise
// use the actor's declared initial state.
func initialForPhase(a Actor, phase *Phase) State {
	if phase == nil {
		return a.Initial
	}

	phaseStates := map[State]bool{}
	for _, s := range phase.States {
		phaseStates[s] = true
	}

	// If the actor's initial state is in this phase, use it.
	if phaseStates[a.Initial] {
		return a.Initial
	}

	// Find the first phase state that's a target from outside the phase.
	for _, t := range a.Transitions {
		if !phaseStates[t.From] && phaseStates[t.To] {
			return t.To
		}
	}

	// Fallback: first phase state.
	for _, s := range phase.States {
		for _, as := range collectStates(a) {
			if as == s {
				return s
			}
		}
	}
	return a.Initial
}

func writeProcesses(b *strings.Builder, p *Protocol, phase *Phase, phaseStates map[State]bool) {
	for i, a := range p.Actors {
		// Filter transitions to those relevant to this phase.
		var transitions []Transition
		for _, t := range a.Transitions {
			if len(phaseStates) > 0 {
				// Include if from OR to is in the phase.
				if !phaseStates[t.From] && !phaseStates[t.To] {
					continue
				}
			}
			transitions = append(transitions, t)
		}

		if len(transitions) == 0 {
			continue
		}

		fmt.Fprintf(b, "fair process %s = %d\n", sanitiseTLA(a.Name), i+1)
		b.WriteString("begin\n")
		fmt.Fprintf(b, "  %s_loop:\n", sanitiseTLA(a.Name))
		if !p.OneShot {
			b.WriteString("  while TRUE do\n")
		}

		b.WriteString("    either\n")

		first := true
		for _, t := range transitions {
			if !first {
				b.WriteString("    or\n")
			}
			first = false

			writeTransitionComment(b, &t)
			writeTransitionAwait(b, p, &a, &t)
			writeTransitionBody(b, p, &a, &t)
		}

		b.WriteString("    end either;\n")
		if !p.OneShot {
			b.WriteString("  end while;\n")
		}
		b.WriteString("end process;\n\n")
	}
}

func writeTransitionComment(b *strings.Builder, t *Transition) {
	fmt.Fprintf(b, "      \\* %s -> %s", t.From, t.To)
	if t.On.Kind == TriggerRecv {
		fmt.Fprintf(b, " on %s", t.On.Msg)
	} else if t.On.Desc != "" {
		fmt.Fprintf(b, " (%s)", t.On.Desc)
	}
	b.WriteString("\n")
}

func writeTransitionAwait(b *strings.Builder, p *Protocol, a *Actor, t *Transition) {
	fmt.Fprintf(b, "      await %s_state = %s_%s",
		sanitiseTLA(a.Name),
		sanitiseTLA(a.Name), sanitiseTLA(string(t.From)))

	if t.On.Kind == TriggerRecv {
		fromActor := msgSender(p, t.On.Msg)
		chanName := channelName(fromActor, a.Name)
		fmt.Fprintf(b, " /\\ Len(%s) > 0 /\\ Head(%s).type = MSG_%s",
			chanName, chanName, sanitiseTLA(string(t.On.Msg)))
	}

	if t.Guard != "" {
		expr := guardExpr(p, t.Guard)
		if t.On.Kind == TriggerRecv {
			fromActor := msgSender(p, t.On.Msg)
			chanName := channelName(fromActor, a.Name)
			expr = strings.ReplaceAll(expr, "recv_msg", "Head("+chanName+")")
		}
		fmt.Fprintf(b, " /\\ (%s)", expr)
	}
	b.WriteString(";\n")
}

func guardExpr(p *Protocol, id GuardID) string {
	for _, g := range p.Guards {
		if g.ID == id {
			return g.Expr
		}
	}
	return string(id)
}

func writeTransitionBody(b *strings.Builder, p *Protocol, a *Actor, t *Transition) {
	if t.On.Kind == TriggerRecv {
		fromActor := msgSender(p, t.On.Msg)
		chanName := channelName(fromActor, a.Name)
		fmt.Fprintf(b, "      recv_msg := Head(%s);\n", chanName)
		fmt.Fprintf(b, "      %s := Tail(%s);\n", chanName, chanName)
	}

	for _, s := range t.Sends {
		chanName := channelName(a.Name, s.To)
		fmt.Fprintf(b, "      %s := Append(%s, ", chanName, chanName)
		writeRecord(b, s.Msg, s.Fields)
		b.WriteString(");\n")
	}

	for _, u := range t.Updates {
		fmt.Fprintf(b, "      %s := %s;\n", sanitiseTLA(u.Var), u.Expr)
	}

	fmt.Fprintf(b, "      %s_state := %s_%s;\n",
		sanitiseTLA(a.Name),
		sanitiseTLA(a.Name), sanitiseTLA(string(t.To)))
}

func writeRecord(b *strings.Builder, msg MsgType, fields map[string]string) {
	b.WriteString("[type |-> MSG_")
	b.WriteString(sanitiseTLA(string(msg)))

	if len(fields) > 0 {
		keys := make([]string, 0, len(fields))
		for k := range fields {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			fmt.Fprintf(b, ", %s |-> %s", k, fields[k])
		}
	}
	b.WriteString("]")
}

func writeAdversary(b *strings.Builder, p *Protocol) {
	channels := channelPairs(p)
	if len(channels) == 0 && len(p.AdvActions) == 0 {
		return
	}

	b.WriteString("\\* Dolev-Yao adversary: controls the network.\n")
	fmt.Fprintf(b, "fair process Adversary = %d\n", len(p.Actors)+1)
	b.WriteString("begin\n")
	b.WriteString("  adv_loop:\n")
	b.WriteString("  while TRUE do\n")

	if p.AdvGuard != "" {
		fmt.Fprintf(b, "    await %s;\n", p.AdvGuard)
	}

	b.WriteString("    either\n")
	b.WriteString("      skip \\* no-op: honest relay\n")

	for _, ch := range channels {
		chanName := fmt.Sprintf("chan_%s_%s", ch.from, ch.to)

		b.WriteString("    or\n")
		fmt.Fprintf(b, "      \\* Eavesdrop on %s -> %s\n", ch.from, ch.to)
		fmt.Fprintf(b, "      await Len(%s) > 0;\n", chanName)
		fmt.Fprintf(b, "      adversary_knowledge := adversary_knowledge \\union {Head(%s)};\n", chanName)

		b.WriteString("    or\n")
		fmt.Fprintf(b, "      \\* Drop from %s -> %s\n", ch.from, ch.to)
		fmt.Fprintf(b, "      await Len(%s) > 0;\n", chanName)
		fmt.Fprintf(b, "      %s := Tail(%s);\n", chanName, chanName)

		b.WriteString("    or\n")
		fmt.Fprintf(b, "      \\* Replay into %s -> %s\n", ch.from, ch.to)
		if p.ChannelBound > 0 {
			fmt.Fprintf(b, "      await adversary_knowledge /= {} /\\ Len(%s) < %d;\n", chanName, p.ChannelBound)
		} else {
			b.WriteString("      await adversary_knowledge /= {};\n")
		}
		fmt.Fprintf(b, "      with msg \\in adversary_knowledge do\n")
		fmt.Fprintf(b, "        %s := Append(%s, msg);\n", chanName, chanName)
		b.WriteString("      end with;\n")
	}

	for _, aa := range p.AdvActions {
		b.WriteString("    or\n")
		fmt.Fprintf(b, "      \\* %s: %s\n", aa.Name, aa.Desc)
		b.WriteString(aa.Code)
		b.WriteString("\n")
	}

	b.WriteString("    end either;\n")
	b.WriteString("  end while;\n")
	b.WriteString("end process;\n\n")
}

func writeProperties(b *strings.Builder, p *Protocol, phase *Phase, phaseVars map[string]bool) {
	if len(p.Properties) == 0 {
		return
	}
	b.WriteString("\\* Verification properties\n")
	for _, prop := range p.Properties {
		// In phase mode, skip properties that reference non-phase variables.
		if phase != nil && !propertyRelevant(prop, phaseVars, p.Vars) {
			continue
		}

		if prop.Desc != "" {
			fmt.Fprintf(b, "\\* %s\n", prop.Desc)
		}
		switch prop.Kind {
		case Invariant:
			fmt.Fprintf(b, "%s == %s\n", sanitiseTLA(prop.Name), prop.Expr)
		case Liveness:
			fmt.Fprintf(b, "%s == <>(%s)\n", sanitiseTLA(prop.Name), prop.Expr)
		case LeadsTo:
			fmt.Fprintf(b, "%s == (%s) ~> (%s)\n", sanitiseTLA(prop.Name), prop.FromExpr, prop.ToExpr)
		}
	}
	b.WriteString("\n")
}

// propertyRelevant returns true if the property only references
// variables that exist in this phase (phase vars + actor state vars +
// channels). Returns false if ANY referenced variable is outside the phase.
func propertyRelevant(prop Property, phaseVars map[string]bool, allVars []VarDef) bool {
	if len(phaseVars) == 0 {
		return true
	}
	expr := prop.Expr + prop.FromExpr + prop.ToExpr

	// Check if any non-phase variable appears in the expression.
	for _, v := range allVars {
		if !phaseVars[v.Name] && v.Name != "recv_msg" && strings.Contains(expr, v.Name) {
			return false
		}
	}
	// Adversary variables are not in p.Vars — check explicitly.
	if strings.Contains(expr, "adversary_knowledge") || strings.Contains(expr, "adversary_keys") {
		// Only relevant if adversary variables exist (i.e., adversary phase).
		if !phaseVars["adversary_keys"] {
			return false
		}
	}
	return true
}

// Helpers.

type channelPair struct{ from, to string }

func channelPairs(p *Protocol) []channelPair {
	seen := map[string]bool{}
	var pairs []channelPair
	for _, m := range p.Messages {
		key := m.From + "_" + m.To
		if !seen[key] {
			seen[key] = true
			pairs = append(pairs, channelPair{from: m.From, to: m.To})
		}
	}
	return pairs
}

func channelName(from, to string) string {
	return "chan_" + from + "_" + to
}

func msgSender(p *Protocol, msg MsgType) string {
	for _, m := range p.Messages {
		if m.Type == msg {
			return m.From
		}
	}
	return "unknown"
}

func sanitiseTLA(s string) string {
	r := strings.NewReplacer(" ", "_", "-", "_", ".", "_")
	return r.Replace(s)
}
