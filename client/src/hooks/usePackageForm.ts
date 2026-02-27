import { useState, useCallback } from 'react';
import { Package, PackagePort, PackageVariable, AddonSource } from '../lib/api';

interface PackageFormData {
  name: string;
  version: string;
  author: string;
  description: string;
  icon: string;
  dockerImage: string;
  installImage: string;
  startup: string;
  installScript: string;
  stopSignal: string;
  stopCommand: string;
  stopTimeout: string;
  startupEditable: boolean;
  dockerImageEditable: boolean;
  ports: PackagePort[];
  variables: PackageVariable[];
  addonSources: AddonSource[];
}

const defaultData: PackageFormData = {
  name: '',
  version: '',
  author: '',
  description: '',
  icon: '',
  dockerImage: '',
  installImage: '',
  startup: '',
  installScript: '',
  stopSignal: 'SIGTERM',
  stopCommand: '',
  stopTimeout: '30',
  startupEditable: false,
  dockerImageEditable: false,
  ports: [],
  variables: [],
  addonSources: [],
};

export function usePackageForm(editPackage?: Package | null) {
  const [data, setData] = useState<PackageFormData>(() => {
    if (editPackage) {
      return {
        name: editPackage.name || '',
        version: editPackage.version || '',
        author: editPackage.author || '',
        description: editPackage.description || '',
        icon: editPackage.icon || '',
        dockerImage: editPackage.docker_image || '',
        installImage: editPackage.install_image || '',
        startup: editPackage.startup || '',
        installScript: editPackage.install_script || '',
        stopSignal: editPackage.stop_signal || 'SIGTERM',
        stopCommand: editPackage.stop_command || '',
        stopTimeout: String(editPackage.stop_timeout || 30),
        startupEditable: editPackage.startup_editable || false,
        dockerImageEditable: editPackage.docker_image_editable || false,
        ports: editPackage.ports || [],
        variables: editPackage.variables || [],
        addonSources: editPackage.addon_sources || [],
      };
    }
    return defaultData;
  });

  const reset = useCallback((pkg?: Package | null) => {
    if (pkg) {
      setData({
        name: pkg.name || '',
        version: pkg.version || '',
        author: pkg.author || '',
        description: pkg.description || '',
        icon: pkg.icon || '',
        dockerImage: pkg.docker_image || '',
        installImage: pkg.install_image || '',
        startup: pkg.startup || '',
        installScript: pkg.install_script || '',
        stopSignal: pkg.stop_signal || 'SIGTERM',
        stopCommand: pkg.stop_command || '',
        stopTimeout: String(pkg.stop_timeout || 30),
        startupEditable: pkg.startup_editable || false,
        dockerImageEditable: pkg.docker_image_editable || false,
        ports: pkg.ports || [],
        variables: pkg.variables || [],
        addonSources: pkg.addon_sources || [],
      });
    } else {
      setData(defaultData);
    }
  }, []);

  const update = (field: keyof PackageFormData, value: any) => {
    setData(prev => ({ ...prev, [field]: value }));
  };

  const toJson = () => JSON.stringify({
    name: data.name,
    version: data.version,
    author: data.author,
    description: data.description,
    icon: data.icon,
    docker_image: data.dockerImage,
    install_image: data.installImage,
    startup: data.startup,
    install_script: data.installScript,
    stop_signal: data.stopSignal,
    stop_command: data.stopCommand,
    stop_timeout: parseInt(data.stopTimeout) || 30,
    startup_editable: data.startupEditable,
    docker_image_editable: data.dockerImageEditable,
    ports: data.ports,
    variables: data.variables,
    addon_sources: data.addonSources,
  }, null, 2);

  const fromJson = (json: string) => {
    try {
      const pkg = JSON.parse(json);
      setData({
        name: pkg.name || '',
        version: pkg.version || '',
        author: pkg.author || '',
        description: pkg.description || '',
        icon: pkg.icon || '',
        dockerImage: pkg.docker_image || '',
        installImage: pkg.install_image || '',
        startup: pkg.startup || '',
        installScript: pkg.install_script || '',
        stopSignal: pkg.stop_signal || 'SIGTERM',
        stopCommand: pkg.stop_command || '',
        stopTimeout: String(pkg.stop_timeout || 30),
        startupEditable: pkg.startup_editable || false,
        dockerImageEditable: pkg.docker_image_editable || false,
        ports: pkg.ports || [],
        variables: pkg.variables || [],
        addonSources: pkg.addon_sources || [],
      });
    } catch { }
  };

  const toApiData = () => ({
    name: data.name,
    version: data.version,
    author: data.author,
    description: data.description,
    icon: data.icon,
    docker_image: data.dockerImage,
    install_image: data.installImage,
    startup: data.startup,
    install_script: data.installScript,
    stop_signal: data.stopSignal,
    stop_command: data.stopCommand,
    stop_timeout: parseInt(data.stopTimeout) || 30,
    startup_editable: data.startupEditable,
    docker_image_editable: data.dockerImageEditable,
    ports: data.ports,
    variables: data.variables,
    config_files: [],
    addon_sources: data.addonSources,
  });

  return { data, update, toJson, fromJson, toApiData, reset };
}
