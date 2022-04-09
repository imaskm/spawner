package azure

import (
	"context"
	"fmt"
	"net/http"

	"github.com/Azure/azure-sdk-for-go/profiles/latest/containerservice/mgmt/containerservice"
	"github.com/Azure/go-autorest/autorest/to"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/pkg/errors"
	"gitlab.com/netbook-devs/spawner-service/pkg/service/constants"
	"gitlab.com/netbook-devs/spawner-service/pkg/service/labels"
	proto "gitlab.com/netbook-devs/spawner-service/proto/netbookdevs/spawnerservice"
)

const testpubkey = "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQDL67TCv+MyUnT0gHUl2xpJF56TjCkcTKkXUjhIaUDY/gv/bFm5pVbvrHovKV/W2MrI5e9Ix2iQIiityWVABFEFWe7m0yx3ds49ZkM3kIflsqmPeywCcN8V2bMsiVwyrLBsboeRcbQyJJIrsb8A0mj3ooWFfT44I42YVCg4FOTsB+wmlthawBlMGKzZb8ITUMaN0VCtXfIslg6ptQHtficL/N1HW7FSXXiZPJaRi3kuCH18e/wCkP4eomWMZ6MQC1CIwGIkfh9K4pfuppfZ9HG+jyw0ha0LZ6utDbEULMPAtvgUZXB7+1vk1NTwi78p558Dk6fxWGRVgSQu7Qk4yddZ nishanth@nishanth-Legion-5-15ACH6"

func (a *AzureController) createAKSCluster(ctx context.Context, req *proto.ClusterRequest) (*proto.ClusterResponse, error) {

	clusterName := req.ClusterName
	account := req.AccountName

	cred, err := getCredentials(ctx, account)
	if err != nil {
		return nil, err
	}

	groupName := cred.ResourceGroup
	region := req.Region

	clientID := cred.ClientID
	clientSecret := cred.ClientSecret

	aksClient, err := getAKSClient(cred)
	if err != nil {
		return nil, errors.Wrap(err, "creaetAKSCluster: cannot to get AKS client")
	}

	a.logger.Infow("creating cluster in AKS", "name", clusterName, "resource-group", groupName)
	tags := labels.DefaultTags()
	for k, v := range req.Labels {
		v := v
		tags[k] = &v

	}
	nodeTags := labels.GetNodeLabel(req.Node)

	//Doc : https://docs.microsoft.com/en-us/rest/api/aks/managed-clusters/create-or-update

	mc := containerservice.ManagedCluster{
		Tags:     tags,
		Name:     &clusterName,
		Location: &region,
		ManagedClusterProperties: &containerservice.ManagedClusterProperties{
			DNSPrefix: &clusterName,
			AgentPoolProfiles: &[]containerservice.ManagedClusterAgentPoolProfile{
				{
					Count:               to.Int32Ptr(1),
					Name:                to.StringPtr(req.Node.Name),
					VMSize:              to.StringPtr(req.Node.Instance),
					OsDiskSizeGB:        to.Int32Ptr(req.Node.DiskSize),
					NodeLabels:          nodeTags,
					Tags:                nodeTags,
					Mode:                containerservice.AgentPoolModeSystem,
					OrchestratorVersion: &constants.AzureKubeVersion,
				},
			},
			ServicePrincipalProfile: &containerservice.ManagedClusterServicePrincipalProfile{
				ClientID: to.StringPtr(clientID),
				Secret:   to.StringPtr(clientSecret),
			},
		},
	}

	future, err := aksClient.CreateOrUpdate(
		ctx,
		groupName,
		clusterName,
		mc,
	)
	if err != nil {
		a.logger.Errorw("failed to create a AKS cluster", "error", err)
		return nil, fmt.Errorf("cannot create AKS cluster: %v", err)
	}

	a.logger.Infow("waiting on the future completion")
	err = future.WaitForCompletionRef(ctx, aksClient.Client)
	if err != nil {
		a.logger.Errorw("failed to get the future response", "error", err)
		return nil, fmt.Errorf("cannot get the AKS cluster create or update future response: %v", err)
	}

	//return future.Result(aksClient)

	return &proto.ClusterResponse{ClusterName: clusterName}, nil
}

func (a *AzureController) getCluster(ctx context.Context, req *proto.GetClusterRequest) (*proto.ClusterSpec, error) {

	clusterName := req.ClusterName
	account := req.AccountName

	cred, err := getCredentials(ctx, account)
	if err != nil {
		return nil, err
	}

	groupName := cred.ResourceGroup
	aksClient, err := getAKSClient(cred)
	if err != nil {
		return nil, errors.Wrap(err, "creaetAKSCluster: cannot to get AKS client")
	}

	a.logger.Infow("fetching cluster information", "cluster", clusterName)
	//Doc : https://docs.microsoft.com/en-us/rest/api/aks/managed-clusters/get
	clstr, err := aksClient.Get(ctx, groupName, clusterName)
	if err != nil {
		a.logger.Errorw("failed to get cluster ", "error", err)
		return nil, err
	}

	node := (*clstr.AgentPoolProfiles)[0]

	state := constants.Inactive
	if clstr.PowerState.Code == containerservice.CodeRunning {
		state = constants.Active
	}

	return &proto.ClusterSpec{
		Name: clusterName,
		NodeSpec: []*proto.NodeSpec{{
			Name:     *node.Name,
			Instance: *node.VMSize,
			Labels:   aws.StringValueMap(node.NodeLabels),
			DiskSize: *node.OsDiskSizeGB,
			State:    state,
		}},
	}, nil
}

func (a *AzureController) deleteCluster(ctx context.Context, req *proto.ClusterDeleteRequest) (*proto.ClusterDeleteResponse, error) {

	clusterName := req.ClusterName

	account := req.AccountName

	cred, err := getCredentials(ctx, account)
	if err != nil {
		return nil, err
	}

	aksClient, err := getAKSClient(cred)
	if err != nil {
		return nil, errors.Wrap(err, "creaetAKSCluster: cannot to get AKS client")
	}

	groupName := cred.ResourceGroup
	//Doc : https://docs.microsoft.com/en-us/rest/api/aks/managed-clusters/delete
	future, err := aksClient.Delete(ctx, groupName, clusterName)

	if err != nil {
		a.logger.Errorw("failed to delete the cluster ", "error", err, "cluster", clusterName)
		return nil, err
	}

	err = future.WaitForCompletionRef(ctx, aksClient.Client)
	if err != nil {
		a.logger.Errorw("failed to get the future response", "error", err)
		return nil, fmt.Errorf("cannot get the AKS cluster create or update future response: %v", err)
	}

	if future.Response().StatusCode == http.StatusNoContent {
		return nil, fmt.Errorf("request resource '%s' not found", clusterName)
	}

	a.logger.Infow("cluster deleted successfully", "cluster", clusterName, "response", future.Status())

	return &proto.ClusterDeleteResponse{}, nil

}

func (a *AzureController) getClusters(ctx context.Context, req *proto.GetClustersRequest) (*proto.GetClustersResponse, error) {

	account := req.AccountName

	cred, err := getCredentials(ctx, account)
	if err != nil {
		return nil, err
	}
	aksClient, err := getAKSClient(cred)
	if err != nil {
		return nil, errors.Wrap(err, "creaetAKSCluster: cannot to get AKS client")
	}

	//Doc : https://docs.microsoft.com/en-us/rest/api/aks/managed-clusters/list
	result, err := aksClient.List(ctx)

	if err != nil {
		a.logger.Errorw("failed to list the cluster ", "error", err)
		return nil, err
	}

	clusters := make([]*proto.ClusterSpec, 0, len(result.Values()))
	for _, cl := range result.Values() {
		mcapp := cl.AgentPoolProfiles
		nodes := make([]*proto.NodeSpec, 0, len(*mcapp))

		for _, app := range *mcapp {
			state := constants.Inactive
			if app.PowerState.Code == containerservice.CodeRunning {
				state = constants.Active
			}
			zones := ""
			if app.AvailabilityZones != nil {
				zones = (*app.AvailabilityZones)[0]
			}
			node := &proto.NodeSpec{
				Name:             *app.Name,
				Instance:         *app.VMSize,
				DiskSize:         *app.OsDiskSizeGB,
				State:            state,
				IpAddr:           "",
				Availabilityzone: zones,
				ClusterId:        *cl.ID,
				Labels:           aws.StringValueMap(app.Tags),
				GpuEnabled:       false,
				//TODO: get health
				Health: &proto.Health{},
			}
			nodes = append(nodes, node)
		}

		spec := &proto.ClusterSpec{
			Name:      *cl.Name,
			ClusterId: *cl.ID,
			NodeSpec:  nodes,
		}
		clusters = append(clusters, spec)
	}

	return &proto.GetClustersResponse{
		Clusters: clusters}, nil
}

func (a *AzureController) clusterStatus(ctx context.Context, req *proto.ClusterStatusRequest) (*proto.ClusterStatusResponse, error) {
	clusterName := req.GetClusterName()

	account := req.AccountName

	cred, err := getCredentials(ctx, account)
	if err != nil {
		return nil, err
	}
	groupName := cred.ResourceGroup
	aksClient, err := getAKSClient(cred)
	if err != nil {
		return nil, errors.Wrap(err, "creaetAKSCluster: cannot to get AKS client")
	}

	a.logger.Infow("fetching cluster information", "cluster", clusterName)
	//Doc : https://docs.microsoft.com/en-us/rest/api/aks/managed-clusters/get
	clstr, err := aksClient.Get(ctx, groupName, clusterName)
	if err != nil {
		a.logger.Errorw("failed to get cluster information", "error", err)
		return nil, err
	}

	state := constants.Inactive
	if clstr.PowerState.Code == containerservice.CodeRunning {
		state = constants.Active
	}
	return &proto.ClusterStatusResponse{
		Status: state,
	}, nil
}