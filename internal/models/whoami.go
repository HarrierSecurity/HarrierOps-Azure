package models

type WhoAmIOutput struct {
	EffectiveScopes []ScopeRef      `json:"effective_scopes"`
	Issues          []Issue         `json:"issues"`
	Metadata        WhoAmIMetadata  `json:"metadata"`
	Principal       Principal       `json:"principal"`
	Subscription    SubscriptionRef `json:"subscription"`
	TenantID        string          `json:"tenant_id"`
}
