package main

import (
	"github.com/gin-gonic/gin"
	"golang_study/2/internal/repository"
	"golang_study/2/internal/repository/cache"
	"golang_study/2/internal/repository/dao"
	"golang_study/2/internal/service"
	"golang_study/2/internal/web"
	"golang_study/2/ioc"
)

func Init() *gin.Engine {
	cmdable := ioc.InitRedis()
	v := ioc.GinMiddlewares(cmdable)
	db := ioc.InitDB()
	userDAO := dao.NewGORMUserDAO(db)
	userCache := cache.NewRedisUserCache(cmdable)
	userRepository := repository.NewCachedUserRepository(userDAO, userCache)
	userService := service.NewUserService(userRepository)
	userHandler := web.NewUserHandler(userService)
	engine := ioc.InitWebServer(v, userHandler)
	return engine

}
