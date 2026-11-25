ARG GO_VERSION=1.26
ARG NODE_VERSION=22
FROM golang:${GO_VERSION}-bookworm AS backend-build
RUN apt-get update && apt-get install -y --no-install-recommends gcc libc6-dev git ca-certificates && rm -rf /var/lib/apt/lists/*
WORKDIR /src/server
COPY server/go.mod server/go.sum ./
RUN go mod download
COPY server/ ./
RUN CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags="-s -w" -o /out/hammond ./main.go
FROM node:${NODE_VERSION}-bookworm-slim AS frontend-build
RUN apt-get update && apt-get install -y --no-install-recommends python3 make g++ && rm -rf /var/lib/apt/lists/*
ENV PYTHON=/usr/bin/python3 NODE_OPTIONS=--openssl-legacy-provider
WORKDIR /src/ui
COPY ui/package.json ui/package-lock.json ./
RUN npm ci --legacy-peer-deps --no-audit --no-fund
COPY ui/ ./
RUN npm run test:unit -- --runInBand && npm run build
FROM debian:bookworm-slim
RUN apt-get update && apt-get install -y --no-install-recommends ca-certificates tzdata tesseract-ocr tesseract-ocr-deu tesseract-ocr-eng imagemagick && rm -rf /var/lib/apt/lists/* && mkdir -p /config /assets /api
ENV CONFIG=/config DATA=/assets GIN_MODE=release TZ=Europe/Berlin
WORKDIR /api
COPY --from=backend-build /out/hammond ./app
COPY --from=frontend-build /src/ui/dist ./dist
VOLUME ["/config", "/assets"]
EXPOSE 3000
ENTRYPOINT ["./app"]
