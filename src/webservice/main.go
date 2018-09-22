package main

import (
	"fmt"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
)

func init() {
	ServiceConfig.LoadConfig()
}

func main() {
	server := echo.New()
	server.Use(middleware.Logger())
	server.Use(middleware.Recover())

	//Routers
	server.PATCH("/group/rename", SetNameGroup)
	server.GET("/group/get", GetGroup)
	server.GET("/group/get/tree", GetTreeGroup)
	server.PUT("/group/add", AddGroup)
	server.PATCH("/group/move", MoveGroup)
	server.DELETE("/group/delete", DeleteGroup)
	server.Logger.Fatal(server.Start(fmt.Sprintf(":%d", ServiceConfig.ServicePort)))
}
