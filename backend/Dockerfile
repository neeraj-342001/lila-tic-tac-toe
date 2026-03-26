FROM heroiclabs/nakama-pluginbuilder:3.26.0 AS builder

WORKDIR /backend

# Copy source before tidy so imports resolve and go.sum is complete.
COPY . .
RUN go mod tidy && go build -trimpath -buildmode=plugin -o ./backend.so

FROM heroiclabs/nakama:3.26.0

COPY --from=builder /backend/backend.so /nakama/data/modules/
COPY local.yml /nakama/data/
COPY entrypoint.sh /nakama/entrypoint.sh
RUN chmod +x /nakama/entrypoint.sh

ENTRYPOINT ["/nakama/entrypoint.sh"]
