# Application層の設計規則

Application層はUseCaseを実現し、Domainと外部I/Oを調停する。

Domain invariantそのものをApplication層へ移さない。

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

---

# references/infrastructure.md
