import { useEffect, useRef, useState, type ChangeEvent, type DragEvent } from 'react';
import { fetchDuplicateRules, importFiles, type DuplicateRules, type ImportResult } from '../api';
import DuplicateReview, { type ReviewItem } from './DuplicateReview';
import DuplicateRulesInfo from './DuplicateRulesInfo';

interface Props {
  /** Called after an import that stored at least one new activity. */
  onImported: () => void;
}

/** One uploaded file and what happened to it. */
interface Entry {
  key: string;
  file: File;
  result: ImportResult;
}

let nextKey = 0;

function possibleDuplicates(n: number): string {
  return `${n} possible duplicate${n === 1 ? '' : 's'}`;
}

export default function ImportPanel({ onImported }: Props) {
  const inputRef = useRef<HTMLInputElement>(null);
  const [rules, setRules] = useState<DuplicateRules | null>(null);
  const [busy, setBusy] = useState(false);
  const [dragging, setDragging] = useState(false);
  const [entries, setEntries] = useState<Entry[] | null>(null);
  // Possible duplicates the user has looked at and decided to skip.
  const [skipped, setSkipped] = useState<ReadonlySet<string>>(new Set());
  const [selected, setSelected] = useState<ReadonlySet<string>>(new Set());
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const controller = new AbortController();
    fetchDuplicateRules(controller.signal)
      .then(setRules)
      .catch(() => {
        // The explanation falls back to approximate wording.
      });
    return () => controller.abort();
  }, []);

  async function upload(files: File[]) {
    const gpx = files.filter((f) => f.name.toLowerCase().endsWith('.gpx'));
    if (gpx.length === 0) {
      setError('Only .gpx files can be imported for now.');
      return;
    }
    setBusy(true);
    setError(null);
    setEntries(null);
    setSkipped(new Set());
    setSelected(new Set());
    try {
      const results = await importFiles(gpx);
      setEntries(results.map((result, i) => ({ key: String(nextKey++), file: gpx[i], result })));
      if (results.some((r) => r.status === 'imported')) onImported();
    } catch (e) {
      setError(e instanceof Error ? e.message : String(e));
    } finally {
      setBusy(false);
    }
  }

  async function importSelectedAnyway() {
    if (!entries) return;
    const chosen = entries.filter((e) => selected.has(e.key));
    setBusy(true);
    setError(null);
    try {
      const results = await importFiles(chosen.map((e) => e.file), true);
      // Results come back in upload order.
      const byKey = new Map(chosen.map((e, i) => [e.key, results[i]]));
      setEntries(entries.map((e) => ({ ...e, result: byKey.get(e.key) ?? e.result })));
      setSelected(new Set());
      if (results.some((r) => r.status === 'imported')) onImported();
    } catch (e) {
      setError(e instanceof Error ? e.message : String(e));
    } finally {
      setBusy(false);
    }
  }

  function toggle(key: string) {
    const next = new Set(selected);
    if (next.has(key)) next.delete(key);
    else next.add(key);
    setSelected(next);
  }

  function skipAll() {
    setSkipped(new Set([...skipped, ...review.map((r) => r.key)]));
    setSelected(new Set());
  }

  function handleChange(e: ChangeEvent<HTMLInputElement>) {
    const files = Array.from(e.target.files ?? []);
    e.target.value = '';
    void upload(files);
  }

  function handleDrop(e: DragEvent<HTMLDivElement>) {
    e.preventDefault();
    setDragging(false);
    void upload(Array.from(e.dataTransfer.files));
  }

  const results = entries?.map((e) => e.result) ?? [];
  const count = (status: ImportResult['status']) => results.filter((r) => r.status === status).length;
  const failed = entries?.filter((e) => e.result.status === 'failed') ?? [];
  const review: ReviewItem[] = (entries ?? [])
    .filter((e) => e.result.status === 'possible_duplicate' && !skipped.has(e.key))
    .flatMap((e) =>
      e.result.activity && e.result.existing
        ? [{ key: e.key, file: e.file.name, activity: e.result.activity, existing: e.result.existing }]
        : [],
    );

  const summary = entries && [
    `${count('imported')} imported`,
    count('already_imported') > 0 && `${count('already_imported')} already imported (exact copies)`,
    review.length > 0 && `${possibleDuplicates(review.length)} to review`,
    skipped.size > 0 && `${possibleDuplicates(skipped.size)} skipped`,
    failed.length > 0 && `${failed.length} failed`,
  ];

  return (
    <div className="import-panel">
      <div
        className={dragging ? 'import dragging' : 'import'}
        onDragOver={(e) => {
          e.preventDefault();
          setDragging(true);
        }}
        onDragLeave={() => setDragging(false)}
        onDrop={handleDrop}
      >
        <input ref={inputRef} type="file" accept=".gpx" multiple hidden onChange={handleChange} />
        <button type="button" className="primary" disabled={busy} onClick={() => inputRef.current?.click()}>
          {busy ? 'Importing…' : 'Import GPX files'}
        </button>
        <span className="hint">or drop files here</span>
        <DuplicateRulesInfo rules={rules} />

        {error && <p className="error">{error}</p>}
        {summary && (
          <div className="import-result">
            {summary.filter(Boolean).join(', ')}
            {failed.length > 0 && (
              <ul>
                {failed.map((e) => (
                  <li key={e.key}>
                    {e.file.name}: {e.result.error}
                  </li>
                ))}
              </ul>
            )}
          </div>
        )}
      </div>

      {review.length > 0 && (
        <DuplicateReview
          items={review}
          selected={selected}
          busy={busy}
          onToggle={toggle}
          onImportSelected={() => void importSelectedAnyway()}
          onSkipAll={skipAll}
        />
      )}
    </div>
  );
}
