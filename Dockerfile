# @impl CON-001-AC1 CON-001-AC2 CON-001-AC3 CON-002 CON-004
FROM golang:1.26.9-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY cmd/ ./cmd/
COPY internal/ ./internal/
# CSS/JS and IANA timezone data are embedded; no runtime toolchain is needed.
# Force the Go WebP backend: scratch has no dynamic loader/system libwebp.
RUN CGO_ENABLED=0 go build -tags nodynamic -trimpath -ldflags="-s -w" -o /out/wishmas ./cmd/wishmas \
    && mkdir -p /out/data

FROM scratch
COPY --from=build /out/wishmas /wishmas
# Public HTTPS image downloads use the system certificate trust store.
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=build --chown=65532:65532 /out/data /data
USER 65532:65532
ENV WISHMAS_ADDR=:8080 WISHMAS_DATA_DIR=/data
EXPOSE 8080
VOLUME ["/data"]
ENTRYPOINT ["/wishmas"]
