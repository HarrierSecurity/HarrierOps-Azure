package providers

import (
	"context"
	"fmt"
	"net/netip"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute/v7"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork"

	"harrierops-azure/internal/models"
)

type computeNetworkClients struct {
	interfaces      *armnetwork.InterfacesClient
	publicIPs       *armnetwork.PublicIPAddressesClient
	securityGroups  *armnetwork.SecurityGroupsClient
	subnets         *armnetwork.SubnetsClient
	disks           *armcompute.DisksClient
	snapshots       *armcompute.SnapshotsClient
	vmExtensions    *armcompute.VirtualMachineExtensionsClient
	virtualMachines *armcompute.VirtualMachinesClient
	vmssExtensions  *armcompute.VirtualMachineScaleSetExtensionsClient
	vmScaleSets     *armcompute.VirtualMachineScaleSetsClient
}

type computeNetworkCollector struct {
	clients computeNetworkClients
	session azureSession
}

func (provider AzureProvider) Endpoints(ctx context.Context, tenant string, subscription string) (EndpointsFacts, error) {
	session, err := provider.session(ctx, tenant, subscription)
	if err != nil {
		return EndpointsFacts{}, err
	}

	vmFacts, err := provider.VMs(ctx, tenant, subscription)
	if err != nil {
		return EndpointsFacts{}, err
	}
	appServiceFacts, err := provider.AppServices(ctx, tenant, subscription)
	if err != nil {
		return EndpointsFacts{}, err
	}
	functionFacts, err := provider.Functions(ctx, tenant, subscription)
	if err != nil {
		return EndpointsFacts{}, err
	}
	containerAppFacts, err := provider.ContainerApps(ctx, tenant, subscription)
	if err != nil {
		return EndpointsFacts{}, err
	}
	containerInstanceFacts, err := provider.ContainerInstances(ctx, tenant, subscription)
	if err != nil {
		return EndpointsFacts{}, err
	}

	endpoints := append([]models.EndpointSummary{}, endpointsFromVMAssets(vmFacts.VMAssets)...)
	endpoints = append(endpoints, endpointsFromAppServices(appServiceFacts.AppServices)...)
	endpoints = append(endpoints, endpointsFromFunctionApps(functionFacts.FunctionApps)...)
	endpoints = append(endpoints, endpointsFromContainerApps(containerAppFacts.ContainerApps)...)
	endpoints = append(endpoints, endpointsFromContainerInstances(containerInstanceFacts.ContainerInstances)...)

	issues := append([]models.Issue{}, vmFacts.Issues...)
	issues = append(issues, appServiceFacts.Issues...)
	issues = append(issues, functionFacts.Issues...)
	issues = append(issues, containerAppFacts.Issues...)
	issues = append(issues, containerInstanceFacts.Issues...)

	return EndpointsFacts{
		TenantID:       session.tenantID,
		SubscriptionID: session.subscription.ID,
		Endpoints:      endpoints,
		Issues:         issues,
	}, nil
}

func (provider AzureProvider) NetworkEffective(ctx context.Context, tenant string, subscription string) (NetworkEffectiveFacts, error) {
	session, err := provider.session(ctx, tenant, subscription)
	if err != nil {
		return NetworkEffectiveFacts{}, err
	}

	networkPortFacts, err := provider.NetworkPorts(ctx, tenant, subscription)
	if err != nil {
		return NetworkEffectiveFacts{}, err
	}
	vmFacts, err := provider.VMs(ctx, tenant, subscription)
	if err != nil {
		return NetworkEffectiveFacts{}, err
	}
	containerInstanceFacts, err := provider.ContainerInstances(ctx, tenant, subscription)
	if err != nil {
		return NetworkEffectiveFacts{}, err
	}

	publicIPEndpoints := append([]models.EndpointSummary{}, publicIPEndpointsFromVMAssets(vmFacts.VMAssets)...)
	publicIPEndpoints = append(publicIPEndpoints, publicIPEndpointsFromContainerInstances(containerInstanceFacts.ContainerInstances)...)

	issues := append([]models.Issue{}, networkPortFacts.Issues...)
	issues = append(issues, vmFacts.Issues...)
	issues = append(issues, containerInstanceFacts.Issues...)

	return NetworkEffectiveFacts{
		TenantID:           session.tenantID,
		SubscriptionID:     session.subscription.ID,
		EffectiveExposures: composeNetworkEffective(publicIPEndpoints, networkPortFacts.NetworkPorts),
		Issues:             issues,
	}, nil
}

func (provider AzureProvider) NetworkPorts(ctx context.Context, tenant string, subscription string) (NetworkPortsFacts, error) {
	session, err := provider.session(ctx, tenant, subscription)
	if err != nil {
		return NetworkPortsFacts{}, err
	}

	state, err := provider.computeNetworkState(session)
	if err != nil {
		return NetworkPortsFacts{}, err
	}

	nics := state.nicSnapshot(ctx)
	vms := state.vmSnapshot(ctx)

	rows := []models.NetworkPortSummary{}
	issues := append([]models.Issue{}, nics.issues...)
	issues = append(issues, vms.issues...)

	subnetNSGCache := map[string]*string{}
	nsgRulesCache := map[string][]networkRule{}
	seen := map[string]struct{}{}

	for _, vm := range vms.assets {
		if len(vm.PublicIPs) == 0 {
			continue
		}
		endpoints := publicIPEndpointsFromVMAssets([]models.VmAsset{vm})
		for _, endpoint := range endpoints {
			for _, nicID := range vm.NICIDs {
				nicKey := armIDJoinKey(nicID)
				nic, ok := nics.byID[nicKey]
				if !ok {
					continue
				}

				nicRows := []models.NetworkPortSummary{}
				visibleNSG := false

				if nic.NetworkSecurityGroupID != nil {
					visibleNSG = true
					rules, ruleIssues := state.collector.resolveNSGInboundAllowRules(ctx, *nic.NetworkSecurityGroupID, nsgRulesCache)
					issues = append(issues, ruleIssues...)
					nicRows = append(nicRows, networkPortRowsFromRules(endpoint, nic, rules, "nic", *nic.NetworkSecurityGroupID)...)
				}

				for _, subnetID := range nic.SubnetIDs {
					subnetNSGID, subnetIssues := state.collector.resolveSubnetNSGID(ctx, subnetID, subnetNSGCache)
					issues = append(issues, subnetIssues...)
					if subnetNSGID == nil || strings.TrimSpace(*subnetNSGID) == "" {
						continue
					}

					visibleNSG = true
					rules, ruleIssues := state.collector.resolveNSGInboundAllowRules(ctx, *subnetNSGID, nsgRulesCache)
					issues = append(issues, ruleIssues...)
					nicRows = append(nicRows, networkPortRowsFromRules(endpoint, nic, rules, "subnet", *subnetNSGID)...)
				}

				if len(nicRows) == 0 && !visibleNSG {
					nicRows = append(nicRows, networkPortRowWithoutNSG(endpoint, nic))
				}

				for _, row := range nicRows {
					rowKey := strings.Join([]string{
						row.AssetID,
						row.Endpoint,
						row.Protocol,
						row.Port,
						row.AllowSourceSummary,
					}, "|")
					if _, exists := seen[rowKey]; exists {
						continue
					}
					seen[rowKey] = struct{}{}
					rows = append(rows, row)
				}
			}
		}
	}

	return NetworkPortsFacts{
		TenantID:       session.tenantID,
		SubscriptionID: session.subscription.ID,
		NetworkPorts:   rows,
		Issues:         issues,
	}, nil
}

func (provider AzureProvider) NICs(ctx context.Context, tenant string, subscription string) (NICsFacts, error) {
	session, err := provider.session(ctx, tenant, subscription)
	if err != nil {
		return NICsFacts{}, err
	}

	state, err := provider.computeNetworkState(session)
	if err != nil {
		return NICsFacts{}, err
	}

	nics := state.nicSnapshot(ctx)
	return NICsFacts{
		TenantID:       session.tenantID,
		SubscriptionID: session.subscription.ID,
		NICAssets:      nics.assets,
		Issues:         nics.issues,
	}, nil
}

func (provider AzureProvider) VMs(ctx context.Context, tenant string, subscription string) (VMsFacts, error) {
	session, err := provider.session(ctx, tenant, subscription)
	if err != nil {
		return VMsFacts{}, err
	}

	state, err := provider.computeNetworkState(session)
	if err != nil {
		return VMsFacts{}, err
	}

	nics := state.nicSnapshot(ctx)
	vms := state.vmSnapshot(ctx)

	return VMsFacts{
		TenantID:       session.tenantID,
		SubscriptionID: session.subscription.ID,
		VMAssets:       vms.assets,
		Issues:         append(append([]models.Issue{}, nics.issues...), vms.issues...),
	}, nil
}

func (provider AzureProvider) SnapshotsDisks(ctx context.Context, tenant string, subscription string) (SnapshotsDisksFacts, error) {
	session, err := provider.session(ctx, tenant, subscription)
	if err != nil {
		return SnapshotsDisksFacts{}, err
	}

	state, err := provider.computeNetworkState(session)
	if err != nil {
		return SnapshotsDisksFacts{}, err
	}

	snapshots := state.snapshotDiskSnapshot(ctx)
	return SnapshotsDisksFacts{
		TenantID:           session.tenantID,
		SubscriptionID:     session.subscription.ID,
		SnapshotDiskAssets: snapshots.assets,
		Issues:             snapshots.issues,
	}, nil
}

func (provider AzureProvider) VMSS(ctx context.Context, tenant string, subscription string) (VMSSFacts, error) {
	session, err := provider.session(ctx, tenant, subscription)
	if err != nil {
		return VMSSFacts{}, err
	}

	state, err := provider.computeNetworkState(session)
	if err != nil {
		return VMSSFacts{}, err
	}

	vmss := state.vmssSnapshot(ctx)
	return VMSSFacts{
		TenantID:       session.tenantID,
		SubscriptionID: session.subscription.ID,
		VMSSAssets:     vmss.assets,
		Issues:         vmss.issues,
	}, nil
}

func (provider AzureProvider) Workloads(ctx context.Context, tenant string, subscription string) (WorkloadsFacts, error) {
	session, err := provider.session(ctx, tenant, subscription)
	if err != nil {
		return WorkloadsFacts{}, err
	}

	vmFacts, err := provider.VMs(ctx, tenant, subscription)
	if err != nil {
		return WorkloadsFacts{}, err
	}
	vmssFacts, err := provider.VMSS(ctx, tenant, subscription)
	if err != nil {
		return WorkloadsFacts{}, err
	}
	appServiceFacts, err := provider.AppServices(ctx, tenant, subscription)
	if err != nil {
		return WorkloadsFacts{}, err
	}
	functionFacts, err := provider.Functions(ctx, tenant, subscription)
	if err != nil {
		return WorkloadsFacts{}, err
	}
	containerAppFacts, err := provider.ContainerApps(ctx, tenant, subscription)
	if err != nil {
		return WorkloadsFacts{}, err
	}
	containerInstanceFacts, err := provider.ContainerInstances(ctx, tenant, subscription)
	if err != nil {
		return WorkloadsFacts{}, err
	}

	endpoints := append([]models.EndpointSummary{}, endpointsFromVMAssets(vmFacts.VMAssets)...)
	endpoints = append(endpoints, endpointsFromAppServices(appServiceFacts.AppServices)...)
	endpoints = append(endpoints, endpointsFromFunctionApps(functionFacts.FunctionApps)...)
	endpoints = append(endpoints, endpointsFromContainerApps(containerAppFacts.ContainerApps)...)
	endpoints = append(endpoints, endpointsFromContainerInstances(containerInstanceFacts.ContainerInstances)...)
	endpointsByAsset := endpointsByAssetID(endpoints)

	workloads := append([]models.WorkloadSummary{}, workloadRowsFromVMs(vmFacts.VMAssets, endpointsByAsset)...)
	workloads = append(workloads, workloadRowsFromAppServices(appServiceFacts.AppServices, endpointsByAsset)...)
	workloads = append(workloads, workloadRowsFromFunctions(functionFacts.FunctionApps, endpointsByAsset)...)
	workloads = append(workloads, workloadRowsFromContainerApps(containerAppFacts.ContainerApps, endpointsByAsset)...)
	workloads = append(workloads, workloadRowsFromContainerInstances(containerInstanceFacts.ContainerInstances, endpointsByAsset)...)
	workloads = append(workloads, workloadRowsFromVMSS(vmssFacts.VMSSAssets)...)
	sort.SliceStable(workloads, func(i int, j int) bool {
		return workloadLess(workloads[i], workloads[j])
	})

	issues := append([]models.Issue{}, vmFacts.Issues...)
	issues = append(issues, vmssFacts.Issues...)
	issues = append(issues, appServiceFacts.Issues...)
	issues = append(issues, functionFacts.Issues...)
	issues = append(issues, containerAppFacts.Issues...)
	issues = append(issues, containerInstanceFacts.Issues...)

	return WorkloadsFacts{
		TenantID:       session.tenantID,
		SubscriptionID: session.subscription.ID,
		Workloads:      workloads,
		Issues:         issues,
	}, nil
}

func (provider AzureProvider) TokensCredentials(ctx context.Context, tenant string, subscription string) (TokensCredentialsFacts, error) {
	session, err := provider.session(ctx, tenant, subscription)
	if err != nil {
		return TokensCredentialsFacts{}, err
	}

	appServiceFacts, err := provider.AppServices(ctx, tenant, subscription)
	if err != nil {
		return TokensCredentialsFacts{}, err
	}
	functionFacts, err := provider.Functions(ctx, tenant, subscription)
	if err != nil {
		return TokensCredentialsFacts{}, err
	}
	containerAppFacts, err := provider.ContainerApps(ctx, tenant, subscription)
	if err != nil {
		return TokensCredentialsFacts{}, err
	}
	containerInstanceFacts, err := provider.ContainerInstances(ctx, tenant, subscription)
	if err != nil {
		return TokensCredentialsFacts{}, err
	}
	envVarFacts, err := provider.EnvVars(ctx, tenant, subscription)
	if err != nil {
		return TokensCredentialsFacts{}, err
	}
	armDeploymentFacts, err := provider.ArmDeployments(ctx, tenant, subscription)
	if err != nil {
		return TokensCredentialsFacts{}, err
	}
	vmFacts, err := provider.VMs(ctx, tenant, subscription)
	if err != nil {
		return TokensCredentialsFacts{}, err
	}
	vmssFacts, err := provider.VMSS(ctx, tenant, subscription)
	if err != nil {
		return TokensCredentialsFacts{}, err
	}

	surfaces := append([]models.TokenCredentialSurfaceSummary{}, tokenCredentialSurfacesFromAppServices(appServiceFacts.AppServices)...)
	surfaces = append(surfaces, tokenCredentialSurfacesFromFunctions(functionFacts.FunctionApps)...)
	surfaces = append(surfaces, tokenCredentialSurfacesFromContainerApps(containerAppFacts.ContainerApps)...)
	surfaces = append(surfaces, tokenCredentialSurfacesFromContainerInstances(containerInstanceFacts.ContainerInstances)...)
	surfaces = append(surfaces, tokenCredentialSurfacesFromEnvVars(envVarFacts.EnvVars)...)
	surfaces = append(surfaces, tokenCredentialSurfacesFromArmDeployments(armDeploymentFacts.Deployments)...)
	surfaces = append(surfaces, tokenCredentialSurfacesFromVMs(vmFacts.VMAssets)...)
	surfaces = append(surfaces, tokenCredentialSurfacesFromVMSS(vmssFacts.VMSSAssets)...)

	issues := append([]models.Issue{}, appServiceFacts.Issues...)
	issues = append(issues, functionFacts.Issues...)
	issues = append(issues, containerAppFacts.Issues...)
	issues = append(issues, containerInstanceFacts.Issues...)
	issues = append(issues, envVarFacts.Issues...)
	issues = append(issues, armDeploymentFacts.Issues...)
	issues = append(issues, vmFacts.Issues...)
	issues = append(issues, vmssFacts.Issues...)

	return TokensCredentialsFacts{
		TenantID:       session.tenantID,
		SubscriptionID: session.subscription.ID,
		Surfaces:       surfaces,
		Issues:         issues,
	}, nil
}

func newComputeNetworkCollector(session azureSession) (computeNetworkCollector, error) {
	virtualMachines, err := armcompute.NewVirtualMachinesClient(session.subscription.ID, session.credential, nil)
	if err != nil {
		return computeNetworkCollector{}, fmt.Errorf("build virtual machines client: %w", err)
	}
	disks, err := armcompute.NewDisksClient(session.subscription.ID, session.credential, nil)
	if err != nil {
		return computeNetworkCollector{}, fmt.Errorf("build disks client: %w", err)
	}
	snapshots, err := armcompute.NewSnapshotsClient(session.subscription.ID, session.credential, nil)
	if err != nil {
		return computeNetworkCollector{}, fmt.Errorf("build snapshots client: %w", err)
	}
	vmScaleSets, err := armcompute.NewVirtualMachineScaleSetsClient(session.subscription.ID, session.credential, nil)
	if err != nil {
		return computeNetworkCollector{}, fmt.Errorf("build vm scale sets client: %w", err)
	}
	vmExtensions, err := armcompute.NewVirtualMachineExtensionsClient(session.subscription.ID, session.credential, nil)
	if err != nil {
		return computeNetworkCollector{}, fmt.Errorf("build vm extensions client: %w", err)
	}
	vmssExtensions, err := armcompute.NewVirtualMachineScaleSetExtensionsClient(session.subscription.ID, session.credential, nil)
	if err != nil {
		return computeNetworkCollector{}, fmt.Errorf("build vm scale set extensions client: %w", err)
	}
	interfaces, err := armnetwork.NewInterfacesClient(session.subscription.ID, session.credential, nil)
	if err != nil {
		return computeNetworkCollector{}, fmt.Errorf("build interfaces client: %w", err)
	}
	publicIPs, err := armnetwork.NewPublicIPAddressesClient(session.subscription.ID, session.credential, nil)
	if err != nil {
		return computeNetworkCollector{}, fmt.Errorf("build public ip client: %w", err)
	}
	securityGroups, err := armnetwork.NewSecurityGroupsClient(session.subscription.ID, session.credential, nil)
	if err != nil {
		return computeNetworkCollector{}, fmt.Errorf("build security groups client: %w", err)
	}
	subnets, err := armnetwork.NewSubnetsClient(session.subscription.ID, session.credential, nil)
	if err != nil {
		return computeNetworkCollector{}, fmt.Errorf("build subnets client: %w", err)
	}

	return computeNetworkCollector{
		session: session,
		clients: computeNetworkClients{
			interfaces:      interfaces,
			publicIPs:       publicIPs,
			securityGroups:  securityGroups,
			subnets:         subnets,
			disks:           disks,
			snapshots:       snapshots,
			vmExtensions:    vmExtensions,
			virtualMachines: virtualMachines,
			vmssExtensions:  vmssExtensions,
			vmScaleSets:     vmScaleSets,
		},
	}, nil
}

func (collector computeNetworkCollector) collectNICAssets(ctx context.Context) ([]models.NicAsset, map[string]models.NicAsset, []models.Issue) {
	nicAssets := []models.NicAsset{}
	nicByID := map[string]models.NicAsset{}
	issues := []models.Issue{}

	pager := collector.clients.interfaces.NewListAllPager(nil)
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			issues = append(issues, issueFromError("nics.list_all", err))
			break
		}
		for _, nic := range page.Value {
			asset := nicAssetFromInterface(nic)
			nicAssets = append(nicAssets, asset)
			nicByID[armIDJoinKey(asset.ID)] = asset
		}
	}

	return nicAssets, nicByID, issues
}

func (collector computeNetworkCollector) collectVMAssets(ctx context.Context, nicByID map[string]models.NicAsset) ([]models.VmAsset, []models.Issue) {
	vmAssets := []models.VmAsset{}
	issues := []models.Issue{}
	publicIPCache := map[string]string{}

	pager := collector.clients.virtualMachines.NewListAllPager(nil)
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			issues = append(issues, issueFromError("vms.list_all", err))
			break
		}
		for _, vm := range page.Value {
			if vm == nil {
				continue
			}

			vmID := stringValue(vm.ID)
			vmName := stringValue(vm.Name)
			if vmName == "" {
				vmName = "unknown"
			}

			nicIDs := []string{}
			privateIPs := []string{}
			publicIPs := []string{}
			for _, nicReference := range vmNetworkInterfaceIDs(vm) {
				nicIDs = append(nicIDs, nicReference)
				nic, ok := nicByID[armIDJoinKey(nicReference)]
				if !ok {
					continue
				}
				privateIPs = append(privateIPs, nic.PrivateIPs...)
				resolvedPublicIPs, publicIPIssues := collector.resolvePublicIPAddresses(ctx, nic, publicIPCache)
				issues = append(issues, publicIPIssues...)
				publicIPs = append(publicIPs, resolvedPublicIPs...)
			}

			vmAssets = append(vmAssets, models.VmAsset{
				ID:            firstNonEmpty(vmID, "/unknown/"+vmName),
				IdentityIDs:   vmIdentityIDs(vm),
				Location:      stringValue(vm.Location),
				Name:          vmName,
				NICIDs:        sortedUniqueStrings(nicIDs),
				PowerState:    collector.vmPowerState(ctx, vm, &issues),
				PrivateIPs:    sortedUniqueStrings(privateIPs),
				PublicIPs:     sortedUniqueStrings(publicIPs),
				ResourceGroup: resourceGroupFromID(vmID),
				VMType:        "vm",
			})
		}
	}

	return vmAssets, issues
}

func (collector computeNetworkCollector) collectVMSSAssets(ctx context.Context) ([]models.VmssAsset, []models.Issue) {
	vmssAssets := []models.VmssAsset{}
	issues := []models.Issue{}

	pager := collector.clients.vmScaleSets.NewListAllPager(nil)
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			issues = append(issues, issueFromError("vmss.list_all", err))
			break
		}
		for _, vmss := range page.Value {
			if vmss == nil {
				continue
			}
			asset, assetIssues := vmssAssetFromResource(vmss)
			vmssAssets = append(vmssAssets, asset)
			issues = append(issues, assetIssues...)
		}
	}

	return vmssAssets, issues
}

type snapshotDiskAttachment struct {
	attachedToID   *string
	attachedToName *string
	diskRole       *string
}

func (collector computeNetworkCollector) collectSnapshotDiskAssets(ctx context.Context) ([]models.SnapshotDiskAsset, []models.Issue) {
	attachmentContext, attachmentIssues := collector.snapshotDiskAttachmentContext(ctx)
	assets := []models.SnapshotDiskAsset{}
	issues := append([]models.Issue{}, attachmentIssues...)

	diskPager := collector.clients.disks.NewListPager(nil)
	for diskPager.More() {
		page, err := diskPager.NextPage(ctx)
		if err != nil {
			issues = append(issues, issueFromError("snapshots_disks.disks", err))
			break
		}
		for _, disk := range page.Value {
			if disk == nil {
				continue
			}
			assets = append(assets, snapshotDiskAssetFromDisk(disk, attachmentContext[armIDJoinKey(stringValue(disk.ID))]))
		}
	}

	snapshotPager := collector.clients.snapshots.NewListPager(nil)
	for snapshotPager.More() {
		page, err := snapshotPager.NextPage(ctx)
		if err != nil {
			issues = append(issues, issueFromError("snapshots_disks.snapshots", err))
			break
		}
		for _, snapshot := range page.Value {
			if snapshot == nil {
				continue
			}
			assets = append(assets, snapshotDiskAssetFromSnapshot(snapshot))
		}
	}

	return assets, issues
}

func (collector computeNetworkCollector) snapshotDiskAttachmentContext(ctx context.Context) (map[string]snapshotDiskAttachment, []models.Issue) {
	attachments := map[string]snapshotDiskAttachment{}
	issues := []models.Issue{}

	pager := collector.clients.virtualMachines.NewListAllPager(nil)
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			issues = append(issues, issueFromError("snapshots_disks.vm_attachment_context", err))
			break
		}
		for _, vm := range page.Value {
			if vm == nil || vm.Properties == nil || vm.Properties.StorageProfile == nil {
				continue
			}

			vmID := stringPtr(stringValue(vm.ID))
			vmName := stringPtr(firstNonEmpty(stringValue(vm.Name), resourceNameFromID(stringValue(vm.ID))))
			storageProfile := vm.Properties.StorageProfile

			if storageProfile.OSDisk != nil && storageProfile.OSDisk.ManagedDisk != nil {
				diskID := stringValue(storageProfile.OSDisk.ManagedDisk.ID)
				if diskID != "" {
					attachments[armIDJoinKey(diskID)] = snapshotDiskAttachment{
						attachedToID:   vmID,
						attachedToName: vmName,
						diskRole:       stringPtr("os-disk"),
					}
				}
			}

			for _, dataDisk := range storageProfile.DataDisks {
				if dataDisk == nil || dataDisk.ManagedDisk == nil {
					continue
				}
				diskID := stringValue(dataDisk.ManagedDisk.ID)
				if diskID == "" {
					continue
				}
				attachments[armIDJoinKey(diskID)] = snapshotDiskAttachment{
					attachedToID:   vmID,
					attachedToName: vmName,
					diskRole:       stringPtr("data-disk"),
				}
			}
		}
	}

	return attachments, issues
}

func snapshotDiskAssetFromDisk(disk *armcompute.Disk, attachment snapshotDiskAttachment) models.SnapshotDiskAsset {
	diskID := stringValue(disk.ID)
	diskName := firstNonEmpty(stringValue(disk.Name), resourceNameFromID(diskID), "unknown")

	var (
		sourceResourceID    *string
		sourceResourceName  *string
		sourceResourceKind  *string
		osType              *string
		sizeGB              *int
		timeCreated         *string
		networkAccessPolicy *string
		publicNetworkAccess *string
		diskAccessID        *string
		maxShares           *int
		encryptionType      *string
		diskEncryptionSetID *string
	)

	if disk.Properties != nil {
		if disk.Properties.CreationData != nil {
			sourceResourceID = stringPtr(stringValue(disk.Properties.CreationData.SourceResourceID))
		}
		sourceResourceName = stringPtr(resourceNameFromID(stringPtrValue(sourceResourceID)))
		sourceResourceKind = snapshotDiskSourceKind(sourceResourceID)
		osType = stringPtr(stringValue(disk.Properties.OSType))
		sizeGB = int32PtrToInt(disk.Properties.DiskSizeGB)
		timeCreated = timePtrString(disk.Properties.TimeCreated)
		networkAccessPolicy = stringPtr(stringValue(disk.Properties.NetworkAccessPolicy))
		publicNetworkAccess = stringPtr(stringValue(disk.Properties.PublicNetworkAccess))
		diskAccessID = stringPtr(stringValue(disk.Properties.DiskAccessID))
		maxShares = int32PtrToInt(disk.Properties.MaxShares)
		if disk.Properties.Encryption != nil {
			encryptionType = stringPtr(stringValue(disk.Properties.Encryption.Type))
			diskEncryptionSetID = stringPtr(stringValue(disk.Properties.Encryption.DiskEncryptionSetID))
		}
	}

	attachedToID := attachment.attachedToID
	if attachedToID == nil {
		attachedToID = stringPtr(stringValue(disk.ManagedBy))
	}
	attachedToName := attachment.attachedToName
	if attachedToName == nil {
		attachedToName = stringPtr(resourceNameFromID(stringPtrValue(attachedToID)))
	}

	attachmentState := "detached"
	if attachedToID != nil {
		attachmentState = "attached"
	}

	asset := models.SnapshotDiskAsset{
		ID:                  firstNonEmpty(diskID, "/unknown/"+diskName),
		Name:                diskName,
		AssetKind:           "disk",
		ResourceGroup:       resourceGroupFromID(diskID),
		Location:            stringPtr(stringValue(disk.Location)),
		DiskRole:            attachment.diskRole,
		AttachmentState:     attachmentState,
		AttachedToID:        attachedToID,
		AttachedToName:      attachedToName,
		SourceResourceID:    sourceResourceID,
		SourceResourceName:  sourceResourceName,
		SourceResourceKind:  sourceResourceKind,
		OSType:              osType,
		SizeGB:              sizeGB,
		TimeCreated:         timeCreated,
		Incremental:         nil,
		NetworkAccessPolicy: networkAccessPolicy,
		PublicNetworkAccess: publicNetworkAccess,
		DiskAccessID:        diskAccessID,
		MaxShares:           maxShares,
		EncryptionType:      encryptionType,
		DiskEncryptionSetID: diskEncryptionSetID,
	}
	asset.Summary = snapshotDiskSummary(asset)
	asset.RelatedIDs = compactStrings(
		stringPtrValue(asset.AttachedToID),
		stringPtrValue(asset.SourceResourceID),
		stringPtrValue(asset.DiskAccessID),
		stringPtrValue(asset.DiskEncryptionSetID),
	)
	return asset
}

func snapshotDiskAssetFromSnapshot(snapshot *armcompute.Snapshot) models.SnapshotDiskAsset {
	snapshotID := stringValue(snapshot.ID)
	snapshotName := firstNonEmpty(stringValue(snapshot.Name), resourceNameFromID(snapshotID), "unknown")

	var (
		sourceResourceID    *string
		sourceResourceName  *string
		sourceResourceKind  *string
		osType              *string
		sizeGB              *int
		timeCreated         *string
		incremental         *bool
		networkAccessPolicy *string
		publicNetworkAccess *string
		diskAccessID        *string
		encryptionType      *string
		diskEncryptionSetID *string
	)

	if snapshot.Properties != nil {
		if snapshot.Properties.CreationData != nil {
			sourceResourceID = stringPtr(stringValue(snapshot.Properties.CreationData.SourceResourceID))
		}
		sourceResourceName = stringPtr(resourceNameFromID(stringPtrValue(sourceResourceID)))
		sourceResourceKind = snapshotDiskSourceKind(sourceResourceID)
		osType = stringPtr(stringValue(snapshot.Properties.OSType))
		sizeGB = int32PtrToInt(snapshot.Properties.DiskSizeGB)
		timeCreated = timePtrString(snapshot.Properties.TimeCreated)
		incremental = snapshot.Properties.Incremental
		networkAccessPolicy = stringPtr(stringValue(snapshot.Properties.NetworkAccessPolicy))
		publicNetworkAccess = stringPtr(stringValue(snapshot.Properties.PublicNetworkAccess))
		diskAccessID = stringPtr(stringValue(snapshot.Properties.DiskAccessID))
		if snapshot.Properties.Encryption != nil {
			encryptionType = stringPtr(stringValue(snapshot.Properties.Encryption.Type))
			diskEncryptionSetID = stringPtr(stringValue(snapshot.Properties.Encryption.DiskEncryptionSetID))
		}
	}

	asset := models.SnapshotDiskAsset{
		ID:                  firstNonEmpty(snapshotID, "/unknown/"+snapshotName),
		Name:                snapshotName,
		AssetKind:           "snapshot",
		ResourceGroup:       resourceGroupFromID(snapshotID),
		Location:            stringPtr(stringValue(snapshot.Location)),
		DiskRole:            nil,
		AttachmentState:     "snapshot",
		AttachedToID:        nil,
		AttachedToName:      nil,
		SourceResourceID:    sourceResourceID,
		SourceResourceName:  sourceResourceName,
		SourceResourceKind:  sourceResourceKind,
		OSType:              osType,
		SizeGB:              sizeGB,
		TimeCreated:         timeCreated,
		Incremental:         incremental,
		NetworkAccessPolicy: networkAccessPolicy,
		PublicNetworkAccess: publicNetworkAccess,
		DiskAccessID:        diskAccessID,
		MaxShares:           nil,
		EncryptionType:      encryptionType,
		DiskEncryptionSetID: diskEncryptionSetID,
	}
	asset.Summary = snapshotDiskSummary(asset)
	asset.RelatedIDs = compactStrings(
		stringPtrValue(asset.SourceResourceID),
		stringPtrValue(asset.DiskAccessID),
		stringPtrValue(asset.DiskEncryptionSetID),
	)
	return asset
}

func snapshotDiskSourceKind(resourceID *string) *string {
	normalized := strings.ToLower(stringPtrValue(resourceID))
	switch {
	case normalized == "":
		return nil
	case strings.Contains(normalized, "/providers/microsoft.compute/disks/"):
		return stringPtr("disk")
	case strings.Contains(normalized, "/providers/microsoft.compute/snapshots/"):
		return stringPtr("snapshot")
	default:
		return stringPtr("resource")
	}
}

func snapshotDiskSummary(asset models.SnapshotDiskAsset) string {
	parts := []string{}

	switch {
	case asset.AssetKind == "snapshot":
		parts = append(parts, "Snapshot of "+firstNonEmpty(stringPtrValue(asset.SourceResourceName), "source resource"))
		if asset.Incremental != nil && *asset.Incremental {
			parts = append(parts, "incremental copy path visible")
		}
	case asset.AttachedToName != nil:
		role := firstNonEmpty(stringPtrValue(asset.DiskRole), "managed disk")
		parts = append(parts, "Attached "+role+" for "+stringPtrValue(asset.AttachedToName))
	default:
		parts = append(parts, "Detached managed disk")
	}

	postureBits := []string{}
	if asset.PublicNetworkAccess != nil {
		postureBits = append(postureBits, "public network "+stringPtrValue(asset.PublicNetworkAccess))
	}
	if asset.NetworkAccessPolicy != nil {
		postureBits = append(postureBits, "network access "+stringPtrValue(asset.NetworkAccessPolicy))
	}
	if asset.MaxShares != nil && *asset.MaxShares != 1 {
		postureBits = append(postureBits, "max shares "+strconv.Itoa(*asset.MaxShares))
	}
	if asset.DiskAccessID != nil {
		postureBits = append(postureBits, "disk access resource visible")
	}
	if len(postureBits) > 0 {
		parts = append(parts, strings.Join(postureBits, ", "))
	}

	encryptionBits := []string{}
	if asset.EncryptionType != nil {
		encryptionBits = append(encryptionBits, stringPtrValue(asset.EncryptionType))
	}
	if asset.DiskEncryptionSetID != nil {
		encryptionBits = append(encryptionBits, "disk encryption set linked")
	}
	if len(encryptionBits) > 0 {
		parts = append(parts, "encryption posture: "+strings.Join(encryptionBits, ", "))
	}

	return strings.Join(parts, "; ") + "."
}

func compactStrings(values ...string) []string {
	items := []string{}
	for _, value := range values {
		if strings.TrimSpace(value) == "" {
			continue
		}
		items = append(items, value)
	}
	return items
}

func int32PtrToInt(value *int32) *int {
	if value == nil {
		return nil
	}
	return intPtr(int(*value))
}

func timePtrString(value *time.Time) *string {
	if value == nil {
		return nil
	}
	formatted := value.UTC().Format("2006-01-02T15:04:05-07:00")
	return &formatted
}

func (collector computeNetworkCollector) resolvePublicIPAddresses(ctx context.Context, nic models.NicAsset, cache map[string]string) ([]string, []models.Issue) {
	publicIPs := []string{}
	issues := []models.Issue{}

	for _, publicIPID := range nic.PublicIPIDs {
		if cached, exists := cache[publicIPID]; exists {
			if cached != "" {
				publicIPs = append(publicIPs, cached)
			}
			continue
		}

		resourceGroup, publicIPName := resourceGroupAndNameFromID(publicIPID)
		if resourceGroup == "" || publicIPName == "" {
			cache[publicIPID] = ""
			continue
		}

		response, err := collector.clients.publicIPs.Get(ctx, resourceGroup, publicIPName, nil)
		if err != nil {
			issues = append(issues, issueFromError("public_ip_addresses["+publicIPID+"]", err))
			continue
		}

		address := ""
		if response.PublicIPAddress.Properties != nil {
			address = stringValue(response.PublicIPAddress.Properties.IPAddress)
		}
		cache[publicIPID] = address
		if address != "" {
			publicIPs = append(publicIPs, address)
		}
	}

	return sortedUniqueStrings(publicIPs), issues
}

func (collector computeNetworkCollector) resolveSubnetNSGID(ctx context.Context, subnetID string, cache map[string]*string) (*string, []models.Issue) {
	if value, exists := cache[subnetID]; exists {
		return value, nil
	}

	resourceGroup, vnetName, subnetName := subnetComponentsFromID(subnetID)
	if resourceGroup == "" || vnetName == "" || subnetName == "" {
		cache[subnetID] = nil
		return nil, nil
	}

	response, err := collector.clients.subnets.Get(ctx, resourceGroup, vnetName, subnetName, nil)
	if err != nil {
		return nil, []models.Issue{issueFromError("subnets["+subnetID+"]", err)}
	}

	nsgID := ""
	if response.Subnet.Properties != nil && response.Subnet.Properties.NetworkSecurityGroup != nil {
		nsgID = stringValue(response.Subnet.Properties.NetworkSecurityGroup.ID)
	}
	cache[subnetID] = stringPtr(nsgID)
	return cache[subnetID], nil
}

func (collector computeNetworkCollector) resolveNSGInboundAllowRules(ctx context.Context, nsgID string, cache map[string][]networkRule) ([]networkRule, []models.Issue) {
	if rules, exists := cache[nsgID]; exists {
		return rules, nil
	}

	resourceGroup, nsgName := resourceGroupAndNameFromID(nsgID)
	if resourceGroup == "" || nsgName == "" {
		cache[nsgID] = []networkRule{}
		return cache[nsgID], nil
	}

	response, err := collector.clients.securityGroups.Get(ctx, resourceGroup, nsgName, nil)
	if err != nil {
		return nil, []models.Issue{issueFromError("network_security_groups["+nsgID+"]", err)}
	}

	rules := inboundAllowRulesFromNSG(&response.SecurityGroup)
	cache[nsgID] = rules
	return rules, nil
}

func (collector computeNetworkCollector) vmPowerState(ctx context.Context, vm *armcompute.VirtualMachine, issues *[]models.Issue) string {
	if vm == nil {
		return ""
	}
	if state := extractPowerState(vm.Properties); state != "" {
		return state
	}

	resourceGroup := resourceGroupFromID(stringValue(vm.ID))
	vmName := stringValue(vm.Name)
	if resourceGroup == "" || vmName == "" {
		return ""
	}

	instanceView, err := collector.clients.virtualMachines.InstanceView(ctx, resourceGroup, vmName, nil)
	if err != nil {
		*issues = append(*issues, issueFromError("vms["+resourceGroup+"/"+vmName+"].instance_view", err))
		return ""
	}

	for _, status := range instanceView.VirtualMachineInstanceView.Statuses {
		code := stringValue(status.Code)
		if strings.HasPrefix(code, "PowerState/") {
			return strings.TrimPrefix(code, "PowerState/")
		}
	}
	return ""
}

func vmNetworkInterfaceIDs(vm *armcompute.VirtualMachine) []string {
	if vm == nil || vm.Properties == nil || vm.Properties.NetworkProfile == nil {
		return nil
	}
	ids := []string{}
	for _, nicReference := range vm.Properties.NetworkProfile.NetworkInterfaces {
		if nicReference == nil {
			continue
		}
		id := stringValue(nicReference.ID)
		if id != "" {
			ids = append(ids, id)
		}
	}
	return sortedUniqueStrings(ids)
}

func vmIdentityIDs(vm *armcompute.VirtualMachine) []string {
	if vm == nil || vm.Identity == nil {
		return nil
	}
	identityIDs := []string{}
	vmID := stringValue(vm.ID)
	if stringValue(vm.Identity.PrincipalID) != "" {
		identityIDs = append(identityIDs, vmID+"/identities/system")
	}
	for identityID := range vm.Identity.UserAssignedIdentities {
		identityIDs = append(identityIDs, identityID)
	}
	return sortedUniqueStrings(identityIDs)
}

func nicAssetFromInterface(nic *armnetwork.Interface) models.NicAsset {
	if nic == nil {
		return models.NicAsset{}
	}

	nicID := stringValue(nic.ID)
	attachedAssetID := (*string)(nil)
	attachedAssetName := (*string)(nil)
	networkSecurityGroupID := (*string)(nil)
	privateIPs := []string{}
	publicIPIDs := []string{}
	subnetIDs := []string{}
	vnetIDs := []string{}

	if nic.Properties != nil {
		if nic.Properties.VirtualMachine != nil {
			attachedID := stringValue(nic.Properties.VirtualMachine.ID)
			attachedAssetID = stringPtr(attachedID)
			attachedAssetName = stringPtr(resourceNameFromID(attachedID))
		}
		if nic.Properties.NetworkSecurityGroup != nil {
			networkSecurityGroupID = stringPtr(stringValue(nic.Properties.NetworkSecurityGroup.ID))
		}
		for _, ipConfiguration := range nic.Properties.IPConfigurations {
			if ipConfiguration == nil || ipConfiguration.Properties == nil {
				continue
			}
			properties := ipConfiguration.Properties
			privateIPs = append(privateIPs, stringValue(properties.PrivateIPAddress))
			if properties.PublicIPAddress != nil {
				publicIPIDs = append(publicIPIDs, stringValue(properties.PublicIPAddress.ID))
			}
			if properties.Subnet != nil {
				subnetID := stringValue(properties.Subnet.ID)
				subnetIDs = append(subnetIDs, subnetID)
				vnetIDs = append(vnetIDs, vnetIDFromSubnetID(subnetID))
			}
		}
	}

	return models.NicAsset{
		AttachedAssetID:        attachedAssetID,
		AttachedAssetName:      attachedAssetName,
		ID:                     firstNonEmpty(nicID, "/unknown/"+stringValue(nic.Name)),
		Name:                   firstNonEmpty(stringValue(nic.Name), resourceNameFromID(nicID), "unknown"),
		NetworkSecurityGroupID: networkSecurityGroupID,
		PrivateIPs:             sortedUniqueStrings(privateIPs),
		PublicIPIDs:            sortedUniqueStrings(publicIPIDs),
		SubnetIDs:              sortedUniqueStrings(subnetIDs),
		VnetIDs:                sortedUniqueStrings(vnetIDs),
	}
}

type networkRule struct {
	Name     string
	Ports    []string
	Protocol string
	Sources  []string
}

func inboundAllowRulesFromNSG(nsg *armnetwork.SecurityGroup) []networkRule {
	if nsg == nil || nsg.Properties == nil {
		return nil
	}

	rules := []networkRule{}
	for _, rule := range nsg.Properties.SecurityRules {
		if rule == nil || rule.Properties == nil {
			continue
		}
		access := strings.ToLower(stringValue(rule.Properties.Access))
		direction := strings.ToLower(stringValue(rule.Properties.Direction))
		if access != "allow" || direction != "inbound" {
			continue
		}
		rules = append(rules, networkRule{
			Name:     firstNonEmpty(stringValue(rule.Name), "allow-rule"),
			Ports:    normalizedDestinationPorts(rule.Properties),
			Protocol: normalizedNetworkProtocol(rule.Properties.Protocol),
			Sources:  normalizedRuleSources(rule.Properties),
		})
	}
	return rules
}

func normalizedNetworkProtocol(value any) string {
	text := strings.TrimSpace(stringValue(value))
	if text == "" || text == "*" {
		return "Any"
	}
	return strings.ToUpper(text)
}

func normalizedDestinationPorts(rule *armnetwork.SecurityRulePropertiesFormat) []string {
	if rule == nil {
		return []string{"any"}
	}
	ports := []string{}
	for _, port := range rule.DestinationPortRanges {
		value := strings.TrimSpace(stringValue(port))
		if value != "" {
			ports = append(ports, value)
		}
	}
	if value := strings.TrimSpace(stringValue(rule.DestinationPortRange)); value != "" {
		ports = append(ports, value)
	}
	if len(ports) == 0 {
		return []string{"any"}
	}
	for index, port := range ports {
		if port == "*" {
			ports[index] = "any"
		}
	}
	return sortedUniqueStrings(ports)
}

func normalizedRuleSources(rule *armnetwork.SecurityRulePropertiesFormat) []string {
	if rule == nil {
		return []string{"Any"}
	}
	sources := []string{}
	for _, source := range rule.SourceAddressPrefixes {
		value := strings.TrimSpace(stringValue(source))
		if value != "" {
			sources = append(sources, value)
		}
	}
	if value := strings.TrimSpace(stringValue(rule.SourceAddressPrefix)); value != "" {
		sources = append(sources, value)
	}
	if len(sources) == 0 {
		return []string{"Any"}
	}
	for index, source := range sources {
		if source == "*" {
			sources[index] = "Any"
		}
	}
	return sortedUniqueStrings(sources)
}

func networkPortRowsFromRules(endpoint models.EndpointSummary, nic models.NicAsset, rules []networkRule, scopeType string, scopeID string) []models.NetworkPortSummary {
	rows := []models.NetworkPortSummary{}
	for _, rule := range rules {
		sourceSummary := networkRuleSourceSummary(rule.Sources)
		confidence := networkPortConfidence(rule.Sources)
		scopeLabel := networkScopeLabel(scopeType, scopeID, rule.Name)
		assetName := firstNonEmpty(endpoint.SourceAssetName, stringPtrValue(nic.AttachedAssetName), nic.Name, "unknown")
		for _, port := range rule.Ports {
			rows = append(rows, models.NetworkPortSummary{
				AllowSourceSummary: sourceSummary + " via " + scopeLabel,
				AssetID:            firstNonEmpty(endpoint.SourceAssetID, stringPtrValue(nic.AttachedAssetID), nic.ID),
				AssetName:          assetName,
				Endpoint:           firstNonEmpty(endpoint.Endpoint, "unknown"),
				ExposureConfidence: confidence,
				Port:               port,
				Protocol:           rule.Protocol,
				RelatedIDs: dedupeStrings(append([]string{
					endpoint.SourceAssetID,
					nic.ID,
					scopeID,
				}, endpoint.RelatedIDs...)),
				Summary: "Asset '" + assetName + "' has inbound " + rule.Protocol + " " + port +
					" allow evidence for endpoint " + firstNonEmpty(endpoint.Endpoint, "unknown") +
					" from " + sourceSummary + " via " + scopeLabel + ".",
			})
		}
	}
	return rows
}

func networkPortRowWithoutNSG(endpoint models.EndpointSummary, nic models.NicAsset) models.NetworkPortSummary {
	assetName := firstNonEmpty(endpoint.SourceAssetName, stringPtrValue(nic.AttachedAssetName), nic.Name, "unknown")
	return models.NetworkPortSummary{
		AllowSourceSummary: "no Azure NSG visible on NIC or subnet",
		AssetID:            firstNonEmpty(endpoint.SourceAssetID, stringPtrValue(nic.AttachedAssetID), nic.ID),
		AssetName:          assetName,
		Endpoint:           firstNonEmpty(endpoint.Endpoint, "unknown"),
		ExposureConfidence: "low",
		Port:               "any",
		Protocol:           "any",
		RelatedIDs:         dedupeStrings(append([]string{endpoint.SourceAssetID, nic.ID}, endpoint.RelatedIDs...)),
		Summary: "Asset '" + assetName + "' exposes endpoint " + firstNonEmpty(endpoint.Endpoint, "unknown") +
			" with no NIC or subnet NSG visible from the current Azure read path. Azure network port restrictions are not evident here, but guest or service controls may still apply.",
	}
}

func networkRuleSourceSummary(sources []string) string {
	if len(sources) == 0 {
		return "unknown sources"
	}
	return strings.Join(sortedUniqueStrings(sources), ", ")
}

func networkPortConfidence(sources []string) string {
	values := []string{}
	lowered := []string{}
	for _, source := range sources {
		value := strings.TrimSpace(source)
		if value == "" {
			continue
		}
		values = append(values, value)
		lowered = append(lowered, strings.ToLower(value))
	}

	for _, value := range lowered {
		if value == "any" || value == "internet" || value == "0.0.0.0/0" || value == "::/0" {
			return "high"
		}
	}
	for _, value := range lowered {
		if value == "azureloadbalancer" {
			return "medium"
		}
	}
	for _, value := range values {
		if strings.Contains(value, "/") && !isPrivateNetworkPrefix(value) {
			return "medium"
		}
	}
	for _, value := range lowered {
		if value == "virtualnetwork" {
			return "low"
		}
	}
	for _, value := range values {
		if isPrivateNetworkPrefix(value) {
			return "low"
		}
	}
	return "medium"
}

func networkScopeLabel(scopeType string, scopeID string, ruleName string) string {
	scopeName := firstNonEmpty(resourceNameFromID(scopeID), scopeID, "unknown")
	resourceGroup := resourceGroupFromID(scopeID)
	label := "subnet-nsg"
	if scopeType == "nic" {
		label = "nic-nsg"
	}
	scopeRef := scopeName
	if resourceGroup != "" {
		scopeRef = resourceGroup + "/" + scopeName
	}
	return label + ":" + scopeRef + "/" + firstNonEmpty(ruleName, "allow-rule")
}

func composeNetworkEffective(endpoints []models.EndpointSummary, networkPorts []models.NetworkPortSummary) []models.NetworkEffectiveSummary {
	rowsByKey := map[string][]models.NetworkPortSummary{}
	for _, row := range networkPorts {
		key := row.AssetID + "|" + row.Endpoint
		rowsByKey[key] = append(rowsByKey[key], row)
	}

	effective := []models.NetworkEffectiveSummary{}
	for _, endpoint := range endpoints {
		if endpoint.EndpointType != "ip" || endpoint.ExposureFamily != "public-ip" {
			continue
		}
		rows := rowsByKey[endpoint.SourceAssetID+"|"+endpoint.Endpoint]
		effective = append(effective, networkEffectiveRowFromEndpoint(endpoint, rows))
	}

	return effective
}

func networkEffectiveRowFromEndpoint(endpoint models.EndpointSummary, networkPorts []models.NetworkPortSummary) models.NetworkEffectiveSummary {
	highest := "low"
	if len(networkPorts) > 0 {
		sort.SliceStable(networkPorts, func(i int, j int) bool {
			return exposureRank(networkPorts[i].ExposureConfidence) < exposureRank(networkPorts[j].ExposureConfidence)
		})
		highest = strings.ToLower(networkPorts[0].ExposureConfidence)
	}

	explicitAllowRows := []models.NetworkPortSummary{}
	for _, row := range networkPorts {
		if !networkPortIsNoNSGObservation(row) {
			explicitAllowRows = append(explicitAllowRows, row)
		}
	}

	internetExposedPorts := []string{}
	constrainedPorts := []string{}
	observedPaths := []string{}
	relatedIDs := append([]string{endpoint.SourceAssetID}, endpoint.RelatedIDs...)

	for _, row := range networkPorts {
		observedPaths = append(observedPaths, row.AllowSourceSummary)
		relatedIDs = append(relatedIDs, row.RelatedIDs...)
	}
	for _, row := range explicitAllowRows {
		label := strings.ToUpper(row.Protocol) + "/" + row.Port
		if networkPortHasBroadInternetSource(row) {
			internetExposedPorts = append(internetExposedPorts, label)
			continue
		}
		constrainedPorts = append(constrainedPorts, label)
	}

	internetExposedPorts = sortedUniqueStrings(internetExposedPorts)
	constrainedPorts = sortedUniqueStrings(constrainedPorts)
	observedPaths = sortedUniqueStrings(observedPaths)
	relatedIDs = dedupeStrings(relatedIDs)

	summary := ""
	switch {
	case len(explicitAllowRows) > 0:
		internetPhrase := "no broad internet allow evidence surfaced"
		if len(internetExposedPorts) > 0 {
			internetPhrase = "internet-facing allow evidence on " + strings.Join(internetExposedPorts, ", ")
		}
		constrainedPhrase := ""
		if len(constrainedPorts) > 0 {
			constrainedPhrase = " and narrower allow evidence on " + strings.Join(constrainedPorts, ", ")
		}
		summary = "Asset '" + endpoint.SourceAssetName + "' endpoint " + endpoint.Endpoint + " has " +
			internetPhrase + constrainedPhrase +
			". Treat this as visible Azure network triage signal, not proof of full effective reachability."
	case len(networkPorts) > 0:
		summary = "Asset '" + endpoint.SourceAssetName + "' endpoint " + endpoint.Endpoint +
			" is visible as a public IP path, but no Azure NSG was visible on the NIC or subnet from the current read path. Treat this as a low-confidence triage clue rather than proof of exposure."
	default:
		summary = "Asset '" + endpoint.SourceAssetName + "' endpoint " + endpoint.Endpoint +
			" is visible as a public IP path, but no inbound-rule evidence was surfaced from the current read path. Treat this as a low-confidence triage clue rather than proof of exposure."
	}

	return models.NetworkEffectiveSummary{
		AssetID:              firstNonEmpty(endpoint.SourceAssetID, "/unknown/"+endpoint.SourceAssetName),
		AssetName:            firstNonEmpty(endpoint.SourceAssetName, "unknown"),
		ConstrainedPorts:     constrainedPorts,
		EffectiveExposure:    highest,
		Endpoint:             firstNonEmpty(endpoint.Endpoint, "unknown"),
		EndpointType:         firstNonEmpty(endpoint.EndpointType, "ip"),
		InternetExposedPorts: internetExposedPorts,
		ObservedPaths:        observedPaths,
		RelatedIDs:           relatedIDs,
		Summary:              summary,
	}
}

func networkPortIsNoNSGObservation(item models.NetworkPortSummary) bool {
	return item.AllowSourceSummary == "no Azure NSG visible on NIC or subnet"
}

func networkPortHasBroadInternetSource(item models.NetworkPortSummary) bool {
	sourceFragment := strings.Split(item.AllowSourceSummary, " via ")[0]
	for _, token := range strings.Split(sourceFragment, ",") {
		value := strings.ToLower(strings.TrimSpace(token))
		if value == "any" || value == "internet" || value == "0.0.0.0/0" || value == "::/0" {
			return true
		}
	}
	return false
}

func endpointsFromVMAssets(vmAssets []models.VmAsset) []models.EndpointSummary {
	endpoints := []models.EndpointSummary{}
	for _, vm := range vmAssets {
		for _, publicIP := range vm.PublicIPs {
			endpoints = append(endpoints, models.EndpointSummary{
				Endpoint:        publicIP,
				EndpointType:    "ip",
				ExposureFamily:  "public-ip",
				IngressPath:     "direct-vm-ip",
				RelatedIDs:      dedupeStrings(append([]string{vm.ID}, append(vm.NICIDs, vm.IdentityIDs...)...)),
				SourceAssetID:   vm.ID,
				SourceAssetKind: "VM",
				SourceAssetName: vm.Name,
				Summary:         "VM '" + vm.Name + "' exposes public IP " + publicIP + ". Review direct ingress path alongside NIC and NSG context.",
			})
		}
	}
	return endpoints
}

func publicIPEndpointsFromVMAssets(vmAssets []models.VmAsset) []models.EndpointSummary {
	return endpointsFromVMAssets(vmAssets)
}

func endpointsFromAppServices(appServices []models.AppServiceAsset) []models.EndpointSummary {
	endpoints := []models.EndpointSummary{}
	for _, app := range appServices {
		if app.DefaultHostname == nil || *app.DefaultHostname == "" {
			continue
		}
		endpoints = append(endpoints, models.EndpointSummary{
			Endpoint:        *app.DefaultHostname,
			EndpointType:    "hostname",
			ExposureFamily:  "managed-web-hostname",
			IngressPath:     "azurewebsites-default-hostname",
			RelatedIDs:      dedupeStrings([]string{app.ID, stringPtrValue(app.WorkloadPrincipalID)}),
			SourceAssetID:   app.ID,
			SourceAssetKind: "AppService",
			SourceAssetName: app.Name,
			Summary:         "AppService '" + app.Name + "' publishes Azure-managed hostname '" + *app.DefaultHostname + "'. Validate whether that ingress path is intended and how it is constrained.",
		})
	}
	return endpoints
}

func endpointsFromFunctionApps(functionApps []models.FunctionAppAsset) []models.EndpointSummary {
	endpoints := []models.EndpointSummary{}
	for _, app := range functionApps {
		if app.DefaultHostname == nil || *app.DefaultHostname == "" {
			continue
		}
		endpoints = append(endpoints, models.EndpointSummary{
			Endpoint:        *app.DefaultHostname,
			EndpointType:    "hostname",
			ExposureFamily:  "managed-web-hostname",
			IngressPath:     "azure-functions-default-hostname",
			RelatedIDs:      dedupeStrings(append([]string{app.ID, stringPtrValue(app.WorkloadPrincipalID)}, app.WorkloadIdentityIDs...)),
			SourceAssetID:   app.ID,
			SourceAssetKind: "FunctionApp",
			SourceAssetName: app.Name,
			Summary:         "FunctionApp '" + app.Name + "' publishes Azure-managed hostname '" + *app.DefaultHostname + "'. Validate whether that ingress path is intended and how it is constrained.",
		})
	}
	return endpoints
}

func endpointsFromContainerApps(containerApps []models.ContainerAppAsset) []models.EndpointSummary {
	endpoints := []models.EndpointSummary{}
	for _, app := range containerApps {
		if app.DefaultHostname == nil || *app.DefaultHostname == "" {
			continue
		}
		endpoints = append(endpoints, models.EndpointSummary{
			Endpoint:        *app.DefaultHostname,
			EndpointType:    "hostname",
			ExposureFamily:  "managed-web-hostname",
			IngressPath:     "azure-container-apps-default-hostname",
			RelatedIDs:      dedupeStrings([]string{app.ID, stringPtrValue(app.WorkloadPrincipalID)}),
			SourceAssetID:   app.ID,
			SourceAssetKind: "ContainerApp",
			SourceAssetName: app.Name,
			Summary:         "ContainerApp '" + app.Name + "' publishes Azure-managed hostname '" + *app.DefaultHostname + "'. Validate whether that ingress path is intended and how it is constrained.",
		})
	}
	return endpoints
}

func endpointsFromContainerInstances(containerInstances []models.ContainerInstanceAsset) []models.EndpointSummary {
	endpoints := []models.EndpointSummary{}
	for _, instance := range containerInstances {
		if instance.PublicIPAddress != nil && *instance.PublicIPAddress != "" {
			endpoints = append(endpoints, models.EndpointSummary{
				Endpoint:        *instance.PublicIPAddress,
				EndpointType:    "ip",
				ExposureFamily:  "public-ip",
				IngressPath:     "azure-container-instances-public-ip",
				RelatedIDs:      dedupeStrings(append([]string{instance.ID, stringPtrValue(instance.WorkloadPrincipalID)}, instance.WorkloadIdentityIDs...)),
				SourceAssetID:   instance.ID,
				SourceAssetKind: "ContainerInstance",
				SourceAssetName: instance.Name,
				Summary:         "ContainerInstance '" + instance.Name + "' exposes public IP " + *instance.PublicIPAddress + ". Review the visible ingress path, ports, and runtime posture together.",
			})
		}
		if instance.FQDN != nil && *instance.FQDN != "" {
			endpoints = append(endpoints, models.EndpointSummary{
				Endpoint:        *instance.FQDN,
				EndpointType:    "hostname",
				ExposureFamily:  "managed-container-fqdn",
				IngressPath:     "azure-container-instances-fqdn",
				RelatedIDs:      dedupeStrings(append([]string{instance.ID, stringPtrValue(instance.WorkloadPrincipalID)}, instance.WorkloadIdentityIDs...)),
				SourceAssetID:   instance.ID,
				SourceAssetKind: "ContainerInstance",
				SourceAssetName: instance.Name,
				Summary:         "ContainerInstance '" + instance.Name + "' publishes hostname '" + *instance.FQDN + "'. Validate whether that ingress path is intended and how it is constrained.",
			})
		}
	}
	return endpoints
}

func publicIPEndpointsFromContainerInstances(containerInstances []models.ContainerInstanceAsset) []models.EndpointSummary {
	endpoints := []models.EndpointSummary{}
	for _, endpoint := range endpointsFromContainerInstances(containerInstances) {
		if endpoint.EndpointType == "ip" && endpoint.ExposureFamily == "public-ip" {
			endpoints = append(endpoints, endpoint)
		}
	}
	return endpoints
}

func extractPowerState(properties *armcompute.VirtualMachineProperties) string {
	if properties == nil || properties.InstanceView == nil {
		return ""
	}
	for _, status := range properties.InstanceView.Statuses {
		code := stringValue(status.Code)
		if strings.HasPrefix(code, "PowerState/") {
			return strings.TrimPrefix(code, "PowerState/")
		}
	}
	return ""
}

func vmssAssetFromResource(vmss *armcompute.VirtualMachineScaleSet) (models.VmssAsset, []models.Issue) {
	if vmss == nil {
		return models.VmssAsset{}, nil
	}

	vmssID := stringValue(vmss.ID)
	vmssName := firstNonEmpty(stringValue(vmss.Name), "unknown")
	identityIDs := []string{}
	identityType := (*string)(nil)
	principalID := (*string)(nil)
	clientID := (*string)(nil)
	if vmss.Identity != nil {
		if stringValue(vmss.Identity.PrincipalID) != "" {
			identityIDs = append(identityIDs, vmssID+"/identities/system")
			principalID = stringPtr(stringValue(vmss.Identity.PrincipalID))
		}
		identityType = stringPtr(stringValue(vmss.Identity.Type))
		for identityID, details := range vmss.Identity.UserAssignedIdentities {
			identityIDs = append(identityIDs, identityID)
			if clientID == nil && details != nil {
				clientID = stringPtr(stringValue(details.ClientID))
			}
		}
	}

	networkCues, issues := vmssNetworkCues(vmss)
	skuName := (*string)(nil)
	instanceCount := (*int)(nil)
	orchestrationMode := (*string)(nil)
	upgradeMode := (*string)(nil)
	overprovision := (*bool)(nil)
	singlePlacementGroup := (*bool)(nil)
	zoneBalance := (*bool)(nil)

	if vmss.SKU != nil {
		skuName = stringPtr(stringValue(vmss.SKU.Name))
		if vmss.SKU.Capacity != nil {
			instanceCount = intPtr(int(*vmss.SKU.Capacity))
		}
	}
	if vmss.Properties != nil {
		orchestrationMode = stringPtr(stringValue(vmss.Properties.OrchestrationMode))
		upgradeMode = stringPtr(stringValue(vmss.Properties.UpgradePolicy.Mode))
		if vmss.Properties.Overprovision != nil {
			value := *vmss.Properties.Overprovision
			overprovision = &value
		}
		if vmss.Properties.SinglePlacementGroup != nil {
			value := *vmss.Properties.SinglePlacementGroup
			singlePlacementGroup = &value
		}
		if vmss.Properties.ZoneBalance != nil {
			value := *vmss.Properties.ZoneBalance
			zoneBalance = &value
		}
	}

	zones := []string{}
	for _, zone := range vmss.Zones {
		zones = append(zones, stringValue(zone))
	}
	zones = sortedUniqueStrings(zones)

	return models.VmssAsset{
		ApplicationGatewayBackendPoolCount: networkCues.applicationGatewayBackendPoolCount,
		ClientID:                           clientID,
		ID:                                 firstNonEmpty(vmssID, "/unknown/"+vmssName),
		IdentityIDs:                        sortedUniqueStrings(identityIDs),
		IdentityType:                       identityType,
		InboundNATPoolCount:                networkCues.inboundNATPoolCount,
		InstanceCount:                      instanceCount,
		LoadBalancerBackendPoolCount:       networkCues.loadBalancerBackendPoolCount,
		Location:                           stringValue(vmss.Location),
		Name:                               vmssName,
		NICConfigurationCount:              networkCues.nicConfigurationCount,
		OrchestrationMode:                  orchestrationMode,
		Overprovision:                      overprovision,
		PrincipalID:                        principalID,
		PublicIPConfigurationCount:         networkCues.publicIPConfigurationCount,
		RelatedIDs: dedupeStrings(append([]string{
			vmssID,
			stringPtrValue(principalID),
		}, append(append(append(identityIDs, networkCues.subnetIDs...), networkCues.loadBalancerBackendPoolIDs...), append(networkCues.applicationGatewayBackendPoolIDs, networkCues.inboundNATPoolIDs...)...)...)),
		ResourceGroup:        resourceGroupFromID(vmssID),
		SinglePlacementGroup: singlePlacementGroup,
		SKUName:              skuName,
		SubnetIDs:            networkCues.subnetIDs,
		Summary:              vmssOperatorSummary(vmssName, skuName, instanceCount, orchestrationMode, upgradeMode, overprovision, singlePlacementGroup, zoneBalance, zones, identityType, networkCues),
		UpgradeMode:          upgradeMode,
		ZoneBalance:          zoneBalance,
		Zones:                zones,
	}, issues
}

type vmssNetworkCueSummary struct {
	applicationGatewayBackendPoolCount int
	applicationGatewayBackendPoolIDs   []string
	inboundNATPoolCount                int
	inboundNATPoolIDs                  []string
	issues                             []models.Issue
	loadBalancerBackendPoolCount       int
	loadBalancerBackendPoolIDs         []string
	networkDetailComplete              bool
	nicConfigurationCount              int
	publicIPConfigurationCount         int
	subnetIDs                          []string
}

func vmssNetworkCues(vmss *armcompute.VirtualMachineScaleSet) (vmssNetworkCueSummary, []models.Issue) {
	vmssID := stringValue(vmss.ID)
	vmssName := firstNonEmpty(stringValue(vmss.Name), "unknown")

	if vmss == nil || vmss.Properties == nil || vmss.Properties.VirtualMachineProfile == nil {
		return vmssNetworkCueSummary{}, []models.Issue{partialCollectionIssue(
			"vmss.network_profile",
			"VM scale set frontend and subnet details were not returned by the current SDK list response; frontend counts may be incomplete.",
			vmssID,
			vmssName,
		)}
	}

	virtualMachineProfile := vmss.Properties.VirtualMachineProfile
	if virtualMachineProfile.NetworkProfile == nil {
		return vmssNetworkCueSummary{}, []models.Issue{partialCollectionIssue(
			"vmss.network_profile",
			"VM scale set network profile details were not returned by the current SDK list response; frontend counts may be incomplete.",
			vmssID,
			vmssName,
		)}
	}

	nicConfigs := virtualMachineProfile.NetworkProfile.NetworkInterfaceConfigurations
	if nicConfigs == nil {
		return vmssNetworkCueSummary{}, []models.Issue{partialCollectionIssue(
			"vmss.network_interface_configurations",
			"VM scale set NIC configuration details were not returned by the current SDK list response; subnet and frontend counts may be incomplete.",
			vmssID,
			vmssName,
		)}
	}

	subnetIDs := []string{}
	loadBalancerBackendPoolIDs := []string{}
	applicationGatewayBackendPoolIDs := []string{}
	inboundNATPoolIDs := []string{}
	publicIPConfigurationCount := 0

	for _, nicConfig := range nicConfigs {
		if nicConfig == nil || nicConfig.Properties == nil {
			continue
		}
		ipConfigurations := nicConfig.Properties.IPConfigurations
		if ipConfigurations == nil {
			return vmssNetworkCueSummary{nicConfigurationCount: len(nicConfigs)}, []models.Issue{partialCollectionIssue(
				"vmss.ip_configurations",
				"VM scale set IP configuration details were not returned by the current SDK list response; subnet and frontend counts may be incomplete.",
				vmssID,
				vmssName,
			)}
		}
		for _, ipConfiguration := range ipConfigurations {
			if ipConfiguration == nil || ipConfiguration.Properties == nil {
				continue
			}
			properties := ipConfiguration.Properties
			if properties.Subnet != nil {
				subnetIDs = append(subnetIDs, stringValue(properties.Subnet.ID))
			}
			for _, pool := range properties.LoadBalancerBackendAddressPools {
				loadBalancerBackendPoolIDs = append(loadBalancerBackendPoolIDs, stringValue(pool.ID))
			}
			for _, pool := range properties.ApplicationGatewayBackendAddressPools {
				applicationGatewayBackendPoolIDs = append(applicationGatewayBackendPoolIDs, stringValue(pool.ID))
			}
			for _, pool := range properties.LoadBalancerInboundNatPools {
				inboundNATPoolIDs = append(inboundNATPoolIDs, stringValue(pool.ID))
			}
			if properties.PublicIPAddressConfiguration != nil {
				publicIPConfigurationCount++
			}
		}
	}

	return vmssNetworkCueSummary{
		applicationGatewayBackendPoolCount: len(sortedUniqueStrings(applicationGatewayBackendPoolIDs)),
		applicationGatewayBackendPoolIDs:   sortedUniqueStrings(applicationGatewayBackendPoolIDs),
		inboundNATPoolCount:                len(sortedUniqueStrings(inboundNATPoolIDs)),
		inboundNATPoolIDs:                  sortedUniqueStrings(inboundNATPoolIDs),
		loadBalancerBackendPoolCount:       len(sortedUniqueStrings(loadBalancerBackendPoolIDs)),
		loadBalancerBackendPoolIDs:         sortedUniqueStrings(loadBalancerBackendPoolIDs),
		networkDetailComplete:              true,
		nicConfigurationCount:              len(nicConfigs),
		publicIPConfigurationCount:         publicIPConfigurationCount,
		subnetIDs:                          sortedUniqueStrings(subnetIDs),
	}, nil
}

func vmssOperatorSummary(
	vmssName string,
	skuName *string,
	instanceCount *int,
	orchestrationMode *string,
	upgradeMode *string,
	overprovision *bool,
	singlePlacementGroup *bool,
	zoneBalance *bool,
	zones []string,
	identityType *string,
	networkCues vmssNetworkCueSummary,
) string {
	identityPhrase := "has no managed identity visible from the current read path"
	if identityType != nil && *identityType != "" {
		identityPhrase = "uses managed identity (" + *identityType + ")"
	}

	footprintParts := []string{}
	if skuName != nil && *skuName != "" {
		footprintParts = append(footprintParts, "SKU "+*skuName)
	}
	if instanceCount != nil {
		footprintParts = append(footprintParts, strconv.Itoa(*instanceCount)+" configured instance(s)")
	}
	footprintPhrase := ""
	if len(footprintParts) > 0 {
		footprintPhrase = strings.Join(footprintParts, ", ") + " and "
	}

	networkParts := []string{}
	if networkCues.publicIPConfigurationCount > 0 {
		networkParts = append(networkParts, strconv.Itoa(networkCues.publicIPConfigurationCount)+" public IP config(s)")
	}
	if networkCues.inboundNATPoolCount > 0 {
		networkParts = append(networkParts, strconv.Itoa(networkCues.inboundNATPoolCount)+" inbound NAT pool ref(s)")
	}
	if networkCues.loadBalancerBackendPoolCount > 0 {
		networkParts = append(networkParts, strconv.Itoa(networkCues.loadBalancerBackendPoolCount)+" LB backend pool ref(s)")
	}
	if networkCues.applicationGatewayBackendPoolCount > 0 {
		networkParts = append(networkParts, strconv.Itoa(networkCues.applicationGatewayBackendPoolCount)+" App Gateway backend pool ref(s)")
	}
	if networkCues.nicConfigurationCount > 0 {
		networkParts = append(networkParts, strconv.Itoa(networkCues.nicConfigurationCount)+" NIC config(s)")
	}
	if len(networkCues.subnetIDs) > 0 {
		networkParts = append(networkParts, strconv.Itoa(len(networkCues.subnetIDs))+" subnet ref(s)")
	}

	networkPhrase := "Visible frontend or network cues are not readable from the current SDK response."
	if len(networkParts) > 0 {
		networkPhrase = "Visible frontend or network cues: " + strings.Join(networkParts, ", ") + "."
	} else if networkCues.networkDetailComplete {
		networkPhrase = "Visible frontend or network cues: none from the current read path."
	}

	postureParts := []string{}
	if orchestrationMode != nil && *orchestrationMode != "" {
		postureParts = append(postureParts, "orchestration "+*orchestrationMode)
	}
	if upgradeMode != nil && *upgradeMode != "" {
		postureParts = append(postureParts, "upgrade "+*upgradeMode)
	}
	if singlePlacementGroup != nil {
		postureParts = append(postureParts, "single-placement-group "+yesNo(*singlePlacementGroup))
	}
	if overprovision != nil {
		postureParts = append(postureParts, "overprovision "+yesNo(*overprovision))
	}
	if zoneBalance != nil {
		postureParts = append(postureParts, "zone-balance "+yesNo(*zoneBalance))
	}
	if len(zones) > 0 {
		postureParts = append(postureParts, "zones "+strings.Join(zones, ","))
	}

	posturePhrase := ""
	if len(postureParts) > 0 {
		posturePhrase = " Visible posture: " + strings.Join(postureParts, ", ") + "."
	}

	return "Virtual Machine Scale Sets (VMSS) asset '" + vmssName + "' carries " + footprintPhrase + identityPhrase + ". " + networkPhrase + posturePhrase
}

func workloadRowsFromAppServices(appServices []models.AppServiceAsset, endpointsByAsset map[string][]models.EndpointSummary) []models.WorkloadSummary {
	workloads := make([]models.WorkloadSummary, 0, len(appServices))
	for _, app := range appServices {
		assetEndpoints := workloadEndpointFacts(endpointsByAsset[armIDJoinKey(app.ID)])
		networkSignals := []string{}
		if app.DefaultHostname != nil && *app.DefaultHostname != "" {
			networkSignals = append(networkSignals, "default-hostname")
		}
		workloads = append(workloads, models.WorkloadSummary{
			AssetID:             app.ID,
			AssetKind:           "AppService",
			AssetName:           app.Name,
			Endpoints:           assetEndpoints.endpoints,
			ExposureFamilies:    assetEndpoints.exposureFamilies,
			IdentityClientID:    app.WorkloadClientID,
			IdentityIDs:         sortedUniqueStrings(app.WorkloadIdentityIDs),
			IdentityPrincipalID: app.WorkloadPrincipalID,
			IdentityType:        app.WorkloadIdentityType,
			IngressPaths:        assetEndpoints.ingressPaths,
			Location:            app.Location,
			RelatedIDs:          dedupeStrings([]string{app.ID, stringPtrValue(app.WorkloadPrincipalID)}),
			ResourceGroup:       app.ResourceGroup,
			Summary: workloadSummaryText(
				"AppService",
				app.Name,
				assetEndpoints.endpoints,
				assetEndpoints.exposureFamilies,
				stringPtrValue(app.WorkloadIdentityType),
				networkSignals,
			),
		})
	}
	return workloads
}

func workloadRowsFromFunctions(functionApps []models.FunctionAppAsset, endpointsByAsset map[string][]models.EndpointSummary) []models.WorkloadSummary {
	workloads := make([]models.WorkloadSummary, 0, len(functionApps))
	for _, app := range functionApps {
		assetEndpoints := workloadEndpointFacts(endpointsByAsset[armIDJoinKey(app.ID)])
		networkSignals := []string{}
		if app.DefaultHostname != nil && *app.DefaultHostname != "" {
			networkSignals = append(networkSignals, "default-hostname")
		}
		if len(app.WorkloadIdentityIDs) > 0 {
			networkSignals = append(networkSignals, "user-assigned="+strconv.Itoa(len(app.WorkloadIdentityIDs)))
		}
		workloads = append(workloads, models.WorkloadSummary{
			AssetID:             app.ID,
			AssetKind:           "FunctionApp",
			AssetName:           app.Name,
			Endpoints:           assetEndpoints.endpoints,
			ExposureFamilies:    assetEndpoints.exposureFamilies,
			IdentityClientID:    app.WorkloadClientID,
			IdentityIDs:         sortedUniqueStrings(app.WorkloadIdentityIDs),
			IdentityPrincipalID: app.WorkloadPrincipalID,
			IdentityType:        app.WorkloadIdentityType,
			IngressPaths:        assetEndpoints.ingressPaths,
			Location:            app.Location,
			RelatedIDs:          dedupeStrings(append([]string{app.ID, stringPtrValue(app.WorkloadPrincipalID)}, app.WorkloadIdentityIDs...)),
			ResourceGroup:       app.ResourceGroup,
			Summary: workloadSummaryText(
				"FunctionApp",
				app.Name,
				assetEndpoints.endpoints,
				assetEndpoints.exposureFamilies,
				stringPtrValue(app.WorkloadIdentityType),
				networkSignals,
			),
		})
	}
	return workloads
}

func workloadRowsFromContainerApps(containerApps []models.ContainerAppAsset, endpointsByAsset map[string][]models.EndpointSummary) []models.WorkloadSummary {
	workloads := make([]models.WorkloadSummary, 0, len(containerApps))
	for _, app := range containerApps {
		assetEndpoints := workloadEndpointFacts(endpointsByAsset[armIDJoinKey(app.ID)])
		networkSignals := []string{}
		if app.DefaultHostname != nil && *app.DefaultHostname != "" {
			networkSignals = append(networkSignals, "default-hostname")
		}
		if app.ExternalIngressEnabled != nil {
			if *app.ExternalIngressEnabled {
				networkSignals = append(networkSignals, "external-ingress")
			} else {
				networkSignals = append(networkSignals, "internal-only")
			}
		}
		if len(app.WorkloadIdentityIDs) > 0 {
			networkSignals = append(networkSignals, "user-assigned="+strconv.Itoa(len(app.WorkloadIdentityIDs)))
		}
		workloads = append(workloads, models.WorkloadSummary{
			AssetID:             app.ID,
			AssetKind:           "ContainerApp",
			AssetName:           app.Name,
			Endpoints:           assetEndpoints.endpoints,
			ExposureFamilies:    assetEndpoints.exposureFamilies,
			IdentityClientID:    app.WorkloadClientID,
			IdentityIDs:         sortedUniqueStrings(app.WorkloadIdentityIDs),
			IdentityPrincipalID: app.WorkloadPrincipalID,
			IdentityType:        app.WorkloadIdentityType,
			IngressPaths:        assetEndpoints.ingressPaths,
			Location:            app.Location,
			RelatedIDs:          dedupeStrings(append([]string{app.ID, stringPtrValue(app.WorkloadPrincipalID)}, app.WorkloadIdentityIDs...)),
			ResourceGroup:       app.ResourceGroup,
			Summary: workloadSummaryText(
				"ContainerApp",
				app.Name,
				assetEndpoints.endpoints,
				assetEndpoints.exposureFamilies,
				stringPtrValue(app.WorkloadIdentityType),
				networkSignals,
			),
		})
	}
	return workloads
}

func workloadRowsFromContainerInstances(containerInstances []models.ContainerInstanceAsset, endpointsByAsset map[string][]models.EndpointSummary) []models.WorkloadSummary {
	workloads := make([]models.WorkloadSummary, 0, len(containerInstances))
	for _, instance := range containerInstances {
		assetEndpoints := workloadEndpointFacts(endpointsByAsset[armIDJoinKey(instance.ID)])
		networkSignals := []string{}
		if instance.PublicIPAddress != nil && *instance.PublicIPAddress != "" {
			networkSignals = append(networkSignals, "public-ip")
		}
		if instance.FQDN != nil && *instance.FQDN != "" {
			networkSignals = append(networkSignals, "fqdn")
		}
		if len(instance.SubnetIDs) > 0 {
			networkSignals = append(networkSignals, "subnets="+strconv.Itoa(len(instance.SubnetIDs)))
		}
		if len(instance.ExposedPorts) > 0 {
			networkSignals = append(networkSignals, "ports="+strconv.Itoa(len(instance.ExposedPorts)))
		}
		if instance.ContainerCount != nil {
			networkSignals = append(networkSignals, "containers="+strconv.Itoa(*instance.ContainerCount))
		}
		if len(instance.WorkloadIdentityIDs) > 0 {
			networkSignals = append(networkSignals, "user-assigned="+strconv.Itoa(len(instance.WorkloadIdentityIDs)))
		}
		workloads = append(workloads, models.WorkloadSummary{
			AssetID:             instance.ID,
			AssetKind:           "ContainerInstance",
			AssetName:           instance.Name,
			Endpoints:           assetEndpoints.endpoints,
			ExposureFamilies:    assetEndpoints.exposureFamilies,
			IdentityClientID:    instance.WorkloadClientID,
			IdentityIDs:         sortedUniqueStrings(instance.WorkloadIdentityIDs),
			IdentityPrincipalID: instance.WorkloadPrincipalID,
			IdentityType:        instance.WorkloadIdentityType,
			IngressPaths:        assetEndpoints.ingressPaths,
			Location:            instance.Location,
			RelatedIDs:          dedupeStrings(append(append([]string{instance.ID, stringPtrValue(instance.WorkloadPrincipalID)}, instance.WorkloadIdentityIDs...), instance.SubnetIDs...)),
			ResourceGroup:       instance.ResourceGroup,
			Summary: workloadSummaryText(
				"ContainerInstance",
				instance.Name,
				assetEndpoints.endpoints,
				assetEndpoints.exposureFamilies,
				stringPtrValue(instance.WorkloadIdentityType),
				networkSignals,
			),
		})
	}
	return workloads
}

func workloadRowsFromVMs(vmAssets []models.VmAsset, endpointsByAsset map[string][]models.EndpointSummary) []models.WorkloadSummary {
	workloads := make([]models.WorkloadSummary, 0, len(vmAssets))
	for _, vm := range vmAssets {
		assetEndpoints := workloadEndpointFacts(endpointsByAsset[armIDJoinKey(vm.ID)])
		networkSignals := []string{}
		if len(vm.PublicIPs) > 0 {
			networkSignals = append(networkSignals, "public-ip="+strconv.Itoa(len(vm.PublicIPs)))
		}
		if len(vm.PrivateIPs) > 0 {
			networkSignals = append(networkSignals, "private-ip="+strconv.Itoa(len(vm.PrivateIPs)))
		}
		if len(vm.NICIDs) > 0 {
			networkSignals = append(networkSignals, "nic="+strconv.Itoa(len(vm.NICIDs)))
		}
		identityType := vmIdentityType(vm.IdentityIDs)
		workloads = append(workloads, models.WorkloadSummary{
			AssetID:             vm.ID,
			AssetKind:           "VM",
			AssetName:           vm.Name,
			Endpoints:           assetEndpoints.endpoints,
			ExposureFamilies:    assetEndpoints.exposureFamilies,
			IdentityClientID:    nil,
			IdentityIDs:         sortedUniqueStrings(vm.IdentityIDs),
			IdentityPrincipalID: nil,
			IdentityType:        stringPtr(identityType),
			IngressPaths:        assetEndpoints.ingressPaths,
			Location:            vm.Location,
			RelatedIDs:          dedupeStrings(append(append([]string{vm.ID}, vm.IdentityIDs...), vm.NICIDs...)),
			ResourceGroup:       vm.ResourceGroup,
			Summary: workloadSummaryText(
				"VM",
				vm.Name,
				assetEndpoints.endpoints,
				assetEndpoints.exposureFamilies,
				identityType,
				networkSignals,
			),
		})
	}
	return workloads
}

func workloadRowsFromVMSS(vmssAssets []models.VmssAsset) []models.WorkloadSummary {
	workloads := make([]models.WorkloadSummary, 0, len(vmssAssets))
	for _, asset := range vmssAssets {
		workloads = append(workloads, models.WorkloadSummary{
			AssetID:             asset.ID,
			AssetKind:           "VMSS",
			AssetName:           asset.Name,
			Endpoints:           []string{},
			ExposureFamilies:    []string{},
			IdentityClientID:    asset.ClientID,
			IdentityIDs:         sortedUniqueStrings(asset.IdentityIDs),
			IdentityPrincipalID: asset.PrincipalID,
			IdentityType:        asset.IdentityType,
			IngressPaths:        []string{},
			Location:            asset.Location,
			RelatedIDs:          dedupeStrings([]string{asset.ID, stringPtrValue(asset.PrincipalID), stringPtrValue(asset.ClientID)}),
			ResourceGroup:       asset.ResourceGroup,
			Summary: workloadSummaryText(
				"VMSS",
				asset.Name,
				nil,
				nil,
				stringPtrValue(asset.IdentityType),
				nil,
			),
		})
	}
	return workloads
}

type workloadEndpointSummary struct {
	endpoints        []string
	exposureFamilies []string
	ingressPaths     []string
}

func workloadEndpointFacts(endpoints []models.EndpointSummary) workloadEndpointSummary {
	return workloadEndpointSummary{
		endpoints:        sortedUniqueStrings(endpointValues(endpoints)),
		exposureFamilies: sortedUniqueStrings(endpointExposureFamilies(endpoints)),
		ingressPaths:     sortedUniqueStrings(endpointIngressPaths(endpoints)),
	}
}

func endpointValues(endpoints []models.EndpointSummary) []string {
	values := make([]string, 0, len(endpoints))
	for _, endpoint := range endpoints {
		values = append(values, endpoint.Endpoint)
	}
	return values
}

func endpointExposureFamilies(endpoints []models.EndpointSummary) []string {
	values := make([]string, 0, len(endpoints))
	for _, endpoint := range endpoints {
		values = append(values, endpoint.ExposureFamily)
	}
	return values
}

func endpointIngressPaths(endpoints []models.EndpointSummary) []string {
	values := make([]string, 0, len(endpoints))
	for _, endpoint := range endpoints {
		values = append(values, endpoint.IngressPath)
	}
	return values
}

func endpointsByAssetID(endpoints []models.EndpointSummary) map[string][]models.EndpointSummary {
	grouped := map[string][]models.EndpointSummary{}
	for _, endpoint := range endpoints {
		key := armIDJoinKey(endpoint.SourceAssetID)
		grouped[key] = append(grouped[key], endpoint)
	}
	return grouped
}

func vmIdentityType(identityIDs []string) string {
	hasSystem := false
	hasUser := false
	for _, identityID := range identityIDs {
		if strings.HasSuffix(identityID, "/identities/system") {
			hasSystem = true
		} else {
			hasUser = true
		}
	}
	switch {
	case hasSystem && hasUser:
		return "SystemAssigned, UserAssigned"
	case hasSystem:
		return "SystemAssigned"
	case hasUser:
		return "UserAssigned"
	default:
		return ""
	}
}

func workloadSummaryText(assetKind string, assetName string, endpoints []string, exposureFamilies []string, identityType string, networkSignals []string) string {
	endpointPhrase := "has no visible endpoint path from the current read path"
	if len(endpoints) > 0 {
		allPublicIP := len(exposureFamilies) > 0
		for _, family := range exposureFamilies {
			if family != "public-ip" {
				allPublicIP = false
				break
			}
		}
		switch {
		case allPublicIP && len(endpoints) == 1:
			endpointPhrase = "exposes reachable endpoint '" + endpoints[0] + "'"
		case allPublicIP:
			endpointPhrase = "exposes " + strconv.Itoa(len(endpoints)) + " reachable endpoints (" + strings.Join(endpoints, ", ") + ")"
		case len(endpoints) == 1:
			endpointPhrase = "publishes visible endpoint hostname '" + endpoints[0] + "'"
		default:
			endpointPhrase = "publishes " + strconv.Itoa(len(endpoints)) + " visible endpoint paths (" + strings.Join(endpoints, ", ") + ")"
		}
	}

	identityPhrase := "has no managed identity context visible from the current read path"
	if identityType != "" {
		identityPhrase = "carries managed identity context (" + identityType + ")"
	}

	signalPhrase := ""
	if len(networkSignals) > 0 {
		signalPhrase = " Visible signals: " + strings.Join(networkSignals, ", ") + "."
	}

	return assetKind + " '" + assetName + "' " + endpointPhrase + " and " + identityPhrase + "." +
		signalPhrase + " Use this as a quick workload census pivot before deeper service-specific review."
}

func workloadLess(left models.WorkloadSummary, right models.WorkloadSummary) bool {
	if (len(left.Endpoints) == 0) != (len(right.Endpoints) == 0) {
		return len(left.Endpoints) > 0
	}
	if (left.IdentityType == nil) != (right.IdentityType == nil) {
		return left.IdentityType != nil
	}
	if workloadKindRank(left.AssetKind) != workloadKindRank(right.AssetKind) {
		return workloadKindRank(left.AssetKind) < workloadKindRank(right.AssetKind)
	}
	return left.AssetName < right.AssetName
}

func workloadKindRank(value string) int {
	switch value {
	case "VM":
		return 0
	case "AppService":
		return 1
	case "FunctionApp":
		return 2
	case "ContainerApp":
		return 3
	case "ContainerInstance":
		return 4
	case "VMSS":
		return 5
	default:
		return 9
	}
}

func tokenCredentialSurfacesFromAppServices(appServices []models.AppServiceAsset) []models.TokenCredentialSurfaceSummary {
	surfaces := make([]models.TokenCredentialSurfaceSummary, 0, len(appServices))
	for _, app := range appServices {
		if surface, ok := tokenCredentialManagedIdentitySurface(
			app.ID,
			"AppService",
			app.Name,
			app.ResourceGroup,
			app.Location,
			app.WorkloadIdentityType,
			app.WorkloadIdentityIDs,
			app.WorkloadPrincipalID,
		); ok {
			surfaces = append(surfaces, surface)
		}
	}
	return surfaces
}

func tokenCredentialSurfacesFromFunctions(functionApps []models.FunctionAppAsset) []models.TokenCredentialSurfaceSummary {
	surfaces := make([]models.TokenCredentialSurfaceSummary, 0, len(functionApps))
	for _, app := range functionApps {
		if surface, ok := tokenCredentialManagedIdentitySurface(
			app.ID,
			"FunctionApp",
			app.Name,
			app.ResourceGroup,
			app.Location,
			app.WorkloadIdentityType,
			app.WorkloadIdentityIDs,
			app.WorkloadPrincipalID,
		); ok {
			surfaces = append(surfaces, surface)
		}
	}
	return surfaces
}

func tokenCredentialSurfacesFromContainerApps(containerApps []models.ContainerAppAsset) []models.TokenCredentialSurfaceSummary {
	surfaces := make([]models.TokenCredentialSurfaceSummary, 0, len(containerApps))
	for _, app := range containerApps {
		if surface, ok := tokenCredentialManagedIdentitySurface(
			app.ID,
			"ContainerApp",
			app.Name,
			app.ResourceGroup,
			app.Location,
			app.WorkloadIdentityType,
			app.WorkloadIdentityIDs,
			app.WorkloadPrincipalID,
		); ok {
			surfaces = append(surfaces, surface)
		}
	}
	return surfaces
}

func tokenCredentialSurfacesFromContainerInstances(containerInstances []models.ContainerInstanceAsset) []models.TokenCredentialSurfaceSummary {
	surfaces := make([]models.TokenCredentialSurfaceSummary, 0, len(containerInstances))
	for _, instance := range containerInstances {
		if surface, ok := tokenCredentialManagedIdentitySurface(
			instance.ID,
			"ContainerInstance",
			instance.Name,
			instance.ResourceGroup,
			instance.Location,
			instance.WorkloadIdentityType,
			instance.WorkloadIdentityIDs,
			instance.WorkloadPrincipalID,
		); ok {
			surfaces = append(surfaces, surface)
		}
	}
	return surfaces
}

func tokenCredentialManagedIdentitySurface(assetID string, assetKind string, assetName string, resourceGroup string, location string, identityType *string, identityIDs []string, principalID *string) (models.TokenCredentialSurfaceSummary, bool) {
	identityLabel := stringPtrValue(identityType)
	if identityLabel == "" {
		return models.TokenCredentialSurfaceSummary{}, false
	}

	operatorSignal := identityLabel
	if len(identityIDs) > 0 {
		operatorSignal += "; user-assigned=" + strconv.Itoa(len(identityIDs))
	}
	nextReviewKind := tokenCredentialNextReviewKind(models.TokenCredentialSurfaceManagedIdentityToken, "workload-identity", operatorSignal)

	return models.TokenCredentialSurfaceSummary{
		AccessPath:     "workload-identity",
		AssetID:        firstNonEmpty(assetID, "/unknown/"+assetName),
		AssetKind:      assetKind,
		AssetName:      firstNonEmpty(assetName, "unknown"),
		Location:       stringPtr(location),
		OperatorSignal: operatorSignal,
		Priority:       "medium",
		RelatedIDs:     dedupeStrings(append(append([]string{assetID, stringPtrValue(principalID)}, identityIDs...), stringPtrValue(principalID))),
		ResourceGroup:  stringPtr(resourceGroup),
		Summary:        assetKind + " '" + firstNonEmpty(assetName, "unknown") + "' can request tokens through attached managed identity (" + identityLabel + "). " + tokenCredentialNextReviewText(nextReviewKind),
		SurfaceType:    models.TokenCredentialSurfaceManagedIdentityToken,

		NextReviewKind: nextReviewKind,
	}, true
}

func tokenCredentialSurfacesFromEnvVars(envVars []models.EnvVarSummary) []models.TokenCredentialSurfaceSummary {
	surfaces := []models.TokenCredentialSurfaceSummary{}
	for _, envVar := range envVars {
		relatedIDs := dedupeStrings(append(append([]string{envVar.AssetID, stringPtrValue(envVar.WorkloadPrincipalID)}, envVar.WorkloadIdentityIDs...), stringPtrValue(envVar.WorkloadPrincipalID)))

		if tokenCredentialPlainTextEnvVar(envVar) {
			nextReviewKind := tokenCredentialNextReviewKind(models.TokenCredentialSurfacePlainTextSecret, "app-setting", "setting="+envVar.SettingName)
			surfaces = append(surfaces, models.TokenCredentialSurfaceSummary{
				AccessPath:     "app-setting",
				AssetID:        envVar.AssetID,
				AssetKind:      envVar.AssetKind,
				AssetName:      envVar.AssetName,
				Location:       stringPtr(envVar.Location),
				OperatorSignal: "setting=" + envVar.SettingName,
				Priority:       "high",
				RelatedIDs:     relatedIDs,
				ResourceGroup:  stringPtr(envVar.ResourceGroup),
				Summary:        envVar.AssetKind + " '" + envVar.AssetName + "' exposes credential-like setting '" + envVar.SettingName + "' as plain-text management-plane app configuration. " + tokenCredentialNextReviewText(nextReviewKind),
				SurfaceType:    models.TokenCredentialSurfacePlainTextSecret,

				NextReviewKind: nextReviewKind,
			})
		}

		if envVar.ValueType == "keyvault-ref" {
			signalParts := []string{"target=" + valueOrUnknown(stringPtrValue(envVar.ReferenceTarget))}
			identitySummary := keyVaultReferenceIdentitySummary(stringPtrValue(envVar.KeyVaultReferenceIdentity))
			if identitySummary != "" {
				signalParts = append(signalParts, "identity="+identitySummary)
			}
			operatorSignal := strings.Join(signalParts, "; ")
			nextReviewKind := tokenCredentialNextReviewKind(models.TokenCredentialSurfaceKeyVaultReference, "app-setting", operatorSignal)

			targetSuffix := ""
			if envVar.ReferenceTarget != nil && *envVar.ReferenceTarget != "" {
				targetSuffix = " (" + *envVar.ReferenceTarget + ")"
			}
			identitySuffix := ""
			if identitySummary != "" {
				identitySuffix = " via " + identitySummary
			}

			surfaces = append(surfaces, models.TokenCredentialSurfaceSummary{
				AccessPath:     "app-setting",
				AssetID:        envVar.AssetID,
				AssetKind:      envVar.AssetKind,
				AssetName:      envVar.AssetName,
				Location:       stringPtr(envVar.Location),
				OperatorSignal: operatorSignal,
				Priority:       "medium",
				RelatedIDs:     relatedIDs,
				ResourceGroup:  stringPtr(envVar.ResourceGroup),
				Summary:        envVar.AssetKind + " '" + envVar.AssetName + "' uses setting '" + envVar.SettingName + "' to reach Key Vault-backed secret material" + targetSuffix + identitySuffix + ". " + tokenCredentialNextReviewText(nextReviewKind),
				SurfaceType:    models.TokenCredentialSurfaceKeyVaultReference,

				NextReviewKind: nextReviewKind,
			})
		}
	}
	return surfaces
}

func tokenCredentialPlainTextEnvVar(envVar models.EnvVarSummary) bool {
	if envVar.ValueType != "plain-text" {
		return false
	}
	if envVar.LooksSensitive {
		return true
	}
	return envVar.AssetKind == "FunctionApp" && envVar.SettingName == "AzureWebJobsStorage"
}

func tokenCredentialSurfacesFromArmDeployments(deployments []models.ArmDeploymentSummary) []models.TokenCredentialSurfaceSummary {
	surfaces := []models.TokenCredentialSurfaceSummary{}
	for _, deployment := range deployments {
		relatedIDs := dedupeStrings([]string{deployment.ID})

		if deployment.OutputsCount > 0 {
			operatorSignal := "outputs=" + strconv.Itoa(deployment.OutputsCount) + "; providers=" + strconv.Itoa(len(deployment.Providers))
			nextReviewKind := tokenCredentialNextReviewKind(models.TokenCredentialSurfaceDeploymentOutput, "deployment-history", operatorSignal)
			surfaces = append(surfaces, models.TokenCredentialSurfaceSummary{
				AccessPath:     "deployment-history",
				AssetID:        firstNonEmpty(deployment.ID, "/unknown/"+deployment.Name),
				AssetKind:      "ArmDeployment",
				AssetName:      firstNonEmpty(deployment.Name, "unknown"),
				Location:       nil,
				OperatorSignal: operatorSignal,
				Priority:       "medium",
				RelatedIDs:     relatedIDs,
				ResourceGroup:  deployment.ResourceGroup,
				Summary:        "Deployment '" + firstNonEmpty(deployment.Name, "unknown") + "' recorded " + strconv.Itoa(deployment.OutputsCount) + " output values in deployment history. " + tokenCredentialNextReviewText(nextReviewKind),
				SurfaceType:    models.TokenCredentialSurfaceDeploymentOutput,

				NextReviewKind: nextReviewKind,
			})
		}

		linkParts := []string{}
		if deployment.TemplateLink != nil && *deployment.TemplateLink != "" {
			linkParts = append(linkParts, "template="+compactLink(*deployment.TemplateLink))
		}
		if deployment.ParametersLink != nil && *deployment.ParametersLink != "" {
			linkParts = append(linkParts, "parameters="+compactLink(*deployment.ParametersLink))
		}
		if len(linkParts) > 0 {
			operatorSignal := strings.Join(linkParts, "; ")
			nextReviewKind := tokenCredentialNextReviewKind(models.TokenCredentialSurfaceLinkedDeploymentAsset, "deployment-history", operatorSignal)
			surfaces = append(surfaces, models.TokenCredentialSurfaceSummary{
				AccessPath:     "deployment-history",
				AssetID:        firstNonEmpty(deployment.ID, "/unknown/"+deployment.Name),
				AssetKind:      "ArmDeployment",
				AssetName:      firstNonEmpty(deployment.Name, "unknown"),
				Location:       nil,
				OperatorSignal: operatorSignal,
				Priority:       "low",
				RelatedIDs:     relatedIDs,
				ResourceGroup:  deployment.ResourceGroup,
				Summary:        "Deployment '" + firstNonEmpty(deployment.Name, "unknown") + "' references remote template or parameter content that may expose reusable configuration or credential context. " + tokenCredentialNextReviewText(nextReviewKind),
				SurfaceType:    models.TokenCredentialSurfaceLinkedDeploymentAsset,

				NextReviewKind: nextReviewKind,
			})
		}
	}
	return surfaces
}

func tokenCredentialSurfacesFromVMs(vmAssets []models.VmAsset) []models.TokenCredentialSurfaceSummary {
	surfaces := []models.TokenCredentialSurfaceSummary{}
	for _, vm := range vmAssets {
		identityIDs := sortedUniqueStrings(vm.IdentityIDs)
		if len(identityIDs) == 0 {
			continue
		}

		assetKind := strings.ToUpper(firstNonEmpty(vm.VMType, "vm"))
		operatorSignal := "public-ip=none; identities=" + strconv.Itoa(len(identityIDs))
		priority := "medium"
		nextReviewKind := models.TokenCredentialReviewManagedIdentityAndPermissions
		summary := assetKind + " '" + vm.Name + "' exposes a token minting path through IMDS for its attached managed identity. " + tokenCredentialNextReviewText(nextReviewKind)
		publiclyReachable := false

		if len(vm.PublicIPs) > 0 {
			operatorSignal = "public-ip=" + vm.PublicIPs[0] + "; identities=" + strconv.Itoa(len(identityIDs))
			priority = "high"
			nextReviewKind = models.TokenCredentialReviewEndpointsIngressAndControl
			summary = assetKind + " '" + vm.Name + "' is publicly reachable and exposes a token minting path through IMDS for its attached managed identity. " + tokenCredentialNextReviewText(nextReviewKind)
			publiclyReachable = true
		}

		surfaces = append(surfaces, models.TokenCredentialSurfaceSummary{
			AccessPath:     "imds",
			AssetID:        firstNonEmpty(vm.ID, "/unknown/"+vm.Name),
			AssetKind:      assetKind,
			AssetName:      firstNonEmpty(vm.Name, "unknown"),
			Location:       stringPtr(vm.Location),
			OperatorSignal: operatorSignal,
			Priority:       priority,
			RelatedIDs:     dedupeStrings(append([]string{vm.ID}, identityIDs...)),
			ResourceGroup:  stringPtr(vm.ResourceGroup),
			Summary:        summary,
			SurfaceType:    models.TokenCredentialSurfaceManagedIdentityToken,

			NextReviewKind:    nextReviewKind,
			PubliclyReachable: publiclyReachable,
		})
	}
	return surfaces
}

func tokenCredentialSurfacesFromVMSS(vmssAssets []models.VmssAsset) []models.TokenCredentialSurfaceSummary {
	surfaces := []models.TokenCredentialSurfaceSummary{}
	for _, asset := range vmssAssets {
		identityIDs := sortedUniqueStrings(asset.IdentityIDs)
		if len(identityIDs) == 0 {
			continue
		}

		operatorSignal := "public-ip=none; identities=" + strconv.Itoa(len(identityIDs))
		if asset.PublicIPConfigurationCount > 0 {
			operatorSignal = "public-ip-configs=" + strconv.Itoa(asset.PublicIPConfigurationCount) + "; identities=" + strconv.Itoa(len(identityIDs))
		}
		nextReviewKind := models.TokenCredentialReviewManagedIdentityAndPermissions

		surfaces = append(surfaces, models.TokenCredentialSurfaceSummary{
			AccessPath:     "imds",
			AssetID:        firstNonEmpty(asset.ID, "/unknown/"+asset.Name),
			AssetKind:      "VMSS",
			AssetName:      firstNonEmpty(asset.Name, "unknown"),
			Location:       stringPtr(asset.Location),
			OperatorSignal: operatorSignal,
			Priority:       "medium",
			RelatedIDs:     dedupeStrings(append([]string{asset.ID, stringPtrValue(asset.PrincipalID)}, identityIDs...)),
			ResourceGroup:  stringPtr(asset.ResourceGroup),
			Summary:        "VMSS '" + firstNonEmpty(asset.Name, "unknown") + "' exposes a token minting path through IMDS for its attached managed identity. " + tokenCredentialNextReviewText(nextReviewKind),
			SurfaceType:    models.TokenCredentialSurfaceManagedIdentityToken,

			NextReviewKind: nextReviewKind,
		})
	}
	return surfaces
}

func tokenCredentialNextReviewKind(surfaceType models.TokenCredentialSurfaceType, accessPath string, operatorSignal string) models.TokenCredentialNextReviewKind {
	switch surfaceType {
	case models.TokenCredentialSurfacePlainTextSecret:
		return models.TokenCredentialReviewEnvVarsSettingContext
	case models.TokenCredentialSurfaceManagedIdentityToken:
		if accessPath == "imds" {
			if publicIP := tokenCredentialSignalValue(operatorSignal, "public-ip"); publicIP != "" && publicIP != "none" {
				return models.TokenCredentialReviewEndpointsIngressAndControl
			}
		}
		return models.TokenCredentialReviewManagedIdentityAndPermissions
	case models.TokenCredentialSurfaceKeyVaultReference:
		if tokenCredentialSignalValue(operatorSignal, "identity") != "" {
			return models.TokenCredentialReviewKeyVaultAndManagedIdentity
		}
		return models.TokenCredentialReviewKeyVaultBoundary
	case models.TokenCredentialSurfaceDeploymentOutput:
		return models.TokenCredentialReviewARMDeploymentOutputs
	case models.TokenCredentialSurfaceLinkedDeploymentAsset:
		return models.TokenCredentialReviewARMDeploymentLinks
	default:
		return models.TokenCredentialReviewWorkloadContext
	}
}

func tokenCredentialNextReviewText(kind models.TokenCredentialNextReviewKind) string {
	switch kind {
	case models.TokenCredentialReviewEnvVarsSettingContext:
		return "Check env-vars for the exact setting context behind this credential clue."
	case models.TokenCredentialReviewEndpointsIngressAndControl:
		return "Check endpoints for the ingress path, then managed-identities and permissions for Azure control."
	case models.TokenCredentialReviewManagedIdentityAndPermissions:
		return "Check managed-identities for the identity path, then permissions for Azure control."
	case models.TokenCredentialReviewKeyVaultAndManagedIdentity:
		return "Check keyvault for the referenced secret boundary, then managed-identities for the backing workload identity."
	case models.TokenCredentialReviewKeyVaultBoundary:
		return "Check keyvault for the referenced secret boundary."
	case models.TokenCredentialReviewARMDeploymentOutputs:
		return "Check arm-deployments for the exact output context behind this credential clue."
	case models.TokenCredentialReviewARMDeploymentLinks:
		return "Check arm-deployments for the linked template or parameter path behind this credential clue."
	default:
		return "Review the surfaced workload context before deeper follow-up."
	}
}

func tokenCredentialSignalValue(signal string, key string) string {
	prefix := key + "="
	for _, part := range strings.Split(signal, ";") {
		value := strings.TrimSpace(part)
		if strings.HasPrefix(value, prefix) {
			candidate := strings.TrimSpace(strings.TrimPrefix(value, prefix))
			if candidate != "" {
				return candidate
			}
			return ""
		}
	}
	return ""
}

func partialCollectionIssue(scope string, message string, assetID string, assetName string) models.Issue {
	contextMap := map[string]string{"collector": scope}
	if assetID != "" {
		contextMap["asset_id"] = assetID
	}
	if assetName != "" {
		contextMap["asset_name"] = assetName
	}
	return models.Issue{
		Kind:    "partial_collection",
		Message: message,
		Scope:   scope,
		Context: contextMap,
	}
}

func resourceGroupAndNameFromID(resourceID string) (string, string) {
	return resourceGroupFromID(resourceID), resourceNameFromID(resourceID)
}

func resourceNameFromID(resourceID string) string {
	parts := armIDParts(resourceID)
	if len(parts) == 0 {
		return ""
	}
	return parts[len(parts)-1]
}

func subnetComponentsFromID(subnetID string) (string, string, string) {
	parts := armIDParts(subnetID)
	resourceGroup := ""
	virtualNetwork := ""
	subnet := ""
	for index := 0; index < len(parts)-1; index++ {
		switch {
		case strings.EqualFold(parts[index], "resourceGroups") && index+1 < len(parts):
			resourceGroup = parts[index+1]
		case strings.EqualFold(parts[index], "virtualNetworks") && index+1 < len(parts):
			virtualNetwork = parts[index+1]
		case strings.EqualFold(parts[index], "subnets") && index+1 < len(parts):
			subnet = parts[index+1]
		}
	}
	return resourceGroup, virtualNetwork, subnet
}

func vnetIDFromSubnetID(subnetID string) string {
	parts := armIDParts(subnetID)
	for index := 0; index < len(parts)-1; index++ {
		if strings.EqualFold(parts[index], "subnets") {
			return "/" + strings.Join(parts[:index], "/")
		}
	}
	return ""
}

func armIDParts(resourceID string) []string {
	if strings.TrimSpace(resourceID) == "" {
		return nil
	}
	parts := strings.Split(strings.Trim(resourceID, "/"), "/")
	filtered := make([]string, 0, len(parts))
	for _, part := range parts {
		if strings.TrimSpace(part) != "" {
			filtered = append(filtered, part)
		}
	}
	return filtered
}

func armIDJoinKey(resourceID string) string {
	return strings.ToLower(strings.TrimSpace(resourceID))
}

func sortedUniqueStrings(values []string) []string {
	deduped := dedupeStrings(values)
	sort.Strings(deduped)
	return deduped
}

func isPrivateNetworkPrefix(value string) bool {
	text := strings.TrimSpace(value)
	if text == "" {
		return false
	}
	if prefix, err := netip.ParsePrefix(text); err == nil {
		return prefix.Addr().IsPrivate()
	}
	if address, err := netip.ParseAddr(text); err == nil {
		return address.IsPrivate()
	}
	return false
}

func exposureRank(value string) int {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "high":
		return 0
	case "medium":
		return 1
	case "low":
		return 2
	default:
		return 9
	}
}

func yesNo(value bool) string {
	if value {
		return "yes"
	}
	return "no"
}
