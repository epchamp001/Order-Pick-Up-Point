package app

import (
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	controller "order-pick-up-point/internal/controller/http"
	"order-pick-up-point/internal/controller/http/middleware"
	"order-pick-up-point/internal/metrics"
	"order-pick-up-point/pkg/jwt"
)

func SetupRoutes(router *gin.Engine, authCtrl controller.AuthController, pvzCtrl controller.PvzController, tokenSvc jwt.TokenService) {
	router.Use(metrics.GinPrometheusMiddleware())

	router.POST("/dummyLogin", authCtrl.DummyLogin)
	router.POST("/register", authCtrl.Register)
	router.POST("/login", authCtrl.Login)

	router.StaticFile("/swagger/spec/http/swagger.json", "./docs/swagger.json")
	router.GET("/swagger/http/*any", ginSwagger.WrapHandler(swaggerFiles.Handler, ginSwagger.URL("/swagger/spec/http/swagger.json")))

	router.StaticFile("/swagger/spec/grpc/swagger.json", "./docs/pvz.swagger.json")
	router.GET("/swagger/grpc/*any", ginSwagger.WrapHandler(swaggerFiles.Handler, ginSwagger.URL("/swagger/spec/grpc/swagger.json")))

	protected := router.Group("/")
	protected.Use(middleware.JWTAuthMiddleware(tokenSvc))
	{
		protected.POST("/pvz", pvzCtrl.CreatePvz)
		protected.POST("/receptions", pvzCtrl.CreateReception)
		protected.POST("/products", pvzCtrl.AddProduct)
		protected.POST("/pvz/:pvzId/delete_last_product", func(c *gin.Context) {
			c.Set("pvzId", c.Param("pvzId"))
			pvzCtrl.DeleteLastProduct(c)
		})
		protected.POST("/pvz/:pvzId/close_last_reception", func(c *gin.Context) {
			c.Set("pvzId", c.Param("pvzId"))
			pvzCtrl.CloseReception(c)
		})
		protected.GET("/pvz", pvzCtrl.GetPvzsInfo)
	}
}
