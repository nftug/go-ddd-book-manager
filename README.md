# go-ddd-book-manager

Go でドメイン駆動設計（DDD）を実践する書籍管理アプリケーションです。

## ディレクトリ構成

```text
cmd/                    アプリケーションとマイグレーションのエントリーポイント
internal/common/        共通のユーティリティ
internal/domain/        書籍・ユーザーなどのドメインモデル
internal/application    アプリケーション層
internal/webapi         Web API (プレゼンテーション層)
internal/infra/         PostgreSQL のリポジトリ、DI、マイグレーション
docker/                 開発用 Docker Compose 設定
```

## 開発環境

- Go（必要なバージョンは `go.mod` を参照）
- [Task](https://taskfile.dev/)
- Docker と Docker Compose（開発用サービスや統合テストで使用）
- PostgreSQL（マイグレーションやアプリケーションのデータ保存で使用）

主なタスクはリポジトリルートで実行します。

```sh
task test              # コード生成とテスト
task test-integration  # Dockertest と PostgreSQL による統合テスト
task fix               # コード生成とフォーマット
task build             # bin/app をビルド
```

`task run` はコード生成と Docker Compose サービスの起動後に `cmd` を実行しますが、アプリケーションサーバーはまだ実装されていません。Compose を使用する場合は、`docker/go-ddd-book-manager/docker-compose.yml` に記載された環境変数を設定してください。

統合テストには Docker デーモンが必要です。`task test-integration` は現在の Docker コンテキストから接続先を取得します。別のソケットを指定する場合は、`DOCKER_HOST` を設定して `go test -tags=integration ./internal/infra/...` を直接実行してください。

## マイグレーション

PostgreSQL の接続文字列を `DATABASE_URL` に設定してから、次のタスクを実行します。

```sh
task migrate-up    # 未適用の SQL マイグレーションを適用
task migrate-down  # 最後に適用したマイグレーショングループをロールバック
```

空の up/down SQL ファイルは次のコマンドで作成できます。

```sh
task migrate-create -- add_example
```

Web API の起動には `DATABASE_URL`、Keycloak の issuer を指定する `OIDC_ISSUER`、API 用 audience の `OIDC_AUDIENCE` が必要です。待受アドレスは `HTTP_ADDR`（省略時 `:8080`）で指定できます。Keycloak の RS256 Access Token を使用し、`realm_access.roles` に `Admin` があるユーザーを管理者として同期します。
