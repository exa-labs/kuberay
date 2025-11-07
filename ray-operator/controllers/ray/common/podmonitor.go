package common

import (
	"context"

	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"

	rayv1 "github.com/ray-project/kuberay/ray-operator/apis/ray/v1"
	"github.com/ray-project/kuberay/ray-operator/controllers/ray/utils"
)

//
//

func BuildPodMonitorForRayCluster(ctx context.Context, cluster rayv1.RayCluster) ([]*monitoringv1.PodMonitor, error) {
	log := ctrl.LoggerFrom(ctx)
	log.Info("Building PodMonitor resources for RayCluster", "cluster", cluster.Name, "namespace", cluster.Namespace)

	podMonitors := []*monitoringv1.PodMonitor{}

	headPodMonitor := buildHeadPodMonitor(cluster)
	podMonitors = append(podMonitors, headPodMonitor)

	workerPodMonitor := buildWorkerPodMonitor(cluster)
	podMonitors = append(podMonitors, workerPodMonitor)

	return podMonitors, nil
}

func buildHeadPodMonitor(cluster rayv1.RayCluster) *monitoringv1.PodMonitor {
	labels := map[string]string{
		utils.RayClusterLabelKey:                cluster.Name,
		utils.RayNodeTypeLabelKey:               string(rayv1.HeadNode),
		utils.KubernetesApplicationNameLabelKey: utils.ApplicationName,
		utils.KubernetesCreatedByLabelKey:       utils.ComponentName,
	}

	selector := metav1.LabelSelector{
		MatchLabels: map[string]string{
			utils.RayClusterLabelKey:  cluster.Name,
			utils.RayNodeTypeLabelKey: string(rayv1.HeadNode),
		},
	}

	metricsPort := "metrics"
	asMetricsPort := "as-metrics"
	dashMetricsPort := "dash-metrics"

	endpoints := []monitoringv1.PodMetricsEndpoint{
		{
			Port: &metricsPort,
			RelabelConfigs: []monitoringv1.RelabelConfig{
				{
					Action:       "replace",
					SourceLabels: []monitoringv1.LabelName{"__meta_kubernetes_pod_label_ray_io_cluster"},
					TargetLabel:  "ray_io_cluster",
				},
			},
		},
		{
			Port: &asMetricsPort,
			RelabelConfigs: []monitoringv1.RelabelConfig{
				{
					Action:       "replace",
					SourceLabels: []monitoringv1.LabelName{"__meta_kubernetes_pod_label_ray_io_cluster"},
					TargetLabel:  "ray_io_cluster",
				},
			},
		},
		{
			Port: &dashMetricsPort,
			RelabelConfigs: []monitoringv1.RelabelConfig{
				{
					Action:       "replace",
					SourceLabels: []monitoringv1.LabelName{"__meta_kubernetes_pod_label_ray_io_cluster"},
					TargetLabel:  "ray_io_cluster",
				},
			},
		},
	}

	return &monitoringv1.PodMonitor{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cluster.Name + "-head-monitor",
			Namespace: cluster.Namespace,
			Labels:    labels,
		},
		Spec: monitoringv1.PodMonitorSpec{
			JobLabel: "ray-head",
			Selector: selector,
			NamespaceSelector: monitoringv1.NamespaceSelector{
				MatchNames: []string{cluster.Namespace},
			},
			PodMetricsEndpoints: endpoints,
		},
	}
}

func buildWorkerPodMonitor(cluster rayv1.RayCluster) *monitoringv1.PodMonitor {
	labels := map[string]string{
		utils.RayClusterLabelKey:                cluster.Name,
		utils.RayNodeTypeLabelKey:               string(rayv1.WorkerNode),
		utils.KubernetesApplicationNameLabelKey: utils.ApplicationName,
		utils.KubernetesCreatedByLabelKey:       utils.ComponentName,
	}

	selector := metav1.LabelSelector{
		MatchLabels: map[string]string{
			utils.RayClusterLabelKey:  cluster.Name,
			utils.RayNodeTypeLabelKey: string(rayv1.WorkerNode),
		},
	}

	metricsPort := "metrics"

	endpoints := []monitoringv1.PodMetricsEndpoint{
		{
			Port: &metricsPort,
			RelabelConfigs: []monitoringv1.RelabelConfig{
				{
					Action:       "replace",
					SourceLabels: []monitoringv1.LabelName{"__meta_kubernetes_pod_label_ray_io_cluster"},
					TargetLabel:  "ray_io_cluster",
				},
			},
		},
	}

	return &monitoringv1.PodMonitor{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cluster.Name + "-worker-monitor",
			Namespace: cluster.Namespace,
			Labels:    labels,
		},
		Spec: monitoringv1.PodMonitorSpec{
			JobLabel: "ray-workers",
			Selector: selector,
			NamespaceSelector: monitoringv1.NamespaceSelector{
				MatchNames: []string{cluster.Namespace},
			},
			PodMetricsEndpoints: endpoints,
		},
	}
}
