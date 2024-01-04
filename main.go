package main

import (
	"sensor-service/cmd"
	"sensor-service/config"
	"sensor-service/internal/platform/cpu"
)

func main() {
	cpu.UtilizeCPU()
	var err error
	err = config.SetConfig("")
	if err != nil {
		panic(err)
	}
	cmd.Execute()
}
