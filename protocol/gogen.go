// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

package protocol

import (
	"fmt"
	"io"
	"strings"
)

// ExportGo writes a Go source file that constructs the Protocol
// value. The generated file imports the protocol package and defines
// a function returning *Protocol.
func (p *Protocol) ExportGo(w io.Writer, pkgName, funcName string) error {
	var b strings.Builder

	b.WriteString("// Copyright 2026 Marcelo Cantos\n")
	b.WriteString("// SPDX-License-Identifier: Apache-2.0\n\n")
	b.WriteString("// Code generated from protocol/*.yaml. DO NOT EDIT.\n\n")
	fmt.Fprintf(&b, "package %s\n\n", pkgName)

	// State constants per actor.
	for _, a := range p.Actors {
		states := collectStates(a)
		prefix := goConstPrefix(a.Name)
		b.WriteString("// " + a.Name + " states.\n")
		b.WriteString("const (\n")
		for _, s := range states {
			fmt.Fprintf(&b, "\t%s%s State = %q\n", prefix, s, s)
		}
		b.WriteString(")\n\n")
	}

	// Message type constants.
	b.WriteString("// Message types.\n")
	b.WriteString("const (\n")
	for _, m := range p.Messages {
		fmt.Fprintf(&b, "\tMsg%s MsgType = %q\n", goCamel(string(m.Type)), m.Type)
	}
	b.WriteString(")\n\n")

	// Guard constants.
	b.WriteString("// Guards.\n")
	b.WriteString("const (\n")
	for _, g := range p.Guards {
		fmt.Fprintf(&b, "\tGuard%s GuardID = %q\n", goCamel(string(g.ID)), g.ID)
	}
	b.WriteString(")\n\n")

	// Action constants.
	actions := map[string]bool{}
	for _, a := range p.Actors {
		for _, t := range a.Transitions {
			if t.Do != "" {
				actions[string(t.Do)] = true
			}
		}
	}
	if len(actions) > 0 {
		b.WriteString("// Actions.\n")
		b.WriteString("const (\n")
		for id := range actions {
			fmt.Fprintf(&b, "\tAction%s ActionID = %q\n", goCamel(id), id)
		}
		b.WriteString(")\n\n")
	}

	fmt.Fprintf(&b, "func %s() *Protocol {\n", funcName)
	b.WriteString("\treturn &Protocol{\n")
	fmt.Fprintf(&b, "\t\tName: %q,\n", p.Name)

	// Actors
	b.WriteString("\t\tActors: []Actor{\n")
	for _, a := range p.Actors {
		fmt.Fprintf(&b, "\t\t\t{Name: %q, Initial: %q, Transitions: []Transition{\n", a.Name, a.Initial)
		for _, t := range a.Transitions {
			b.WriteString("\t\t\t\t{")
			fmt.Fprintf(&b, "From: %q, To: %q, ", t.From, t.To)
			if t.On.Kind == TriggerRecv {
				fmt.Fprintf(&b, "On: Recv(%q)", t.On.Msg)
			} else {
				fmt.Fprintf(&b, "On: Internal(%q)", t.On.Desc)
			}
			if t.Guard != "" {
				fmt.Fprintf(&b, ", Guard: %q", t.Guard)
			}
			if t.Do != "" {
				fmt.Fprintf(&b, ", Do: %q", t.Do)
			}
			if len(t.Sends) > 0 {
				b.WriteString(", Sends: []Send{")
				for _, s := range t.Sends {
					fmt.Fprintf(&b, "{To: %q, Msg: %q", s.To, s.Msg)
					if len(s.Fields) > 0 {
						b.WriteString(", Fields: map[string]string{")
						for k, v := range s.Fields {
							fmt.Fprintf(&b, "%q: %q, ", k, v)
						}
						b.WriteString("}")
					}
					b.WriteString("}, ")
				}
				b.WriteString("}")
			}
			if len(t.Updates) > 0 {
				b.WriteString(", Updates: []VarUpdate{")
				for _, u := range t.Updates {
					fmt.Fprintf(&b, "{Var: %q, Expr: %q}, ", u.Var, u.Expr)
				}
				b.WriteString("}")
			}
			b.WriteString("},\n")
		}
		b.WriteString("\t\t\t}},\n")
	}
	b.WriteString("\t\t},\n")

	// Messages
	b.WriteString("\t\tMessages: []Message{\n")
	for _, m := range p.Messages {
		fmt.Fprintf(&b, "\t\t\t{Type: %q, From: %q, To: %q, Desc: %q},\n", m.Type, m.From, m.To, m.Desc)
	}
	b.WriteString("\t\t},\n")

	// Vars
	b.WriteString("\t\tVars: []VarDef{\n")
	for _, v := range p.Vars {
		fmt.Fprintf(&b, "\t\t\t{Name: %q, Initial: %q, Desc: %q},\n", v.Name, v.Initial, v.Desc)
	}
	b.WriteString("\t\t},\n")

	// Guards
	b.WriteString("\t\tGuards: []GuardDef{\n")
	for _, g := range p.Guards {
		fmt.Fprintf(&b, "\t\t\t{ID: %q, Expr: %q},\n", g.ID, g.Expr)
	}
	b.WriteString("\t\t},\n")

	// Operators
	b.WriteString("\t\tOperators: []Operator{\n")
	for _, op := range p.Operators {
		fmt.Fprintf(&b, "\t\t\t{Name: %q, Params: %q, Expr: %q, Desc: %q},\n",
			op.Name, op.Params, op.Expr, op.Desc)
	}
	b.WriteString("\t\t},\n")

	// AdvActions
	b.WriteString("\t\tAdvActions: []AdvAction{\n")
	for _, aa := range p.AdvActions {
		fmt.Fprintf(&b, "\t\t\t{Name: %q, Desc: %q, Code: %q},\n", aa.Name, aa.Desc, aa.Code)
	}
	b.WriteString("\t\t},\n")

	// Properties
	b.WriteString("\t\tProperties: []Property{\n")
	for _, prop := range p.Properties {
		kindStr := "Invariant"
		if prop.Kind == Liveness {
			kindStr = "Liveness"
		}
		fmt.Fprintf(&b, "\t\t\t{Name: %q, Kind: %s, Expr: %q, Desc: %q},\n",
			prop.Name, kindStr, prop.Expr, prop.Desc)
	}
	b.WriteString("\t\t},\n")

	fmt.Fprintf(&b, "\t\tChannelBound: %d,\n", p.ChannelBound)
	fmt.Fprintf(&b, "\t\tOneShot: %v,\n", p.OneShot)

	b.WriteString("\t}\n}\n\n")

	// --- Structs ---
	needsFrozen := false
	for _, sd := range p.Structs {
		for _, f := range sd.Fields {
			if f.Type == VarSetString {
				needsFrozen = true
			}
		}
	}

	if len(p.Structs) > 0 {
		b.WriteString("// --- Structs ---\n\n")
		for _, sd := range p.Structs {
			if sd.Desc != "" {
				fmt.Fprintf(&b, "// %s\n", sd.Desc)
			}
			fmt.Fprintf(&b, "type %s struct {\n", goCamel(sd.Name))
			for _, f := range sd.Fields {
				fmt.Fprintf(&b, "\t%s %s", goCamel(f.Name), goType(f.Type))
				if f.Desc != "" {
					fmt.Fprintf(&b, " // %s", f.Desc)
				}
				b.WriteString("\n")
			}
			b.WriteString("}\n\n")

			// Converter: vars map -> struct.
			structName := goCamel(sd.Name)
			fmt.Fprintf(&b, "// %sFromVars constructs a %s from a flat variable map.\n", structName, structName)
			fmt.Fprintf(&b, "func %sFromVars(m map[string]any) %s {\n", structName, structName)
			fmt.Fprintf(&b, "\treturn %s{\n", structName)
			for _, f := range sd.Fields {
				fmt.Fprintf(&b, "\t\t%s: %s,\n", goCamel(f.Name), goVarCast(f.Name, f.Type))
			}
			b.WriteString("\t}\n}\n\n")

			// Converter: struct -> vars map.
			fmt.Fprintf(&b, "// ToVars writes a %s into a flat variable map.\n", structName)
			fmt.Fprintf(&b, "func (s %s) ToVars(vars map[string]any) {\n", structName)
			for _, f := range sd.Fields {
				fmt.Fprintf(&b, "\tvars[%q] = s.%s\n", f.Name, goCamel(f.Name))
			}
			b.WriteString("}\n\n")
		}
	}

	// --- Typed variable state ---
	if len(p.Vars) > 0 {
		b.WriteString("// --- Variable state ---\n\n")
		b.WriteString("// Vars holds the complete protocol variable state.\n")
		b.WriteString("type Vars struct {\n")
		for _, v := range p.Vars {
			fmt.Fprintf(&b, "\t%s %s", goCamel(v.Name), goType(v.Type))
			if v.Desc != "" {
				fmt.Fprintf(&b, " // %s", v.Desc)
			}
			b.WriteString("\n")
		}
		b.WriteString("}\n\n")

		// FromMap converter.
		b.WriteString("// VarsFromMap constructs Vars from a flat variable map.\n")
		b.WriteString("func VarsFromMap(m map[string]any) Vars {\n")
		b.WriteString("\treturn Vars{\n")
		for _, v := range p.Vars {
			fmt.Fprintf(&b, "\t\t%s: %s,\n", goCamel(v.Name), goVarCast(v.Name, v.Type))
		}
		b.WriteString("\t}\n}\n\n")

		// ToMap converter.
		b.WriteString("// ToMap writes Vars into a flat variable map.\n")
		b.WriteString("func (v Vars) ToMap(m map[string]any) {\n")
		for _, v := range p.Vars {
			fmt.Fprintf(&b, "\tm[%q] = v.%s\n", v.Name, goCamel(v.Name))
		}
		b.WriteString("}\n\n")
	}

	// Import annotation for frozen if needed.
	if needsFrozen {
		b.WriteString("// Note: this package requires github.com/arr-ai/frozen\n")
	}

	_, err := io.WriteString(w, b.String())
	return err
}

// goType maps a VarType to its Go type string.
func goType(t VarType) string {
	switch t {
	case VarInt:
		return "int"
	case VarBool:
		return "bool"
	case VarSetString:
		return "frozen.Set[string]"
	default:
		return "string"
	}
}

// goVarCast generates a type assertion expression for reading from a map[string]any named "m".
func goVarCast(varName string, t VarType) string {
	key := fmt.Sprintf("m[%q]", varName)
	// The map parameter in generated functions is always named "m".
	switch t {
	case VarInt:
		return key + ".(int)"
	case VarBool:
		return key + ".(bool)"
	case VarSetString:
		return key + ".(frozen.Set[string])"
	default:
		return key + ".(string)"
	}
}

// goConstPrefix maps actor names to Go constant prefixes.
// "server" -> "Server", "ios" -> "App", "cli" -> "CLI"
func goConstPrefix(actor string) string {
	switch actor {
	case "ios":
		return "App"
	case "cli":
		return "CLI"
	default:
		if len(actor) == 0 {
			return ""
		}
		return strings.ToUpper(actor[:1]) + actor[1:]
	}
}

// goCamel converts a snake_case identifier to GoCamelCase.
// "pair_begin" -> "PairBegin", "token_valid" -> "TokenValid"
func goCamel(s string) string {
	var b strings.Builder
	upper := true
	for _, r := range s {
		if r == '_' || r == '-' {
			upper = true
			continue
		}
		if upper {
			b.WriteRune(rune(strings.ToUpper(string(r))[0]))
			upper = false
		} else {
			b.WriteRune(r)
		}
	}
	return b.String()
}
