# Application層の設計規則

Application層はUseCaseを実現し、Domainと外部I/Oを調停する。

Domain invariantそのものをApplication層へ移さない。

## Package構成と依存方向

対象は`internal/application/`とする。

Feature単位で分け、Feature内では更新系を`command`、参照系を`query` packageに分ける。

```text
internal/application/
├── book/
│   ├── command/
│   │   ├── create_book.go
│   │   └── update_book.go
│   ├── query/
│   │   ├── get_book_details.go
│   │   ├── get_book_list.go
│   │   └── queryservice.go
│   └── dto/
│       ├── identity.go
│       ├── query.go
│       ├── request.go
│       └── response.go
├── user/
│   ├── command/
│   │   └── sync_user.go
│   ├── query/
│   │   ├── get_user_details.go
│   │   └── queryservice.go
│   └── dto/
│       ├── request.go
│       └── response.go
└── common/
    ├── errors.go
    └── pagination.go
```

上記は配置例とし、必要になったファイルのみ追加する。

`command`はDomain操作とRepositoryによる更新を担当し、`query`は参照UseCaseとQueryServiceを担当する。

Featureごとに`dto` packageを設け、作成・更新などの入力DTOは`dto/request.go`、返却DTOは`dto/response.go`へ置く。

CommandとQueryで同じ表現が必要な場合は、`dto` packageの型を共有する。

検索条件は`dto/query.go`、共通のページング型は`common/pagination.go`、QueryService interfaceは`query/queryservice.go`へ置く。

親のFeature packageに子package間で共有する型を置く場合は、`command`・`query`へ依存させない。

`command`と`query`は`dto`に依存してよいが、互いに依存させない。`dto`から`command`・`query`へは依存させない。

ApplicationはDomainへ依存してよいが、Infrastructure、Bun、Echoへ依存しない。

InfrastructureはApplicationが定義したQueryService interfaceを実装してよい。

## UseCase Service

原則としてUseCaseごとの小さなServiceを定義し、`Execute`で実行する。

更新系は以下を調停する。

1. 入力をDomainのconstructorでValue Objectへ変換する
2. 必要なAggregateをRepositoryから取得する
3. 不在をUseCaseの意味に応じて扱う
4. Domain methodを実行する
5. Repositoryへ保存する

処理順はUseCaseに応じて決め、Domainの認可やvalidationをApplicationへ複製しない。

Repository interfaceは既存のDomain packageのものを使い、同じ目的のinterfaceをApplicationへ重複定義しない。現在のRepository契約は`FindByID`・`Create`・`Update`・`Delete`で、UseCaseは操作に応じて呼び分ける。

## 入力DTOとIdentity

path parameterに対応する対象識別情報、query parameterに対応する検索条件、作成・更新入力は、Applicationの入力契約として定義する。

対象識別情報はFeatureの`dto/identity.go`へ置き、`command`・`query`で共有する。

```go
type BookIdentity struct {
	BookID uuid.UUID `param:"book_id"`
}
```

詳細取得・更新・削除・貸出・返却など、同じ対象を識別するUseCaseで`BookIdentity`を受け取る。

入力DTOはバインディング用structとしても使用し、HTTP入力との対応を示す`param`・`query`・`json`タグを持ってよい。

| 情報                                          | 所有する層   |
| --------------------------------------------- | ------------ |
| `dto.BookIdentity`、一覧検索条件、更新入力    | Application  |
| `/books/:book_id`などのroute定義              | Presentation |
| `param`・`query`・`json`タグ、binding用struct | Application  |
| 入力元を指定したbinding                       | Presentation |
| path/query文字列からUUIDや数値へのparse       | Presentation |

PresentationはApplicationの入力DTOへ入力元を明示してbindする。タグは許容するが、ApplicationはEchoの型やHTTP requestに依存しない。

ページ番号・ページサイズ・検索条件の組み合わせなど、UseCaseとしての入力検証はApplicationで行う。

`common.Pagination.Normalize`で未指定のpage/page_sizeを1/10に補い、負数、100を超えるpage_size、offset計算のoverflowを拒否する。QueryServiceには正規化済みの値を渡す。

タイトルやISBNなどのDomain validationは、Domainのconstructorに委ねる。

optionalな入力は既存の`option.Option[T]`を利用してよい。

部分更新を導入する場合は「未指定」と「値の削除」を区別する入力契約を定義する。

## 参照系とQueryService

参照系は、ApplicationのFeature内の`query` packageにQueryService interface、`dto` packageに返却DTOを定義し、InfrastructureでQueryServiceを実装する。

一覧・詳細表示のためだけにAggregate全体を復元する必要はない。

検索・ソート・ページングはInfrastructureのqueryへ委ね、必要なデータをDTOへ投影する。Book一覧は件数と一覧を同一のread-only repeatable-read transactionで取得する。

一覧の順序は同値の場合のID順なども含めて安定させる。

Domain EntityやBun modelをレスポンスとして直接公開しない。

Domainの`Option[T]`などはDTO境界で必要な表現へ変換する。

詳細取得の不在と、検索結果が空であることは区別して扱う。

## ActorとDomainContext

`context.Context`はRepository、QueryServiceなどのI/Oへ渡す。

Domain操作には、実行時のActorと注入されたClock・UUIDGeneratorから`common.DomainContext`を作成して渡す。

ActorはUseCase実行ごとに引数で受け取り、共有Serviceのfieldへ保持しない。

書籍作成時のownerは原則として実行者から作り、任意のowner指定や所有者変更は別のUseCaseとして認可する。

## 認可

Domain状態を変更する操作の認可はDomain methodで保証する。

貸出履歴の管理者限定公開など、参照UseCaseへのアクセス制御はApplicationで行う。

DTOに`canUpdate`・`canDelete`などを含める場合は、Domainと同じ認可規則から算出する。

そのために必要な読み取り用の認可APIはDomainに設け、QueryService実装へ認可規則を複製しない。

表示用の権限情報があっても、実際の更新時にはDomainで認可する。

## UserとIdentity Providerの同期

JWTの検証やclaimのparseはPresentation・認証adapter側の責務とする。

ApplicationのSyncUserは検証済みIdentity情報を受け取り、Userの投影を同期して`UserResponseDTO`を返す。

`UserResponseDTO`は`user/dto/response.go`へ置き、必要に応じてUser詳細取得などでも共有する。

```text
検証済みIdentity情報
  → user.Repository.FindByID
  → 不在ならSystemActorでUser.New
  → 存在すればSystemActorでUser.Sync
  → 必要な保存
  → 同期後のUserからUserResponseDTOを作成して返す
```

通常の業務API向けのActor解決は同期とは別の読み取り処理とし、既存Userを取得して`User.Actor()`を返す。Userを自動作成・同期しない。

name・email・roleのすべてを同期対象とし、emailだけの変更も取り込む。

変更のない同期でDomainの更新監査情報を変更しない。

通常のクライアント入力からSystemActorや任意のidentity roleを選ばせない。

IdP所有fieldを編集する一般ユーザー向けの`User.Update` UseCaseを追加しない。

## Repository結果の意味付け

Repositoryの`FindByID`が返す、

```text
nil, nil
```

は「storage上、そのEntityが存在しない」ことを意味する。

それをUseCase上の`NotFound`として扱うかはApplication層で判断する。

Infrastructure層で`sql.ErrNoRows`をDomain/Applicationの`ErrNotFound`へ変換しない。

## Error変換

lower layerのerrorにApplication上の文脈を追加する場合は`%w`でwrapする。

DomainのValidationErrorやErrForbiddenを文字列へ潰さない。

HTTP statusやProblem Detailsへの変換はPresentation層の責務とする。

## Domain操作

Application層はDomain Entityのfieldを直接変更しない。

Domain methodを通して操作する。

Domainで保証すべきinvariantや認可をApplicationだけで保証しない。

## Transaction

Aggregate内部の永続化transactionはRepository実装側で完結させる。

複数Aggregateや複数RepositoryにまたがるUseCase transactionが必要になった場合のみ、Application側でtransaction boundaryを検討する。

必要になる前からUnit of Work等を導入しない。

Repositoryの更新処理内のtransactionだけでは、取得から保存までの同時更新は保護されない。

現在のRepositoryは`version`を使った楽観ロックを実装し、更新・削除が0行なら`common.ErrConflict`を返す。Applicationはこのerrorを保持して返す。

## 依存注入

Serviceのproviderは`do.Injector`から必要なinterfaceを解決し、Serviceのfieldに設定する。

既存の`samber/do`によるprovider登録は`internal/application/injector.go`へまとめる。各providerは依存を解決してServiceへ保持し、Serviceの処理中にcontainerから依存を取得しない。

Rustの`Arc`やRegistry構造をそのまま再現するためだけの抽象を追加しない。

## テスト

Applicationのunit testでは、必要に応じてRepository・QueryServiceのfakeと固定Clock・UUIDGeneratorを使う。

主に以下を確認する。

- 不在をUseCaseのNotFoundへ変換すること
- Domainの認可・validation失敗時に保存しないこと
- I/Oのerrorを保持して返すこと
- 参照UseCaseのアクセス制御と検索条件の検証
- Userの新規取り込み、emailのみの変更、変更なしの同期

Domainの個々のinvariantはDomain testで、SQL・DTO投影・競合検出はInfrastructureのintegration testで検証する。
