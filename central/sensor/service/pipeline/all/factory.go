package all

import (
	"context"

	"github.com/stackrox/rox/central/sensor/service/pipeline"
	"github.com/stackrox/rox/central/sensor/service/pipeline/alerts"
	"github.com/stackrox/rox/central/sensor/service/pipeline/auditlogstateupdate"
	"github.com/stackrox/rox/central/sensor/service/pipeline/clusterhealthupdate"
	"github.com/stackrox/rox/central/sensor/service/pipeline/clustermetrics"
	"github.com/stackrox/rox/central/sensor/service/pipeline/clusterstatusupdate"
	"github.com/stackrox/rox/central/sensor/service/pipeline/complianceoperatorprofiles"
	"github.com/stackrox/rox/central/sensor/service/pipeline/complianceoperatorresults"
	"github.com/stackrox/rox/central/sensor/service/pipeline/complianceoperatorrules"
	"github.com/stackrox/rox/central/sensor/service/pipeline/complianceoperatorscans"
	"github.com/stackrox/rox/central/sensor/service/pipeline/complianceoperatorscansettingbinding"
	"github.com/stackrox/rox/central/sensor/service/pipeline/deploymentevents"
	"github.com/stackrox/rox/central/sensor/service/pipeline/imageintegrations"
	"github.com/stackrox/rox/central/sensor/service/pipeline/namespaces"
	"github.com/stackrox/rox/central/sensor/service/pipeline/networkflowupdate"
	"github.com/stackrox/rox/central/sensor/service/pipeline/networkpolicies"
	"github.com/stackrox/rox/central/sensor/service/pipeline/nodeinventory"
	"github.com/stackrox/rox/central/sensor/service/pipeline/nodes"
	"github.com/stackrox/rox/central/sensor/service/pipeline/podevents"
	"github.com/stackrox/rox/central/sensor/service/pipeline/processindicators"
	"github.com/stackrox/rox/central/sensor/service/pipeline/processlisteningonport"
	"github.com/stackrox/rox/central/sensor/service/pipeline/reprocessing"
	"github.com/stackrox/rox/central/sensor/service/pipeline/rolebindings"
	"github.com/stackrox/rox/central/sensor/service/pipeline/roles"
	"github.com/stackrox/rox/central/sensor/service/pipeline/secrets"
	"github.com/stackrox/rox/central/sensor/service/pipeline/serviceaccounts"
	"github.com/stackrox/rox/pkg/features"
)

// NewFactory returns a new instance of a Factory that produces a pipeline handling all message types.
func NewFactory() pipeline.Factory {
	return &factoryImpl{}
}

type factoryImpl struct{}

// PipelineForCluster grabs items from the queue, processes them, and potentially sends them back to sensor.
func (s *factoryImpl) PipelineForCluster(ctx context.Context, clusterID string) (pipeline.ClusterPipeline, error) {
	flowUpdateFragment, err := networkflowupdate.Singleton().GetFragment(ctx, clusterID)
	if err != nil {
		return nil, err
	}

	pipelines := []pipeline.Fragment{
		deploymentevents.GetPipeline(),
		podevents.GetPipeline(),
		processindicators.GetPipeline(),
		processlisteningonport.GetPipeline(),
		networkpolicies.GetPipeline(),
		namespaces.GetPipeline(),
		secrets.GetPipeline(),
		nodes.GetPipeline(),
		flowUpdateFragment,
		imageintegrations.GetPipeline(),
		clusterstatusupdate.GetPipeline(),
		clusterhealthupdate.GetPipeline(),
		clustermetrics.GetPipeline(),
		serviceaccounts.GetPipeline(),
		roles.GetPipeline(),
		rolebindings.GetPipeline(),
		reprocessing.GetPipeline(),
		alerts.GetPipeline(),
		auditlogstateupdate.GetPipeline(),
	}
	if features.RHCOSNodeScanning.Enabled() {
		pipelines = append(pipelines, nodeinventory.GetPipeline())
	}
	if features.ComplianceOperatorCheckResults.Enabled() {
		pipelines = append(pipelines,
			complianceoperatorresults.GetPipeline(),
			complianceoperatorprofiles.GetPipeline(),
			complianceoperatorscansettingbinding.GetPipeline(),
			complianceoperatorrules.GetPipeline(),
			complianceoperatorscans.GetPipeline(),
		)
	}

	return NewClusterPipeline(clusterID, pipelines...), nil
}
