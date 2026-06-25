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
		fmt.Fprintf(&b, "%-14s %-22s %s\n", "Camera", r.SpecA.Camera, r.SpecB.Camera)
		fmt.Fprintf(&b, "%-14s %-22s %s\n", "Price", r.SpecA.Price, r.SpecB.Price)
	}
	return b.String()
}
