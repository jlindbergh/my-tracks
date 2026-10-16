export interface Activity {
  id: string;
  name: string;
  sport: string;
  startedAt: string | null;
  durationS: number | null;
  distanceM: number;
  /** [west, south, east, north] in WGS 84. */
  bbox: [number, number, number, number];
}

export interface ImportResult {
  file: string;
  status: 'imported' | 'duplicate' | 'failed';
  activityId?: string;
  error?: string;
}

/** Absolute URL template for the signed-in user's activity tiles. */
export const tilesUrl = `${window.location.origin}/api/tiles/{z}/{x}/{y}`;

export async function fetchActivities(signal?: AbortSignal): Promise<Activity[]> {
  const res = await fetch('/api/activities', { signal });
  if (!res.ok) throw new Error(`Loading activities failed (${res.status})`);
  return res.json();
}

export async function importFiles(files: File[]): Promise<ImportResult[]> {
  const body = new FormData();
  for (const f of files) body.append('files', f, f.name);
  const res = await fetch('/api/activities/import', { method: 'POST', body });
  if (!res.ok) throw new Error(`Import failed (${res.status})`);
  return res.json();
}
