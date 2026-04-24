####################### Build stage #######################
# golang:<version>-alpine は、Alpine Linux プロジェクトをベースにしている。
# イメージサイズを最小にするため、git、gcc、bash などは、Alpine-based のイメージには含まれていない。
FROM golang:1.24.6-alpine3.22 AS builder
# 作業ディレクトリの定義をする。今回は、app ディレクトリとした。
WORKDIR /app
# go.mod と go.sum を app ディレクトリにコピー
COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .
# 指定されたモジュールをダウンロードする。
RUN go mod download && go mod verify
# ルートディレクトリの中身を app フォルダにコピーする
# 実行ファイルの作成
# -o はアウトプットの名前を指定。
# ビルドするファイル名を指定（今回は main.go）。
RUN go build -o /bin/main ./cmd/main.go


####################### Run stage #######################
# Goで作成したバイナリは Alpine Linux 上で動く。
# alpineLinux とは軽量でセキュアな Linux であり、とにかく軽量。
FROM alpine:latest
# 作業ディレクトリの定義
WORKDIR /app

RUN apk --update add tzdata && \
    cp /usr/share/zoneinfo/Asia/Tokyo /etc/localtime && \
    apk del tzdata && \
    rm -rf /var/cache/apk/*
RUN apk --no-cache add ca-certificates
# Build stage からビルドされた main だけを Run stage にコピーする。（重要）
COPY --from=builder /bin/main /bin/main
# ローカルの .env をコンテナ側の app フォルダにコピーする
COPY .env ./
# EXPOSE 命令は、実際にポートを公開するわけではない。
# これは、イメージを構築する人とコンテナを実行する人の間で、どのポートを公開するかについての一種の文書として機能する。
# 今回、docker-compose.yml において、api コンテナは 8080 ポートを解放するため「8080」とする。
EXPOSE 8080
# バイナリファイルの実行
CMD [ "/bin/main" ]