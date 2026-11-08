package interpreter

import (
	stderrors "errors"

	lyerr "lycka/internal/errors"
	"lycka/internal/parser"
	"lycka/internal/types"
)

// attach はエラーに位置情報を付ける。まだ位置のないエラーにだけ付くので、
// 最も内側(原因に近い)の位置が残る。*errors.Error 以外は内部エラーに包む。
func attach(err error, pos parser.Pos) error {
	var le *lyerr.Error
	if stderrors.As(err, &le) {
		le.WithPos(pos.Line, pos.Col)
		return le
	}
	return lyerr.Internal(err.Error()).WithPos(pos.Line, pos.Col)
}

// Run はコードを上から順に実行する。
func (c *Context) Run(code *types.Code) error {
	return c.runNodes(code.Body)
}

func (c *Context) runNodes(body []parser.Node) error {
	c.depth++
	defer func() { c.depth-- }()
	if c.depth > maxDepth {
		return lyerr.RecursionLimit()
	}
	for _, n := range body {
		call, ok := n.(*parser.Call)
		if !ok {
			return attach(lyerr.NotAFunctionCall(), n.Position())
		}
		if _, err := c.evalCall(call); err != nil {
			return attach(err, call.Position())
		}
	}
	return nil
}

// Eval は式を評価して値を返す。
func (c *Context) Eval(n parser.Node) (types.Value, error) {
	v, err := c.eval(n)
	if err != nil {
		return nil, attach(err, n.Position())
	}
	return v, nil
}

func (c *Context) eval(n parser.Node) (types.Value, error) {
	switch x := n.(type) {
	case *parser.IntLit:
		return types.NewInt(x.Value), nil
	case *parser.StrLit:
		return &types.Str{V: x.Value}, nil
	case *parser.BoolLit:
		return types.FromBool(x.Value), nil
	case *parser.NullLit:
		return types.NullValue, nil
	case *parser.CodeLit:
		return &types.Code{Body: x.Body, Src: x.Src}, nil
	case *parser.ListLit:
		items := make([]types.Value, len(x.Elems))
		for i, e := range x.Elems {
			v, err := c.Eval(e)
			if err != nil {
				return nil, err
			}
			items[i] = v
		}
		return types.NewList(items), nil
	case *parser.VarRef:
		v, ok := c.GetVar(x.Name)
		if !ok {
			return nil, lyerr.UndefinedVariable(x.Name)
		}
		return v, nil
	case *parser.Call:
		return c.evalCall(x)
	case *parser.UnaryMinus:
		v, err := c.Eval(x.Operand)
		if err != nil {
			return nil, err
		}
		i, ok := v.(*types.Int)
		if !ok {
			return nil, lyerr.OperatorType("neg", v.Type())
		}
		return i.Neg(), nil
	case *parser.BinaryOp:
		return c.evalBinary(x)
	}
	return nil, lyerr.Internal("unknown AST node")
}

// ---- 関数呼び出し ----

func (c *Context) evalCall(n *parser.Call) (types.Value, error) {
	fn, found := c.sh.reg.Lookup(n.Name)
	if !found {
		return nil, lyerr.UndefinedFunction(n.Name)
	}
	if ok, tooMany := fn.checkArity(len(n.Args)); !ok {
		return nil, lyerr.FunctionArgCount(fn.Name, tooMany)
	}

	args := make([]Arg, len(n.Args))
	for i, a := range n.Args {
		kind := fn.kindAt(i)
		args[i].Kind = kind
		switch kind {
		case ParamValue:
			expr, err := singleExpr(fn, a)
			if err != nil {
				return nil, err
			}
			v, err := c.Eval(expr)
			if err != nil {
				return nil, err
			}
			args[i].Value = v
		case ParamLazy:
			expr, err := singleExpr(fn, a)
			if err != nil {
				return nil, err
			}
			args[i].Expr = expr
		case ParamCode:
			code, err := c.codeFromArg(fn, a)
			if err != nil {
				return nil, err
			}
			args[i].Code = code
		}
	}
	return fn.Fn(c, args)
}

func singleExpr(fn *Func, a *parser.Arg) (parser.Node, error) {
	if len(a.Items) != 1 {
		return nil, attach(lyerr.FunctionArgForm(fn.Name), a.Pos)
	}
	return a.Items[0], nil
}

// codeFromArg は cd 型の引数を作る。cd[] は省略できる(仕様書 3.4)。
func (c *Context) codeFromArg(fn *Func, a *parser.Arg) (*types.Code, error) {
	if len(a.Items) == 1 {
		switch x := a.Items[0].(type) {
		case *parser.CodeLit:
			return &types.Code{Body: x.Body, Src: x.Src}, nil
		case *parser.VarRef:
			v, ok := c.GetVar(x.Name)
			if !ok {
				return nil, attach(lyerr.UndefinedVariable(x.Name), x.Position())
			}
			code, ok := v.(*types.Code)
			if !ok {
				return nil, attach(lyerr.FunctionType(fn.Name, v.Type()), x.Position())
			}
			return code, nil
		}
	}
	return &types.Code{Body: a.Items, Src: a.Src}, nil
}

// ---- 演算 ----

func (c *Context) evalBinary(n *parser.BinaryOp) (types.Value, error) {
	l, err := c.Eval(n.Left)
	if err != nil {
		return nil, err
	}
	r, err := c.Eval(n.Right)
	if err != nil {
		return nil, err
	}
	switch n.Op {
	case "+":
		return opAdd(l, r)
	case "-":
		return opSub(l, r)
	case "/":
		return opDiv(l, r)
	case "=":
		t, err := types.Equal(l, r)
		if err != nil {
			return nil, err
		}
		switch t {
		case types.TriTrue:
			return types.True, nil
		case types.TriFalse:
			return types.False, nil
		}
		return types.NewSmthng(), nil
	}
	return nil, lyerr.Internal("unknown operator " + n.Op)
}

// rejectTypes は operands のうち banned に含まれる型があればエラーを返す。
func rejectTypes(op string, banned []string, operands ...types.Value) error {
	for _, v := range operands {
		for _, b := range banned {
			if v.Type() == b {
				return lyerr.OperatorType(op, v.Type())
			}
		}
	}
	return nil
}

// 仕様書 5.1: 加算演算子。bl 型には使えず、異なるデータ型は混ぜられない。
func opAdd(l, r types.Value) (types.Value, error) {
	if err := rejectTypes("+", []string{types.TypeBool}, l, r); err != nil {
		return nil, err
	}
	if l.Type() != r.Type() {
		return nil, lyerr.OperatorMixed("+", l.Type(), r.Type())
	}
	switch a := l.(type) {
	case *types.Int:
		return a.Add(r.(*types.Int)), nil
	case *types.List:
		return types.Union(a, r.(*types.List)), nil
	case *types.Str:
		return a.Concat(r.(*types.Str)), nil
	case *types.Code:
		return types.ConcatCode(a, r.(*types.Code)), nil
	case *types.Null:
		return types.NullValue, nil
	case *types.Smthng:
		return types.NewSmthng(), nil
	}
	return nil, lyerr.Internal("unhandled type in +")
}

// 仕様書 5.2: 減算演算子。bl 型と cd 型には使えない。
func opSub(l, r types.Value) (types.Value, error) {
	if err := rejectTypes("-", []string{types.TypeBool, types.TypeCode}, l, r); err != nil {
		return nil, err
	}
	if l.Type() != r.Type() {
		return nil, lyerr.OperatorMixed("-", l.Type(), r.Type())
	}
	switch a := l.(type) {
	case *types.Int:
		return a.Sub(r.(*types.Int)), nil
	case *types.List:
		return types.Difference(a, r.(*types.List)), nil
	case *types.Str:
		return a.Remove(r.(*types.Str)), nil
	case *types.Null:
		return types.NullValue, nil
	case *types.Smthng:
		return types.NewSmthng(), nil
	}
	return nil, lyerr.Internal("unhandled type in -")
}

// 除算演算子 ÷。仕様書 3.7 の例(0 ÷ 0 が smthng になる)から導入。
// int 同士は 0 方向への切り捨て除算。0 ÷ 0 は smthng、0 以外 ÷ 0 はエラー。
// null と smthng は + - と同じ扱い(同型どうしで同型を返す)。
func opDiv(l, r types.Value) (types.Value, error) {
	banned := []string{types.TypeBool, types.TypeStr, types.TypeList, types.TypeCode}
	if err := rejectTypes("÷", banned, l, r); err != nil {
		return nil, err
	}
	if l.Type() != r.Type() {
		return nil, lyerr.OperatorMixed("÷", l.Type(), r.Type())
	}
	switch a := l.(type) {
	case *types.Int:
		b := r.(*types.Int)
		if b.IsZero() {
			if a.IsZero() {
				return types.NewSmthng(), nil
			}
			return nil, lyerr.DivisionByZero()
		}
		return a.Quo(b), nil
	case *types.Null:
		return types.NullValue, nil
	case *types.Smthng:
		return types.NewSmthng(), nil
	}
	return nil, lyerr.Internal("unhandled type in ÷")
}
