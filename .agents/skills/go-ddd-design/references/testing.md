# テスト設計規則

テストは単なるcoverage確保ではなく、DomainやArchitecture上の判断を固定するために書く。

## Domain Test

原則としてexternal test packageを使用する。

```go
package book_test
```

private fieldではなくobservable behaviorをテストする。

特に以下をテストする。

- constructor validation
- Domain invariant
- Permission behavior
- defensive copy
- Option semantics
- no-op時のAudit
- Role invariant
- Checkout / Return
- Restore behavior

設計判断として重要な振る舞いは明示的なtestとして残す。

## Infrastructure Test

DB behaviorが関係するものはintegration testで確認する。

実際のPostgreSQLを使用する。

integration testにはbuild tagを使用する。

```go
//go:build integration
```

Repository testでは必要に応じて以下を確認する。

- Migration適用
- Create / FindByID round-trip
- Update
- absence
- collection ordering
- transaction rollback
- Delete
- database constraint
- Domain / Persistence変換

Repositoryから取得したAggregateが保存前のDomain状態と等価になることを確認する。

## Transaction Test

複数tableを更新するRepositoryでは、途中の失敗でtransaction全体がrollbackされることを確認する。

単に正常系CRUDだけをテストして終わらせない。

## Test Utility

共通のintegration test setupは、

```text
internal/infra/testutil
```

へ置いてよい。

`testutil`はProduction codeではなくtest support codeとして扱う。

coverage評価から除外してよい。

## Generated Code

generated codeそのものの実装詳細を重複してテストしない。

生成元となる宣言と、Application/Domainから必要な振る舞いをテストする。

## 実行

Go sourceを変更した場合はリポジトリルートで先に `task fix` を実行し、その後に編集したGoファイルが属する各packageのテストを実行する。

```sh
go test ./internal/application/book/query
```

InfrastructureのDB動作を確認する場合は、対象packageのintegration testを実行する。

```sh
go test -tags=integration ./internal/infra/book
```

`task test` は `go test ./...` を実行するため、明示的な指示がない限り使わない。
