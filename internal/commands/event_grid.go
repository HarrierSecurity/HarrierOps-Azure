package commands

import (
	"context"
	"strings"
	"time"

	"harrierops-azure/internal/models"
	"harrierops-azure/internal/providers"
)

func eventGridHandler(provider providers.Provider, now func() time.Time) Handler {
	return func(ctx context.Context, request Request) (any, error) {
		facts, err := provider.EventGrid(ctx, request.Tenant, request.Subscription)
		if err != nil {
			return nil, err
		}

		routes := sortedByLess(facts.Routes, eventGridLess)
		for index := range routes {
			routes[index] = decorateEventGridArtifact(routes[index])
		}

		return models.EventGridOutput{
			Findings: []models.Finding{},
			Issues:   facts.Issues,
			Metadata: withRuntimeArtifactContext(runtimeCommandMetadata("event-grid", now, facts.TenantID, facts.SubscriptionID), request, facts.CurrentPrincipal, facts.AuthMode, facts.TokenSource),
			Routes:   routes,
		}, nil
	}
}

func eventGridLess(left models.EventGridRouteAsset, right models.EventGridRouteAsset) bool {
	leftRank := eventGridClassificationRank(left.Classification)
	rightRank := eventGridClassificationRank(right.Classification)
	if leftRank != rightRank {
		return leftRank < rightRank
	}

	if left.ExternalDelivery != right.ExternalDelivery {
		return left.ExternalDelivery
	}

	leftAzureSource := strings.HasPrefix(strings.ToLower(left.SourceID), "/subscriptions/")
	rightAzureSource := strings.HasPrefix(strings.ToLower(right.SourceID), "/subscriptions/")
	if leftAzureSource != rightAzureSource {
		return leftAzureSource
	}

	if left.Name != right.Name {
		return left.Name < right.Name
	}
	return left.ID < right.ID
}

func eventGridClassificationRank(classification string) int {
	switch classification {
	case "execution-capable":
		return 0
	case "external-callback":
		return 1
	default:
		return 2
	}
}

func decorateEventGridArtifact(route models.EventGridRouteAsset) models.EventGridRouteAsset {
	route.Source = compactArtifactValue(eventGridArtifactSource(route))
	route.Destination = compactArtifactValue(eventGridArtifactDestination(route))
	return route
}

func eventGridArtifactSource(route models.EventGridRouteAsset) string {
	if route.SourceID == "" {
		return "-"
	}
	if route.SourceType != "" {
		return resourceNameFromScopedID(route.SourceID) + " (" + route.SourceType + ")"
	}
	return resourceNameFromScopedID(route.SourceID)
}

func eventGridArtifactDestination(route models.EventGridRouteAsset) string {
	if route.ExternalDelivery {
		return "webhook (redacted)"
	}
	if route.DestinationTargetID != nil && *route.DestinationTargetID != "" {
		return resourceNameFromScopedID(*route.DestinationTargetID)
	}
	if route.DestinationType == "" {
		return "-"
	}
	return route.DestinationType
}

func resourceNameFromScopedID(resourceID string) string {
	text := strings.Trim(strings.TrimSpace(resourceID), "/")
	if text == "" {
		return ""
	}
	parts := strings.Split(text, "/")
	if len(parts) == 0 {
		return text
	}
	if len(parts) == 2 && strings.EqualFold(parts[0], "subscriptions") {
		return "subscription root"
	}
	for index := 0; index < len(parts)-1; index++ {
		if strings.EqualFold(parts[index], "sites") && index+3 < len(parts) && strings.EqualFold(parts[index+2], "functions") {
			return parts[index+1] + "/" + parts[index+3]
		}
	}
	return parts[len(parts)-1]
}
