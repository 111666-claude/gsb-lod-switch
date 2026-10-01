// Command lod-switch 跑 LOD 样例。
package main

import (
	"flag"
	"fmt"
	"io"
	"os"

	"example.com/lod"
)

// Run 执行一次命令行调用，返回退出码。
func Run(args []string, stdout, stderr io.Writer) int {
	flags := flag.NewFlagSet("lod-switch", flag.ContinueOnError)
	flags.SetOutput(stderr)
	sample := flags.String("sample", "hysteresis", "hysteresis / event / boundary / scan")
	if err := flags.Parse(args); err != nil {
		return 2
	}
	switch *sample {
	case "hysteresis":
		// 距离在 22 与 18 之间来回抖，恰好落在滞回带里。
		manager := lod.NewManager()
		manager.Add("e1", 10)
		manager.Tick("e1", 10)
		start := manager.Switches()
		for index := 0; index < 10; index++ {
			manager.Tick("e1", 22)
			manager.Tick("e1", 18)
		}
		fmt.Fprintf(stdout, "switches=%d\n", manager.Switches()-start)
	case "event":
		manager := lod.NewManager()
		manager.Add("e1", 10)
		for index := 0; index < 10; index++ {
			manager.Tick("e1", 10)
		}
		fmt.Fprintf(stdout, "switches=%d\n", manager.Switches())
	case "boundary":
		manager := lod.NewManager()
		manager.Add("e1", 10)
		fmt.Fprintf(stdout, "level=%d\n", manager.Tick("e1", 20))
	case "scan":
		manager := lod.NewManager()
		for index := 0; index < 3000; index++ {
			manager.Add(fmt.Sprintf("e-%d", index), 10)
		}
		for index := 0; index < 3000; index++ {
			manager.Tick("e-2999", 10)
		}
		fmt.Fprintf(stdout, "scanned=%d\n", manager.Scanned())
	default:
		fmt.Fprintln(stderr, "需要 --sample hysteresis|event|boundary|scan")
		return 2
	}
	return 0
}

func main() {
	os.Exit(Run(os.Args[1:], os.Stdout, os.Stderr))
}
