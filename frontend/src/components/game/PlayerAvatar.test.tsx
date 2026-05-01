import { describe, expect, it } from "vite-plus/test";
import { render, screen } from "@testing-library/react";
import { PlayerAvatar } from "./PlayerAvatar";

// 1x1 transparent PNG, used so the AvatarImage loads in browser-mode tests.
const tinyPng =
  "data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR42mNkAAIAAAoAAv/lxKUAAAAASUVORK5CYII=";

describe("PlayerAvatar", () => {
  it("renders the username's initial when no avatarUrl is provided", () => {
    render(<PlayerAvatar username="alice" />);
    expect(screen.getByText("A")).toBeTruthy();
  });

  it("renders the AvatarImage with the supplied src once it loads", async () => {
    render(<PlayerAvatar username="bob" avatarUrl={tinyPng} />);
    const img = (await screen.findByAltText("bob's avatar")) as HTMLImageElement;
    expect(img.src).toBe(tinyPng);
  });

  it("uses '?' as the initial when the username is empty", () => {
    render(<PlayerAvatar username="" />);
    expect(screen.getByText("?")).toBeTruthy();
  });
});
