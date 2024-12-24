package main

import "render-service/services"

const port = 3000

func main() {
	(&services.RenderService{}).Start(port)
}
