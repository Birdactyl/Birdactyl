import { api } from './client';
import type { Node } from './admin';

export interface PackagePort { name: string; default: number; protocol: string; primary?: boolean; }
export interface PackageVariable { name: string; description: string; default: string; user_editable: boolean; rules?: string; }
export interface PackageConfigFile { path: string; template: string; }

export interface AddonSourceMapping {
  results?: string;
  id?: string;
  name?: string;
  description?: string;
  icon?: string;
  author?: string;
  downloads?: string;
  version_id?: string;
  version_name?: string;
  download_url?: string;
  file_name?: string;
}

export interface AddonSource {
  id: string;
  name: string;
  icon?: string;
  type?: string;
  search_url: string;
  versions_url?: string;
  download_url?: string;
  install_path: string;
  file_filter?: string;
  headers?: Record<string, string>;
  mapping: AddonSourceMapping;
}

export interface Package {
  id: string; name: string; version: string; author: string; description: string; icon?: string;
  docker_image: string; install_image: string; startup: string; install_script: string;
  stop_signal: string; stop_command: string; stop_timeout: number;
  startup_editable: boolean; docker_image_editable: boolean;
  ports: PackagePort[]; variables: PackageVariable[]; config_files: PackageConfigFile[];
  addon_sources?: AddonSource[];
  created_at: string; updated_at: string;
}

export const getAvailableNodes = () => api.get<Node[]>('/nodes');
export const getAvailablePackages = () => api.get<Package[]>('/packages');
