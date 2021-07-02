package runtime

import (
	"fmt"
	"net"
	"strings"
	"time"

	corev1 "k8s.io/api/core/v1"
)

const portScanTimeout = 100 * time.Millisecond

func (rt *DaprRuntime) ProbeApplicationAvailability() (bool, error) {
	if rt.runtimeConfig.Pod != nil && rt.runtimeConfig.AppContainer != nil {
		podStatus := rt.runtimeConfig.Pod.Status

		appContainerState := getContainerStatusByName(&podStatus, rt.runtimeConfig.AppContainer.Name)
		if appContainerState == nil {
			return false, fmt.Errorf("cannot find container %v in pod %v", rt.runtimeConfig.AppContainer.Name, rt.runtimeConfig.Pod.Name)
		} else if appContainerState.Running == nil {
			return false, nil
		}
	}

	if rt.runtimeConfig.ApplicationProbingPort != 0 {
		return scanLocalPort(rt.runtimeConfig.ApplicationProbingPort, portScanTimeout), nil
	} else {
		return true, nil
	}
}

func getContainerStatusByName(podStatus *corev1.PodStatus, containerName string) *corev1.ContainerState {
	for _, status := range podStatus.ContainerStatuses {
		if status.Name == containerName {
			return &status.State
		}
	}
	return nil
}

func scanLocalPort(port int, timeout time.Duration) bool {
	target := fmt.Sprintf("%s:%d", "127.0.0.1", port)
	conn, err := net.DialTimeout("tcp", target, timeout)

	if err != nil {
		if strings.Contains(err.Error(), "too many open files") {
			time.Sleep(timeout)
			return scanLocalPort(port, timeout)
		}
		return false
	}

	conn.Close()
	return true
}
