package multitrack

import (
	"fmt"
	"github.com/werf/kubedog/pkg/tracker/pod"
	"k8s.io/client-go/kubernetes"
)

func (mt *multitracker) TrackPod(kube kubernetes.Interface, spec MultitrackSpec, opts MultitrackOptions) error {
	feed := pod.NewFeed()

	feed.OnAdded(func() error {
		mt.mux.Lock()
		defer mt.mux.Unlock()

		mt.PodsStatuses[spec.ResourceName] = feed.GetStatus()

		return mt.podAdded(spec, feed)
	})
	feed.OnSucceeded(func() error {
		mt.mux.Lock()
		defer mt.mux.Unlock()

		mt.PodsStatuses[spec.ResourceName] = feed.GetStatus()

		return mt.podSucceeded(spec, feed)
	})
	feed.OnFailed(func(reason string) error {
		mt.mux.Lock()
		defer mt.mux.Unlock()

		mt.PodsStatuses[spec.ResourceName] = feed.GetStatus()

		return mt.podFailed(spec, feed, reason)
	})
	feed.OnReady(func() error {
		mt.mux.Lock()
		defer mt.mux.Unlock()

		mt.PodsStatuses[spec.ResourceName] = feed.GetStatus()

		return mt.podReady(spec, feed)
	})
	feed.OnEventMsg(func(msg string) error {
		mt.mux.Lock()
		defer mt.mux.Unlock()

		mt.PodsStatuses[spec.ResourceName] = feed.GetStatus()

		return mt.podEventMsg(spec, feed, msg)
	})
	//feed.OnContainerLogChunk(func(chunk *pod.ContainerLogChunk) error {
	//	mt.mux.Lock()
	//	defer mt.mux.Unlock()
	//	var podChunk *pod.PodLogChunk
	//	podChunk.PodName = spec.ResourceName
	//	podChunk.ContainerLogChunk = chunk
	//
	//	mt.PodsStatuses[spec.ResourceName] = feed.GetStatus()
	//
	//	return mt.podPodLogChunk(spec, feed, podChunk)
	//})
	//feed.OnContainerError(func(ContainerError pod.ContainerError) error {
	//	mt.mux.Lock()
	//	defer mt.mux.Unlock()
	//
	//	mt.PodsStatuses[spec.ResourceName] = feed.GetStatus()
	//
	//	return mt.podContainerError(spec, feed, ContainerError)
	//})
	feed.OnStatus(func(status pod.PodStatus) error {
		mt.mux.Lock()
		defer mt.mux.Unlock()

		mt.PodsStatuses[spec.ResourceName] = status

		return nil
	})

	return feed.Track(spec.ResourceName, spec.Namespace, kube, opts.Options)
}

func (mt *multitracker) podAdded(spec MultitrackSpec, feed pod.Feed) error {
	mt.displayResourceTrackerMessageF("pod", spec, "added")

	return nil
}

func (mt *multitracker) podSucceeded(spec MultitrackSpec, feed pod.Feed) error {
	mt.displayResourceTrackerMessageF("pod", spec, "succeeded")

	return mt.handleResourceReadyCondition(mt.TrackingPods, spec)
}

func (mt *multitracker) podFailed(spec MultitrackSpec, feed pod.Feed, reason string) error {
	mt.displayResourceErrorF("pod", spec, "%s", reason)
	return mt.handleResourceFailure(mt.TrackingPods, "pod", spec, reason)
}

func (mt *multitracker) podReady(spec MultitrackSpec, feed pod.Feed) error {
	mt.displayResourceTrackerMessageF("pod", spec, "ready")

	return mt.handleResourceReadyCondition(mt.TrackingPods, spec)
}

func (mt *multitracker) podEventMsg(spec MultitrackSpec, feed pod.Feed, msg string) error {
	mt.displayResourceEventF("pod", spec, "%s", msg)
	return nil
}

func (mt *multitracker) podPodLogChunk(spec MultitrackSpec, feed pod.Feed, chunk *pod.PodLogChunk) error {
	mt.displayResourceLogChunk("pod", spec, podContainerLogChunkHeader(chunk.PodName, chunk.ContainerLogChunk), chunk.ContainerLogChunk)
	return nil
}

func (mt *multitracker) podPodError(spec MultitrackSpec, feed pod.Feed, podError pod.PodError) error {
	reason := fmt.Sprintf("po/%s container/%s: %s", podError.PodName, podError.ContainerName, podError.Message)

	mt.displayResourceErrorF("pod", spec, "%s", reason)

	return mt.handleResourceFailure(mt.TrackingPods, "pod", spec, reason)
}
