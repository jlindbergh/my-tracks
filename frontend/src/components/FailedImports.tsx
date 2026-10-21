import { formatCount, plural } from '../plural';

export interface FailedItem {
  key: string;
  file: string;
  /** Why the backend could not import the file. */
  error: string;
}

interface Props {
  items: FailedItem[];
}

/** Lists the files that could not be imported and why. */
export default function FailedImports({ items }: Props) {
  const n = items.length;
  return (
    <section className="failed" role="alert" aria-labelledby="failed-heading">
      <h2 id="failed-heading">{formatCount(n, 'file', 'files')} could not be imported</h2>
      <p className="hint">
        Nothing was imported from {plural(n, 'this file', 'these files')}. The rest of the import was not affected. If a
        file is damaged, try exporting the activity again.
      </p>
      <ul>
        {items.map((item) => (
          <li key={item.key}>
            <span className="failed-file">{item.file}</span>
            <span className="failed-reason">{item.error}</span>
          </li>
        ))}
      </ul>
    </section>
  );
}
