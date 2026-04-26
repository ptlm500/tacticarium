import { http, HttpResponse } from "msw";
import { worker } from "../mocks/browser";
import { api, HttpError } from "./client";

const API_URL = "http://localhost:8080";

describe("api.get retry", () => {
  it("retries 5xx responses and succeeds when a later attempt returns 200", async () => {
    let attempt = 0;
    worker.use(
      http.get(`${API_URL}/api/retry-test`, () => {
        attempt += 1;
        if (attempt < 3) {
          return new HttpResponse("upstream blew up", { status: 503 });
        }
        return HttpResponse.json({ ok: true });
      }),
    );

    const result = await api.get<{ ok: boolean }>("/api/retry-test");
    expect(result).toEqual({ ok: true });
    expect(attempt).toBe(3);
  });

  it("does not retry on 4xx responses", async () => {
    let attempt = 0;
    worker.use(
      http.get(`${API_URL}/api/no-retry`, () => {
        attempt += 1;
        return HttpResponse.json({ error: "nope" }, { status: 404 });
      }),
    );

    await expect(api.get("/api/no-retry")).rejects.toBeInstanceOf(HttpError);
    expect(attempt).toBe(1);
  });

  it("gives up after exhausting all attempts on persistent 5xx", async () => {
    let attempt = 0;
    worker.use(
      http.get(`${API_URL}/api/always-fails`, () => {
        attempt += 1;
        return new HttpResponse("server is on fire", { status: 500 });
      }),
    );

    await expect(api.get("/api/always-fails")).rejects.toMatchObject({
      status: 500,
    });
    expect(attempt).toBe(3);
  });
});

describe("api.post (no retry)", () => {
  it("does not retry POST on 5xx", async () => {
    let attempt = 0;
    worker.use(
      http.post(`${API_URL}/api/post-fail`, () => {
        attempt += 1;
        return new HttpResponse("nope", { status: 502 });
      }),
    );

    await expect(api.post("/api/post-fail")).rejects.toMatchObject({
      status: 502,
    });
    expect(attempt).toBe(1);
  });
});
