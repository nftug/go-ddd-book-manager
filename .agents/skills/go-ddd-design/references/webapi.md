# Web API層の設計規則

この層は現時点で未実装。以下は実装時の方針であり、既存コードの挙動を記述したものではない。Applicationの入力DTOなど既に実装された境界を優先して確認する。

対象は`internal/webapi/`とする。

Echoを利用し、HTTPとApplicationの入出力を接続する。

Domain invariantやUseCaseの処理をHandlerへ移さない。

## Package構成と依存方向

```text
internal/webapi/
├── server.go
├── router.go
├── error_handler.go
├── auth/
│   ├── middleware.go
│   ├── claims.go
│   ├── context.go
│   └── errors.go
├── book/
│   ├── router.go
│   └── handler.go
└── user/
    ├── router.go
    └── handler.go
```

上記は配置例とし、必要なファイルのみ追加する。

Feature単位でHandlerとroute登録をまとめる。HTTP HandlerをCommand・Queryのsubpackageへ分ける必要はない。

Web APIはApplicationと必要なDomain型を利用する。ApplicationやDomainからEchoへ依存させない。

Handlerや認証middlewareからBunやInfrastructureの具象Repositoryを操作しない。

依存の生成・登録はcomposition boundaryで行い、HandlerはconstructorでUseCaseなどの依存を受け取る。

## Handlerとbinding

Handlerの責務は以下とする。

1. path・query・bodyを読み取る
2. ApplicationのIdentity・検索条件・Requestへbindする
3. 認証済みActorとrequestの`context.Context`を渡してUseCaseを実行する
4. 成功レスポンスを返す

エラーは共通のHTTPErrorHandlerへ返す。

bindingに使用するstructと`param`・`query`・`json`タグはApplicationの入力DTOに置く。Web APIはこれらの型を使って入力元ごとにbindする。

| 入力元 | Application側 |
| --- | --- |
| path | Featureの`dto/identity.go` |
| query | Featureの`dto/query.go`、共通のページング型は`common/pagination.go` |
| body | Featureの`dto/request.go` |

UUIDや数値へのparse、不正なJSONなどのHTTP入力エラーはWeb APIで扱う。

ページサイズの範囲や検索条件の組み合わせはApplication、タイトルやISBNのvalidationはDomainに委ねる。

path・query・bodyの入力元を明示してbindし、bodyでpathの識別情報や認証済みIdentityを上書きさせない。

Domain EntityやBun modelへ直接bindしない。

Applicationの返却DTOはそのままJSONへ変換してよい。HTTP固有の表現が必要な場合はWeb APIの`response.go`で変換する。

## ルーティングと成功レスポンス

HTTP method・path・status・middlewareの対応はWeb APIで定義する。

ApplicationのIdentityは対象の識別情報を表し、`/books/:book_id`などのroute文字列を持たない。

基本的なrouteは以下とする。共通のAPI prefixを付ける場合はルート登録側でまとめる。

| Method | Path                        | UseCase            | 成功status |
| ------ | --------------------------- | ------------------ | ---------- |
| POST   | `/users/me`                 | SyncUser           | 200        |
| GET    | `/users/me`                 | 自分のUser詳細取得 | 200        |
| GET    | `/books`                    | 書籍一覧取得       | 200        |
| GET    | `/books/:book_id`           | 書籍詳細取得       | 200        |
| POST   | `/books`                    | 書籍作成           | 201        |
| PUT    | `/books/:book_id`           | 書籍更新           | 204        |
| DELETE | `/books/:book_id`           | 書籍削除           | 204        |
| POST   | `/books/:book_id/checkouts` | 貸出               | 204        |
| POST   | `/books/:book_id/return`    | 返却               | 204        |
| GET    | `/books/:book_id/checkouts` | 貸出履歴取得       | 200        |

SyncUserは新規取り込み・既存同期で共通の成功契約とし、Applicationが返す同期後の`UserResponseDTO`を返す。

204ではbodyを返さない。

## 認証middleware

認証は以下の2段階に分ける。

- `RequireIdentity`: Bearer Tokenを検証し、検証済みIdentityをrequest内で保持する
- `RequireUser`: 検証済みIdentityに対応するDBのUserを取得し、Actorをrequest内で保持する

`RequireUser`は`RequireIdentity`の後に実行する。

| API              | RequireIdentity | RequireUser |
| ---------------- | --------------- | ----------- |
| `POST /users/me` | 必須            | 適用しない  |
| その他の業務API  | 必須            | 必須        |

書籍一覧・詳細を含め、業務APIに匿名アクセスを許可しない。

SyncUserはHTTP methodとrouteで明示的に登録する。同じ`/users/me`でもGETは登録済みUserを必要とする。

同期routeはIdentity認証のみのgroupへ、その他はIdentity認証とUser解決を適用するgroupへ登録するなど、middlewareの適用範囲を明示する。

RequireUserはApplicationのActor解決処理を呼び出し、Userの取得・Actorへの変換を委ねる。

User不在は認証エラー（401）とし、通常の対象EntityのNotFound（404）と区別する。

DB障害はUser不在と扱わず、元のerrorを保持してサーバーエラーへ渡す。

Actor・Identityを共有Serviceのfieldやglobal変数へ保存しない。request内の値は型付きaccessorで取得し、型不一致などの内部不整合は500として扱う。

## SyncUser

Userの新規取り込み・同期は`POST /users/me`からApplicationのSyncUser UseCaseを呼び出して行う。

```text
POST /users/me
  → RequireIdentity
  → 検証済みIdentityからSyncUser入力を作成
  → SystemActorでUser.NewまたはUser.Sync
  → 保存し、同期後のUserResponseDTOを返す

その他の業務API
  → RequireIdentity
  → RequireUser
  → DBのUserがなければ401
  → DBのUser.Actor()でUseCaseを実行
```

通常の業務APIでUserを自動作成・同期しない。

name・email・roleは検証済みclaimまたはIdPのUserInfoから取得する。クライアントのbodyで任意のUser ID・role・SystemActorを指定させない。

roleはIdP固有のclaimから許可したDomain roleへ明示的に変換する。通常のUserにSystemRoleを割り当てない。

SyncUserに必要なprofile情報の検証は同期処理で行い、その他のAPIで不要なname・email claimまで必須にしない。

通常APIのActorはDBのUserから作るため、name・email・roleの変更は次のSyncUser成功時に反映される。

## OIDCとAccess Token検証

クライアントがIdPでログインし、このAPIへBearer Access Tokenを送る構成を基本とする。

API向けAccess Tokenと、ログイン結果を表すID Tokenを区別する。ID TokenをそのままAPIの認証資格として受け入れない。

移植元と同様にKeycloakのJWT Access Tokenを前提とする。別のIdPやopaque tokenへ対応するときは検証方式を見直す。

検証部品の第一候補は`auth0/go-jwt-middleware`の`validator`と`jwks`とし、薄いEcho middlewareから利用する。

ライブラリのHTTP middlewareにエラーレスポンスを直接書かせず、検証結果をEchoのerror処理へ接続する。

以下を設定・検証する。

- 設定で固定したissuer
- このAPIを対象とするaudience（必須）
- 許可した署名アルゴリズムと署名
- 有効期限（expを必須とする）と有効開始時刻などの時間条件
- Userを識別するsub
- IdPのtoken種別を区別するために必要なclaim等の規則

未検証のtokenからissuerや鍵取得先を自由に選ばせない。

OIDC discoveryとJWKSの取得・キャッシュにはライブラリの機能を使う。通信timeoutや鍵更新時の挙動を確認し、毎requestでJWKSを取得しない。

現在のDomainはUser IDにUUIDを使うため、単一issuerのKeycloak subがUUID形式であることを前提としてparseする。一般のOIDC subはUUIDとは限らない。

署名・issuer・audience等の検証完了前のclaimを、Actor生成やSyncUserへ利用しない。

`coreos/go-oidc`はOIDCログイン・ID Token検証が必要になった場合の候補とする。Echo専用のJWT middlewareを使う場合も、discovery・JWKS管理・Access Token検証の責務を確認する。

## Echoのエラーハンドリング

`e.HTTPErrorHandler`へHTTPエラー変換とレスポンス整形を集約する。

Handler・認証middlewareはerrorを返す。Application・DomainにHTTP statusやEchoのerror型を持ち込まない。

`errors.Is`・`errors.As`でerror chainを判定し、`Error()`の文字列で分岐しない。

| エラー                                 | HTTP status |
| -------------------------------------- | ----------- |
| Tokenなし・無効・期限切れ              | 401         |
| 有効なIdentityに対応するDBのUserが不在 | 401         |
| Domain・UseCaseの認可失敗              | 403         |
| 対象Bookなどの不在                     | 404         |
| binding・parse失敗、入力validation失敗 | 400         |
| 同時更新の競合                         | 409         |
| 想定外のDB障害や内部不整合             | 500         |

認証に必要なclaimの不正は401として扱い、通常のHTTP入力validation失敗と区別する。

IdP通信・JWKS取得などの外部障害を、資格情報の不正と区別できる場合はサーバー側の障害として扱う。

401には`WWW-Authenticate: Bearer`を付ける。

Echo由来の404・405・413・415などのHTTPエラーも尊重する。

500は汎用メッセージを返し、SQL・token・内部errorの詳細をレスポンスへ露出しない。

DomainのValidationErrorを公開する場合はfield・codeを利用し、内部error文字列をそのまま返さない。

レスポンス送信済みの場合は二重書き込みを避ける。HEADなどのHTTP上のbody制約も守る。

errorの文脈を追加するときは`%w`で元のerrorを保持する。

## middlewareとログ

共通middlewareには必要に応じてRequest ID、Recover、Request Logger、Body Limit、CORSを登録する。

Recoverによるpanic処理も共通のエラーハンドリングへ接続する。

ログはRequest ID・method・route・statusなどを使い、Authorization headerやtokenを記録しない。

## テスト

`httptest`とUseCase・Token検証のfakeを使い、実際のEcho routerを通して次を確認する。

- 未登録Userでも有効なIdentityでPOST /users/meへ到達できること
- SyncUser以外はTokenと登録済みUserの両方を必要とすること
- GET /users/meはSyncUserの例外対象にならないこと
- 通常APIでUserを自動作成・同期しないこと
- User不在は401、DB障害は500になること
- path・query・bodyからApplication入力への変換が正しいこと
- 認可失敗・validation・NotFoundなどが共通のエラーハンドラを通ること
- 401の認証header、204の空body、500の詳細非公開

Token検証adapterは、ローカルのdiscovery・JWKS serverと署名済みtokenを使い、署名・issuer・audience・期限・鍵更新時の挙動を検証する。

## 参考

- [移植元のroute定義](https://github.com/nftug/rust-ddd-book-manager/blob/main/api/src/router/book/router.rs)
- [移植元のHandler](https://github.com/nftug/rust-ddd-book-manager/blob/main/api/src/router/book/handlers.rs)
- [Echo Error Handling](https://echo.labstack.com/guide/error-handling/)
- [Echo JWT middleware](https://github.com/labstack/echo-jwt)
- [Auth0 Go JWT middleware](https://github.com/auth0/go-jwt-middleware)
- [coreos/go-oidc](https://github.com/coreos/go-oidc)
