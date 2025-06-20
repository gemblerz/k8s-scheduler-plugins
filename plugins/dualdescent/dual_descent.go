package dualdescent

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"time"

	"golang.org/x/xerrors"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/klog/v2"
	"k8s.io/kubernetes/pkg/scheduler/framework"
)

type DualdescentOptimizer struct {
}

var (
	_ framework.ScorePlugin    = &DualdescentOptimizer{}
	_ framework.PreScorePlugin = &DualdescentOptimizer{}
)

const (
	// Name is the name of the plugin used in the plugin registry and configurations.
	Name             = "DualDescentOptimizer"
	preScoreStateKey = "PreScore" + Name
)

// New initializes a new plugin and returns it.
func New(ctx context.Context, arg runtime.Object, h framework.Handle) (framework.Plugin, error) {
	klog.Info("initialize DualDescentOptimizer plugin")
	return &DualdescentOptimizer{}, nil
}

// Name returns the name of the plugin. It is used in logs, etc.
func (*DualdescentOptimizer) Name() string {
	return Name
}

// preScoreState computed at PreScore and used at Score.
type preScoreState struct {
	OptimizedScores map[string]int64
}

// Clone implements the mandatory Clone interface. We don't really copy the data since
// there is no need for that.
func (s *preScoreState) Clone() framework.StateData {
	return s
}

func (opt *DualdescentOptimizer) PreScore(ctx context.Context, state *framework.CycleState, pod *v1.Pod, nodes []*framework.NodeInfo) *framework.Status {
	klog.Info(fmt.Sprintf("execute PreScore on DualDescentOptimizer plugin %+v", pod))

	// Cost value of the Pod.
	var totalRequestedMemory int64
	for _, container := range pod.Spec.Containers {
		totalRequestedMemory += container.Resources.Requests.Memory().Value()
	}
	nodeConstraints := make(map[string]int64, len(nodes))

	// Calculate the available memory for each node.
	for _, n := range nodes {
		nodeConstraints[n.Node().Name] = opt.calculateAvailableMemory(n)
	}

	// Run the optimizer for the Pod.
	scores := opt.optimize(totalRequestedMemory, nodeConstraints)

	s := &preScoreState{
		OptimizedScores: scores,
	}
	state.Write(preScoreStateKey, s)

	return nil
}

func (*DualdescentOptimizer) EventsToRegister() []framework.ClusterEvent {
	return []framework.ClusterEvent{
		{Resource: framework.Node, ActionType: framework.Add},
	}
}

var ErrNotExpectedPreScoreState = errors.New("unexpected pre score state")

// Score invoked at the score extension point.
func (*DualdescentOptimizer) Score(ctx context.Context, state *framework.CycleState, pod *v1.Pod, nodeName string) (int64, *framework.Status) {
	klog.Info(fmt.Sprintf("execute Score on NodeNumber plugin %+v", pod))
	data, err := state.Read(preScoreStateKey)
	if err != nil {
		// return success even if there is no value in preScoreStateKey, since the
		// suffix of pod name maybe non-number.
		return 0, nil
	}

	s, ok := data.(*preScoreState)
	if !ok {
		err = xerrors.Errorf("fetched pre score state is not *preScoreState, but %T, %w", data, ErrNotExpectedPreScoreState)
		return 0, framework.AsStatus(err)
	}

	if score, ok := s.OptimizedScores[nodeName]; ok {
		return score, framework.NewStatus(framework.Success, fmt.Sprintf("Here is the score for the node: %d", score))
	} else {
		return 0, framework.NewStatus(framework.Error, fmt.Sprintf("Node %s not found in pre score state", nodeName))
	}
}

// ScoreExtensions of the Score plugin.
func (*DualdescentOptimizer) ScoreExtensions() framework.ScoreExtensions {
	return nil
}

// optimize optimizes the primal solution for the nodes.
func (opt *DualdescentOptimizer) optimize(podCost int64, nodeConstraints map[string]int64) (scores map[string]int64) {
	scores = make(map[string]int64, len(nodeConstraints))
	// TODO: We need to implement the optimization algorithm here:
	//   calculating the primal solution and sharing the lambda value
	//   with the nodes.

	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	for nodeName := range nodeConstraints {
		scores[nodeName] = rng.Int63n(11) // Random number between 0 and 10
	}
	return
}

// calculateAvailableMemory returns the available memory for the node.
func (*DualdescentOptimizer) calculateAvailableMemory(node *framework.NodeInfo) int64 {
	return node.Allocatable.Memory - node.Requested.Memory
}
