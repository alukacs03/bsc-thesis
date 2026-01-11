import { apiClient } from './client';
import type { KubernetesClusterResponse, KubernetesWorkloadsResponse, ApplyManifestResponse, GetResourceYAMLResponse, KubernetesNetworkingResponse, DeleteResourceResponse } from '../types/kubernetes';

export const kubernetesAPI = {
  getCluster: () => apiClient.get<KubernetesClusterResponse>('/admin/kubernetes/cluster'),
  refreshJoinTokens: () => apiClient.post<{ message: string }>('/admin/kubernetes/refresh-join'),
  getWorkloads: () => apiClient.get<KubernetesWorkloadsResponse>('/admin/kubernetes/workloads'),
  getNetworking: () => apiClient.get<KubernetesNetworkingResponse>('/admin/kubernetes/networking'),
  applyManifest: (yaml: string) => apiClient.post<ApplyManifestResponse>('/admin/kubernetes/apply', { yaml }),
  getResourceYAML: (namespace: string, kind: string, name: string) =>
    apiClient.get<GetResourceYAMLResponse>(`/admin/kubernetes/resource?namespace=${encodeURIComponent(namespace)}&kind=${encodeURIComponent(kind)}&name=${encodeURIComponent(name)}`),
  deleteResource: (namespace: string, kind: string, name: string) =>
    apiClient.delete<DeleteResourceResponse>(`/admin/kubernetes/resource?namespace=${encodeURIComponent(namespace)}&kind=${encodeURIComponent(kind)}&name=${encodeURIComponent(name)}`),
};
