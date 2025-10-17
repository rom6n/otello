package http

import (
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	flog "github.com/gofiber/fiber/v2/log"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/swagger"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/rom6n/otello/internal/app/adapters/handler"
	"github.com/rom6n/otello/internal/app/config"
	"github.com/rom6n/otello/internal/app/domain/user"
	"github.com/rom6n/otello/internal/utils/httputils"
	"github.com/rom6n/otello/internal/utils/jwtutils"
)

func NewFiberApp(cfg config.Config) *fiber.App {
	app := fiber.New()
	app.Use(logger.New())
	app.Use(limiter.New(limiter.Config{
		Max:               60,
		Expiration:        60 * time.Second,
		LimiterMiddleware: limiter.SlidingWindow{},
	}))
	app.Use(MaxBody)

	app.Get("/docs/*", swagger.HandlerDefault)

	api := app.Group("/api")

	userApi := api.Group("/user")
	hotelApi := api.Group("/hotel")
	hotelRoomApi := api.Group("/hotel-room")
	flightTicketApi := api.Group("/flight-ticket")

	adminApi := api.Group("/admin", CheckJwtMiddleware(cfg.JWTRepo, true))
	adminHotelApi := adminApi.Group("/hotel")
	adminHotelRoomApi := adminApi.Group("/hotel-room")
	adminFlightTicketApi := adminApi.Group("/flight-ticket")

	userHandler := handler.UserHandler{
		UserUsecase: cfg.UserUsecases,
	}

	hotelHandler := handler.HotelHandler{
		HotelUsecase: cfg.HotelUsecases,
	}

	hotelRoomHandler := handler.HotelRoomHandler{
		HotelRoomUsecase: cfg.HotelRoomUsecases,
	}

	rentHandler := handler.RentHandler{
		RentUsecase: cfg.RentUsecases,
	}

	flightTicketHandler := handler.FlightTicketHandler{
		FlightTicketUsecase: cfg.FlightTicketUsecases,
	}

	userApi.Post("/register", userHandler.Register())
	userApi.Post("/login", userHandler.Login())
	userApi.Post("/get-admin", CheckJwtMiddleware(cfg.JWTRepo, false), userHandler.GetAdminRole())
	userApi.Post("/get-user", CheckJwtMiddleware(cfg.JWTRepo, false), userHandler.GetUserRole())
	userApi.Put("/rename", CheckJwtMiddleware(cfg.JWTRepo, false), userHandler.ChangeName())

	hotelApi.Get("/find", hotelHandler.Find())

	hotelRoomApi.Get("/find", hotelRoomHandler.Find())
	hotelRoomApi.Post("/rent", CheckJwtMiddleware(cfg.JWTRepo, false), rentHandler.Create())
	hotelRoomApi.Post("/unrent", CheckJwtMiddleware(cfg.JWTRepo, false), rentHandler.Delete())

	flightTicketApi.Get("/find", flightTicketHandler.Find())
	flightTicketApi.Post("/buy", flightTicketHandler.Buy())

	// 5da2255a-1ce7-4427-ad44-862165ebf9d7
	adminHotelApi.Post("/create", hotelHandler.Create())
	adminHotelApi.Put("/update", hotelHandler.Update())
	adminHotelApi.Delete("/delete", hotelHandler.Delete())

	adminHotelRoomApi.Post("/create", hotelRoomHandler.Create())
	adminHotelRoomApi.Put("/update", hotelRoomHandler.Update())
	adminHotelRoomApi.Delete("/delete", hotelRoomHandler.Delete())

	adminFlightTicketApi.Post("/create", flightTicketHandler.Create())
	adminFlightTicketApi.Put("/update", flightTicketHandler.Update())
	adminFlightTicketApi.Delete("/delete", flightTicketHandler.Delete())

	return app
}

func CheckJwtMiddleware(jwtRepo jwtutils.JwtRepository, adminOnly bool) fiber.Handler {
	return func(c *fiber.Ctx) error {
		jwtToken := c.Cookies("jwtToken")

		var userUuid string
		var userRole user.UserRole
		if jwtToken == "" {
			var buildErr error
			userUuid, userRole, buildErr = buildAccessJwtTokenByRefreshJwtToken(c, jwtRepo)
			if buildErr != nil {
				return buildErr
			}
		}

		if userUuid == "" || userRole == "" {
			claims, err := jwtRepo.VerifyJwt(jwtToken)
			if err != nil {
				return httputils.HandleUnsuccess(c, "jwt token not verified or accepted", fmt.Sprintf("%v-%v", err, claims["iss"]), nil, fiber.StatusForbidden)
			}

			userUuid = claims["user_id"].(string)
			userRole = user.UserRole(claims["role"].(string))
		}

		if adminOnly && userRole != user.RoleAdmin {
			return httputils.HandleUnsuccess(c, "no permission", "", nil, fiber.StatusForbidden)
		}

		c.Locals("id", userUuid)
		c.Locals("role", string(userRole))
		flog.Info("MIDDLEWARE PASSED IP: ", c.IP())

		return c.Next()
	}
}

func buildAccessJwtTokenByRefreshJwtToken(c *fiber.Ctx, jwtRepo jwtutils.JwtRepository) (string, user.UserRole, error) {
	jwtRefreshToken := c.Cookies("jwtRefreshToken")
	if jwtRefreshToken == "" {
		return "", user.RoleUser, httputils.HandleUnsuccess(c, "login before be processed", "unauthorized", nil, fiber.StatusUnauthorized)
	}

	claims, err := jwtRepo.VerifyJwt(jwtRefreshToken)
	if err != nil {
		return "", user.RoleUser, httputils.HandleUnsuccess(c, "jwt refresh token not verified or accepted", fmt.Sprintf("%v-%v", err, claims["iss"]), nil, fiber.StatusForbidden)
	}

	userUuid, userRole, parseErr := parseUserDataFromClaims(c, claims)
	if parseErr != nil {
		return "", user.RoleUser, parseErr
	}

	jwtAccessToken, jwtErr := jwtRepo.NewJwt(userUuid, userRole, httputils.JwtAccessToken)
	if jwtErr != nil {
		flog.Warnf("Failed to create jwt access token: %v", jwtErr)
		return "", user.RoleUser, httputils.HandleUnsuccess(c, "login before be processed", "unauthorized", nil, fiber.StatusUnauthorized)
	}

	jwtAccessCookie := httputils.BuildCookie(jwtAccessToken, httputils.JwtAccessToken)
	c.Cookie(jwtAccessCookie)

	return userUuid.String(), userRole, nil
}

func parseUserDataFromClaims(c *fiber.Ctx, claims jwt.MapClaims) (uuid.UUID, user.UserRole, error) {
	userUuidStr := claims["user_id"].(string)
	userUuid, parseErr := uuid.Parse(userUuidStr)
	if parseErr != nil {
		flog.Warnf("Failed to parse user UUID: %v", parseErr)
		return uuid.Nil, user.RoleUser, httputils.HandleUnsuccess(c, "login before be processed", "unauthorized", nil, fiber.StatusUnauthorized)
	}

	var userRole user.UserRole
	switch claims["role"] {
	case "user":
		userRole = user.RoleUser
	case "admin":
		userRole = user.RoleAdmin
	default:
		flog.Warnf("User has not existed role: %v", claims["role"])
		return uuid.Nil, user.RoleUser, httputils.HandleUnsuccess(c, "login before be processed", "unauthorized", nil, fiber.StatusUnauthorized)
	}

	return userUuid, userRole, nil
}

func MaxBody(c *fiber.Ctx) error {
	maxSize := 1024 * 20
	if len(c.Body()) > maxSize {
		return c.Status(fiber.StatusRequestEntityTooLarge).SendString("Payload too large")
	}
	return c.Next()
}
