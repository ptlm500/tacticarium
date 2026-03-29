import { useAuth } from "../hooks/useAuth";

export function LoginPage() {
  const { login } = useAuth();

  return (
    <div className="min-h-screen flex items-center justify-center bg-gray-900">
      <div className="text-center p-8">
        <h1 className="text-4xl font-bold text-white mb-2">Tacticarium</h1>
        <p className="text-gray-400 mb-8">Warhammer 40K 10th Edition</p>
        <button
          onClick={login}
          className="bg-indigo-600 hover:bg-indigo-700 text-white font-semibold py-3 px-8 rounded-lg text-lg transition-colors"
        >
          Sign in with Discord
        </button>
      </div>
    </div>
  );
}
