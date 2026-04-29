package council

import (
	"context"
	"fmt"

	"github.com/AnimusHQ/news/internal/models/adapters"
)

// Runner executes a panel of model providers and aggregates their reviews.
type Runner struct {
	Providers []adapters.Provider
}

func NewRunner(providers []adapters.Provider) Runner {
	return Runner{Providers: append([]adapters.Provider(nil), providers...)}
}

func (r Runner) Run(ctx context.Context, req adapters.Request) (Report, error) {
	if len(r.Providers) == 0 {
		return Report{}, fmt.Errorf("council runner requires at least one provider")
	}

	reviews := make([]ModelReview, 0, len(r.Providers))
	for _, provider := range r.Providers {
		response, err := provider.Run(ctx, req)
		if err != nil {
			return Report{}, fmt.Errorf("provider %s failed: %w", provider.ID(), err)
		}
		reviews = append(reviews, response.Review)
	}

	return Aggregate(reviews)
}
