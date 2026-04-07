FROM node:20-slim AS builder
ENV PNPM_HOME="/pnpm"
ENV PATH="$PNPM_HOME:$PATH"
RUN corepack enable
WORKDIR /app/admin
COPY admin/package.json admin/pnpm-lock.yaml admin/pnpm-workspace.yaml ./
RUN pnpm install --frozen-lockfile
COPY shared/ /app/shared/
COPY admin/ .
ARG VITE_API_URL
RUN pnpm build

FROM nginx:alpine-slim
COPY --from=builder /app/admin/dist /usr/share/nginx/html
COPY admin/nginx.conf /etc/nginx/conf.d/default.conf
EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]
