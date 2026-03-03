package cli

import "io"

// Run 是 CLI 主入口，串联所有模块完成转换。
// stdout 和 stderr 通过参数注入，便于测试。
func Run(args Args, stdout io.Writer, stderr io.Writer) error {
	return nil
}
