package models

type OutputMode string
type RoleTrustsMode string
type RoleTrustFollowOnKind string
type WorkloadExposure string
type EnvVarTargetService string
type TokenCredentialSurfaceType string
type TokenCredentialNextReviewKind string

const (
	OutputTable OutputMode = "table"
	OutputJSON  OutputMode = "json"
	OutputCSV   OutputMode = "csv"

	RoleTrustsModeFast    RoleTrustsMode = "fast"
	RoleTrustsModeFull    RoleTrustsMode = "full"
	RoleTrustsModeFastOld RoleTrustsMode = "fast-old"
	RoleTrustsModeFullOld RoleTrustsMode = "full-old"

	RoleTrustFollowOnPrivilegeConfirmation RoleTrustFollowOnKind = "privilege-confirmation"
	RoleTrustFollowOnOwnershipReview       RoleTrustFollowOnKind = "ownership-review"
	RoleTrustFollowOnOutsideTenant         RoleTrustFollowOnKind = "outside-tenant"

	WorkloadExposureNone    WorkloadExposure = ""
	WorkloadExposurePublic  WorkloadExposure = "public"
	WorkloadExposureExposed WorkloadExposure = "exposed"

	EnvVarTargetServiceNone     EnvVarTargetService = ""
	EnvVarTargetServiceStorage  EnvVarTargetService = "storage"
	EnvVarTargetServiceDatabase EnvVarTargetService = "database"

	TokenCredentialSurfacePlainTextSecret       TokenCredentialSurfaceType = "plain-text-secret"
	TokenCredentialSurfaceKeyVaultReference     TokenCredentialSurfaceType = "keyvault-reference"
	TokenCredentialSurfaceManagedIdentityToken  TokenCredentialSurfaceType = "managed-identity-token"
	TokenCredentialSurfaceDeploymentOutput      TokenCredentialSurfaceType = "deployment-output"
	TokenCredentialSurfaceLinkedDeploymentAsset TokenCredentialSurfaceType = "linked-deployment-content"

	TokenCredentialReviewEnvVarsSettingContext         TokenCredentialNextReviewKind = "env-vars-setting-context"
	TokenCredentialReviewEndpointsIngressAndControl    TokenCredentialNextReviewKind = "endpoints-ingress-and-control"
	TokenCredentialReviewManagedIdentityAndPermissions TokenCredentialNextReviewKind = "managed-identity-and-permissions"
	TokenCredentialReviewKeyVaultAndManagedIdentity    TokenCredentialNextReviewKind = "keyvault-and-managed-identity"
	TokenCredentialReviewKeyVaultBoundary              TokenCredentialNextReviewKind = "keyvault-boundary"
	TokenCredentialReviewARMDeploymentOutputs          TokenCredentialNextReviewKind = "arm-deployment-outputs"
	TokenCredentialReviewARMDeploymentLinks            TokenCredentialNextReviewKind = "arm-deployment-links"
	TokenCredentialReviewWorkloadContext               TokenCredentialNextReviewKind = "workload-context"
)

func (mode OutputMode) Valid() bool {
	switch mode {
	case OutputTable, OutputJSON, OutputCSV:
		return true
	default:
		return false
	}
}

func (mode RoleTrustsMode) Valid() bool {
	switch mode {
	case RoleTrustsModeFast, RoleTrustsModeFull, RoleTrustsModeFastOld, RoleTrustsModeFullOld:
		return true
	default:
		return false
	}
}

func (mode RoleTrustsMode) Semantic() RoleTrustsMode {
	switch mode {
	case RoleTrustsModeFastOld:
		return RoleTrustsModeFast
	case RoleTrustsModeFullOld:
		return RoleTrustsModeFull
	default:
		return mode
	}
}

func (mode RoleTrustsMode) Legacy() bool {
	switch mode {
	case RoleTrustsModeFastOld, RoleTrustsModeFullOld:
		return true
	default:
		return false
	}
}

type Metadata struct {
	Command            string  `json:"command"`
	DevOpsOrganization *string `json:"devops_organization"`
	GeneratedAt        string  `json:"generated_at"`
	SchemaVersion      string  `json:"schema_version"`
	SubscriptionID     *string `json:"subscription_id"`
	TenantID           *string `json:"tenant_id"`
	TokenSource        *string `json:"token_source"`
}

type RuntimeCommandMetadata struct {
	Command        string  `json:"command"`
	GeneratedAt    string  `json:"generated_at"`
	SchemaVersion  string  `json:"schema_version"`
	SubscriptionID *string `json:"subscription_id"`
	TenantID       *string `json:"tenant_id"`
	TokenSource    *string `json:"token_source"`
}

type WhoAmIMetadata struct {
	AuthMode *string `json:"auth_mode"`
	Metadata
}

type ScopedCommandMetadata struct {
	SchemaVersion      string  `json:"schema_version"`
	Command            string  `json:"command"`
	GeneratedAt        string  `json:"generated_at"`
	TenantID           *string `json:"tenant_id"`
	SubscriptionID     *string `json:"subscription_id"`
	DevOpsOrganization *string `json:"devops_organization"`
	TokenSource        *string `json:"token_source"`
	AuthMode           *string `json:"auth_mode"`
}

type PermissionsMetadata = ScopedCommandMetadata

type Issue struct {
	Kind    string            `json:"kind"`
	Message string            `json:"message"`
	Scope   string            `json:"scope,omitempty"`
	Context map[string]string `json:"context,omitempty"`
}

type Finding struct {
	ID          string   `json:"id"`
	Title       string   `json:"title"`
	Severity    string   `json:"severity"`
	Description string   `json:"description"`
	RelatedIDs  []string `json:"related_ids,omitempty"`
}

type SubscriptionRef struct {
	DisplayName string `json:"display_name,omitempty"`
	ID          string `json:"id"`
	State       string `json:"state,omitempty"`
}

type ScopeRef struct {
	DisplayName string `json:"display_name,omitempty"`
	ID          string `json:"id"`
	ScopeType   string `json:"scope_type"`
}

type Principal struct {
	DisplayName   string `json:"display_name,omitempty"`
	ID            string `json:"id"`
	PrincipalType string `json:"principal_type"`
	TenantID      string `json:"tenant_id,omitempty"`
}

type TopResourceTypes map[string]int

func StringPtr(value string) *string {
	if value == "" {
		return nil
	}
	return &value
}
