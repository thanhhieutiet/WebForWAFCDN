import type { RoutingHostConfig, RoutingMode } from '@/types';

const HOST_CONFIG_KEY = 'labshop_host_config';

export const DEFAULT_HOST_CONFIG: RoutingHostConfig = {
  direct: 'localhost:8080',
  waap: 'waap.labshop.local',
  cdn: 'cdn.labshop.local',
  full: 'shop.labshop.local',
};

export function getHostConfig(): RoutingHostConfig {
  try {
    const raw = localStorage.getItem(HOST_CONFIG_KEY);
    if (raw) {
      return { ...DEFAULT_HOST_CONFIG, ...JSON.parse(raw) };
    }
  } catch (e) {
    console.warn('Failed to load host config from localStorage:', e);
  }
  return DEFAULT_HOST_CONFIG;
}

export function saveHostConfig(config: RoutingHostConfig): void {
  try {
    localStorage.setItem(HOST_CONFIG_KEY, JSON.stringify(config));
    window.dispatchEvent(new Event('labshop_host_config_changed'));
  } catch (e) {
    console.warn('Failed to save host config to localStorage:', e);
  }
}

export function detectCurrentRouting(config = getHostConfig()): {
  host: string;
  routing: RoutingMode;
  label: string;
  isUnknown: boolean;
} {
  const currentHost = typeof window !== 'undefined' ? window.location.host : 'localhost:5173';
  const cleanCurrent = currentHost.toLowerCase().split(':')[0] || '';

  const cleanDirect = config.direct.toLowerCase().split(':')[0] || '';
  const cleanWaap = config.waap.toLowerCase().split(':')[0] || '';
  const cleanCdn = config.cdn.toLowerCase().split(':')[0] || '';
  const cleanFull = config.full.toLowerCase().split(':')[0] || '';

  if (currentHost === config.direct || cleanCurrent === cleanDirect || currentHost.includes('5173') || currentHost.includes('8080')) {
    return { host: currentHost, routing: 'direct', label: 'Direct Origin (Trực tiếp EC2)', isUnknown: false };
  }
  if (currentHost === config.waap || cleanCurrent === cleanWaap) {
    return { host: currentHost, routing: 'waap', label: 'WAAP Only (Client -> WAAP -> Origin)', isUnknown: false };
  }
  if (currentHost === config.cdn || cleanCurrent === cleanCdn) {
    return { host: currentHost, routing: 'cdn', label: 'CDN Only (Client -> CDN -> Origin)', isUnknown: false };
  }
  if (currentHost === config.full || cleanCurrent === cleanFull) {
    return { host: currentHost, routing: 'full', label: 'Full Topology (Client -> CDN -> WAAP -> Origin)', isUnknown: false };
  }

  return { host: currentHost, routing: 'unknown', label: 'Tùy chỉnh / Chưa xác định', isUnknown: true };
}
