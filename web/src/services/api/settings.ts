import { apiClient } from './client';
import type { DeploymentSettings, DeploymentSettingsUpdate } from '../types/settings';

export const settingsAPI = {
  getDeploymentSettings: () => apiClient.get<DeploymentSettings>('/admin/deployment/settings'),
  updateDeploymentSettings: (settings: DeploymentSettingsUpdate) =>
    apiClient.put<DeploymentSettings>('/admin/deployment/settings', settings),
};
