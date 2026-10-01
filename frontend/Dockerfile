FROM node:20-bookworm-slim AS build

WORKDIR /app

COPY package.json package-lock.json ./
RUN npm ci

# NEXT_PUBLIC_* vars are inlined into the client bundle at build time, so
# this must be a build arg, not just a runtime env var on the final image.
ARG NEXT_PUBLIC_API_BASE_URL=http://localhost:8080
ENV NEXT_PUBLIC_API_BASE_URL=$NEXT_PUBLIC_API_BASE_URL

COPY . .
RUN npm run build

FROM node:20-bookworm-slim AS runtime

ENV NODE_ENV=production
# The server is plain `node server.js`; the npm bundled in the base image is
# removed below because it is unused here and ships its own copy of node-tar.
# apt-get upgrade picks up Debian security fixes published after the base
# image was built, which the Trivy gate would otherwise flag.
RUN apt-get update && apt-get upgrade -y --no-install-recommends \
    && rm -rf /var/lib/apt/lists/* \
    && rm -rf /usr/local/lib/node_modules/npm /usr/local/bin/npm /usr/local/bin/npx \
    && useradd --system --create-home --uid 10001 app
WORKDIR /app

# `output: "standalone"` traces only the deps actually used at runtime into
# .next/standalone, so the runtime image doesn't need node_modules at all.
COPY --from=build /app/.next/standalone ./
COPY --from=build /app/.next/static ./.next/static

USER app
EXPOSE 3000
CMD ["node", "server.js"]
