//go:build !windows

package plugin

import "os"

func killTree(p *os.Process) { _ = p.Kill() }
