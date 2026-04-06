import { ws } from "msw";

export const gameWs = ws.link("ws://localhost:8080/ws/game/*");

export const wsHandlers = [gameWs.addEventListener("connection", () => {})];
