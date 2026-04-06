import { useState, useRef } from "react";
import { ImportResult } from "../api/admin";

interface ImportDialogProps {
  title: string;
  accept: string;
  onImport: (file: File) => Promise<ImportResult>;
  onClose: () => void;
  onSuccess: () => void;
}

export function ImportDialog({ title, accept, onImport, onClose, onSuccess }: ImportDialogProps) {
  const [file, setFile] = useState<File | null>(null);
  const [loading, setLoading] = useState(false);
  const [result, setResult] = useState<ImportResult | null>(null);
  const [error, setError] = useState<string | null>(null);
  const inputRef = useRef<HTMLInputElement>(null);

  const handleImport = async () => {
    if (!file) return;
    setLoading(true);
    setError(null);
    try {
      const res = await onImport(file);
      setResult(res);
      onSuccess();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Import failed");
    } finally {
      setLoading(false);
    }
  };

  return (
    <div
      className="fixed inset-0 bg-black/50 flex items-center justify-center z-50"
      onClick={onClose}
    >
      <div
        className="bg-gray-800 rounded-lg p-6 w-full max-w-md"
        onClick={(e) => e.stopPropagation()}
      >
        <h2 className="text-lg font-semibold mb-4">{title}</h2>

        <div
          className="border-2 border-dashed border-gray-600 rounded-lg p-8 text-center cursor-pointer hover:border-amber-500 transition-colors"
          onClick={() => inputRef.current?.click()}
        >
          <input
            ref={inputRef}
            type="file"
            accept={accept}
            className="hidden"
            onChange={(e) => setFile(e.target.files?.[0] || null)}
          />
          {file ? (
            <p className="text-sm text-amber-400">
              {file.name} ({(file.size / 1024).toFixed(1)} KB)
            </p>
          ) : (
            <p className="text-sm text-gray-400">Click to select a file ({accept})</p>
          )}
        </div>

        {error && <p className="mt-3 text-sm text-red-400">{error}</p>}

        {result && (
          <div className="mt-3 p-3 bg-green-900/30 border border-green-700 rounded text-sm">
            <p className="text-green-400 font-medium">Import successful</p>
            {Object.entries(result)
              .filter(([k]) => k !== "entity")
              .map(([k, v]) => (
                <p key={k} className="text-green-300">
                  {k}: {String(v)}
                </p>
              ))}
          </div>
        )}

        <div className="mt-4 flex justify-end gap-2">
          <button
            onClick={onClose}
            className="px-3 py-1.5 text-sm text-gray-400 hover:text-gray-300"
          >
            Close
          </button>
          {!result && (
            <button
              onClick={handleImport}
              disabled={!file || loading}
              className="px-3 py-1.5 text-sm bg-amber-600 hover:bg-amber-500 rounded disabled:opacity-50 disabled:cursor-not-allowed"
            >
              {loading ? "Importing..." : "Import"}
            </button>
          )}
        </div>
      </div>
    </div>
  );
}
