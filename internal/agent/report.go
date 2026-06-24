package agent

import "github.com/GatosTheDog/versous/internal/rag"

type Report struct {
	ProductA string
	ProductB string
	Aspects  []rag.Verdict
	Winner   string
}
