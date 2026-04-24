package providers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strings"

	"harrierops-azure/internal/models"
)

const (
	graphScope            = "https://graph.microsoft.com/.default"
	graphEndpoint         = "https://graph.microsoft.com/v1.0"
	graphBatchMaxRequests = 20
)

type graphBatchRequest struct {
	Key string
	URL string
}

func (provider AzureProvider) RoleTrusts(ctx context.Context, tenant string, subscription string, mode models.RoleTrustsMode) (RoleTrustsFacts, error) {
	if !mode.Valid() {
		mode = models.RoleTrustsModeFast
	}

	semanticMode := mode.Semantic()

	session, err := provider.session(ctx, tenant, subscription)
	if err != nil {
		return RoleTrustsFacts{}, err
	}

	graphToken, err := accessToken(ctx, session.credential, graphScope)
	if err != nil {
		return RoleTrustsFacts{}, err
	}

	if mode.Legacy() {
		return provider.roleTrustsSerial(ctx, session, graphToken, semanticMode)
	}

	return provider.roleTrustsBatched(ctx, session, graphToken, semanticMode)
}

func (provider AzureProvider) roleTrustsSerial(ctx context.Context, session azureSession, graphToken string, mode models.RoleTrustsMode) (RoleTrustsFacts, error) {
	issues := []models.Issue{}
	trusts := []models.RoleTrustSummary{}

	servicePrincipals, err := graphListObjects(ctx, graphToken, graphCollectionURL("/servicePrincipals", map[string]string{
		"$select": "id,appId,displayName,servicePrincipalType",
	}))
	if err != nil {
		issues = append(issues, issueFromError("role_trusts.service_principals", err))
		servicePrincipals = []map[string]any{}
	}

	servicePrincipalByID := map[string]map[string]any{}
	servicePrincipalByAppID := map[string]map[string]any{}
	for _, item := range servicePrincipals {
		if id := mapStringValue(item, "id"); id != "" {
			servicePrincipalByID[id] = item
		}
		if appID := mapStringValue(item, "appId"); appID != "" {
			servicePrincipalByAppID[appID] = item
		}
	}

	applications := []map[string]any{}
	applicationByAppID := map[string]map[string]any{}
	if mode == models.RoleTrustsModeFull {
		applications, err = graphListObjects(ctx, graphToken, graphCollectionURL("/applications", map[string]string{
			"$select": "id,appId,displayName",
		}))
		if err != nil {
			issues = append(issues, issueFromError("role_trusts.applications", err))
			applications = []map[string]any{}
		}
	}
	for _, application := range applications {
		if appID := mapStringValue(application, "appId"); appID != "" {
			applicationByAppID[appID] = application
		}
	}

	seededAppIDs := make([]string, 0, len(servicePrincipalByAppID))
	for appID := range servicePrincipalByAppID {
		if appID != "" && applicationByAppID[appID] == nil {
			seededAppIDs = append(seededAppIDs, appID)
		}
	}
	sort.Strings(seededAppIDs)

	for _, appID := range seededAppIDs {
		application, lookupErr := graphGetApplicationByAppID(ctx, graphToken, appID)
		if lookupErr != nil {
			issues = append(issues, issueFromError("role_trusts.applications.by_app_id["+appID+"]", lookupErr))
			continue
		}
		if len(application) == 0 {
			continue
		}
		applicationByAppID[appID] = application
		applications = append(applications, application)
	}

	for _, application := range applications {
		appObjectID := mapStringValue(application, "id")
		if appObjectID == "" {
			continue
		}
		appDisplayName := firstNonEmpty(mapStringValue(application, "displayName"), mapStringValue(application, "appId"), appObjectID)
		backingSP := servicePrincipalByAppID[mapStringValue(application, "appId")]

		federatedCredentials, credentialErr := graphListObjects(ctx, graphToken, graphCollectionURL("/applications/"+url.PathEscape(appObjectID)+"/federatedIdentityCredentials", map[string]string{
			"$select": "id,issuer,subject",
		}))
		if credentialErr != nil {
			issues = append(issues, issueFromError("role_trusts.applications["+appObjectID+"].federated_credentials", credentialErr))
			federatedCredentials = []map[string]any{}
		}
		for _, credential := range federatedCredentials {
			trusts = append(trusts, federatedCredentialTrust(application, backingSP, credential, appDisplayName))
		}

		owners, ownerErr := graphListObjects(ctx, graphToken, graphCollectionURL("/applications/"+url.PathEscape(appObjectID)+"/owners", map[string]string{
			"$select": "id,displayName,userPrincipalName,appId,servicePrincipalType",
		}))
		if ownerErr != nil {
			issues = append(issues, issueFromError("role_trusts.applications["+appObjectID+"].owners", ownerErr))
			owners = []map[string]any{}
		}
		for _, owner := range owners {
			ownerID := mapStringValue(owner, "id")
			if ownerID == "" {
				continue
			}
			trusts = append(trusts, applicationOwnerTrust(owner, application, backingSP, appDisplayName))
		}
	}

	for _, servicePrincipal := range servicePrincipals {
		spID := mapStringValue(servicePrincipal, "id")
		if spID == "" {
			continue
		}
		servicePrincipalName := firstNonEmpty(mapStringValue(servicePrincipal, "displayName"), spID)

		owners, ownerErr := graphListObjects(ctx, graphToken, graphCollectionURL("/servicePrincipals/"+url.PathEscape(spID)+"/owners", map[string]string{
			"$select": "id,displayName,userPrincipalName,appId,servicePrincipalType",
		}))
		if ownerErr != nil {
			issues = append(issues, issueFromError("role_trusts.service_principals["+spID+"].owners", ownerErr))
			owners = []map[string]any{}
		}
		for _, owner := range owners {
			ownerID := mapStringValue(owner, "id")
			if ownerID == "" {
				continue
			}
			trusts = append(trusts, servicePrincipalOwnerTrust(owner, servicePrincipal, servicePrincipalName))
		}

		assignments, assignmentErr := graphListObjects(ctx, graphToken, graphCollectionURL("/servicePrincipals/"+url.PathEscape(spID)+"/appRoleAssignments", map[string]string{
			"$select": "id,resourceId",
		}))
		if assignmentErr != nil {
			issues = append(issues, issueFromError("role_trusts.service_principals["+spID+"].app_role_assignments", assignmentErr))
			assignments = []map[string]any{}
		}
		for _, assignment := range assignments {
			resourceID := mapStringValue(assignment, "resourceId")
			resource := servicePrincipalByID[resourceID]
			if resource == nil && resourceID != "" {
				resource, err = graphGetObject(ctx, graphToken, graphObjectURL("/servicePrincipals/"+url.PathEscape(resourceID), map[string]string{
					"$select": "id,appId,displayName,servicePrincipalType",
				}))
				if err != nil {
					issues = append(issues, issueFromError("role_trusts.service_principals["+spID+"].resource["+resourceID+"]", err))
					resource = map[string]any{}
				} else if mapStringValue(resource, "id") != "" {
					servicePrincipalByID[mapStringValue(resource, "id")] = resource
				}
			}
			trusts = append(trusts, appRoleAssignmentTrust(servicePrincipal, servicePrincipalName, assignment, resource))
		}
	}

	return finalizeRoleTrustFacts(session, mode, trusts, issues), nil
}

func (provider AzureProvider) roleTrustsBatched(ctx context.Context, session azureSession, graphToken string, mode models.RoleTrustsMode) (RoleTrustsFacts, error) {
	issues := []models.Issue{}
	trusts := []models.RoleTrustSummary{}

	servicePrincipals, err := graphListObjects(ctx, graphToken, graphCollectionURL("/servicePrincipals", map[string]string{
		"$select": "id,appId,displayName,servicePrincipalType",
	}))
	if err != nil {
		issues = append(issues, issueFromError("role_trusts.service_principals", err))
		servicePrincipals = []map[string]any{}
	}

	servicePrincipalByID := map[string]map[string]any{}
	servicePrincipalByAppID := map[string]map[string]any{}
	for _, item := range servicePrincipals {
		if id := mapStringValue(item, "id"); id != "" {
			servicePrincipalByID[id] = item
		}
		if appID := mapStringValue(item, "appId"); appID != "" {
			servicePrincipalByAppID[appID] = item
		}
	}

	applications := []map[string]any{}
	applicationByAppID := map[string]map[string]any{}
	if mode == models.RoleTrustsModeFull {
		applications, err = graphListObjects(ctx, graphToken, graphCollectionURL("/applications", map[string]string{
			"$select": "id,appId,displayName",
		}))
		if err != nil {
			issues = append(issues, issueFromError("role_trusts.applications", err))
			applications = []map[string]any{}
		}
	}
	for _, application := range applications {
		if appID := mapStringValue(application, "appId"); appID != "" {
			applicationByAppID[appID] = application
		}
	}

	seededAppIDs := make([]string, 0, len(servicePrincipalByAppID))
	for appID := range servicePrincipalByAppID {
		if appID != "" && applicationByAppID[appID] == nil {
			seededAppIDs = append(seededAppIDs, appID)
		}
	}
	sort.Strings(seededAppIDs)

	seededApplicationRequests := make([]graphBatchRequest, 0, len(seededAppIDs))
	for _, appID := range seededAppIDs {
		seededApplicationRequests = append(seededApplicationRequests, graphBatchRequest{
			Key: appID,
			URL: graphCollectionURL("/applications", map[string]string{
				"$filter": "appId eq '" + strings.ReplaceAll(appID, "'", "''") + "'",
				"$select": "id,appId,displayName",
			}),
		})
	}

	seededApplicationsByAppID, seededApplicationErrs := graphBatchListObjectsByKey(ctx, graphToken, seededApplicationRequests)
	for _, appID := range seededAppIDs {
		if lookupErr, ok := seededApplicationErrs[appID]; ok {
			issues = append(issues, issueFromError("role_trusts.applications.by_app_id["+appID+"]", lookupErr))
			continue
		}
		if items := seededApplicationsByAppID[appID]; len(items) > 0 {
			applicationByAppID[appID] = items[0]
			applications = append(applications, items[0])
		}
	}

	applicationFederatedRequests := []graphBatchRequest{}
	applicationOwnerRequests := []graphBatchRequest{}
	for _, application := range applications {
		appObjectID := mapStringValue(application, "id")
		if appObjectID == "" {
			continue
		}
		applicationFederatedRequests = append(applicationFederatedRequests, graphBatchRequest{
			Key: appObjectID,
			URL: graphCollectionURL("/applications/"+url.PathEscape(appObjectID)+"/federatedIdentityCredentials", map[string]string{
				"$select": "id,issuer,subject",
			}),
		})
		applicationOwnerRequests = append(applicationOwnerRequests, graphBatchRequest{
			Key: appObjectID,
			URL: graphCollectionURL("/applications/"+url.PathEscape(appObjectID)+"/owners", map[string]string{
				"$select": "id,displayName,userPrincipalName,appId,servicePrincipalType",
			}),
		})
	}

	federatedByApplicationID, federatedErrs := graphBatchListObjectsByKey(ctx, graphToken, applicationFederatedRequests)
	applicationOwnersByID, applicationOwnerErrs := graphBatchListObjectsByKey(ctx, graphToken, applicationOwnerRequests)

	for _, application := range applications {
		appObjectID := mapStringValue(application, "id")
		if appObjectID == "" {
			continue
		}
		appDisplayName := firstNonEmpty(mapStringValue(application, "displayName"), mapStringValue(application, "appId"), appObjectID)
		backingSP := servicePrincipalByAppID[mapStringValue(application, "appId")]

		if credentialErr, ok := federatedErrs[appObjectID]; ok {
			issues = append(issues, issueFromError("role_trusts.applications["+appObjectID+"].federated_credentials", credentialErr))
		} else {
			for _, credential := range federatedByApplicationID[appObjectID] {
				trusts = append(trusts, federatedCredentialTrust(application, backingSP, credential, appDisplayName))
			}
		}

		if ownerErr, ok := applicationOwnerErrs[appObjectID]; ok {
			issues = append(issues, issueFromError("role_trusts.applications["+appObjectID+"].owners", ownerErr))
		} else {
			for _, owner := range applicationOwnersByID[appObjectID] {
				ownerID := mapStringValue(owner, "id")
				if ownerID == "" {
					continue
				}
				trusts = append(trusts, applicationOwnerTrust(owner, application, backingSP, appDisplayName))
			}
		}
	}

	servicePrincipalOwnerRequests := []graphBatchRequest{}
	servicePrincipalAssignmentRequests := []graphBatchRequest{}
	for _, servicePrincipal := range servicePrincipals {
		spID := mapStringValue(servicePrincipal, "id")
		if spID == "" {
			continue
		}
		servicePrincipalOwnerRequests = append(servicePrincipalOwnerRequests, graphBatchRequest{
			Key: spID,
			URL: graphCollectionURL("/servicePrincipals/"+url.PathEscape(spID)+"/owners", map[string]string{
				"$select": "id,displayName,userPrincipalName,appId,servicePrincipalType",
			}),
		})
		servicePrincipalAssignmentRequests = append(servicePrincipalAssignmentRequests, graphBatchRequest{
			Key: spID,
			URL: graphCollectionURL("/servicePrincipals/"+url.PathEscape(spID)+"/appRoleAssignments", map[string]string{
				"$select": "id,resourceId",
			}),
		})
	}

	servicePrincipalOwnersByID, servicePrincipalOwnerErrs := graphBatchListObjectsByKey(ctx, graphToken, servicePrincipalOwnerRequests)
	assignmentsByServicePrincipalID, assignmentErrs := graphBatchListObjectsByKey(ctx, graphToken, servicePrincipalAssignmentRequests)

	missingResourceIDs := []string{}
	missingResourceSeen := map[string]struct{}{}
	for _, servicePrincipal := range servicePrincipals {
		spID := mapStringValue(servicePrincipal, "id")
		if spID == "" {
			continue
		}
		for _, assignment := range assignmentsByServicePrincipalID[spID] {
			resourceID := mapStringValue(assignment, "resourceId")
			if resourceID == "" || servicePrincipalByID[resourceID] != nil {
				continue
			}
			if _, exists := missingResourceSeen[resourceID]; exists {
				continue
			}
			missingResourceSeen[resourceID] = struct{}{}
			missingResourceIDs = append(missingResourceIDs, resourceID)
		}
	}
	sort.Strings(missingResourceIDs)

	resourceRequests := make([]graphBatchRequest, 0, len(missingResourceIDs))
	for _, resourceID := range missingResourceIDs {
		resourceRequests = append(resourceRequests, graphBatchRequest{
			Key: resourceID,
			URL: graphObjectURL("/servicePrincipals/"+url.PathEscape(resourceID), map[string]string{
				"$select": "id,appId,displayName,servicePrincipalType",
			}),
		})
	}

	resourceByID, resourceErrs := graphBatchGetObjectsByKey(ctx, graphToken, resourceRequests)
	for _, resourceID := range missingResourceIDs {
		resource := resourceByID[resourceID]
		if mapStringValue(resource, "id") == "" {
			continue
		}
		servicePrincipalByID[mapStringValue(resource, "id")] = resource
	}

	for _, servicePrincipal := range servicePrincipals {
		spID := mapStringValue(servicePrincipal, "id")
		if spID == "" {
			continue
		}
		servicePrincipalName := firstNonEmpty(mapStringValue(servicePrincipal, "displayName"), spID)

		if ownerErr, ok := servicePrincipalOwnerErrs[spID]; ok {
			issues = append(issues, issueFromError("role_trusts.service_principals["+spID+"].owners", ownerErr))
		} else {
			for _, owner := range servicePrincipalOwnersByID[spID] {
				ownerID := mapStringValue(owner, "id")
				if ownerID == "" {
					continue
				}
				trusts = append(trusts, servicePrincipalOwnerTrust(owner, servicePrincipal, servicePrincipalName))
			}
		}

		if assignmentErr, ok := assignmentErrs[spID]; ok {
			issues = append(issues, issueFromError("role_trusts.service_principals["+spID+"].app_role_assignments", assignmentErr))
			continue
		}

		for _, assignment := range assignmentsByServicePrincipalID[spID] {
			resourceID := mapStringValue(assignment, "resourceId")
			resource := servicePrincipalByID[resourceID]
			if resource == nil && resourceID != "" {
				if resourceErr, ok := resourceErrs[resourceID]; ok {
					issues = append(issues, issueFromError("role_trusts.service_principals["+spID+"].resource["+resourceID+"]", resourceErr))
					resource = map[string]any{}
				} else if resolved := resourceByID[resourceID]; len(resolved) > 0 {
					resource = resolved
					if resolvedID := mapStringValue(resolved, "id"); resolvedID != "" {
						servicePrincipalByID[resolvedID] = resolved
					}
				}
			}
			trusts = append(trusts, appRoleAssignmentTrust(servicePrincipal, servicePrincipalName, assignment, resource))
		}
	}

	return finalizeRoleTrustFacts(session, mode, trusts, issues), nil
}

func finalizeRoleTrustFacts(session azureSession, mode models.RoleTrustsMode, trusts []models.RoleTrustSummary, issues []models.Issue) RoleTrustsFacts {
	trusts = dedupeRoleTrusts(trusts)
	sort.SliceStable(trusts, func(i int, j int) bool {
		left := trusts[i]
		right := trusts[j]
		switch {
		case left.Confidence != right.Confidence:
			return left.Confidence == "confirmed"
		case left.TrustType != right.TrustType:
			return left.TrustType < right.TrustType
		case stringPtrValue(left.SourceName) != stringPtrValue(right.SourceName):
			return stringPtrValue(left.SourceName) < stringPtrValue(right.SourceName)
		case stringPtrValue(left.TargetName) != stringPtrValue(right.TargetName):
			return stringPtrValue(left.TargetName) < stringPtrValue(right.TargetName)
		default:
			return left.SourceObjectID < right.SourceObjectID
		}
	})

	return RoleTrustsFacts{
		TenantID:       session.tenantID,
		SubscriptionID: session.subscription.ID,
		Mode:           mode,
		Trusts:         trusts,
		Issues:         issues,
	}
}

func federatedCredentialTrust(application map[string]any, backingSP map[string]any, credential map[string]any, appDisplayName string) models.RoleTrustSummary {
	relatedIDs := []string{mapStringValue(application, "id"), mapStringValue(credential, "id")}
	if backingID := mapStringValue(backingSP, "id"); backingID != "" {
		relatedIDs = append(relatedIDs, backingID)
	}

	targetObjectID := mapStringValue(application, "id")
	targetName := stringPtr(appDisplayName)
	targetType := "Application"
	nextReview := "Check permissions for the backing identity behind application '" + appDisplayName + "'."
	escalation := "Application '" + appDisplayName + "' already has a federated trust path."
	var usableIdentityResult *string
	if backingSP != nil {
		targetObjectID = mapStringValue(backingSP, "id")
		targetName = stringPtr(graphDisplayName(backingSP))
		targetType = "ServicePrincipal"
		nextReview = "Check permissions for Azure control on service principal '" + graphDisplayName(backingSP) + "'."
		escalation = "Application '" + appDisplayName + "' already has federated trust that can yield service principal '" + graphDisplayName(backingSP) + "' access."
		usableIdentityResult = stringPtr("Federated sign-in can yield service principal '" + graphDisplayName(backingSP) + "' access.")
	}

	return models.RoleTrustSummary{
		TrustType:            "federated-credential",
		SourceObjectID:       mapStringValue(application, "id"),
		SourceName:           stringPtr(appDisplayName),
		SourceType:           "Application",
		TargetObjectID:       firstNonEmpty(targetObjectID, mapStringValue(application, "id")),
		TargetName:           targetName,
		TargetType:           targetType,
		EvidenceType:         "graph-federated-credential",
		Confidence:           "confirmed",
		ControlPrimitive:     stringPtr("existing-federated-credential"),
		ControlledObjectType: stringPtr("Application"),
		ControlledObjectName: stringPtr(appDisplayName),
		EscalationMechanism:  stringPtr(escalation),
		UsableIdentityResult: usableIdentityResult,
		DefenderCutPoint:     stringPtr("Remove or tighten the federated credential on application '" + appDisplayName + "'."),
		OperatorSignal:       stringPtr("Trust expansion visible; privilege confirmation next."),
		NextReview:           stringPtr(nextReview),
		Summary:              "Application '" + appDisplayName + "' trusts federated subject '" + firstNonEmpty(mapStringValue(credential, "subject"), "unknown subject") + "' from issuer '" + firstNonEmpty(mapStringValue(credential, "issuer"), "unknown issuer") + "'. This row shows trust expansion into the target identity rather than direct Azure privilege by itself. " + nextReview,
		RelatedIDs:           dedupeStrings(relatedIDs),
		FollowOnKind:         models.RoleTrustFollowOnPrivilegeConfirmation,
	}
}

func applicationOwnerTrust(owner map[string]any, application map[string]any, backingSP map[string]any, appDisplayName string) models.RoleTrustSummary {
	nextReview := "Review ownership around application '" + appDisplayName + "'; if it backs an Azure-facing identity, confirm that identity in permissions."
	escalation := "Control of application '" + appDisplayName + "' could change authentication material Azure accepts for identities backed by that application."
	var backingServicePrincipalID *string
	var backingServicePrincipalName *string
	var usableIdentityResult *string
	if backingSP != nil && mapStringValue(backingSP, "id") != "" {
		backingServicePrincipalID = stringPtr(mapStringValue(backingSP, "id"))
		backingServicePrincipalName = stringPtr(graphDisplayName(backingSP))
		escalation = "Control of application '" + appDisplayName + "' could change authentication material that makes service principal '" + graphDisplayName(backingSP) + "' usable."
		usableIdentityResult = stringPtr("Control of application '" + appDisplayName + "' could make service principal '" + graphDisplayName(backingSP) + "' usable.")
	}

	return models.RoleTrustSummary{
		TrustType:                   "app-owner",
		SourceObjectID:              mapStringValue(owner, "id"),
		SourceName:                  stringPtr(graphDisplayName(owner)),
		SourceType:                  graphObjectType(owner),
		TargetObjectID:              mapStringValue(application, "id"),
		TargetName:                  stringPtr(appDisplayName),
		TargetType:                  "Application",
		EvidenceType:                "graph-owner",
		Confidence:                  "confirmed",
		ControlPrimitive:            stringPtr("change-auth-material"),
		ControlledObjectType:        stringPtr("Application"),
		ControlledObjectName:        stringPtr(appDisplayName),
		BackingServicePrincipalID:   backingServicePrincipalID,
		BackingServicePrincipalName: backingServicePrincipalName,
		EscalationMechanism:         stringPtr(escalation),
		UsableIdentityResult:        usableIdentityResult,
		DefenderCutPoint:            stringPtr("Remove the ownership path that lets the source control application '" + appDisplayName + "'."),
		OperatorSignal:              stringPtr("Indirect control visible; ownership review next."),
		NextReview:                  stringPtr(nextReview),
		Summary:                     "Owner '" + graphDisplayName(owner) + "' can modify application '" + appDisplayName + "'. This is an indirect-control row: ownership is the visible trust path, not direct Azure privilege by itself. " + nextReview,
		RelatedIDs:                  dedupeStrings([]string{mapStringValue(owner, "id"), mapStringValue(application, "id")}),
		FollowOnKind:                models.RoleTrustFollowOnOwnershipReview,
	}
}

func servicePrincipalOwnerTrust(owner map[string]any, servicePrincipal map[string]any, servicePrincipalName string) models.RoleTrustSummary {
	nextReview := "Check permissions for Azure control on service principal '" + servicePrincipalName + "'."
	return models.RoleTrustSummary{
		TrustType:            "service-principal-owner",
		SourceObjectID:       mapStringValue(owner, "id"),
		SourceName:           stringPtr(graphDisplayName(owner)),
		SourceType:           graphObjectType(owner),
		TargetObjectID:       mapStringValue(servicePrincipal, "id"),
		TargetName:           stringPtr(servicePrincipalName),
		TargetType:           "ServicePrincipal",
		EvidenceType:         "graph-owner",
		Confidence:           "confirmed",
		ControlPrimitive:     stringPtr("owner-control"),
		ControlledObjectType: stringPtr("ServicePrincipal"),
		ControlledObjectName: stringPtr(servicePrincipalName),
		EscalationMechanism:  stringPtr("Owner-level control over service principal '" + servicePrincipalName + "' could add or replace authentication material Azure accepts for service principal '" + servicePrincipalName + "'."),
		UsableIdentityResult: stringPtr("That could make service principal '" + servicePrincipalName + "' usable."),
		DefenderCutPoint:     stringPtr("Remove the owner-level control path over service principal '" + servicePrincipalName + "'."),
		OperatorSignal:       stringPtr("Trust expansion visible; privilege confirmation next."),
		NextReview:           stringPtr(nextReview),
		Summary:              "Owner '" + graphDisplayName(owner) + "' can modify service principal '" + servicePrincipalName + "'. This row shows a service-principal takeover path rather than direct Azure privilege by itself. " + nextReview,
		RelatedIDs:           dedupeStrings([]string{mapStringValue(owner, "id"), mapStringValue(servicePrincipal, "id")}),
		FollowOnKind:         models.RoleTrustFollowOnPrivilegeConfirmation,
	}
}

func appRoleAssignmentTrust(servicePrincipal map[string]any, servicePrincipalName string, assignment map[string]any, resource map[string]any) models.RoleTrustSummary {
	resourceID := firstNonEmpty(mapStringValue(assignment, "resourceId"), mapStringValue(resource, "id"), "unknown")
	resourceName := firstNonEmpty(mapStringValue(resource, "displayName"), resourceID)
	nextReview := "Check permissions for Azure control on service principal '" + servicePrincipalName + "'."
	return models.RoleTrustSummary{
		TrustType:            "app-to-service-principal",
		SourceObjectID:       mapStringValue(servicePrincipal, "id"),
		SourceName:           stringPtr(servicePrincipalName),
		SourceType:           "ServicePrincipal",
		TargetObjectID:       resourceID,
		TargetName:           stringPtr(resourceName),
		TargetType:           "ServicePrincipal",
		EvidenceType:         "graph-app-role-assignment",
		Confidence:           "confirmed",
		ControlPrimitive:     stringPtr("existing-app-role-assignment"),
		ControlledObjectType: stringPtr("ServicePrincipal"),
		ControlledObjectName: stringPtr(resourceName),
		EscalationMechanism:  stringPtr("Service principal '" + servicePrincipalName + "' already holds an application-permission path into service principal '" + resourceName + "'."),
		UsableIdentityResult: stringPtr("Service principal '" + servicePrincipalName + "' already has application-permission reach to '" + resourceName + "'."),
		DefenderCutPoint:     stringPtr("Remove the app-role assignment path from service principal '" + servicePrincipalName + "' to '" + resourceName + "'."),
		OperatorSignal:       stringPtr("Trust expansion visible; privilege confirmation next."),
		NextReview:           stringPtr(nextReview),
		Summary:              "Service principal '" + servicePrincipalName + "' holds an application permission or app-role assignment to '" + resourceName + "'. This row is a trust-edge and application-permission cue; confirm whether the same identity also holds Azure control. " + nextReview,
		RelatedIDs:           dedupeStrings([]string{mapStringValue(servicePrincipal, "id"), mapStringValue(assignment, "id"), resourceID}),
		FollowOnKind:         models.RoleTrustFollowOnPrivilegeConfirmation,
	}
}

func graphGetApplicationByAppID(ctx context.Context, token string, appID string) (map[string]any, error) {
	items, err := graphListObjects(ctx, token, graphCollectionURL("/applications", map[string]string{
		"$filter": "appId eq '" + strings.ReplaceAll(appID, "'", "''") + "'",
		"$select": "id,appId,displayName",
	}))
	if err != nil || len(items) == 0 {
		return map[string]any{}, err
	}
	return items[0], nil
}

func graphListObjects(ctx context.Context, token string, rawURL string) ([]map[string]any, error) {
	nextURL := rawURL
	items := []map[string]any{}
	for nextURL != "" {
		payload, err := authorizedJSONGetWithToken(ctx, token, nextURL)
		if err != nil {
			return nil, err
		}
		for _, item := range listValue(payload, "value") {
			if mapped, ok := item.(map[string]any); ok {
				items = append(items, mapped)
			}
		}
		nextURL = mapStringValue(payload, "@odata.nextLink")
	}
	return items, nil
}

func graphListObjectsLimit(ctx context.Context, token string, rawURL string, limit int) ([]map[string]any, bool, error) {
	if limit <= 0 {
		items, err := graphListObjects(ctx, token, rawURL)
		return items, false, err
	}
	nextURL := rawURL
	items := []map[string]any{}
	for nextURL != "" {
		payload, err := authorizedJSONGetWithToken(ctx, token, nextURL)
		if err != nil {
			return nil, false, err
		}
		pageItems := listValue(payload, "value")
		for index, item := range pageItems {
			if mapped, ok := item.(map[string]any); ok {
				items = append(items, mapped)
				if len(items) >= limit {
					truncated := index < len(pageItems)-1 || mapStringValue(payload, "@odata.nextLink") != ""
					return items, truncated, nil
				}
			}
		}
		nextURL = mapStringValue(payload, "@odata.nextLink")
	}
	return items, false, nil
}

func graphGetObject(ctx context.Context, token string, rawURL string) (map[string]any, error) {
	return authorizedJSONGetWithToken(ctx, token, rawURL)
}

func graphBatchListObjectsByKey(ctx context.Context, token string, requests []graphBatchRequest) (map[string][]map[string]any, map[string]error) {
	pending := append([]graphBatchRequest{}, requests...)
	partial := map[string][]map[string]any{}
	results := map[string][]map[string]any{}
	errs := map[string]error{}

	for len(pending) > 0 {
		chunkSize := graphBatchMaxRequests
		if len(pending) < chunkSize {
			chunkSize = len(pending)
		}
		chunk := append([]graphBatchRequest{}, pending[:chunkSize]...)
		pending = pending[chunkSize:]

		bodies, bodyErrs, err := graphBatchExecute(ctx, token, chunk)
		if err != nil {
			for _, request := range chunk {
				errs[request.Key] = err
				delete(partial, request.Key)
			}
			continue
		}

		for _, request := range chunk {
			if requestErr, ok := bodyErrs[request.Key]; ok {
				errs[request.Key] = requestErr
				delete(partial, request.Key)
				continue
			}

			body, ok := bodies[request.Key]
			if !ok {
				errs[request.Key] = fmt.Errorf("GET %s: missing batch response", request.URL)
				delete(partial, request.Key)
				continue
			}

			for _, item := range listValue(body, "value") {
				if mapped, ok := item.(map[string]any); ok {
					partial[request.Key] = append(partial[request.Key], mapped)
				}
			}

			nextURL := mapStringValue(body, "@odata.nextLink")
			if nextURL != "" {
				pending = append(pending, graphBatchRequest{Key: request.Key, URL: nextURL})
				continue
			}

			if len(partial[request.Key]) == 0 {
				results[request.Key] = []map[string]any{}
			} else {
				results[request.Key] = append([]map[string]any{}, partial[request.Key]...)
			}
			delete(partial, request.Key)
		}
	}

	return results, errs
}

func graphBatchGetObjectsByKey(ctx context.Context, token string, requests []graphBatchRequest) (map[string]map[string]any, map[string]error) {
	results := map[string]map[string]any{}
	errs := map[string]error{}

	for len(requests) > 0 {
		chunkSize := graphBatchMaxRequests
		if len(requests) < chunkSize {
			chunkSize = len(requests)
		}
		chunk := append([]graphBatchRequest{}, requests[:chunkSize]...)
		requests = requests[chunkSize:]

		bodies, bodyErrs, err := graphBatchExecute(ctx, token, chunk)
		if err != nil {
			for _, request := range chunk {
				errs[request.Key] = err
			}
			continue
		}

		for _, request := range chunk {
			if requestErr, ok := bodyErrs[request.Key]; ok {
				errs[request.Key] = requestErr
				continue
			}

			body, ok := bodies[request.Key]
			if !ok {
				errs[request.Key] = fmt.Errorf("GET %s: missing batch response", request.URL)
				continue
			}

			results[request.Key] = body
		}
	}

	return results, errs
}

func graphBatchExecute(ctx context.Context, token string, requests []graphBatchRequest) (map[string]map[string]any, map[string]error, error) {
	if len(requests) == 0 {
		return map[string]map[string]any{}, map[string]error{}, nil
	}

	type batchRequestEnvelope struct {
		ID      string            `json:"id"`
		Method  string            `json:"method"`
		URL     string            `json:"url"`
		Headers map[string]string `json:"headers,omitempty"`
	}
	type batchResponseEnvelope struct {
		ID     string         `json:"id"`
		Status int            `json:"status"`
		Body   map[string]any `json:"body"`
	}
	type batchRequestPayload struct {
		Requests []batchRequestEnvelope `json:"requests"`
	}
	type batchResponsePayload struct {
		Responses []batchResponseEnvelope `json:"responses"`
	}

	requestPayload := batchRequestPayload{
		Requests: make([]batchRequestEnvelope, 0, len(requests)),
	}
	requestByID := map[string]graphBatchRequest{}
	for index, request := range requests {
		requestID := fmt.Sprintf("%d", index)
		requestPayload.Requests = append(requestPayload.Requests, batchRequestEnvelope{
			ID:     requestID,
			Method: http.MethodGet,
			URL:    graphBatchURL(request.URL),
			Headers: map[string]string{
				"Accept": "application/json",
			},
		})
		requestByID[requestID] = request
	}

	bodyBytes, err := json.Marshal(requestPayload)
	if err != nil {
		return nil, nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, graphEndpoint+"/$batch", bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, nil, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, nil, fmt.Errorf("POST %s: %s", graphEndpoint+"/$batch", resp.Status)
	}

	var responsePayload batchResponsePayload
	if err := json.NewDecoder(resp.Body).Decode(&responsePayload); err != nil {
		return nil, nil, err
	}

	results := map[string]map[string]any{}
	errs := map[string]error{}
	for _, response := range responsePayload.Responses {
		request, ok := requestByID[response.ID]
		if !ok {
			continue
		}
		if response.Status < 200 || response.Status >= 300 {
			errs[request.Key] = graphBatchRequestError(request.URL, response.Status, response.Body)
			continue
		}
		if response.Body == nil {
			results[request.Key] = map[string]any{}
			continue
		}
		results[request.Key] = response.Body
	}

	return results, errs, nil
}

func graphBatchRequestError(rawURL string, status int, body map[string]any) error {
	errorBody := mapValue(body, "error")
	code := mapStringValue(errorBody, "code")
	message := mapStringValue(errorBody, "message")
	switch {
	case code != "" && message != "":
		return fmt.Errorf("GET %s: status %d %s: %s", rawURL, status, code, message)
	case message != "":
		return fmt.Errorf("GET %s: status %d %s", rawURL, status, message)
	default:
		return fmt.Errorf("GET %s: status %d", rawURL, status)
	}
}

func graphBatchURL(rawURL string) string {
	if strings.HasPrefix(rawURL, graphEndpoint) {
		trimmed := strings.TrimPrefix(rawURL, graphEndpoint)
		if trimmed == "" {
			return "/"
		}
		return trimmed
	}
	return rawURL
}

func graphCollectionURL(path string, query map[string]string) string {
	return graphObjectURL(path, query)
}

func graphObjectURL(path string, query map[string]string) string {
	values := url.Values{}
	for key, value := range query {
		if strings.TrimSpace(value) != "" {
			values.Set(key, value)
		}
	}
	if len(values) == 0 {
		return graphEndpoint + path
	}
	return graphEndpoint + path + "?" + values.Encode()
}

func graphDisplayName(item map[string]any) string {
	return firstNonEmpty(
		mapStringValue(item, "displayName"),
		mapStringValue(item, "userPrincipalName"),
		mapStringValue(item, "appId"),
		mapStringValue(item, "id"),
		"unknown",
	)
}

func graphObjectType(item map[string]any) string {
	odataType := mapStringValue(item, "@odata.type")
	if odataType != "" {
		parts := strings.Split(odataType, ".")
		value := parts[len(parts)-1]
		return strings.ToUpper(value[:1]) + value[1:]
	}
	if mapStringValue(item, "servicePrincipalType") != "" || mapStringValue(item, "appId") != "" {
		return "ServicePrincipal"
	}
	if mapStringValue(item, "userPrincipalName") != "" {
		return "User"
	}
	return "DirectoryObject"
}

func dedupeRoleTrusts(items []models.RoleTrustSummary) []models.RoleTrustSummary {
	deduped := []models.RoleTrustSummary{}
	seen := map[string]struct{}{}
	for _, item := range items {
		key := strings.Join([]string{
			item.TrustType,
			item.SourceObjectID,
			item.TargetObjectID,
			item.EvidenceType,
			item.Summary,
		}, "\x00")
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		deduped = append(deduped, item)
	}
	return deduped
}
