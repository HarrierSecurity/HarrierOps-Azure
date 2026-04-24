package render

import "harrierops-azure/internal/models"

func payloadFindings(payload any) []models.Finding {
	switch out := payload.(type) {
	case models.AcrOutput:
		return cloneFindings(out.Findings)
	case models.AksOutput:
		return cloneFindings(out.Findings)
	case models.ApiMgmtOutput:
		return cloneFindings(out.Findings)
	case models.AppServicesOutput:
		return cloneFindings(out.Findings)
	case models.ApplicationGatewayOutput:
		return cloneFindings(out.Findings)
	case models.ArmDeploymentsOutput:
		return armDeploymentFindings(out.Findings)
	case models.AuthPoliciesOutput:
		return authPolicyFindings(out.Findings)
	case models.AutomationOutput:
		return cloneFindings(out.Findings)
	case models.ContainerAppsOutput:
		return cloneFindings(out.Findings)
	case models.ContainerAppsJobsOutput:
		return cloneFindings(out.Findings)
	case models.ContainerInstancesOutput:
		return cloneFindings(out.Findings)
	case models.CrossTenantOutput:
		return cloneFindings(out.Findings)
	case models.DatabasesOutput:
		return cloneFindings(out.Findings)
	case models.DevopsOutput:
		return cloneFindings(out.Findings)
	case models.DnsOutput:
		return cloneFindings(out.Findings)
	case models.EndpointsOutput:
		return cloneFindings(out.Findings)
	case models.EnvVarsOutput:
		return envVarFindings(out.Findings)
	case models.FunctionsOutput:
		return cloneFindings(out.Findings)
	case models.WebJobsOutput:
		return cloneFindings(out.Findings)
	case models.KeyVaultOutput:
		return keyVaultFindings(out.Findings)
	case models.LighthouseOutput:
		return cloneFindings(out.Findings)
	case models.ManagedIdentitiesOutput:
		return managedIdentityFindings(out.Findings)
	case models.NetworkEffectiveOutput:
		return cloneFindings(out.Findings)
	case models.NetworkPortsOutput:
		return cloneFindings(out.Findings)
	case models.NicsOutput:
		return cloneFindings(out.Findings)
	case models.ResourceTrustsOutput:
		return resourceTrustFindings(out.Findings)
	case models.SnapshotsDisksOutput:
		return cloneFindings(out.Findings)
	case models.StorageOutput:
		return storageFindings(out.Findings)
	case models.TokensCredentialsOutput:
		return tokenCredentialFindings(out.Findings)
	case models.VmsOutput:
		return vmFindings(out.Findings)
	case models.VMExtensionsOutput:
		return cloneFindings(out.Findings)
	case models.VmssOutput:
		return cloneFindings(out.Findings)
	case models.WorkloadsOutput:
		return cloneFindings(out.Findings)
	default:
		return nil
	}
}

func payloadIssues(payload any) []models.Issue {
	switch out := payload.(type) {
	case models.AcrOutput:
		return cloneIssues(out.Issues)
	case models.AksOutput:
		return cloneIssues(out.Issues)
	case models.ApiMgmtOutput:
		return cloneIssues(out.Issues)
	case models.AppCredentialsOutput:
		return cloneIssues(out.Issues)
	case models.AppServicesOutput:
		return cloneIssues(out.Issues)
	case models.ApplicationGatewayOutput:
		return cloneIssues(out.Issues)
	case models.ArmDeploymentsOutput:
		return cloneIssues(out.Issues)
	case models.AuthPoliciesOutput:
		return cloneIssues(out.Issues)
	case models.AutomationOutput:
		return cloneIssues(out.Issues)
	case models.ChainsOutput:
		return cloneIssues(out.Issues)
	case models.ChainsOverviewOutput:
		return cloneIssues(out.Issues)
	case models.ContainerAppsOutput:
		return cloneIssues(out.Issues)
	case models.ContainerAppsJobsOutput:
		return cloneIssues(out.Issues)
	case models.ContainerInstancesOutput:
		return cloneIssues(out.Issues)
	case models.CrossTenantOutput:
		return cloneIssues(out.Issues)
	case models.DatabasesOutput:
		return cloneIssues(out.Issues)
	case models.DevopsOutput:
		return cloneIssues(out.Issues)
	case models.DnsOutput:
		return cloneIssues(out.Issues)
	case models.EndpointsOutput:
		return cloneIssues(out.Issues)
	case models.EnvVarsOutput:
		return cloneIssues(out.Issues)
	case models.FunctionsOutput:
		return cloneIssues(out.Issues)
	case models.WebJobsOutput:
		return cloneIssues(out.Issues)
	case models.InventoryOutput:
		return cloneIssues(out.Issues)
	case models.KeyVaultOutput:
		return cloneIssues(out.Issues)
	case models.LighthouseOutput:
		return cloneIssues(out.Issues)
	case models.ManagedIdentitiesOutput:
		return cloneIssues(out.Issues)
	case models.NetworkEffectiveOutput:
		return cloneIssues(out.Issues)
	case models.NetworkPortsOutput:
		return cloneIssues(out.Issues)
	case models.NicsOutput:
		return cloneIssues(out.Issues)
	case models.PermissionsOutput:
		return cloneIssues(out.Issues)
	case models.PersistenceAutomationOutput:
		return cloneIssues(out.Issues)
	case models.PersistenceAppServiceOutput:
		return cloneIssues(out.Issues)
	case models.PersistenceWebJobsOutput:
		return cloneIssues(out.Issues)
	case models.PersistenceContainerAppsJobsOutput:
		return cloneIssues(out.Issues)
	case models.PersistenceVMExtensionsOutput:
		return cloneIssues(out.Issues)
	case models.PersistenceAzureMLOutput:
		return cloneIssues(out.Issues)
	case models.PersistenceFunctionsOutput:
		return cloneIssues(out.Issues)
	case models.PersistenceLogicAppsOutput:
		return cloneIssues(out.Issues)
	case models.PersistenceOverviewOutput:
		return cloneIssues(out.Issues)
	case models.PrincipalsOutput:
		return cloneIssues(out.Issues)
	case models.PrivescOutput:
		return cloneIssues(out.Issues)
	case models.RbacOutput:
		return cloneIssues(out.Issues)
	case models.ResourceTrustsOutput:
		return cloneIssues(out.Issues)
	case models.RoleTrustsOutput:
		return cloneIssues(out.Issues)
	case models.SnapshotsDisksOutput:
		return cloneIssues(out.Issues)
	case models.StorageOutput:
		return cloneIssues(out.Issues)
	case models.TokensCredentialsOutput:
		return cloneIssues(out.Issues)
	case models.VmsOutput:
		return cloneIssues(out.Issues)
	case models.VMExtensionsOutput:
		return cloneIssues(out.Issues)
	case models.VmssOutput:
		return cloneIssues(out.Issues)
	case models.WhoAmIOutput:
		return cloneIssues(out.Issues)
	case models.WorkloadsOutput:
		return cloneIssues(out.Issues)
	default:
		return nil
	}
}

func cloneFindings(findings []models.Finding) []models.Finding {
	return append([]models.Finding{}, findings...)
}

func cloneIssues(issues []models.Issue) []models.Issue {
	return append([]models.Issue{}, issues...)
}

func envVarFindings(findings []models.EnvVarFinding) []models.Finding {
	rows := make([]models.Finding, 0, len(findings))
	for _, finding := range findings {
		rows = append(rows, models.Finding{
			ID:          finding.ID,
			Title:       finding.Title,
			Severity:    finding.Severity,
			Description: finding.Description,
			RelatedIDs:  append([]string{}, finding.RelatedIDs...),
		})
	}
	return rows
}

func tokenCredentialFindings(findings []models.TokenCredentialFinding) []models.Finding {
	rows := make([]models.Finding, 0, len(findings))
	for _, finding := range findings {
		rows = append(rows, models.Finding{
			ID:          finding.ID,
			Title:       finding.Title,
			Severity:    finding.Severity,
			Description: finding.Description,
			RelatedIDs:  append([]string{}, finding.RelatedIDs...),
		})
	}
	return rows
}

func storageFindings(findings []models.StorageFinding) []models.Finding {
	rows := make([]models.Finding, 0, len(findings))
	for _, finding := range findings {
		rows = append(rows, models.Finding{
			ID:          finding.ID,
			Title:       finding.Title,
			Severity:    finding.Severity,
			Description: finding.Description,
			RelatedIDs:  append([]string{}, finding.RelatedIDs...),
		})
	}
	return rows
}

func keyVaultFindings(findings []models.KeyVaultFinding) []models.Finding {
	rows := make([]models.Finding, 0, len(findings))
	for _, finding := range findings {
		rows = append(rows, models.Finding{
			ID:          finding.ID,
			Title:       finding.Title,
			Severity:    finding.Severity,
			Description: finding.Description,
			RelatedIDs:  append([]string{}, finding.RelatedIDs...),
		})
	}
	return rows
}

func vmFindings(findings []models.VmsFinding) []models.Finding {
	rows := make([]models.Finding, 0, len(findings))
	for _, finding := range findings {
		rows = append(rows, models.Finding{
			ID:          finding.ID,
			Title:       finding.Title,
			Severity:    finding.Severity,
			Description: finding.Description,
			RelatedIDs:  append([]string{}, finding.RelatedIDs...),
		})
	}
	return rows
}

func resourceTrustFindings(findings []models.ResourceTrustFinding) []models.Finding {
	rows := make([]models.Finding, 0, len(findings))
	for _, finding := range findings {
		rows = append(rows, models.Finding{
			ID:          finding.ID,
			Title:       finding.Title,
			Severity:    finding.Severity,
			Description: finding.Description,
			RelatedIDs:  append([]string{}, finding.RelatedIDs...),
		})
	}
	return rows
}

func managedIdentityFindings(findings []models.ManagedIdentityFinding) []models.Finding {
	rows := make([]models.Finding, 0, len(findings))
	for _, finding := range findings {
		rows = append(rows, models.Finding{
			ID:          finding.ID,
			Title:       finding.Title,
			Severity:    finding.Severity,
			Description: finding.Description,
			RelatedIDs:  append([]string{}, finding.RelatedIDs...),
		})
	}
	return rows
}

func authPolicyFindings(findings []models.AuthPolicyFinding) []models.Finding {
	rows := make([]models.Finding, 0, len(findings))
	for _, finding := range findings {
		rows = append(rows, models.Finding{
			ID:          finding.ID,
			Title:       finding.Title,
			Severity:    finding.Severity,
			Description: finding.Description,
			RelatedIDs:  append([]string{}, finding.RelatedIDs...),
		})
	}
	return rows
}

func armDeploymentFindings(findings []models.ArmDeploymentFinding) []models.Finding {
	rows := make([]models.Finding, 0, len(findings))
	for _, finding := range findings {
		rows = append(rows, models.Finding{
			ID:          finding.ID,
			Title:       finding.Title,
			Severity:    finding.Severity,
			Description: finding.Description,
			RelatedIDs:  append([]string{}, finding.RelatedIDs...),
		})
	}
	return rows
}
