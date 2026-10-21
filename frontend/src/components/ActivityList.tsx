import type { Activity } from '../api';
import { formatDate, formatDistance, formatDuration, sportColor, sportLabel } from '../format';

interface Props {
  activities: Activity[];
  selectedId: string | null;
  onSelect: (activity: Activity) => void;
}

export default function ActivityList({ activities, selectedId, onSelect }: Props) {
  if (activities.length === 0) {
    return <p className="empty">No activities yet. Import some GPX files to get started.</p>;
  }
  return (
    <ul className="activity-list">
      {activities.map((a) => (
        <li key={a.id}>
          <button
            type="button"
            className={a.id === selectedId ? 'activity selected' : 'activity'}
            onClick={() => onSelect(a)}
          >
            <span className="swatch" style={{ background: sportColor(a.sport) }} aria-hidden />
            <span className="activity-text">
              <span className="activity-name">{a.name}</span>
              <span className="activity-meta">
                {formatDate(a.startedAt)} · {sportLabel(a.sport)} · {formatDistance(a.distanceM)}
                {a.durationS != null && ` · ${formatDuration(a.durationS)}`}
              </span>
            </span>
          </button>
        </li>
      ))}
    </ul>
  );
}
