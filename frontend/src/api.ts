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

/** The values used to tell activities apart. */
export interface ActivityFacts {
  id?: string;
  name: string;
  startedAt: string | null;
  durationS: number | null;
  distanceM: number;
}

export interface ImportResult {
  file: string;
  status: 'imported' | 'already_imported' | 'possible_duplicate' | 'failed';
  activityId?: string;
  error?: string;
  /** The activity read from the file. */
  activity?: ActivityFacts;
  /** The already imported activity it matched. */
  existing?: ActivityFacts;
}

/** How the backend detects duplicates; see backend/internal/importer/duplicates.go. */
export interface DuplicateRules {
  startToleranceS: number;
  /** Fraction, e.g. 0.02 for 2%. */
  relativeTolerance: number;
}

/** Absolute URL template for the signed-in user's activity tiles. */
export const tilesUrl = `${window.location.origin}/api/tiles/{z}/{x}/{y}`;

export async function fetchActivities(signal?: AbortSignal): Promise<Activity[]> {
  const res = await fetch('/api/activities', { signal });
  if (!res.ok) throw new Error(`Loading activities failed (${res.status})`);
  return res.json();
}

export async function fetchDuplicateRules(signal?: AbortSignal): Promise<DuplicateRules> {
  const res = await fetch('/api/import/rules', { signal });
  if (!res.ok) throw new Error(`Loading import rules failed (${res.status})`);
  return res.json();
}

/**
 * Uploads files for import. Results are returned in the same order as files.
 * With allowDuplicates, possible duplicates are imported instead of reported.
 */
export async function importFiles(files: File[], allowDuplicates = false): Promise<ImportResult[]> {
  const body = new FormData();
  for (const f of files) body.append('files', f, f.name);
  const url = allowDuplicates ? '/api/activities/import?allowDuplicates=true' : '/api/activities/import';
  const res = await fetch(url, { method: 'POST', body });
  if (!res.ok) throw new Error(`Import failed (${res.status})`);
  return res.json();
}
