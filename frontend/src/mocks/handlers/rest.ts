import { http, HttpResponse } from "msw";
import {
  mockUser,
  mockFactions,
  mockDetachments,
  mockMissions,
  mockSecondaryCards,
} from "../../test/fixtures";

const API_URL = "http://localhost:8080";

export const restHandlers = [
  // Auth
  http.get(`${API_URL}/api/auth/me`, () => {
    return HttpResponse.json(mockUser);
  }),

  http.post(`${API_URL}/api/auth/logout`, () => {
    return HttpResponse.json(null);
  }),

  // Factions
  http.get(`${API_URL}/api/factions`, () => {
    return HttpResponse.json(mockFactions);
  }),

  http.get(`${API_URL}/api/factions/:factionId/detachments`, () => {
    return HttpResponse.json(mockDetachments);
  }),

  http.get(`${API_URL}/api/factions/:factionId/stratagems`, () => {
    return HttpResponse.json([]);
  }),

  // Missions / reference data
  http.get(`${API_URL}/api/missions`, () => {
    return HttpResponse.json(mockMissions);
  }),

  http.get(`${API_URL}/api/secondary-cards`, () => {
    return HttpResponse.json(mockSecondaryCards);
  }),

  http.get(`${API_URL}/api/force-dispositions`, () => {
    return HttpResponse.json([]);
  }),

  http.get(`${API_URL}/api/mission-matchups`, () => {
    return HttpResponse.json([]);
  }),

  http.get(`${API_URL}/api/deployment-patterns`, () => {
    return HttpResponse.json([]);
  }),

  // Games
  http.get(`${API_URL}/api/games`, () => {
    return HttpResponse.json([]);
  }),

  http.post(`${API_URL}/api/games`, () => {
    return HttpResponse.json({ id: "game-new", inviteCode: "XYZ789" });
  }),

  http.post(`${API_URL}/api/games/join/:code`, ({ params }) => {
    const code = params.code as string;
    if (code === "INVALID") {
      return HttpResponse.json({ error: "Invalid code" }, { status: 404 });
    }
    return HttpResponse.json({ id: "game-joined", inviteCode: code });
  }),

  http.get(`${API_URL}/api/games/:id`, () => {
    return HttpResponse.json(null);
  }),

  http.get(`${API_URL}/api/games/:id/events`, () => {
    return HttpResponse.json([]);
  }),

  // User history
  http.get(`${API_URL}/api/users/me/history`, () => {
    return HttpResponse.json([]);
  }),
];
