package models

type EventGridRouteAsset struct {
	ID                  string   `json:"id"`
	Name                string   `json:"route"`
	Source              *string  `json:"source,omitempty"`
	Destination         *string  `json:"destination,omitempty"`
	DestinationType     string   `json:"destination_type"`
	Classification      string   `json:"classification"`
	SourceID            string   `json:"source_id"`
	SourceType          string   `json:"source_type"`
	DestinationTargetID *string  `json:"destination_target_id,omitempty"`
	ExternalDelivery    bool     `json:"external_delivery"`
	ProvisioningState   *string  `json:"provisioning_state,omitempty"`
	IdentityType        *string  `json:"identity_type,omitempty"`
	IdentityID          *string  `json:"identity_id,omitempty"`
	EventDeliverySchema *string  `json:"event_delivery_schema,omitempty"`
	IncludedEventTypes  []string `json:"included_event_types"`
	Summary             string   `json:"summary"`
	RelatedIDs          []string `json:"related_ids"`
}

type EventGridMetadata = RuntimeCommandMetadata

type EventGridOutput struct {
	Findings []Finding             `json:"findings"`
	Issues   []Issue               `json:"issues"`
	Metadata EventGridMetadata     `json:"metadata"`
	Routes   []EventGridRouteAsset `json:"routes"`
}
