# go-ddd-book-manager

Go でドメイン駆動設計（DDD）を実践する書籍管理アプリケーションです。

## 開発を始める

Go（バージョンは `go.mod` を参照）、[Task](https://taskfile.dev/)、Docker Compose が必要です。

```sh
task run         # 開発用サービスを起動して API を実行
task dev         # 同じ構成で変更を監視し、API を再起動
task migrate-up  # 未適用のマイグレーションを適用
```

API は `localhost:8080`、PostgreSQL は `localhost:5432`、Keycloak は `localhost:8081` で利用できます。`task run` と `task dev` はコード生成と Docker Compose の起動も行います。`task dev` の監視中に `go:generate` の対象を変更した場合は、別途 `task generate` を実行してください。

開発用の接続先は `DATABASE_URL`、`OIDC_ISSUER`、`OIDC_AUDIENCE`、`CORS_ALLOWED_ORIGINS` で上書きできます。API の待受アドレスは `HTTP_ADDR`（初期値 `:8080`）で指定します。直接起動する場合は `DATABASE_URL`、`OIDC_ISSUER`、`OIDC_AUDIENCE` が必要です。CORS は `CORS_ALLOWED_ORIGINS` にカンマ区切りで許可するオリジンを指定すると有効になります。

Compose のポートや PostgreSQL の認証情報は [docker-compose.yml](docker/go-ddd-book-manager/docker-compose.yml) の環境変数で変更できます。既存の PostgreSQL ボリュームには認証情報の変更は自動反映されません。認証情報に URL エンコードが必要な文字を使う場合は、エンコード済みの `DATABASE_URL` を指定してください。

## 開発用認証

Keycloak の開発用 realm は `book-manager`、client ID と API audience は `book-manager-api` です。`user` と `admin` のパスワードはいずれも `password` です。これらは開発専用のアカウントです。

```sh
curl -X POST 'http://localhost:8081/realms/book-manager/protocol/openid-connect/token' \
  -d 'client_id=book-manager-api' -d 'grant_type=password' \
  -d 'username=user' -d 'password=password'
```

API は Keycloak の RS256 Access Token を使用し、realm role `Admin` を持つユーザーを管理者として同期します。Keycloak は既存の realm を起動時インポートで上書きしません。[realm.json](docker/go-ddd-book-manager/realm.json) の変更を反映するには、必要なデータを退避し、管理画面で realm を削除してから Keycloak を再起動してください。

## 主なタスク

```sh
task fix                 # コード生成とフォーマット
task test                # コード生成と全テスト
task test-integration    # Docker と PostgreSQL を使う統合テスト
task openapi             # 試作用 swagger.json を生成
task build               # bin/app をビルド
task migrate-down        # 最後のマイグレーショングループをロールバック
task migrate-create -- add_example  # 空のマイグレーションを作成
```

`migrate-down` には対象 DB の `DATABASE_URL` を明示してください。統合テストは現在の Docker コンテキストを使用します。別のソケットを使う場合は `DOCKER_HOST` を設定し、`go test -tags=integration ./internal/infra/...` を直接実行してください。

OpenAPI 定義はハンドラーのコメントから生成されます。コメントと実装の一致は自動では保証されません。

## ログとトレース

`task run` と `task dev` はログとトレースを OTLP/HTTP に送信します。[Grafana](http://localhost:3001)（ユーザー名・パスワードは `admin`）の Explore で Loki のログと Tempo のトレースを検索できます。送信先は `OTEL_EXPORTER_OTLP_ENDPOINT` で変更できます。直接起動時に送信先が未設定ならログは標準出力のみです。

アクセスログとトレースにはクエリ文字列が含まれます。`access_token`、`token`、`password`、`client_secret` の値はマスクしますが、それ以外のクエリ値は記録されます。リクエスト本文と Authorization ヘッダーは記録しません。

## ディレクトリ構成

```text
cmd/                   アプリケーションとマイグレーションの入口
internal/domain/       ドメインモデル
internal/application/  ユースケース
internal/webapi/       HTTP API
internal/infra/        永続化、DI、マイグレーション
internal/common/       共通ユーティリティ
docker/                開発用 Docker Compose 設定
```
