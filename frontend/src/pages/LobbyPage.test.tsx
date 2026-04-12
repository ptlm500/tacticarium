import { screen, act } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { http, HttpResponse } from "msw";
import { worker } from "../mocks/browser";
import { renderWithProviders } from "../test/renderWithProviders";
import { mockUser } from "../test/fixtures";
import { LobbyPage } from "./LobbyPage";
import { GameSummary } from "../types/game";

const API_URL = "http://localhost:8080";

const mockGames: GameSummary[] = [
  {
    id: "game-1",
    inviteCode: "ABC123",
    status: "active",
    missionName: "Supply Drop",
    createdAt: "2025-01-01T00:00:00Z",
    players: [
      {
        userId: "user-1",
        username: "TestPlayer",
        factionName: "Space Marines",
        playerNumber: 1,
        totalVp: 10,
      },
      { userId: "user-2", username: "Opponent", factionName: "Orks", playerNumber: 2, totalVp: 5 },
    ],
  },
  {
    id: "game-2",
    inviteCode: "DEF456",
    status: "setup",
    missionName: "",
    createdAt: "2025-01-02T00:00:00Z",
    players: [
      { userId: "user-1", username: "TestPlayer", factionName: "", playerNumber: 1, totalVp: 0 },
    ],
  },
];

function setupMocks(games: GameSummary[] = mockGames) {
  worker.use(
    http.get(`${API_URL}/api/games`, () => {
      return HttpResponse.json(games);
    }),
  );
}

describe("LobbyPage - Remove Game", () => {
  beforeEach(() => {
    localStorage.setItem("token", "test-token");
  });

  it("shows remove buttons next to each game", async () => {
    setupMocks();

    await act(async () => {
      renderWithProviders(<LobbyPage />, { user: mockUser });
    });

    await vi.waitFor(() => {
      expect(screen.getByText("Supply Drop")).toBeTruthy();
    });

    const removeButtons = screen.getAllByRole("button", { name: "Remove game" });
    expect(removeButtons).toHaveLength(2);
  });

  it("shows confirmation modal when remove button is clicked", async () => {
    setupMocks();
    const user = userEvent.setup();

    await act(async () => {
      renderWithProviders(<LobbyPage />, { user: mockUser });
    });

    await vi.waitFor(() => {
      expect(screen.getByText("Supply Drop")).toBeTruthy();
    });

    const removeButtons = screen.getAllByRole("button", { name: "Remove game" });
    await user.click(removeButtons[0]);

    expect(screen.getByText("Remove Game")).toBeTruthy();
    expect(screen.getByText(/Are you sure you want to remove this game/)).toBeTruthy();
    expect(screen.getByRole("button", { name: "Remove" })).toBeTruthy();
    expect(screen.getByRole("button", { name: "Cancel" })).toBeTruthy();
  });

  it("closes modal when cancel is clicked", async () => {
    setupMocks();
    const user = userEvent.setup();

    await act(async () => {
      renderWithProviders(<LobbyPage />, { user: mockUser });
    });

    await vi.waitFor(() => {
      expect(screen.getByText("Supply Drop")).toBeTruthy();
    });

    const removeButtons = screen.getAllByRole("button", { name: "Remove game" });
    await user.click(removeButtons[0]);

    expect(screen.getByText("Remove Game")).toBeTruthy();

    await user.click(screen.getByRole("button", { name: "Cancel" }));

    expect(screen.queryByText("Remove Game")).toBeNull();
    // Game should still be visible
    expect(screen.getByText("Supply Drop")).toBeTruthy();
  });

  it("removes game from list after confirming", async () => {
    setupMocks();
    const user = userEvent.setup();

    let hideCalled = false;
    worker.use(
      http.post(`${API_URL}/api/games/:id/hide`, () => {
        hideCalled = true;
        return new HttpResponse(null, { status: 204 });
      }),
    );

    await act(async () => {
      renderWithProviders(<LobbyPage />, { user: mockUser });
    });

    await vi.waitFor(() => {
      expect(screen.getByText("Supply Drop")).toBeTruthy();
    });

    const removeButtons = screen.getAllByRole("button", { name: "Remove game" });
    await user.click(removeButtons[0]);
    await user.click(screen.getByRole("button", { name: "Remove" }));

    await vi.waitFor(() => {
      expect(hideCalled).toBe(true);
      expect(screen.queryByText("Supply Drop")).toBeNull();
    });

    // Second game should still be visible
    expect(screen.getByText("No mission selected")).toBeTruthy();
  });

  it("shows error when hide API fails", async () => {
    setupMocks();
    const user = userEvent.setup();

    worker.use(
      http.post(`${API_URL}/api/games/:id/hide`, () => {
        return HttpResponse.json({ error: "not found" }, { status: 404 });
      }),
    );

    await act(async () => {
      renderWithProviders(<LobbyPage />, { user: mockUser });
    });

    await vi.waitFor(() => {
      expect(screen.getByText("Supply Drop")).toBeTruthy();
    });

    const removeButtons = screen.getAllByRole("button", { name: "Remove game" });
    await user.click(removeButtons[0]);
    await user.click(screen.getByRole("button", { name: "Remove" }));

    await vi.waitFor(() => {
      expect(screen.getByText("Failed to remove game")).toBeTruthy();
    });

    // Game should still be visible
    expect(screen.getByText("Supply Drop")).toBeTruthy();
  });
});
