# Lycka

**Lycka** は、演算と条件式を除くすべての処理を「関数」で行うプログラミング言語です。
条件分岐(`if()`)も繰り返し(`while()`)も並行実行(`concur()`)も、専用の構文ではなく関数です。
このリポジトリは Go で書かれたインタプリタ(拡張子 `.ly`)です。

```
set("n") to(3)
while(n = 0 = FALSE)(print(n) set("n") to(n - 1))
```

仕様書: [docs/SPEC.md](docs/SPEC.md)　／　仕様にない部分の解釈: [docs/implementation-notes.md](docs/implementation-notes.md)

## ビルドと実行

Go 1.24 以上が必要です。外部依存はありません。

```sh
go build -o lycka ./cmd/lycka
./lycka examples/hello.ly
echo 'print(1 + 2)' | ./lycka -     # 標準入力から実行
```

終了コード: `0` 成功 / `1` 構文エラー・実行エラー / `2` 使い方の誤り。
エラーは標準エラー出力に、仕様書 6 章の階層形式で表示されます。

```
ERROR: Violated The Function Rule (1)
    Violated The "print" Function Rule (1/1)
        Types That Cannot Be Used (1/1/1)
            "cd" Type Error (1/1/1/cd)
    at line 1, column 1
```

## 言語の早見表

| 種類 | 内容 |
|---|---|
| データ型 | `int` `list` `str` `cd` `bl` `null` `smthng` |
| 関数 | `print` `if` `else` `isitsth` `while` `concur` `c` `set` `to` |
| 演算 | `+` `-` `=`、および `÷`(`0 ÷ 0` が `smthng` になる) |

```
c("関数を並べて書く。改行してもしなくてもよい")
set("x") to(1)

c("条件分岐: cd[] は省略できる")
if(x = 1)(print("one")) else(print("other"))

c("コードは値。変数に入れて後から実行できる")
set("hello") to(cd[print("hi")])
if(TRUE)(hello)

c("smthng: 同じ「何か」同士でも等しいとは言えない")
set("s") to(0 ÷ 0)
print(isitsth(s = s))

c("並行実行")
concur(print("A"))(print("B"))
```

その他のサンプルは [examples/](examples/) を参照してください。

## テスト

```sh
go test -race ./...
```

- 各パッケージの単体テスト(字句解析・構文解析・型・関数・エラー形式)
- [`tests/test.ly`](tests/test.ly): Lycka で書かれた自己検査プログラム(各チェックが `ok` / `FAIL` を表示)。
  [`tests/run_test.go`](tests/run_test.go) が実行して `FAIL` が無いことを検証します。

## 構成

```
cmd/lycka/            起動処理(引数の取得・ソースの読み込み・実行)
internal/lexer/       字句解析
internal/parser/      構文解析と AST(c() の除去もここ)
internal/interpreter/ evaluator(AST の評価)・context(変数と実行状態)・registry(関数の登録)
internal/types/       7種のデータ型と、比較・集合演算
internal/functions/   組み込み関数(1関数1ファイル)
internal/errors/      階層構造のエラー
examples/ tests/ docs/
```

関数を追加するには、`internal/functions/` に関数定義を書き、`functions.go` の `RegisterAll` に加えます。
引数は `ParamValue`(評価済みの値)・`ParamLazy`(未評価の式)・`ParamCode`(cd 型)から選べます。
