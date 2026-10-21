/** The message to show for a caught error. */
export function formatError(e: unknown): string {
  return e instanceof Error ? e.message : String(e);
}

export function formatDistance(m: number): string {
  return m >= 1000 ? `${(m / 1000).toFixed(1)} km` : `${Math.round(m)} m`;
}

export function formatDuration(s: number | null): string {
  if (s == null) return '';
  const h = Math.floor(s / 3600);
  const m = Math.floor((s % 3600) / 60);
  return h > 0 ? `${h} h ${m} min` : `${m} min`;
}

export function formatDate(value: string | number | null): string {
  if (value == null) return 'Unknown date';
  const d = typeof value === 'number' ? new Date(value * 1000) : new Date(value);
  return d.toLocaleDateString(undefined, { year: 'numeric', month: 'short', day: 'numeric' });
}

export function formatDateTime(value: string | null): string {
  if (value == null) return 'Unknown time';
  return new Date(value).toLocaleString(undefined, {
    year: 'numeric',
    month: 'short',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit',
  });
}

/** Turns a sport id like "virtual_ride" into "Virtual ride". */
export function formatSport(sport: string): string {
  const words = sport.replace(/_/g, ' ');
  return words.charAt(0).toUpperCase() + words.slice(1);
}

export function formatPercent(fraction: number): string {
  return `${Number((fraction * 100).toFixed(1))}%`;
}

/** Line colors per sport, shared by the map and the activity list. */
export const sportColors: Record<string, string> = {
  running: '#e4572e',
  cycling: '#2e86ab',
  walking: '#76b041',
  hiking: '#a1662f',
  swimming: '#17bebb',
};
export const defaultSportColor = '#7d5ba6';

export function sportColor(sport: string): string {
  return sportColors[sport] ?? defaultSportColor;
}
