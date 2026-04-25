package commands

import (
	"context"
	"time"

	"harrierops-azure/internal/artifacts"
	"harrierops-azure/internal/models"
	"harrierops-azure/internal/providers"
)

type identityControlFutures struct {
	permissions asyncCommandOutput[models.PermissionsOutput]
	rbac        asyncCommandOutput[models.RbacOutput]
}

type identityControlData struct {
	permissions      models.PermissionsOutput
	rbac             models.RbacOutput
	evidence         persistencePrincipalEvidence
	sessionArtifacts []models.SessionArtifact
}

func startIdentityControlFutures(
	group commandOutputGroup,
	ctx context.Context,
	provider providers.Provider,
	now func() time.Time,
	request Request,
) identityControlFutures {
	expected := helperArtifactExpectedSessions(ctx, request, provider, now, "permissions", "rbac")
	return startIdentityControlFuturesWithExpected(group, ctx, provider, now, request, expected)
}

func startIdentityControlFuturesWithExpected(
	group commandOutputGroup,
	ctx context.Context,
	provider providers.Provider,
	now func() time.Time,
	request Request,
	expected map[string]artifacts.ExpectedSession,
) identityControlFutures {
	return identityControlFutures{
		permissions: runPermissionsOutput(group, ctx, request, provider, now, expected),
		rbac:        runRBACOutput(group, ctx, request, provider, now, expected),
	}
}

func (futures identityControlFutures) wait() (identityControlData, error) {
	permissions, rbac, evidence, sessionArtifacts, err := waitPermissionsRBACBundle(futures.permissions, futures.rbac)
	if err != nil {
		return identityControlData{}, err
	}
	return identityControlData{
		permissions:      permissions,
		rbac:             rbac,
		evidence:         evidence,
		sessionArtifacts: sessionArtifacts,
	}, nil
}

func startPermissionsFuture(
	group commandOutputGroup,
	ctx context.Context,
	provider providers.Provider,
	now func() time.Time,
	request Request,
) asyncCommandOutput[models.PermissionsOutput] {
	expected := helperArtifactExpectedSessions(ctx, request, provider, now, "permissions")
	return runPermissionsOutput(group, ctx, request, provider, now, expected)
}
