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

## 開発時におけるDockerの使用

`task docker-up` は開発用の PostgreSQL（localhost:5432）と Keycloak（localhost:8081）を起動します。

Compose のポートや PostgreSQL のユーザー名・パスワード・DB名は `docker/go-ddd-book-manager/docker-compose.yml` の環境変数で上書きできます。初期値の認証情報は開発専用です。既存の PostgreSQL ボリュームには、初回起動後の認証情報の変更は自動反映されません。

`task run` はコード生成と Docker Compose サービスの起動後に Web API を実行します。開発用の接続設定が自動的に使われ、`DATABASE_URL`、`OIDC_ISSUER`、`OIDC_AUDIENCE`、`CORS_ALLOWED_ORIGINS` を設定すると上書きできます。初期設定に対応する値は以下のとおりです。

`DATABASE_USERNAME` や `DATABASE_PASSWORD` にURL上でエンコードが必要な文字を使う場合は、エンコード済みの `DATABASE_URL` を明示してください。

```sh
export DATABASE_URL='postgres://bookmanager:bookmanager@localhost:5432/bookmanager?sslmode=disable'
export OIDC_ISSUER='http://localhost:8081/realms/book-manager'
export OIDC_AUDIENCE='book-manager-api'
export CORS_ALLOWED_ORIGINS='http://localhost:3000'
```

統合テストには Docker デーモンが必要です。`task test-integration` は現在の Docker コンテキストから接続先を取得します。別のソケットを指定する場合は、`DOCKER_HOST` を設定して `go test -tags=integration ./internal/infra/...` を直接実行してください。

## マイグレーション

PostgreSQL を起動してから、次のタスクを実行します。`migrate-up` は `DATABASE_URL` が未設定の場合、Composeの開発用接続設定を使います。`migrate-down` には対象DBの `DATABASE_URL` を明示してください。

```sh
task migrate-up    # 未適用の SQL マイグレーションを適用
task migrate-down  # 最後に適用したマイグレーショングループをロールバック
```

空の up/down SQL ファイルは次のコマンドで作成できます。

```sh
task migrate-create -- add_example
```

Web API の起動には `DATABASE_URL`、Keycloak の issuer を指定する `OIDC_ISSUER`、API 用 audience の `OIDC_AUDIENCE` が必要です。待受アドレスは `HTTP_ADDR`（省略時 `:8080`）で指定できます。Keycloak の RS256 Access Token を使用し、`realm_access.roles` に `Admin` があるユーザーを管理者として同期します。

ブラウザーから別オリジンでアクセスする場合は、`CORS_ALLOWED_ORIGINS` に許可するオリジンをカンマ区切りで指定します（例: `https://app.example.com,https://admin.example.com`）。直接起動時に未設定なら CORS を有効にしません。`task run` の開発用デフォルトは `http://localhost:3000` です。Bearer トークンの `Authorization` と `Content-Type` ヘッダーに対応し、Cookie を使うクロスオリジン認証は許可しません。

### 開発用 Keycloak

Docker Compose の Keycloak は初回起動時に `docker/go-ddd-book-manager/realm.json` をインポートします。開発用の realm は `book-manager`、API の audience とトークン取得用 client ID は `book-manager-api` です。`OIDC_ISSUER` には `http://localhost:<KEYCLOAK_PORT_OUTER>/realms/book-manager`、`OIDC_AUDIENCE` には `book-manager-api` を設定してください。

| ユーザー名 | パスワード | 権限                         |
| ---------- | ---------- | ---------------------------- |
| `user`     | `password` | 通常ユーザー                 |
| `admin`    | `password` | 管理者（realm role `Admin`） |

どちらも開発専用のアカウントです。たとえば次のコマンドでアクセストークンを取得できます。

```sh
curl -X POST "http://localhost:${KEYCLOAK_PORT_OUTER:-8081}/realms/book-manager/protocol/openid-connect/token" \
  -d 'client_id=book-manager-api' \
  -d 'grant_type=password' \
  -d 'username=user' \
  -d 'password=password'
```

Keycloak は既に存在する realm を起動時インポートで上書きしません。`realm.json` の変更を既存環境に反映するには、必要なデータを退避したうえで Keycloak の管理画面から該当 realm を削除し、Keycloak を再起動してください。
