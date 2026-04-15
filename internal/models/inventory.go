package models

type InventoryOutput struct {
	Issues             []Issue          `json:"issues"`
	Metadata           Metadata         `json:"metadata"`
	ResourceCount      int              `json:"resource_count"`
	ResourceGroupCount int              `json:"resource_group_count"`
	Subscription       SubscriptionRef  `json:"subscription"`
	TopResourceTypes   TopResourceTypes `json:"top_resource_types"`
}
