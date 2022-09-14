package cli

import (
	"context"

	"github.com/funlessdev/fl-cli/pkg/deploy"
)

func NewDeployer(ctx context.Context) (deploy.DockerDeployer, error) {
	deployer := deploy.NewLocalDeployer("fl-core", "fl-worker", "fl_net", "fl_runtime_net")
	err := deployer.Setup(ctx)
	if err != nil {
		return nil, err
	}
	return deployer, nil
}

func DeployDev(ctx context.Context, deployer deploy.DockerDeployer) error {

	if err := deployer.CreateFLNetworks(ctx); err != nil {
		return err
	}

	if err := deployer.PullCoreImage(ctx); err != nil {
		return err
	}

	if err := deployer.PullWorkerImage(ctx); err != nil {
		return err
	}

	if err := deployer.StartCore(ctx); err != nil {
		return err
	}

	if err := deployer.StartWorker(ctx); err != nil {
		return err
	}

	return nil
}

func DestroyDev(ctx context.Context, deployer deploy.DockerDeployer) error {

	if err := deployer.RemoveWorkerContainer(ctx); err != nil {
		return err
	}

	if err := deployer.RemoveCoreContainer(ctx); err != nil {
		return err
	}

	if err := deployer.RemoveFunctionContainers(ctx); err != nil {
		return err
	}

	if err := deployer.RemoveFLNetworks(ctx); err != nil {
		return err
	}

	return nil
}
