package models

type WebJobAsset struct {
	DetailedStatus     *string  `json:"detailed_status,omitempty"`
	ID                 string   `json:"id"`
	JobType            *string  `json:"job_type,omitempty"`
	LatestRunStatus    *string  `json:"latest_run_status,omitempty"`
	LatestRunTrigger   *string  `json:"latest_run_trigger,omitempty"`
	Location           string   `json:"location"`
	Mode               string   `json:"mode"`
	Name               string   `json:"name"`
	ParentAppID        string   `json:"parent_app_id"`
	ParentAppName      string   `json:"parent_app_name"`
	ParentHostname     *string  `json:"parent_hostname,omitempty"`
	ParentIdentityIDs  []string `json:"parent_identity_ids,omitempty"`
	ParentIdentityType *string  `json:"parent_identity_type,omitempty"`
	RelatedIDs         []string `json:"related_ids"`
	ResourceGroup      string   `json:"resource_group"`
	RunCommand         *string  `json:"run_command,omitempty"`
	ScheduleExpression *string  `json:"schedule_expression,omitempty"`
	SchedulerLogsURL   *string  `json:"scheduler_logs_url,omitempty"`
	Status             *string  `json:"status,omitempty"`
	Summary            string   `json:"summary"`
}

type WebJobsMetadata = RuntimeCommandMetadata

type WebJobsOutput struct {
	Findings []Finding       `json:"findings"`
	Issues   []Issue         `json:"issues"`
	Metadata WebJobsMetadata `json:"metadata"`
	WebJobs  []WebJobAsset   `json:"webjobs"`
}
