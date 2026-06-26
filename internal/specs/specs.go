package specs

import (
	"context"
	"fmt"
	"strings"

	"github.com/GatosTheDog/versous/internal/llm"
)

type Spec struct {
	Display, Processor, RAM, Battery, Camera, Price string
}

func Fetch(ctx context.Context, llmClient *llm.Client, product string) (Spec, error) {
	prompt := fmt.Sprintf(`Give key specs for %s in exactly this format, no extra text:
		Display: ...
		Processor: ...
		RAM: ...
		Battery: ...
		Camera: ...
		Price: ...`, product)
	result, err := llmClient.Generate(ctx, prompt)
	if err != nil {
		return Spec{}, fmt.Errorf("fetch specs: %w", err)
	}

	spec := Spec{}
	for _, line := range strings.Split(result, "\n") {
		parts := strings.SplitN(line, ": ", 2)
		if len(parts) != 2 {
			continue
		}
		switch strings.TrimSpace(parts[0]) {
		case "Display":
			spec.Display = strings.TrimSpace(parts[1])
		case "Processor":
			spec.Processor = strings.TrimSpace(parts[1])
		case "RAM":
			spec.RAM = strings.TrimSpace(parts[1])
		case "Battery":
			spec.Battery = strings.TrimSpace(parts[1])
		case "Camera":
			spec.Camera = strings.TrimSpace(parts[1])
		case "Price":
			spec.Price = strings.TrimSpace(parts[1])
		}
	}
	return spec, nil
}
