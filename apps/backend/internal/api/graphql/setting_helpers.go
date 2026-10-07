package graphql

import (
	"github.com/playingwithclouds/veil/internal/api/graphql/model"
	"github.com/playingwithclouds/veil/internal/settings"
)

// mergeKindLimits overlays per-kind rate-limit input onto the current settings,
// returning the merged map ready to persist.
func mergeKindLimits(current map[string]settings.KindRateLimit, inputs []*model.KindRateLimitInput) map[string]any {
	out := map[string]any{}
	for kind, limit := range current {
		out[kind] = map[string]any{
			"maxConcurrent":   limit.MaxConcurrent,
			"retryInitialMs":  limit.RetryInitialMs,
			"retryMultiplier": limit.RetryMultiplier,
			"retryMaxMs":      limit.RetryMaxMs,
		}
	}
	for _, in := range inputs {
		base, ok := current[in.Kind]
		if !ok {
			base = settings.KindRateLimit{}
		}
		if in.MaxConcurrent != nil {
			base.MaxConcurrent = *in.MaxConcurrent
		}
		if in.RetryInitialMs != nil {
			base.RetryInitialMs = *in.RetryInitialMs
		}
		if in.RetryMultiplier != nil {
			base.RetryMultiplier = *in.RetryMultiplier
		}
		if in.RetryMaxMs != nil {
			base.RetryMaxMs = *in.RetryMaxMs
		}
		out[in.Kind] = map[string]any{
			"maxConcurrent":   base.MaxConcurrent,
			"retryInitialMs":  base.RetryInitialMs,
			"retryMultiplier": base.RetryMultiplier,
			"retryMaxMs":      base.RetryMaxMs,
		}
	}
	return out
}
