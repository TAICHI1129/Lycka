c("変数操作のサンプル: set で宣言し、to で代入する")

set("x") to(10)
print(x)

c("演算子の結果をそのまま代入できる")
set("y") to(x + 5)
print(y)

c("cd 型の値は変数に入れておき、後から実行できる")
set("greet") to(cd[print("こんにちは")])
if(TRUE)(greet)

c("list 型・str 型の演算")
set("xs") to(list[1, 2, 3] + list[3, 4])
print(xs)
set("word") to("lycka" + "-lang")
print(word)

c("条件分岐: if / else")
if(y = 15)(print("y は 15 です"))
else(print("y は 15 ではありません"))

c("繰り返し: while (3 から 1 まで数える)")
set("n") to(3)
while(n = 0 = FALSE)(print(n) set("n") to(n - 1))

c("smthng: 0 ÷ 0 は「何か」になり、x = x でさえ TRUE/FALSE と言えない")
set("s") to(0 ÷ 0)
print(isitsth(s))
print(isitsth(s = s))
print(isitsth(1))

c("concur: 2 つのコードを並行実行(変数への代入は別々の名前に)")
concur(set("a") to(1))(set("b") to(2))
print(a + b)
