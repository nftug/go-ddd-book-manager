---
name: go-ddd-design
description: このリポジトリでGoによるDDD設計・実装・レビューを行う際に使用する。Domain、Application、Infrastructure、Repository、Bun永続化、マイグレーション、テストに関する既存の設計規則を適用する。
---

# Go DDD 設計

このSkillは、このリポジトリでDDDに関わるコードを追加・変更・レビューするときに使用する。

現在の実装を設計判断の基準とする。記述とコードが異なる場合は、関連する実装とテストを確認してから記述を更新する。一般的なGoプロジェクトでよく使われるという理由だけで別の設計パターンへ置き換えない。

## 作業前に読むもの

変更対象に応じて以下を読む。

- Domain層を変更する場合: `references/domain.md`
- Application層を変更する場合: `references/application.md`
- Infrastructure層、Repository実装、Bun、DB、Migrationを変更する場合: `references/infrastructure.md`
- Web API層を変更する場合: `references/webapi.md`
- テストを追加・変更する場合: `references/testing.md`

複数レイヤーにまたがる変更では、該当するすべてを読むこと。

## 基本方針

変更前に以下を確認する。

1. その責務を持つべきレイヤーはどこか
2. Domain invariantは何か
3. 認可をどこで保証するか
4. 値・Entity・Aggregate Rootのどれとして扱うか
5. optionalityを誰が所有するか
6. 永続化都合がDomainへ漏れていないか

既存の設計を拡張し、同じ目的の別抽象を並立させない。

必要以上に汎用化・抽象化しない。

## 共通Go規則

### 命名規則

Goの命名規則に従う。

- `ID`, `UUID`, `URL` を使う
- `Id`, `Uuid`, `Url` は使わない
- getterに `Get` を付けない
- interfaceに `IRepository` のような `I` prefixを付けない

### interfaceと依存

interfaceは利用側のpackageで定義し、小さく保つ。

### 依存性注入

依存性注入には `github.com/samber/do/v2` を使用する。

providerの登録は `internal/infra/injector.go` と `internal/application/injector.go` にまとめる。

各providerが依存を解決してServiceのfieldへ渡す。`Execute`などの業務処理中にcontainerから依存を取得しない。Domain objectにcontainerを持ち込まない。

interfaceを通して利用する依存は、そのinterface型で登録・解決する。injectorを作成した呼び出し元は、終了時に `Shutdown()` を呼ぶ。

既存のproviderと登録の例:

```go
func NewRepository(i do.Injector) (domain.Repository, error) {
	return &repository{db: do.MustInvoke[*bun.DB](i)}, nil
}

func NewGetBookDetails(i do.Injector) (*GetBookDetails, error) {
	return &GetBookDetails{q: do.MustInvoke[QueryService](i)}, nil
}
```

ルートの `Inject` でproviderを登録する。Featureごとに `injector.go` を増やさない。

```go
func Inject(i do.Injector) do.Injector {
    do.Provide(i, bookinfra.NewRepository)
    do.Provide(i, bookinfra.NewQueryService)
    return i
}
```

構築時にInfrastructureとApplicationの各 `Inject` を呼び、作成元で `Shutdown()` を呼ぶ。

### その他

`samber/lo` は処理が明確になる場合に使用してよいが、通常のloopの方が明快ならloopを使う。

## 検証

Goコードを変更した後は、リポジトリルートで `task fix` を実行し、その後に編集したGoファイルが属する各packageだけを `go test` する。`task test` は `go test ./...` を実行するため、明示的な指示がない限り使わない。

```sh
task fix
go test ./internal/application/book/command
```

上記のpackage pathは例。複数packageを編集した場合は各packageをテストする。InfrastructureやDBの動作を確認する必要がある場合は、対象packageのintegration testを実行する。

```sh
go test -tags=integration ./internal/infra/book
```

DockerやPostgreSQLを利用できずintegration testを実行できなかった場合は、その旨を明示する。

作業完了前に以下を確認する。

- DomainからInfrastructureへの依存がない
- generated codeが最新
- 永続化のround-tripでDomain状態が失われない
- Domain invariantと認可が維持されている
- 無関係な変更が混入していない
