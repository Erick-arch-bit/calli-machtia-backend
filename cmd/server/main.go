package main

import (
	"context"
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/calli-machtia/backend/internal/config"
	"github.com/calli-machtia/backend/internal/database"
	"github.com/calli-machtia/backend/internal/handlers"
	"github.com/calli-machtia/backend/internal/middleware"
	"github.com/calli-machtia/backend/internal/repository"
	"github.com/calli-machtia/backend/internal/services"
)

func main() {
	cfg := config.Load()

	pg := database.InitPostgres(cfg.DATABASE_URL)
	defer pg.Close()

	mdb := database.InitMongoDB(cfg.MONGODB_URI, cfg.MONGODB_DATABASE)
	rdb := database.InitRedis(cfg.REDIS_URL)
	if rdb != nil {
		defer rdb.Close()
	}

	userRepo := repository.NewUserRepo(pg)
	courseRepo := repository.NewCourseRepo(pg)
	enrollRepo := repository.NewEnrollmentRepo(pg)
	paymentRepo := repository.NewPaymentRepo(pg)
	lessonRepo := repository.NewLessonRepo(mdb)
	resetRepo := repository.NewPasswordResetRepo(pg)

	authSvc := services.NewAuthService(cfg, rdb)
	stripeSvc := services.NewStripeService(cfg, paymentRepo, enrollRepo)

	authHandler := handlers.NewAuthHandler(authSvc, userRepo, resetRepo)
	courseHandler := handlers.NewCourseHandler(courseRepo, lessonRepo)
	enrollHandler := handlers.NewEnrollmentHandler(enrollRepo, courseRepo)
	lessonHandler := handlers.NewLessonHandler(lessonRepo)
	paymentHandler := handlers.NewPaymentHandler(stripeSvc, paymentRepo)
	adminHandler := handlers.NewAdminHandler(userRepo, courseRepo, paymentRepo, enrollRepo)

	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(gin.Logger())
	r.Use(middleware.CORSMiddleware(cfg.CORS_ORIGIN))
	r.Use(middleware.RateLimitMiddleware(rdb))

	r.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "Calli Machtia API", "version": "1.0.0"})
	})

	r.GET("/health", func(c *gin.Context) {
		dbOK := false
		if err := pg.Ping(context.Background()); err == nil {
			dbOK = true
		}
		mdbOK := mdb != nil
		rdbOK := rdb != nil
		if rdb != nil {
			if err := rdb.Ping(context.Background()).Err(); err != nil {
				rdbOK = false
			}
		}
		c.JSON(200, gin.H{
			"status":   "ok",
			"version":  "1.0.0",
			"database": dbOK,
			"mongodb":  mdbOK,
			"redis":    rdbOK,
		})
	})

	api := r.Group("/api")
	{
		auth := api.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
			auth.GET("/me", middleware.AuthMiddleware(cfg.JWT_SECRET), authHandler.Me)
			auth.POST("/refresh", authHandler.Refresh)
			auth.PUT("/profile", middleware.AuthMiddleware(cfg.JWT_SECRET), authHandler.UpdateProfile)
			auth.POST("/logout", middleware.AuthMiddleware(cfg.JWT_SECRET), authHandler.Logout)
			auth.POST("/forgot-password", authHandler.ForgotPassword)
			auth.POST("/reset-password", authHandler.ResetPassword)
		}

		api.GET("/categories", courseHandler.ListCategories)

		courses := api.Group("/courses")
		{
			courses.GET("/instructor/mine", middleware.AuthMiddleware(cfg.JWT_SECRET), middleware.RoleMiddleware("instructor"), courseHandler.MyCourses)
			courses.GET("", courseHandler.List)
			courses.GET("/:slug", courseHandler.GetBySlug)
			courses.POST("", middleware.AuthMiddleware(cfg.JWT_SECRET), middleware.RoleMiddleware("instructor", "admin"), courseHandler.Create)
			courses.PUT("/:id", middleware.AuthMiddleware(cfg.JWT_SECRET), middleware.RoleMiddleware("instructor", "admin"), courseHandler.Update)
			courses.DELETE("/:id", middleware.AuthMiddleware(cfg.JWT_SECRET), middleware.RoleMiddleware("instructor", "admin"), courseHandler.Delete)
		}

		enrollments := api.Group("/enrollments")
		{
			enrollments.POST("", middleware.AuthMiddleware(cfg.JWT_SECRET), enrollHandler.Enroll)
			enrollments.GET("/mine", middleware.AuthMiddleware(cfg.JWT_SECRET), enrollHandler.MyEnrollments)
			enrollments.GET("/course/:courseId", middleware.AuthMiddleware(cfg.JWT_SECRET), middleware.RoleMiddleware("instructor", "admin"), enrollHandler.CourseEnrollments)
			enrollments.PUT("/:id/progress", middleware.AuthMiddleware(cfg.JWT_SECRET), enrollHandler.UpdateProgress)
			enrollments.DELETE("/:id", middleware.AuthMiddleware(cfg.JWT_SECRET), enrollHandler.Unenroll)
		}

		content := api.Group("/courses/:courseId/modules")
		{
			content.GET("", lessonHandler.GetModules)
			content.GET("/:moduleId", lessonHandler.GetModule)
			content.POST("", middleware.AuthMiddleware(cfg.JWT_SECRET), middleware.RoleMiddleware("instructor", "admin"), lessonHandler.CreateModule)
			content.PUT("/:moduleId", middleware.AuthMiddleware(cfg.JWT_SECRET), middleware.RoleMiddleware("instructor", "admin"), lessonHandler.UpdateModule)
			content.DELETE("/:moduleId", middleware.AuthMiddleware(cfg.JWT_SECRET), middleware.RoleMiddleware("instructor", "admin"), lessonHandler.DeleteModule)
			content.POST("/:moduleId/lessons", middleware.AuthMiddleware(cfg.JWT_SECRET), middleware.RoleMiddleware("instructor", "admin"), lessonHandler.AddLesson)
			content.PUT("/:moduleId/lessons/:lessonId", middleware.AuthMiddleware(cfg.JWT_SECRET), middleware.RoleMiddleware("instructor", "admin"), lessonHandler.UpdateLesson)
			content.DELETE("/:moduleId/lessons/:lessonId", middleware.AuthMiddleware(cfg.JWT_SECRET), middleware.RoleMiddleware("instructor", "admin"), lessonHandler.DeleteLesson)
		}

		payments := api.Group("/payments")
		{
			payments.POST("/create-intent", middleware.AuthMiddleware(cfg.JWT_SECRET), paymentHandler.CreateIntent)
			payments.POST("/webhook", paymentHandler.Webhook)
			payments.GET("", middleware.AuthMiddleware(cfg.JWT_SECRET), paymentHandler.History)
		}

		admin := api.Group("/admin", middleware.AuthMiddleware(cfg.JWT_SECRET), middleware.RoleMiddleware("admin"))
		{
			admin.GET("/users", adminHandler.ListUsers)
			admin.PUT("/users/:id/role", adminHandler.UpdateUserRole)
			admin.GET("/courses", adminHandler.ListCourses)
			admin.GET("/stats", adminHandler.Stats)
			admin.DELETE("/courses/:id", adminHandler.ForceDeleteCourse)
		}
	}

	addr := fmt.Sprintf(":%s", cfg.PORT)
	log.Printf("Server starting on %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
