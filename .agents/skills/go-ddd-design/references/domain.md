# Domain層の設計規則

対象:

```text
internal/domain/
```

## 1. Package構成

packageは原則としてFeatureまたはAggregateの境界とする。

例:

```text
internal/domain/
├── book/
├── user/
└── common/
```

以下のような技術分類だけを目的としたsubpackageは作らない。

```text
entities/
values/
repositories/
services/
```

同一package内では、型の種類ではなくDomain conceptごとにファイルを分ける。

例:

```text
book/
├── book.go
├── author.go
├── checkout.go
├── values.go
├── permission_policy.go
├── repository.go
└── errors.go
```

Aggregate Root直属の小さなscalar Value Objectは`values.go`へまとめてよい。

独立した概念や振る舞いを持つものは、その概念のファイルへ置く。

`common`には複数Featureで本当に共有されるDomain conceptだけを置く。

単なる重複排除を理由に`common`へ移動しない。

## 2. Value Object

Value Objectは値として扱う。

内部fieldは不変性を守るためにprivateにする。

```go
type Title struct {
	value string
}

func (t Title) Value() string {
	return t.value
}
```

外部入力から作成するときはvalidationを行うconstructorを使う。

```go
NewTitle(...)
NewISBN(...)
NewDescription(...)
```

生成済みValue Objectは常に有効な値を表す。

永続化データから復元するときは`RestoreX`を使用する。

```go
RestoreTitle(...)
RestoreAuthor(...)
```

Value Object自身の中に「値が存在しない状態」を持たせない。

## 3. Optionality

optionalかどうかは、その値を所有するEntity/Aggregateの関係として表現する。

```go
type ISBN struct {
	value string
}

type Book struct {
	isbn option.Option[ISBN]
}
```

Value Objectが存在するなら有効な値を持つ。

Domain上のoptionalには原則としてrepository-localの`option.Option[T]`を使う。

optionalを表すためだけに`*T`を使用しない。

## 4. Entity / Aggregate Root

Value Objectはvalue semanticsを基本とする。

状態変更を行うEntity/Aggregate Rootのmethodにはpointer receiverを使用する。

```go
func (b *Book) Update(...) error
```

「ポインタ禁止」ではなく、identityまたはmutationの意味がある場所でポインタを使う。

Entityのfieldはprivateにする。

collection getterでは内部sliceを直接返さない。

必要に応じてcloneして返す。

```go
func (a AuthorList) Values() []Author {
	return lo.Clone(a.values)
}
```

## 5. NewとRestore

`New`は通常のDomain上の作成操作に使う。

作成時に必要なinvariantと認可を保証する。

`Restore`は永続化された状態のrehydrationに使う。

`Restore`ではユーザー操作に対する認可を行わない。

ただし、Domainとして存在不能な状態を黙って受け入れてはならない。

例:

- 不正なenum
- audit情報の一部だけが存在する状態

これらは復元時にもerrorにしてよい。

## 6. DomainContext

以下が必要なDomain操作には`common.DomainContext`を使う。

- Actor
- 現在時刻
- UUID生成

純粋なDomain methodへ`context.Context`を渡さない。

`context.Context`はRepository、HTTP、DB、外部サービスなどI/O境界で使用する。

時刻とUUID生成はport越しに行い、Domainをテスト可能にする。

## 7. 認可

Domain状態を保護する認可はDomain behaviorとして扱う。

Application層だけのチェックに依存しない。

共通interface:

```go
type PermissionPolicy interface {
	CanCreate() bool
	CanUpdate() bool
	CanDelete() bool
}
```

具体的Policyは対象Featureの近くに置く。

例:

```text
bookPermission
ownerChangePermission
userProvisioningPermission
```

可能で自然な場合は、認可をbusiness validationより先に行う。

ただし、認可自体がAggregateの現在状態に依存する場合まで、無理に順序を統一しない。

重要なのは、権限のないActorがDomain状態を変更できないことである。

## 8. SystemActor

`SystemActor`は内部の自動処理を表す。

通常のUserとは別の概念として扱う。

```go
common.NewSystemActor()
```

通常Userに`SystemRole`を保存してはならない。

zero UUIDが一致しただけで通常操作の権限を得ないようにする。

SystemActorの権限は明示的に定義する。

## 9. User

認証・Identity ProviderをUser identityのSSoTとする。

Applicationは以下の値を独自に編集しない。

- name
- email
- identity role

Domainの`User`は、Identity Provider上の状態をApplicationへ投影したものとして扱う。

基本的なlifecycleは以下。

```text
New    -> IdentityをApplicationへ取り込む
Sync   -> Identity Providerの状態と同期する
Delete -> Application側の投影を削除する
```

これらはSystemActorによって行う。

Identity Provider所有fieldを変更する一般的な`User.Update`を追加しない。

将来Application固有の編集可能プロフィールが必要になった場合は、Identityとは別概念としてmodelingする。

## 10. Audit

Entity/Aggregateの通常の作成・更新履歴には`common.EntityAudit`を使う。

updated auditは以下の全項目が揃っているか、すべて存在しない状態でなければならない。

```text
updated_by_id
updated_by_name
updated_at
```

すべてのDomain eventで`EntityAudit`を更新する必要はない。

BookのCheckout/Returnは独自のCheckout履歴を持つため、Bookの一般的な`EntityAudit`を更新しない。

この違いを単なる一貫性のために統一しない。

## 11. Domain Error

validation情報をprogramから扱う必要がある場合は`ValidationError`を使用する。

```text
field
code
```

errorの判定には`errors.As`を使う。

`Error()`文字列をmachine-readableな判定に利用しない。

例:

```text
book.title      / empty
book.author[1]  / too_long
user.role       / invalid
```

Domain上の認可失敗には`common.ErrForbidden`を使う。

lower layerのerrorへ文脈を追加する場合は`%w`でwrapする。

## 12. Repository interface

Repository interfaceはDomainのFeature packageに定義する。

```go
type Repository interface {
	FindByID(ctx context.Context, id uuid.UUID) (*Book, error)
	Save(ctx context.Context, book Book) error
}
```

mutableなAggregate Rootはpointerで返してよい。

`FindByID`のcontractは以下とする。

```text
entity != nil, err == nil -> found
entity == nil, err == nil -> absent
entity == nil, err != nil -> storage/infrastructure failure
```

「存在しない」というstorage上の事実を、InfrastructureでUseCase上のNotFoundへ変換しない。

NotFoundとして扱うかはApplication層が判断する。
