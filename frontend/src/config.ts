import { createContext, use } from 'react';
import type { AppConfig } from './api';

/** The backend's configuration, provided by ConfigProvider once it has loaded. */
export const ConfigContext = createContext<AppConfig | null>(null);

/** Returns the backend's configuration. Only use it inside ConfigProvider. */
export function useConfig(): AppConfig {
  const config = use(ConfigContext);
  if (!config) throw new Error('useConfig must be used inside ConfigProvider');
  return config;
}
