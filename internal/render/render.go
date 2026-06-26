package render

import (
	"fmt"
	"strings"

	"github.com/GatosTheDog/versous/internal/agent"
	"github.com/GatosTheDog/versous/internal/specs"
)

func Render(r agent.Report) string {
	var b strings.Builder
	fmt.Fprintf(&b, "=== Versous: %s vs %s ===\n\n", r.ProductA, r.ProductB)
	for _, aspect := range r.Aspects {
		fmt.Fprintf(&b, "[%s]  → %s\n", aspect.Aspect, aspect.Winner)
		fmt.Fprintf(&b, "%s\n\n", aspect.Summary)
	}
	fmt.Fprintf(&b, "OVERALL WINNER: %s\n", r.Winner)

	if r.SpecA != (specs.Spec{}) {
		fmt.Fprintf(&b, "\n── Specs ──────────────────────────────\n")
		fmt.Fprintf(&b, "%-14s %-22s %s\n", "", r.ProductA, r.ProductB)
		fmt.Fprintf(&b, "%-14s %-22s %s\n", "Processor", r.SpecA.Processor, r.SpecB.Processor)
		fmt.Fprintf(&b, "%-14s %-22s %s\n", "RAM", r.SpecA.RAM, r.SpecB.RAM)
		fmt.Fprintf(&b, "%-14s %-22s %s\n", "Battery", r.SpecA.Battery, r.SpecB.Battery)
		linesA := wrapSpec(r.SpecA.Camera, 22)
		linesB := wrapSpec(r.SpecB.Camera, 22)
		for i := range max(len(linesA), len(linesB)) {
			a, bv := "", ""
			if i < len(linesA) {
				a = linesA[i]
			}
			if i < len(linesB) {
				bv = linesB[i]
			}
			label := ""
			if i == 0 {
				label = "Camera"
			}
			fmt.Fprintf(&b, "%-14s %-22s %s\n", label, a, bv)
		}
		fmt.Fprintf(&b, "%-14s %-22s %s\n", "Price", r.SpecA.Price, r.SpecB.Price)
	}

	return b.String()
}

func wrapSpec(s string, width int) []string {
	if len(s) <= width {
		return []string{s}
	}
	var lines []string
	for len(s) > width {
		cut := width
		for cut > 0 && s[cut] != ' ' {
			cut--
		}
		if cut == 0 {
			cut = width
		}
		lines = append(lines, strings.TrimSpace(s[:cut]))
		s = strings.TrimSpace(s[cut:])
	}
	if s != "" {
		lines = append(lines, s)
	}
	return lines
}
