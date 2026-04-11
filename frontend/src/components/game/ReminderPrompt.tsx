import { ReactNode } from "react";

interface Props {
  title: string;
  description: string;
  confirmLabel: string;
  cancelLabel: string;
  onConfirm: () => void;
  onCancel: () => void;
  children: ReactNode;
}

export function ReminderPrompt({
  title,
  description,
  confirmLabel,
  cancelLabel,
  onConfirm,
  onCancel,
  children,
}: Props) {
  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/70">
      <div className="bg-gray-800 border border-gray-600 rounded-xl max-w-lg w-full mx-4 max-h-[80vh] overflow-auto">
        <div className="px-5 py-4 border-b border-gray-700">
          <h2 className="text-lg font-bold text-white">{title}</h2>
          <p className="text-sm text-gray-400 mt-1">{description}</p>
        </div>

        <div className="px-5 py-4 space-y-4">{children}</div>

        <div className="px-5 py-4 border-t border-gray-700 flex gap-3">
          <button
            onClick={onCancel}
            className="flex-1 bg-gray-700 hover:bg-gray-600 text-white font-semibold py-2.5 rounded-lg transition-colors"
          >
            {cancelLabel}
          </button>
          <button
            onClick={onConfirm}
            className="flex-1 bg-indigo-600 hover:bg-indigo-700 text-white font-semibold py-2.5 rounded-lg transition-colors"
          >
            {confirmLabel}
          </button>
        </div>
      </div>
    </div>
  );
}
