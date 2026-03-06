package cli

import (
	"flag"
	"fmt"
	"io"
)

// Args 表示解析后的命令行参数
type Args struct {
	URL           string // GitHub Issue/PR/Discussion URL（必填）
	Output        string // 输出文件路径，空字符串表示 stdout
	WithReactions bool   // 是否显示 Reactions
	Verbose       bool   // 是否输出调试日志到 stderr
}

// InputError 表示用户输入导致的错误（退出码 1）
type InputError struct {
	Err error
}

func (e *InputError) Error() string { return e.Err.Error() }
func (e *InputError) Unwrap() error { return e.Err }

// ParseArgs 解析命令行参数。args 为 os.Args[1:]。
// URL 可以出现在 flags 之前或之后。
func ParseArgs(args []string) (Args, error) {
	fs := flag.NewFlagSet("issue2md", flag.ContinueOnError)
	fs.SetOutput(io.Discard)

	var a Args
	fs.StringVar(&a.Output, "o", "", "output file path")
	fs.StringVar(&a.Output, "output", "", "output file path")
	fs.BoolVar(&a.WithReactions, "with-reactions", false, "show reactions")
	fs.BoolVar(&a.Verbose, "verbose", false, "verbose output to stderr")

	// 将 positional arg (URL) 移到末尾，使 flag 解析能处理混合顺序
	flagArgs, positional := separateArgs(args, fs)
	if err := fs.Parse(flagArgs); err != nil {
		return Args{}, fmt.Errorf("parsing flags: %w", err)
	}

	// 合并 flag.Parse 剩余的位置参数
	positional = append(positional, fs.Args()...)

	if len(positional) < 1 {
		return Args{}, fmt.Errorf("usage: issue2md <url> [-o <file>] [--with-reactions] [--verbose]")
	}

	a.URL = positional[0]
	return a, nil
}

// separateArgs 将 args 中的 flag 参数和 positional 参数分离。
// Go 的 flag 包在遇到第一个非 flag 参数后停止解析，
// 因此需要手动分离以支持 `issue2md <url> --verbose` 这样 URL 在前的用法。
func separateArgs(args []string, fs *flag.FlagSet) (flagArgs, positional []string) {
	i := 0
	for i < len(args) {
		arg := args[i]
		if arg == "--" {
			positional = append(positional, args[i+1:]...)
			return
		}
		if len(arg) > 1 && arg[0] == '-' {
			flagArgs = append(flagArgs, arg)
			// 检查该 flag 是否带值（非 bool 类型）
			name := arg
			if len(name) > 2 && name[:2] == "--" {
				name = name[2:]
			} else {
				name = name[1:]
			}
			if f := fs.Lookup(name); f != nil {
				if _, ok := f.Value.(interface{ IsBoolFlag() bool }); !ok {
					// 非 bool flag，下一个参数是值
					if i+1 < len(args) {
						i++
						flagArgs = append(flagArgs, args[i])
					}
				}
			}
		} else {
			positional = append(positional, arg)
		}
		i++
	}
	return
}
