# Infrastructure層の設計規則

対象:

```text
internal/infra/
```

InfrastructureはDomainのRepository・Clock・UUIDGeneratorと、ApplicationのQueryServiceを実装する。

Domain packageからInfrastructure packageへ依存させてはならない。

## Package構成

Feature単位でInfrastructureを分ける。

例:

```text
internal/infra/
├── book/
│   ├── model.go
│   ├── converter.go
│   ├── repository.go
│   ├── queryservice.go
│   └── repository_test.go
├── user/
├── common/
├── database/
├── migrations/
└── testutil/
```

DB modelやRepository実装を技術分類だけの巨大packageへ集約しない。

## Bun Model

Bun modelはInfrastructure専用とする。

Domain EntityへBun tagを付けない。

```text
Domain Model
    ↕ converter
Bun Model
    ↕
PostgreSQL
```

Bun modelはDB mappingだけを担当する。

nullable DB columnにはpointerを使用してよい。

```go
ISBN *string
```

Domainとの境界で`Option[T]`へ変換する。

Persistence modelのtypeは外部packageから必要とされない限りprivateにする。

## Converter

DomainとPersistence modelの変換を明示的に書く。

ORM都合でDomain modelを変更しない。

Persistenceから復元するときはDomainの`Restore`を使用する。

collectionの意味的な順序を保持する。

例:

```text
authors   -> order_index
checkouts -> position
```

round-tripによってDomain状態が変化しないことをintegration testで保証する。

## Repository

Domain packageに定義されたRepository interfaceを実装する。

compile-time assertionを使用してよい。

```go
var _ domain.Repository = (*repository)(nil)
```

`sql.ErrNoRows`は、

```go
return nil, nil
```

へ変換する。

Domain/ApplicationのNotFound errorへ変換しない。

DB errorは可能な限り元のerrorを保持して返す。

`Create`・`Update`・`Delete`を別の操作として実装する。重複キーによる作成失敗、またはversion条件付きの更新・削除が0行だった場合は`common.ErrConflict`を返す。`database.ConflictOnDuplicate`と`database.ValidateNoConflict`に共通処理がある。

## Aggregateの永続化

Aggregateは一単位として永続化する。

複数tableへ書き込む場合はtransactionを使う。

Bookは`Create`で親と子をinsertし、`Update`でversion条件付きの親更新後に既存の子をdeleteしてから現在の子をinsertする。

更新時の子の全置換は同じtransaction内で行う。`Delete`は親をversion条件付きで削除し、子はDBのcascadeで削除する。

実際の性能問題が発生する前から差分永続化を導入しない。

atomic correctnessを優先する。

複数queryでAggregateを復元する場合、一貫したsnapshotが必要なら適切なread transaction / isolation levelを使用する。

## QueryService

Applicationの`query/queryservice.go`に定義したinterfaceを実装し、DTOを直接投影する。Bookの詳細・一覧は、関連データの読取りや件数と一覧の整合性を保つため、read-onlyのrepeatable-read transactionを使用する。検索条件は`dto.BookQuery`、ページングは正規化済みの`application/common.Pagination`を受け取り、一覧の同値順序にはIDを使って安定させる。参照時の権限表示はDomainの読取り用Policyに従う。

## Audit Persistence

Domainの`EntityAudit`とDB audit columnの変換はInfrastructureで行う。

以下のnullable tupleは、

```text
updated_by_id
updated_by_name
updated_at
```

すべてNULLまたはすべてNOT NULLであることをDB constraintでも保証してよい。

## Migration

Bun Migratorによる明示的なSQL migrationを使用する。

Migration fileはGit管理する。

Bun modelからruntimeで自動的にproduction schemaを変更しない。

```text
Bun Model ≠ Migration History
```

schema変更はreview可能なSQLとして残す。

DB constraintはPersistence layerで意味のあるinvariantに使用する。

例:

- PRIMARY KEY
- FOREIGN KEY
- UNIQUE
- NOT NULL
- intentionalなON DELETE CASCADE
- audit tupleの整合性

すべてのDomain ruleをDB constraintへ重複実装する必要はない。

## Infrastructure Service

以下のDomain portをInfrastructureで実装する。

```text
Clock
UUIDGenerator
Repository
```

DI containerはcomposition/infrastructure boundaryで使用してよい。

Domain objectへDI containerを持ち込まない。
