import { useConfig } from '../config';
import { formatPercent } from '../format';
import { formatCount } from '../plural';

/** Explains how duplicates are detected, using the backend's actual values. */
export default function DuplicateRulesInfo() {
  const rule = useConfig().import.duplicateRule;
  const seconds = formatCount(rule.startToleranceS, 'second', 'seconds');
  const percent = formatPercent(rule.relativeTolerance);

  return (
    <details className="rules">
      <summary>How duplicates are detected</summary>
      <ul>
        <li>
          <strong>Already imported:</strong> an exact copy of a file you have imported before is always skipped.
        </li>
        <li>
          <strong>Possible duplicate:</strong> a file is held back for you to review when one of your activities
          starts within {seconds} of it <em>and</em> its distance and duration are both within {percent}.
        </li>
        <li>
          GPS points are not compared, since apps like Garmin, Strava and Runkeeper record them differently.
        </li>
        <li>Files without timestamps can only be detected as exact copies.</li>
      </ul>
    </details>
  );
}
