FROM node:latest AS build-vue

WORKDIR /app/frontend

COPY frontend/package*.json ./

RUN npm install

COPY frontend/. .

RUN npm run build-only

FROM --platform=$BUILDPLATFORM golang:alpine AS build-go

ARG TARGETOS
ARG TARGETARCH

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -o /server

# Debian (glibc) is required because the official Ookla CLI is a dynamically-linked binary — it will not run on a scratch/musl image
FROM debian:bookworm-slim

# Install the official Ookla speedtest CLI from Ookla's package repository (https://www.speedtest.net/apps/cli)
RUN apt-get update \
    && apt-get install -y --no-install-recommends ca-certificates curl gnupg \
    && curl -fsSL https://packagecloud.io/install/repositories/ookla/speedtest-cli/script.deb.sh | bash \
    && apt-get install -y --no-install-recommends speedtest \
    && apt-get purge -y --auto-remove curl gnupg \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /

COPY --from=build-go /server /server
COPY --from=build-vue /app/frontend/dist /frontend/dist

EXPOSE 8080

COPY --from=ghcr.io/tarampampam/microcheck:1 /bin/httpcheck /bin/httpcheck
HEALTHCHECK --interval=30s --timeout=3s --retries=3 --start-period=10s CMD ["httpcheck", "http://localhost:8080/"]

ENTRYPOINT ["/server"]
