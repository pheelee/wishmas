# @impl CON-001-AC1 CON-001-AC2 CON-001-AC3 CON-002 CON-004
FROM golang:1.26.9-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY cmd/ ./cmd/
COPY internal/ ./internal/
# CSS/JS and IANA timezone data are embedded; no runtime toolchain is needed.
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/wishmas ./cmd/wishmas \
    && mkdir -p /out/data

FROM scratch
COPY --from=build /out/wishmas /wishmas
COPY --from=build --chown=65532:65532 /out/data /data
USER 65532:65532
ENV WISHMAS_ADDR=:8080 WISHMAS_DATA_DIR=/data
EXPOSE 8080
VOLUME ["/data"]
ENTRYPOINT ["/wishmas"]
