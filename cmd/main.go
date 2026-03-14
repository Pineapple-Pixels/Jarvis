package main

import "asistente/config"

func main() {
	app := NewApp(config.Load())
	defer app.Close()
	app.Run()
}
