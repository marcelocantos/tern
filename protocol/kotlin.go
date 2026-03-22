// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

package protocol

import (
	"fmt"
	"io"
	"strings"
	"unicode"
)

// ExportKotlin writes a Kotlin source file with enum classes for states and
// message types, and table-driven state machine classes for each actor.
func (p *Protocol) ExportKotlin(w io.Writer, pkg string) error {
	var b strings.Builder

	b.WriteString("// Copyright 2026 Marcelo Cantos\n")
	b.WriteString("// SPDX-License-Identifier: Apache-2.0\n\n")
	b.WriteString("// Auto-generated from protocol definition. Do not edit.\n")
	b.WriteString("// Source of truth: protocol/*.yaml\n\n")
	fmt.Fprintf(&b, "package %s\n\n", pkg)

	// Message type enum.
	b.WriteString("enum class MessageType(val value: String) {\n")
	for i, m := range p.Messages {
		comma := ","
		if i == len(p.Messages)-1 {
			comma = ";"
		}
		fmt.Fprintf(&b, "    %s(\"%s\")%s\n", kotlinPascalCase(string(m.Type)), m.Type, comma)
	}
	b.WriteString("}\n\n")

	// Per-actor state enum + machine.
	for _, a := range p.Actors {
		typeName := kotlinTypeName(a.Name)
		states := collectStates(a)

		// State enum.
		fmt.Fprintf(&b, "enum class %sState(val value: String) {\n", typeName)
		for i, s := range states {
			comma := ","
			if i == len(states)-1 {
				comma = ";"
			}
			fmt.Fprintf(&b, "    %s(\"%s\")%s\n", string(s), s, comma)
		}
		b.WriteString("}\n\n")

		// Machine class.
		fmt.Fprintf(&b, "class %sMachine {\n", typeName)
		fmt.Fprintf(&b, "    var state: %sState = %sState.%s\n", typeName, typeName, a.Initial)
		b.WriteString("        private set\n\n")

		// handleMessage — use `when {}` with boolean conditions to support guards.
		b.WriteString("    /** Process a received message. Returns the new state, or null if rejected. */\n")
		fmt.Fprintf(&b, "    fun handleMessage(msg: MessageType, guard: (String) -> Boolean = { true }): %sState? {\n", typeName)
		b.WriteString("        val newState = when {\n")
		for _, t := range a.Transitions {
			if t.On.Kind != TriggerRecv {
				continue
			}
			guardExpr := ""
			if t.Guard != "" {
				guardExpr = fmt.Sprintf(" && guard(\"%s\")", t.Guard)
			}
			fmt.Fprintf(&b, "            state == %sState.%s && msg == MessageType.%s%s ->\n",
				typeName, t.From,
				kotlinPascalCase(string(t.On.Msg)),
				guardExpr)
			fmt.Fprintf(&b, "                %sState.%s\n", typeName, t.To)
		}
		b.WriteString("            else -> null\n")
		b.WriteString("        }\n")
		b.WriteString("        if (newState != null) state = newState\n")
		b.WriteString("        return newState\n")
		b.WriteString("    }\n\n")

		// step
		b.WriteString("    /** Attempt an internal transition. Returns the new state, or null if none available. */\n")
		fmt.Fprintf(&b, "    fun step(guard: (String) -> Boolean = { true }): %sState? {\n", typeName)

		// Group internal transitions by from-state.
		byFrom := map[State][]Transition{}
		for _, t := range a.Transitions {
			if t.On.Kind == TriggerInternal {
				byFrom[t.From] = append(byFrom[t.From], t)
			}
		}

		b.WriteString("        val newState = when {\n")
		for _, s := range states {
			ts := byFrom[s]
			if len(ts) == 0 {
				continue
			}
			for _, t := range ts {
				guardExpr := ""
				if t.Guard != "" {
					guardExpr = fmt.Sprintf(" && guard(\"%s\")", t.Guard)
				}
				fmt.Fprintf(&b, "            state == %sState.%s%s ->\n",
					typeName, s, guardExpr)
				fmt.Fprintf(&b, "                %sState.%s\n", typeName, t.To)
			}
		}
		b.WriteString("            else -> null\n")
		b.WriteString("        }\n")
		b.WriteString("        if (newState != null) state = newState\n")
		b.WriteString("        return newState\n")
		b.WriteString("    }\n")

		b.WriteString("}\n\n")
	}

	_, err := io.WriteString(w, b.String())
	return err
}

func kotlinTypeName(name string) string {
	if len(name) == 0 {
		return name
	}
	return strings.ToUpper(name[:1]) + name[1:]
}

// kotlinPascalCase converts snake_case to PascalCase for Kotlin enum values.
// "pair_begin" -> "PairBegin", "token_valid" -> "TokenValid"
func kotlinPascalCase(s string) string {
	var result []rune
	nextUpper := true
	for _, r := range s {
		if r == '_' {
			nextUpper = true
			continue
		}
		if nextUpper {
			result = append(result, unicode.ToUpper(r))
			nextUpper = false
		} else {
			result = append(result, r)
		}
	}
	return string(result)
}
