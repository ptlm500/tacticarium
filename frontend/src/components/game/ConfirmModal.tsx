interface Props {
  title: string;
  message: string;
  confirmLabel?: string;
  cancelLabel?: string;
  onConfirm: () => void;
  onCancel: () => void;
  variant?: "danger" | "default";
}

export function ConfirmModal({
  title,
  message,
  confirmLabel = "Confirm",
  cancelLabel = "Cancel",
  onConfirm,
  onCancel,
  variant = "default",
}: Props) {
  const confirmColors =
    variant === "danger" ? "bg-red-600 hover:bg-red-700" : "bg-indigo-600 hover:bg-indigo-700";

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/70">
      <div className="bg-gray-800 border border-gray-600 rounded-xl max-w-sm w-full mx-4">
        <div className="px-5 py-4 border-b border-gray-700">
          <h2 className="text-lg font-bold text-white">{title}</h2>
        </div>
        <div className="px-5 py-4">
          <p className="text-sm text-gray-300">{message}</p>
        </div>
        <div className="px-5 py-4 border-t border-gray-700 flex gap-3">
          <button
            onClick={onCancel}
            className="flex-1 bg-gray-700 hover:bg-gray-600 text-white font-semibold py-2.5 rounded-lg transition-colors"
          >
            {cancelLabel}
          </button>
          <button
            onClick={onConfirm}
            className={`flex-1 ${confirmColors} text-white font-semibold py-2.5 rounded-lg transition-colors`}
          >
            {confirmLabel}
          </button>
        </div>
      </div>
    </div>
  );
}
