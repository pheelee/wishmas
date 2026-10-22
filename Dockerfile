# @impl CON-001-AC1 CON-001-AC2 CON-001-AC3 CON-002 CON-004
FROM golang:1.26.9-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY cmd/ ./cmd/
COPY internal/ ./internal/
# CSS/JS are embedded; no runtime toolchain is needed.
# Force Go-only image backends: scratch has no shared WebP or AVIF libraries.
RUN CGO_ENABLED=0 go build -tags nodynamic -trimpath -ldflags="-s -w" -o /out/wishmas ./cmd/wishmas \
    && CGO_ENABLED=0 go build -tags nodynamic -trimpath -ldflags="-s -w" -o /out/wishmas-admin ./cmd/wishmas-admin \
    && mkdir -p /out/data

FROM scratch
COPY --from=build /out/wishmas /wishmas
# @impl CON-007-AC2 CON-008-AC1: offline helper, invoked explicitly.
COPY --from=build /out/wishmas-admin /wishmas-admin
# Public HTTPS image downloads use the system certificate trust store.
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=build --chown=65532:65532 /out/data /data
USER 65532:65532
ENV WISHMAS_ADDR=:8080 WISHMAS_DATA_DIR=/data
EXPOSE 8080
VOLUME ["/data"]
ENTRYPOINT ["/wishmas"]
