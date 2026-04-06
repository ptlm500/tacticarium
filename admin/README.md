# Tacticarium Admin

Web interface for managing Tacticarium reference data (factions, stratagems, missions, secondaries, etc.). Authenticates via GitHub OAuth, separate from the player Discord auth.

## Environment Variables

### Admin Frontend (build-time)

| Variable       | Default                 | Description     |
| -------------- | ----------------------- | --------------- |
| `VITE_API_URL` | `http://localhost:8080` | Backend API URL |

Pass as a Docker build arg:

```bash
docker build --build-arg VITE_API_URL=https://api.example.com -t tacticarium-admin .
```

### Backend (required for admin auth)

These are set on the **backend** service, not the admin frontend:

| Variable               | Default                                          | Description                                              |
| ---------------------- | ------------------------------------------------ | -------------------------------------------------------- |
| `GITHUB_CLIENT_ID`     | `""`                                             | GitHub OAuth App client ID                               |
| `GITHUB_CLIENT_SECRET` | `""`                                             | GitHub OAuth App client secret                           |
| `GITHUB_REDIRECT_URI`  | `http://localhost:8080/api/auth/github/callback` | GitHub OAuth callback URL                                |
| `ADMIN_GITHUB_IDS`     | `""`                                             | Comma-separated GitHub **user IDs** allowed admin access |
| `ADMIN_FRONTEND_URL`   | `http://localhost:5174`                          | Admin frontend URL (used for CORS and OAuth redirect)    |

### Finding your GitHub user ID

```bash
curl -s https://api.github.com/users/YOUR_USERNAME | grep '"id"'
```

### Setting up a GitHub OAuth App

1. Go to GitHub Settings > Developer settings > OAuth Apps > New OAuth App
2. Set **Authorization callback URL** to your `GITHUB_REDIRECT_URI` value
3. Copy the Client ID and generate a Client Secret

## Local Development

```bash
# Install dependencies
vp install

# Start dev server (port 5174)
vp dev
```

Requires the backend running on port 8080 with the GitHub OAuth env vars configured.

## Production Build

```bash
vp build    # Output in dist/
```

Or via Docker:

```bash
docker build --build-arg VITE_API_URL=https://api.example.com -t tacticarium-admin .
docker run -p 80:80 tacticarium-admin
```
