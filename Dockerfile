# @impl CON-001-AC1 CON-001-AC2 CON-001-AC3 CON-002-AC2 CON-004
# @impl REL-001-AC2 REL-001-AC4: cross-compile both targets on the native builder.
FROM --platform=$BUILDPLATFORM golang:1.27.3-alpine AS build
ARG TARGETOS
ARG TARGETARCH
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY cmd/ ./cmd/
COPY internal/ ./internal/
# CSS/JS are embedded; no runtime toolchain is needed.
# Force Go-only image backends: scratch has no shared WebP or AVIF libraries.
RUN GOOS=$TARGETOS GOARCH=$TARGETARCH CGO_ENABLED=0 go build -tags nodynamic -trimpath -ldflags="-s -w" -o /out/wishmas ./cmd/wishmas \
    && GOOS=$TARGETOS GOARCH=$TARGETARCH CGO_ENABLED=0 go build -tags nodynamic -trimpath -ldflags="-s -w" -o /out/wishmas-admin ./cmd/wishmas-admin \
    && mkdir -p /out/data

FROM scratch
# @impl REL-001-AC3 REL-001-AC4: metadata identifies the tested source/version.
ARG VERSION=dev
ARG REVISION=unknown
LABEL org.opencontainers.image.source="https://github.com/pheelee/wishmas" \
      org.opencontainers.image.version=$VERSION \
      org.opencontainers.image.revision=$REVISION
COPY --from=build /out/wishmas /wishmas
# @impl CON-007-AC2 CON-008-AC1: offline helper, invoked explicitly.
COPY --from=build /out/wishmas-admin /wishmas-admin
# Public HTTPS image downloads use the system certificate trust store.
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
# @impl CON-001-AC5: Compose user overrides need no image rebuild.
COPY --from=build --chown=1000:1000 /out/data /data
USER 1000:1000
ENV WISHMAS_ADDR=:8080 WISHMAS_DATA_DIR=/data
EXPOSE 8080
VOLUME ["/data"]
ENTRYPOINT ["/wishmas"]
