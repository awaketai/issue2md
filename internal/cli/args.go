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
	fs.StringVar(&a.Output, "o", "", "output file path or directory")
	fs.StringVar(&a.Output, "output", "", "output file path or directory")
	fs.BoolVar(&a.WithReactions, "with-reactions", false, "show reaction stats")
	fs.BoolVar(&a.Verbose, "verbose", false, "print debug logs to stderr")

	// Go 的 flag 包遇到第一个非 flag 参数就停止解析。
	// 为了支持 `issue2md <url> --verbose` 这样 URL 在 flag 前面的用法，
	// 先将 args 分为 flags 和 positional，再解析 flags。
	var flags []string
	var positional []string
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if arg == "--" {
			positional = append(positional, args[i+1:]...)
			break
		}
		if len(arg) > 0 && arg[0] == '-' {
			flags = append(flags, arg)
			// 如果是带值的 flag（-o），下一个参数是值
			if (arg == "-o" || arg == "--output") && i+1 < len(args) {
				i++
				flags = append(flags, args[i])
			}
		} else {
			positional = append(positional, arg)
		}
	}

	if err := fs.Parse(flags); err != nil {
		return Args{}, fmt.Errorf("parsing flags: %w", err)
	}

	if len(positional) < 1 {
		return Args{}, fmt.Errorf("usage: issue2md <url> [-o <file>] [--with-reactions] [--verbose]")
	}

	a.URL = positional[0]
	return a, nil
}
