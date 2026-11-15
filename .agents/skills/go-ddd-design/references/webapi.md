# Web API層の設計規則

この層は実装済み。現在のコードとテストを設計判断の基準とする。

対象は`internal/webapi/`とする。

Echoを利用し、HTTPとApplicationの入出力を接続する。

Domain invariantやUseCaseの処理をHandlerへ移さない。

## Package構成と依存方向

```text
internal/webapi/
├── server.go          # Echoの構築とroute登録
├── injector.go        # Web APIのprovider登録
├── auth/              # Token検証、認証middleware、Identity
├── bind/              # 入力元別のbinding
├── handler/           # User同期・UseCase実行・レスポンス・エラー処理
├── book/              # 書籍のrouteとHandler
└── user/              # UserのrouteとHandler
```

Feature単位でHandlerとroute登録をまとめる。HTTP HandlerをCommand・Queryのsubpackageへ分ける必要はない。

Web APIはApplicationと必要なDomain型を利用する。ApplicationやDomainからEchoへ依存させない。

Handlerや認証middlewareからBunやInfrastructureの具象Repositoryを操作しない。

`webapi.Inject` でproviderを登録し、HandlerのconstructorでUseCaseなどの依存を解決して保持する。

## Handlerとbinding

Handlerの責務は以下とする。

1. path・query・bodyを読み取る
2. ApplicationのIdentity・検索条件・Requestへbindする
3. 認証済みActorとrequestの`context.Context`を渡してUseCaseを実行する
4. 成功レスポンスを返す

エラーは共通のHTTPErrorHandlerへ返す。

`bind.PathParams`・`bind.QueryParams`・`bind.Body` はEchoのDefaultBinderを入力元ごとに呼ぶ。bindingに使用するstructと`param`・`query`・`json`タグはApplicationの入力DTOに置く。Web APIはこれらの型を使って入力元ごとにbindする。

| 入力元 | Application側 |
| --- | --- |
| path | Featureの`dto/identity.go` |
| query | Featureの`dto/query.go`、共通のページング型は`common/pagination.go` |
| body | Featureの`dto/request.go` |

UUIDや数値へのparse、不正なJSONなどのHTTP入力エラーはWeb APIで扱う。

ページサイズの範囲や検索条件の組み合わせはApplication、タイトルやISBNのvalidationはDomainに委ねる。

path・query・bodyの入力元を明示してbindし、bodyでpathの識別情報や認証済みIdentityを上書きさせない。

Domain EntityやBun modelへ直接bindしない。

`handler.Executor` の `OK`・`Created`・`NoContent` はUser同期によるActorの生成、UseCaseの実行、成功レスポンスを共通化する。Applicationの返却DTOはそのままJSONへ変換する。

## ルーティングと成功レスポンス

HTTP method・path・status・middlewareの対応はWeb APIで定義する。

ApplicationのIdentityは対象の識別情報を表し、`/books/:book_id`などのroute文字列を持たない。

現在のrouteは以下のとおり。共通のAPI prefixはない。

| Method | Path                        | UseCase            | 成功status |
| ------ | --------------------------- | ------------------ | ---------- |
| GET    | `/users/me`                 | 自分のUser詳細取得 | 200        |
| GET    | `/books`                    | 書籍一覧取得       | 200        |
| GET    | `/books/:book_id`           | 書籍詳細取得       | 200        |
| POST   | `/books`                    | 書籍作成           | 201        |
| PUT    | `/books/:book_id`           | 書籍更新           | 204        |
| DELETE | `/books/:book_id`           | 書籍削除           | 204        |
| PUT    | `/books/:book_id/owner`     | 所有者変更         | 204        |
| POST   | `/books/:book_id/checkouts` | 貸出               | 204        |
| POST   | `/books/:book_id/return`    | 返却               | 204        |
| GET    | `/books/:book_id/checkouts` | 貸出履歴取得       | 200        |

書籍作成は`CreatedResponse`を返す。`POST /users/me`は登録されていない。

204ではbodyを返さない。

## 認証middleware

現在の業務routeには`RequireIdentity`を適用し、Bearer Tokenを検証してIdentityをrequestの`context.Context`に保持する。

各Handlerは`handler.Executor`を呼ぶ。Executorは`IdentityFrom`でIdentityを取り出し、`SyncUser`にID・name・email・roleを渡す。Userがなければ作成し、あれば同期したうえで、返されたActorとrequestの`context.Context`をUseCaseに渡す。同期は入力のbindingより先に行われる。

同期時のDomain ValidationErrorは`auth.ErrUnauthorized`に変換し、DB障害などは元のerrorを返す。Identityの欠落など内部不整合は500として扱う。ActorやIdentityを共有Serviceのfieldやglobal変数へ保存しない。

name・email・roleは検証済みclaimから取得する。クライアントのbodyで任意のUser ID・role・SystemActorを指定させない。roleはIdP固有のclaimから許可したDomain roleへ変換し、通常のUserにSystemRoleを割り当てない。

## OIDCとAccess Token検証

クライアントがIdPでログインし、このAPIへBearer Access Tokenを送る構成を基本とする。

API向けAccess Tokenと、ログイン結果を表すID Tokenを区別する。ID TokenをそのままAPIの認証資格として受け入れない。

移植元と同様にKeycloakのJWT Access Tokenを前提とする。別のIdPやopaque tokenへ対応するときは検証方式を見直す。

`auth.Config` のissuerとaudienceを使い、`auth0/go-jwt-middleware/v2`のvalidatorとJWKS caching providerで検証する。issuerはHTTPまたはHTTPSのURL、audienceは空でない必要がある。

- 署名アルゴリズムはRS256。署名・issuer・audience・時間条件を検証する。
- `exp`は必須、`sub`はnilでないUUID、claimの`typ`は`Bearer`を要求する。
- `realm_access.roles`に`Admin`があればAdminRole、それ以外はUserRoleに変換する。nameとemailは同期時に利用する。
- JWKSのキャッシュは5分、HTTP clientのtimeoutは5秒で、同期refreshを使う。
- 無効なTokenやclaimは`ErrUnauthorized`、鍵取得などprovider側の失敗は`ErrProvider`として区別する。後者は500になる。

署名などの検証前のclaimをActor生成やSyncUserに使わない。UUID形式のsubとclaimの扱いはKeycloakを前提とするため、別のIdPやopaque tokenへの対応時は見直す。

## Echoのエラーハンドリング

`e.HTTPErrorHandler`へHTTPエラー変換とレスポンス整形を集約する。

Handler・認証middlewareはerrorを返す。Application・DomainにHTTP statusやEchoのerror型を持ち込まない。

`errors.Is`・`errors.As`でerror chainを判定し、`Error()`の文字列で分岐しない。

| エラー                                 | HTTP status |
| -------------------------------------- | ----------- |
| Tokenなし・無効・期限切れ              | 401         |
| Domain・UseCaseの認可失敗              | 403         |
| 対象Bookなどの不在                     | 404         |
| binding・parse失敗、入力validation失敗 | 400         |
| 同時更新の競合                         | 409         |
| 想定外のDB障害や内部不整合             | 500         |

認証に必要なclaimや同期時のprofileの不正は401として扱い、通常のHTTP入力validation失敗と区別する。

IdP通信・JWKS取得などの外部障害を、資格情報の不正と区別できる場合はサーバー側の障害として扱う。

401には`WWW-Authenticate: Bearer`を付ける。

Echo由来の404・405・413・415などのHTTPエラーも尊重する。

500は汎用メッセージを返し、SQL・token・内部errorの詳細をレスポンスへ露出しない。

JSONは`message`を返し、DomainのValidationErrorには`field`・`code`も含める。内部error文字列は返さない。

レスポンス送信済みの場合は二重書き込みを避ける。HEADなどのHTTP上のbody制約も守る。

errorの文脈を追加するときは`%w`で元のerrorを保持する。

## middlewareとログ

共通middlewareとしてRequest IDとRecoverを登録している。

Recoverによるpanic処理も共通のエラーハンドリングへ接続する。

ログはRequest ID・method・route・statusを含め、5xxと送信済みレスポンスのエラーでは元のerrorも記録する。Authorization headerやtokenを記録しない。

## テスト

`httptest`とUseCase・Token検証のfakeを使い、実際のEcho routerを通して次を確認する。

- 業務APIが有効なTokenを必要とし、Handler実行前にSyncUserを呼ぶこと
- 同期時のvalidation失敗は401、DB障害は500になること
- POST /users/meが登録されていないこと
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
