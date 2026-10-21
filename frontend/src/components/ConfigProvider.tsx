import { useEffect, useState, type ReactNode } from 'react';
import { fetchConfig, type AppConfig } from '../api';
import { ConfigContext } from '../config';
import { formatError } from '../format';

interface Props {
  children: ReactNode;
}

/** Loads the backend's configuration and renders children once it is available. */
export default function ConfigProvider({ children }: Props) {
  const [config, setConfig] = useState<AppConfig | null>(null);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const controller = new AbortController();
    fetchConfig(controller.signal)
      .then(setConfig)
      .catch((e: unknown) => {
        if (!controller.signal.aborted) setError(formatError(e));
      });
    return () => controller.abort();
  }, []);

  if (error) return <p className="error app-message">{error}</p>;
  if (!config) return null;
  return <ConfigContext value={config}>{children}</ConfigContext>;
}
