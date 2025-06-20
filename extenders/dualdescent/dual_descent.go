package dualdescent_extender

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/gemblerz/k8s-scheduler-plugins/plugins/dualdescent"
	v1 "k8s.io/api/core/v1"
	"k8s.io/klog"
	"k8s.io/kubernetes/pkg/scheduler/framework"
	"sigs.k8s.io/kube-scheduler-simulator/simulator/scheduler/plugin"
)

// DualDescentExtender implements the Kubernetes scheduler extender interface.
type dualdescentPreScorePluginExtender struct {
	handle plugin.SimulatorHandle
}

// BeforePreScore implements the plugin.PreScorePluginExtender interface.
func (e *dualdescentPreScorePluginExtender) BeforePreScore(ctx context.Context, state *framework.CycleState, pod *v1.Pod, nodes []*framework.NodeInfo) *framework.Status {
	// Add any logic needed after PreScore phase here.
	return nil
}

// AfterPreScore implements the plugin.PreScorePluginExtender interface.
func (e *dualdescentPreScorePluginExtender) AfterPreScore(ctx context.Context, state *framework.CycleState, pod *v1.Pod, nodes []*framework.NodeInfo, preScoreStatus *framework.Status) *framework.Status {
	klog.Info(fmt.Sprintf("execute AfterPreScore on dualdescentPreScorePluginExtender %+v", pod))

	c, err := state.Read("PreScore" + dualdescent.Name)
	if err != nil {
		// klog.Info("no state data", "pod", klog.KObj(pod))
		return preScoreStatus
	}

	j, err := json.Marshal(c)
	if err != nil {
		// klog.Info("json marshal failed in extender", "pod", klog.KObj(pod))
		return preScoreStatus
	}

	preScoreData := string(j)

	// store data via plugin.SimulatorHandle
	e.handle.AddCustomResult(pod.Namespace, pod.Name, "dualdescent-prescore-data", preScoreData)
	return preScoreStatus
}

// NewDualDescentExtender creates a new DualDescentExtender.
func New(handle plugin.SimulatorHandle) plugin.PluginExtenders {
	e := &dualdescentPreScorePluginExtender{
		handle: handle,
	}
	return plugin.PluginExtenders{PreScorePluginExtender: e}
}
