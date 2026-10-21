import { formatSport } from '../format';
import { formatCount, plural } from '../plural';

export interface NoTrackItem {
  key: string;
  file: string;
  name: string;
  sport: string;
}

interface Props {
  items: NoTrackItem[];
}

/** Tells the user which files were skipped because they have no GPS track. */
export default function NoTrackNotice({ items }: Props) {
  const bySport = new Map<string, number>();
  for (const item of items) bySport.set(item.sport, (bySport.get(item.sport) ?? 0) + 1);
  const sports = [...bySport].sort(([a], [b]) => a.localeCompare(b));
  const n = items.length;

  return (
    <section className="notice" role="status" aria-labelledby="no-track-heading">
      <h2 id="no-track-heading">
        {formatCount(n, 'activity was', 'activities were')} not imported
      </h2>
      <p className="hint">
        {plural(n, 'It has', 'They have')} no GPS track, so there is nothing to show on the map. This is usual for
        indoor activities like treadmill runs, trainer rides or gym sessions.
      </p>
      <p className="notice-types">
        <span className="hint">Activity {plural(sports.length, 'type', 'types')}:</span>{' '}
        {sports.map(([sport, count], i) => (
          <span key={sport}>
            {i > 0 && ', '}
            {formatSport(sport)} ({count})
          </span>
        ))}
      </p>
      <details>
        <summary>Show files</summary>
        <ul>
          {items.map((item) => (
            <li key={item.key}>
              {item.file}
              {item.name !== item.file.replace(/\.gpx$/i, '') && <span className="hint"> – {item.name}</span>}
              <span className="hint"> · {formatSport(item.sport)}</span>
            </li>
          ))}
        </ul>
      </details>
    </section>
  );
}
