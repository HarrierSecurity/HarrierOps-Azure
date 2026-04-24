package commands

import (
	"context"
	"sort"
	"time"

	"harrierops-azure/internal/models"
	"harrierops-azure/internal/providers"
)

func relayHandler(provider providers.Provider, now func() time.Time) Handler {
	return func(ctx context.Context, request Request) (any, error) {
		facts, err := provider.Relay(ctx, request.Tenant, request.Subscription)
		if err != nil {
			return nil, err
		}
		namespaces := append([]models.RelayNamespaceAsset{}, facts.Namespaces...)
		sort.SliceStable(namespaces, func(i, j int) bool {
			if relayPriority(namespaces[i]) != relayPriority(namespaces[j]) {
				return relayPriority(namespaces[i]) > relayPriority(namespaces[j])
			}
			if relayIntValue(namespaces[i].HybridConnectionCount) != relayIntValue(namespaces[j].HybridConnectionCount) {
				return relayIntValue(namespaces[i].HybridConnectionCount) > relayIntValue(namespaces[j].HybridConnectionCount)
			}
			return namespaces[i].Name < namespaces[j].Name
		})
		return models.RelayOutput{
			Findings:   []models.Finding{},
			Issues:     facts.Issues,
			Metadata:   runtimeCommandMetadata("relay", now, facts.TenantID, facts.SubscriptionID),
			Namespaces: namespaces,
		}, nil
	}
}

func relayPriority(namespace models.RelayNamespaceAsset) int {
	score := 0
	if relayIntValue(namespace.HybridConnectionCount) > 0 {
		score += 3
	}
	if relayIntValue(namespace.AuthorizationRuleCount) > 0 {
		score += 1
	}
	for _, connection := range namespace.HybridConnections {
		if relayIntValue(connection.ListenerCount) > 0 {
			score += 2
			break
		}
	}
	return score
}

func relayIntValue(value *int) int {
	if value == nil {
		return 0
	}
	return *value
}
