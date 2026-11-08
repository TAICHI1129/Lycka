// lycka は Lycka プログラミング言語のインタプリタ。
//
// 使い方:
//
//	lycka <file.ly>    ファイルを実行する
//	lycka -            標準入力のソースを実行する
package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	lyerr "lycka/internal/errors"
	"lycka/internal/functions"
)

// utf8BOM は UTF-8 の BOM(Windows のエディタが付けることがある)。
const utf8BOM = "\xef\xbb\xbf"

const usage = `Usage: lycka <file.ly>
       lycka -          read source from standard input
`

func main() {
	os.Exit(run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr))
}

// run は終了コードを返す(0: 成功, 1: 構文・実行エラー, 2: 使い方の誤り)。
func run(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	if len(args) == 1 && (args[0] == "-h" || args[0] == "--help") {
		fmt.Fprint(stdout, usage)
		return 0
	}
	if len(args) != 1 {
		fmt.Fprint(stderr, usage)
		return 2
	}

	path := args[0]
	var (
		src []byte
		err error
	)
	if path == "-" {
		src, err = io.ReadAll(stdin)
	} else {
		src, err = os.ReadFile(path)
	}
	if err != nil {
		fmt.Fprintln(stderr, lyerr.CannotReadFile(path, err))
		return 1
	}

	// 出力はまとめて書き込み、エラー表示の前に必ずフラッシュして順序を保つ。
	out := bufio.NewWriter(stdout)
	it := functions.NewInterpreter(out)
	runErr := it.Run(strings.TrimPrefix(string(src), utf8BOM)) // BOM は無視
	out.Flush()
	if runErr != nil {
		fmt.Fprintln(stderr, runErr)
		return 1
	}
	return 0
}
