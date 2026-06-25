package agent

import (
	"github.com/GatosTheDog/versous/internal/rag"
	"github.com/GatosTheDog/versous/internal/specs"
)

type Report struct {
	ProductA     string
	ProductB     string
	Aspects      []rag.Verdict
	SpecA, SpecB specs.Spec
	Winner       string
}
