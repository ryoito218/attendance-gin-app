# ============================
# ビルド用ステージ
# ============================
FROM golang:1.25 AS builder

WORKDIR /app

# 依存関係だけ先にコピーして go mod download
COPY go.mod go.sum ./
RUN go mod download

# ソースコードをコピー
COPY . .

# 本番用バイナリをビルド
# -o /app/app という名前の実行ファイルを作成
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/app ./cmd/app

# ============================
# 実行用ステージ
# ============================
FROM gcr.io/distroless/base-debian12

WORKDIR /app

# ビルド済みバイナリをコピー
COPY --from=builder /app/app /app/app

# public ディレクトリ（フロント用）もコピー
COPY --from=builder /app/public /app/public

# 環境変数（デフォルト値）
ENV DB_HOST=db
ENV DB_USER=appuser
ENV DB_PASSWORD=apppass
ENV DB_NAME=attendance_db

# アプリのポート
EXPOSE 8080

# 実行コマンド
ENTRYPOINT ["/app/app"]