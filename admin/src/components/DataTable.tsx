import { useState } from "react";

interface Column<T> {
  key: string;
  label: string;
  render?: (item: T) => React.ReactNode;
}

interface DataTableProps<T> {
  columns: Column<T>[];
  data: T[];
  getKey: (item: T) => string;
  onEdit?: (item: T) => void;
  onDelete?: (item: T) => void;
  searchField?: string;
}

// eslint-disable-next-line @typescript-eslint/no-explicit-any
export function DataTable<T extends Record<string, any>>({
  columns,
  data,
  getKey,
  onEdit,
  onDelete,
  searchField,
}: DataTableProps<T>) {
  const [search, setSearch] = useState("");
  const [deleteConfirm, setDeleteConfirm] = useState<string | null>(null);

  const filtered =
    searchField && search
      ? data.filter((item) => {
          const val = item[searchField];
          return typeof val === "string" && val.toLowerCase().includes(search.toLowerCase());
        })
      : data;

  return (
    <div>
      {searchField && (
        <input
          type="text"
          placeholder="Search..."
          value={search}
          onChange={(e) => setSearch(e.target.value)}
          className="mb-4 w-full max-w-sm px-3 py-2 bg-gray-700 border border-gray-600 rounded text-sm text-gray-100 placeholder-gray-400 focus:outline-none focus:border-amber-500"
        />
      )}
      <div className="overflow-x-auto">
        <table className="w-full text-sm">
          <thead>
            <tr className="border-b border-gray-700">
              {columns.map((col) => (
                <th key={col.key} className="text-left px-3 py-2 text-gray-400 font-medium">
                  {col.label}
                </th>
              ))}
              {(onEdit || onDelete) && (
                <th className="text-right px-3 py-2 text-gray-400 font-medium">Actions</th>
              )}
            </tr>
          </thead>
          <tbody>
            {filtered.map((item) => {
              const key = getKey(item);
              return (
                <tr key={key} className="border-b border-gray-800 hover:bg-gray-800/50">
                  {columns.map((col) => (
                    <td key={col.key} className="px-3 py-2">
                      {col.render ? col.render(item) : String(item[col.key] ?? "")}
                    </td>
                  ))}
                  {(onEdit || onDelete) && (
                    <td className="px-3 py-2 text-right space-x-2">
                      {onEdit && (
                        <button
                          onClick={() => onEdit(item)}
                          className="text-amber-400 hover:text-amber-300 text-xs"
                        >
                          Edit
                        </button>
                      )}
                      {onDelete &&
                        (deleteConfirm === key ? (
                          <>
                            <button
                              onClick={() => {
                                onDelete(item);
                                setDeleteConfirm(null);
                              }}
                              className="text-red-400 hover:text-red-300 text-xs"
                            >
                              Confirm
                            </button>
                            <button
                              onClick={() => setDeleteConfirm(null)}
                              className="text-gray-400 hover:text-gray-300 text-xs"
                            >
                              Cancel
                            </button>
                          </>
                        ) : (
                          <button
                            onClick={() => setDeleteConfirm(key)}
                            className="text-red-400 hover:text-red-300 text-xs"
                          >
                            Delete
                          </button>
                        ))}
                    </td>
                  )}
                </tr>
              );
            })}
            {filtered.length === 0 && (
              <tr>
                <td colSpan={columns.length + 1} className="px-3 py-8 text-center text-gray-500">
                  No data found
                </td>
              </tr>
            )}
          </tbody>
        </table>
      </div>
      <p className="mt-2 text-xs text-gray-500">{filtered.length} items</p>
    </div>
  );
}
