import { useEffect, useState } from "react";
import { useParams, useNavigate } from "react-router-dom";
import { gamesApi } from "../api/games";

export function JoinRedirect() {
  const { code } = useParams<{ code: string }>();
  const navigate = useNavigate();
  const [error, setError] = useState("");

  useEffect(() => {
    if (!code) return;
    gamesApi
      .join(code)
      .then(({ id }) => navigate(`/game/${id}/setup`))
      .catch(() => setError("Failed to join game. Invalid or expired code."));
  }, [code, navigate]);

  if (error) {
    return (
      <div className="min-h-screen bg-gray-900 text-white flex items-center justify-center">
        <div className="text-center">
          <p className="text-red-400 mb-4">{error}</p>
          <a href="/" className="text-indigo-400 hover:text-indigo-300">
            Back to Lobby
          </a>
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-gray-900 text-white flex items-center justify-center">
      <p>Joining game...</p>
    </div>
  );
}
