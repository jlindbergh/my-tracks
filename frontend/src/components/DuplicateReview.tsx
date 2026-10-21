import type { ActivityFacts, Comparison } from '../api';
import { formatDateTime, formatDistance, formatDuration, formatPercent } from '../format';

export interface ReviewItem {
  key: string;
  file: string;
  activity: ActivityFacts;
  existing: ActivityFacts;
  comparison: Comparison;
}

interface Props {
  items: ReviewItem[];
  selected: ReadonlySet<string>;
  busy: boolean;
  onToggle: (key: string) => void;
  onImportSelected: () => void;
  onSkipAll: () => void;
}

function differences(c: Comparison): string {
  const parts: string[] = [];
  if (c.startDiffS != null) {
    parts.push(c.startDiffS === 0 ? 'same start time' : `starts ${Number(c.startDiffS.toFixed(1))} s apart`);
  }
  parts.push(`distance ${formatPercent(c.distanceDiff)} different`);
  if (c.durationDiff != null) {
    parts.push(`duration ${formatPercent(c.durationDiff)} different`);
  }
  return parts.join(' · ');
}

export default function DuplicateReview({ items, selected, busy, onToggle, onImportSelected, onSkipAll }: Props) {
  return (
    <section className="review" aria-labelledby="review-heading">
      <h2 id="review-heading">Possible duplicates ({items.length})</h2>
      <p className="hint">
        These files look like activities you already have, so they were not imported. Only import them if they are
        really separate activities, otherwise they will show up twice.
      </p>
      <ul>
        {items.map((item) => (
          <li key={item.key}>
            <label className="review-item">
              <input
                type="checkbox"
                aria-label={`Import ${item.file} anyway`}
                checked={selected.has(item.key)}
                onChange={() => onToggle(item.key)}
              />
              <span>
                <span className="review-file">{item.file}</span>
                <table>
                  <thead>
                    <tr>
                      <th />
                      <th>In this file</th>
                      <th>Already imported</th>
                    </tr>
                  </thead>
                  <tbody>
                    <tr>
                      <th>Name</th>
                      <td>{item.activity.name}</td>
                      <td>{item.existing.name}</td>
                    </tr>
                    <tr>
                      <th>Start</th>
                      <td>{formatDateTime(item.activity.startedAt)}</td>
                      <td>{formatDateTime(item.existing.startedAt)}</td>
                    </tr>
                    <tr>
                      <th>Distance</th>
                      <td>{formatDistance(item.activity.distanceM)}</td>
                      <td>{formatDistance(item.existing.distanceM)}</td>
                    </tr>
                    <tr>
                      <th>Duration</th>
                      <td>{formatDuration(item.activity.durationS) || '–'}</td>
                      <td>{formatDuration(item.existing.durationS) || '–'}</td>
                    </tr>
                  </tbody>
                </table>
                <span className="hint">{differences(item.comparison)}</span>
              </span>
            </label>
          </li>
        ))}
      </ul>
      <div className="review-actions">
        <button type="button" className="primary" disabled={busy || selected.size === 0} onClick={onImportSelected}>
          Import {selected.size > 0 ? selected.size : ''} selected anyway
        </button>
        <button type="button" className="secondary" disabled={busy} onClick={onSkipAll}>
          Skip all
        </button>
      </div>
    </section>
  );
}
