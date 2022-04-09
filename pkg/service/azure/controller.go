package azure

import (
	"context"

	proto "gitlab.com/netbook-devs/spawner-service/proto/netbookdevs/spawnerservice"
	"go.uber.org/zap"
)

type AzureController struct {
	logger *zap.SugaredLogger
}

func NewController(logger *zap.SugaredLogger) *AzureController {
	return &AzureController{
		logger: logger,
	}
}

func (a *AzureController) CreateCluster(ctx context.Context, req *proto.ClusterRequest) (*proto.ClusterResponse, error) {
	return a.createAKSCluster(ctx, req)
}

func (a *AzureController) GetCluster(ctx context.Context, req *proto.GetClusterRequest) (*proto.ClusterSpec, error) {
	return a.getCluster(ctx, req)
}

func (a *AzureController) GetClusters(ctx context.Context, req *proto.GetClustersRequest) (*proto.GetClustersResponse, error) {
	return a.getClusters(ctx, req)
}

func (a *AzureController) ClusterStatus(ctx context.Context, req *proto.ClusterStatusRequest) (*proto.ClusterStatusResponse, error) {
	return a.clusterStatus(ctx, req)
}

func (a *AzureController) AddNode(ctx context.Context, req *proto.NodeSpawnRequest) (*proto.NodeSpawnResponse, error) {
	return a.addNode(ctx, req)
}

func (a *AzureController) DeleteCluster(ctx context.Context, req *proto.ClusterDeleteRequest) (*proto.ClusterDeleteResponse, error) {
	return a.deleteCluster(ctx, req)
}

func (a *AzureController) DeleteNode(ctx context.Context, req *proto.NodeDeleteRequest) (*proto.NodeDeleteResponse, error) {
	return a.deleteNode(ctx, req)
}

func (a *AzureController) AddToken(ctx context.Context, req *proto.AddTokenRequest) (*proto.AddTokenResponse, error) {
	panic("not implemented") // TODO: Implement
}

func (a *AzureController) GetToken(ctx context.Context, req *proto.GetTokenRequest) (*proto.GetTokenResponse, error) {
	return a.getToken(ctx, req)
}

func (a *AzureController) AddRoute53Record(ctx context.Context, req *proto.AddRoute53RecordRequest) (*proto.AddRoute53RecordResponse, error) {
	panic("not implemented") // TODO: Implement
}

func (a *AzureController) CreateVolume(ctx context.Context, req *proto.CreateVolumeRequest) (*proto.CreateVolumeResponse, error) {
	return a.createVolume(ctx, req)
}

func (a *AzureController) DeleteVolume(ctx context.Context, req *proto.DeleteVolumeRequest) (*proto.DeleteVolumeResponse, error) {
	return a.deleteVolume(ctx, req)
}

func (a *AzureController) CreateSnapshot(ctx context.Context, req *proto.CreateSnapshotRequest) (*proto.CreateSnapshotResponse, error) {
	return a.createSnapshot(ctx, req)
}

func (a *AzureController) CreateSnapshotAndDelete(ctx context.Context, req *proto.CreateSnapshotAndDeleteRequest) (*proto.CreateSnapshotAndDeleteResponse, error) {
	return a.createSnapshotAndDelete(ctx, req)
}

func (a *AzureController) GetWorkspacesCost(_ context.Context, _ *proto.GetWorkspacesCostRequest) (*proto.GetWorkspacesCostResponse, error) {
	panic("not implemented") // TODO: Implement
}