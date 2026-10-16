import { useRef, useState, type ChangeEvent, type DragEvent } from 'react';
import { importFiles, type ImportResult } from '../api';

interface Props {
  /** Called after an import that stored at least one new activity. */
  onImported: (results: ImportResult[]) => void;
}

export default function ImportPanel({ onImported }: Props) {
  const inputRef = useRef<HTMLInputElement>(null);
  const [busy, setBusy] = useState(false);
  const [dragging, setDragging] = useState(false);
  const [results, setResults] = useState<ImportResult[] | null>(null);
  const [error, setError] = useState<string | null>(null);

  async function upload(files: File[]) {
    const gpx = files.filter((f) => f.name.toLowerCase().endsWith('.gpx'));
    if (gpx.length === 0) {
      setError('Only .gpx files can be imported for now.');
      return;
    }
    setBusy(true);
    setError(null);
    setResults(null);
    try {
      const res = await importFiles(gpx);
      setResults(res);
      if (res.some((r) => r.status === 'imported')) onImported(res);
    } catch (e) {
      setError(e instanceof Error ? e.message : String(e));
    } finally {
      setBusy(false);
    }
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

  const counts = results && {
    imported: results.filter((r) => r.status === 'imported').length,
    duplicate: results.filter((r) => r.status === 'duplicate').length,
    failed: results.filter((r) => r.status === 'failed'),
  };

  return (
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

      {error && <p className="error">{error}</p>}
      {counts && (
        <div className="import-result">
          {counts.imported} imported
          {counts.duplicate > 0 && `, ${counts.duplicate} already imported`}
          {counts.failed.length > 0 && `, ${counts.failed.length} failed`}
          {counts.failed.length > 0 && (
            <ul>
              {counts.failed.map((r) => (
                <li key={r.file}>
                  {r.file}: {r.error}
                </li>
              ))}
            </ul>
          )}
        </div>
      )}
    </div>
  );
}
