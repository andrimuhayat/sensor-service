package cpu

import "runtime"

func UtilizeCPU() {
	runtime.GOMAXPROCS(runtime.NumCPU())
}
