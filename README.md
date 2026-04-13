# チャレキャラ Backend

## プロジェクト概要

部員の作成した作品やブログの投稿や閲覧ができる Web アプリ、Toybox のバックエンドです。  
既に動いている Web アプリのリプレイスを目指すものになります。(https://github.com/Kyutech-C3/toybox-server)
/docs/design-doc.md に[設計ドキュメント](docs/design-doc.md)があります。

## 開発環境の準備

### 開発に必要な CLI ツールのインストール

以下のコマンド実行後、wire, migrate, swag mockgenコマンドがそれぞれ実行できるようになることを確認してください。実行できない場合 `go env GOPATH`で表示されるディレクトリ配下の bin へのパスが通ってない可能性が高いです。

```
$ go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
$ go install github.com/google/wire/cmd/wire@latest
$ go install github.com/swaggo/swag/cmd/swag@latest
$ go install go.uber.org/mock/mockgen@latest
```

### プロジェクトのセットアップ

```
# レポジトリをクローンする
$ git clone git@github.com:simesaba80/toybox-back.git

# 作業ディレクトリに移動する
$ cd toybox-back

# 依存モジュールをインストールする
$ go mod download

# .envの作成と書き込み
$ cp .env.example .env

# サーバーを起動
$ docker compose up -d

# 本番環境の場合
$ docker compose -f compose.yaml -f compose.app.yaml up -d --build

# マイグレーションの実行
$ migrate --path db/migrations --database '<DBの接続文字列>' up

# データの挿入(初回のみ)
# バックアップのSQLをプロジェクト直下に配置後
$ ./scripts/movedata.sh

# 以下のURLを開き、起動しているか確認
http://localhost:8080/

#サーバーの停止
$ docker compose down

```

※依存先が増えた場合は適切に/internal/di/wire.go に記述し、`wire ./internal/di`を実行すること

### データベースのセットアップ

データベースのマイグレーションには golang-migrate を使っています。

基本的な使い方は以下の記事を参考にしてください。
https://zenn.dev/farstep/books/f74e6b76ea7456/viewer/4cd440
マイグレーション時 DSN の指定が必要になります。コンテナのネットワーク外から実行するので.env で記入した DSN と HOST の値が異なることに注意してください

### API ドキュメントの更新

http://localhost:8080/swagger/index.html にアクセスすることで API ドキュメントを確認できる。

API ドキュメントの更新

```
$ swag init -g cmd/main.go --parseDependency
```

## ブランチ命名規則

ブランチの命名規則は以下のようにすること

```
[接頭辞]/issue番号-内容-内容
```

内容はケバブケースを使用し、接頭辞には以下のような内容を使用すること。

| 接頭辞   | 内容                                                                                                                       |
| :------- | :------------------------------------------------------------------------------------------------------------------------- |
| fix      | 既存の機能の問題を修正する場合に使用します。                                                                               |
| hotfix   | 緊急の変更を追加する場合に使用します。                                                                                     |
| feat     | 新しい機能やファイルを追加する場合に使用します。                                                                           |
| update   | 既存の機能に問題がないが、修正を加えたい場合に使用します。                                                                 |
| change   | 仕様変更により、既存の機能に修正を加えた場合に使用します。                                                                 |
| refactor | コードの改善をする場合に使用します。                                                                                       |
| delete   | ファイルを削除する場合や、機能を削除する場合に使用します。                                                                 |
| rename   | ファイル名を変更する場合に使用します。                                                                                     |
| move     | ファイルを移動する場合に使用します。                                                                                       |
| upgrade  | バージョンをアップグレードする場合に使用します。                                                                           |
| revert   | 以前のコミットに戻す場合に使用します。                                                                                     |
| docs     | ドキュメントを修正する場合に使用します。                                                                                   |
| style    | コーディングスタイルの修正をする場合に使用します。                                                                         |
| test     | テストコードを修正する場合や、テストコードを追加する場合に使用します。                                                     |
| chore    | ビルドツールやライブラリで自動生成されたものをコミットする場合や、上記の接頭辞に当てはまらない修正をする場合に使用します。 |

## コミット命名規則

```
[接頭辞]: #issue番号 やったこと
```

上記の命名規則でコミットすること、接頭辞はブランチ命名規則のものと同じ
