package cli

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
func ParseArgs(args []string) (Args, error) {
	return Args{}, nil
}
