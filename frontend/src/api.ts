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
  sport: string;
  startedAt: string | null;
  durationS: number | null;
  distanceM: number;
}

export interface ImportResult {
  file: string;
  status: 'imported' | 'already_imported' | 'possible_duplicate' | 'no_gps_track' | 'failed';
  activityId?: string;
  error?: string;
  /** The activity read from the file. For no_gps_track only name and sport are meaningful. */
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

/** Fetches JSON from url. what names the data in the error message. */
async function getJSON<T>(url: string, what: string, signal?: AbortSignal): Promise<T> {
  const res = await fetch(url, { signal });
  if (!res.ok) throw new Error(`Loading ${what} failed (${res.status})`);
  return res.json() as Promise<T>;
}

export function fetchActivities(signal?: AbortSignal): Promise<Activity[]> {
  return getJSON('/api/activities', 'activities', signal);
}

export function fetchDuplicateRules(signal?: AbortSignal): Promise<DuplicateRules> {
  return getJSON('/api/import/rules', 'import rules', signal);
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
