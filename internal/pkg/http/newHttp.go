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

const (
	maxBodySize = 1024 * 20 // 20 KB

	maxRequestsPerWindow = 90
	timeWindow           = 60 * time.Second
)

type handlers struct {
	userHandler         handler.UserHandler
	hotelHandler        handler.HotelHandler
	hotelRoomHandler    handler.HotelRoomHandler
	rentHandler         handler.RentHandler
	flightTicketHandler handler.FlightTicketHandler
}

type adminAPIs struct {
	hotelApi        fiber.Router
	hotelRoomApi    fiber.Router
	flightTicketApi fiber.Router
}

type userAPIs struct {
	userApi         fiber.Router
	hotelApi        fiber.Router
	hotelRoomApi    fiber.Router
	flightTicketApi fiber.Router
}

func NewFiberApp(cfg config.Config) *fiber.App {
	app := fiber.New()
	app.Use(logger.New())
	app.Use(limiter.New(limiter.Config{
		Max:               maxRequestsPerWindow,
		Expiration:        timeWindow,
		LimiterMiddleware: limiter.SlidingWindow{},
	}))
	app.Use(maxBodyLimitMiddleware)
	CheckAuthorized := checkJwtMiddleware(cfg.JWTRepo, false)

	createdAdminAPI, createdUserAPI := createAPIGroups(app, cfg)
	createdHandlers := createHandlers(cfg)
	connectUserRoutes(app, createdHandlers, createdUserAPI, CheckAuthorized)
	connectAdminRoutes(createdHandlers, createdAdminAPI)

	return app
}

func createAPIGroups(app *fiber.App, cfg config.Config) (adminAPIs, userAPIs) {
	api := app.Group("/api")

	userApi := api.Group("/user")
	hotelApi := api.Group("/hotel")
	hotelRoomApi := api.Group("/hotel-room")
	flightTicketApi := api.Group("/flight-ticket")

	adminApi := api.Group("/admin", checkJwtMiddleware(cfg.JWTRepo, true))

	adminHotelApi := adminApi.Group("/hotel")
	adminHotelRoomApi := adminApi.Group("/hotel-room")
	adminFlightTicketApi := adminApi.Group("/flight-ticket")

	newAdminApi := adminAPIs{
		hotelApi:        adminHotelApi,
		hotelRoomApi:    adminHotelRoomApi,
		flightTicketApi: adminFlightTicketApi,
	}

	newUserApi := userAPIs{
		userApi:         userApi,
		hotelApi:        hotelApi,
		hotelRoomApi:    hotelRoomApi,
		flightTicketApi: flightTicketApi,
	}

	return newAdminApi, newUserApi
}

func connectAdminRoutes(handlers handlers, createdAdminApi adminAPIs) {
	createdAdminApi.hotelApi.Post("/create", handlers.hotelHandler.Create())
	createdAdminApi.hotelApi.Put("/update", handlers.hotelHandler.Update())
	createdAdminApi.hotelApi.Delete("/delete", handlers.hotelHandler.Delete())

	createdAdminApi.hotelRoomApi.Post("/create", handlers.hotelRoomHandler.Create())
	createdAdminApi.hotelRoomApi.Put("/update", handlers.hotelRoomHandler.Update())
	createdAdminApi.hotelRoomApi.Delete("/delete", handlers.hotelRoomHandler.Delete())

	createdAdminApi.flightTicketApi.Post("/create", handlers.flightTicketHandler.Create())
	createdAdminApi.flightTicketApi.Put("/update", handlers.flightTicketHandler.Update())
	createdAdminApi.flightTicketApi.Delete("/delete", handlers.flightTicketHandler.Delete())
}

func connectUserRoutes(app *fiber.App, handlers handlers, createdUserApi userAPIs, CheckAuthorized fiber.Handler) {
	// Swagger docs route
	app.Get("/docs/*", swagger.HandlerDefault)

	createdUserApi.userApi.Post("/register", handlers.userHandler.Register())
	createdUserApi.userApi.Post("/login", handlers.userHandler.Login())
	createdUserApi.userApi.Post("/get-admin", CheckAuthorized, handlers.userHandler.GetAdminRole())
	createdUserApi.userApi.Post("/get-user", CheckAuthorized, handlers.userHandler.GetUserRole())
	createdUserApi.userApi.Put("/rename", CheckAuthorized, handlers.userHandler.ChangeName())

	createdUserApi.hotelApi.Get("/find", handlers.hotelHandler.Find())

	createdUserApi.hotelRoomApi.Get("/find", handlers.hotelRoomHandler.Find())
	createdUserApi.hotelRoomApi.Post("/rent", CheckAuthorized, handlers.rentHandler.Create())
	createdUserApi.hotelRoomApi.Post("/unrent", CheckAuthorized, handlers.rentHandler.Delete())

	createdUserApi.flightTicketApi.Get("/find", handlers.flightTicketHandler.Find())
	createdUserApi.flightTicketApi.Post("/buy", handlers.flightTicketHandler.Buy())
}

func createHandlers(cfg config.Config) handlers {
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
	return handlers{
		userHandler:         userHandler,
		hotelHandler:        hotelHandler,
		hotelRoomHandler:    hotelRoomHandler,
		rentHandler:         rentHandler,
		flightTicketHandler: flightTicketHandler,
	}
}

func checkJwtMiddleware(jwtRepo jwtutils.JwtRepository, adminOnly bool) fiber.Handler {
	return func(c *fiber.Ctx) error {
		errLoginBeforeBeProcessed := httputils.HandleUnsuccess(c, "login before be processed", "unauthorized", nil, fiber.StatusUnauthorized)

		jwtToken := c.Cookies("jwtToken")

		var userUuid string
		var userRole string
		if jwtToken == "" {
			gottenUserUuid, gottenUserRole, buildErr := buildAccessJwtTokenByRefreshJwtToken(c, jwtRepo)
			if buildErr != nil {
				return buildErr
			}
			userUuid = gottenUserUuid
			userRole = string(gottenUserRole)
		}

		if userUuid == "" || userRole == "" {
			claims, err := jwtRepo.VerifyJwt(jwtToken)
			if err != nil {
				return httputils.HandleUnsuccess(c, "jwt token not verified or accepted", fmt.Sprintf("%v-%v", err, claims["iss"]), nil, fiber.StatusForbidden)
			}

			var ok bool
			if userUuid, ok = claims["user_id"].(string); !ok {
				flog.Warnf("User has not string-type user_id in claims")
				return errLoginBeforeBeProcessed
			}

			if userRole, ok = claims["role"].(string); !ok {
				flog.Warnf("User has not string-type role")
				return errLoginBeforeBeProcessed
			}
		}

		if adminOnly && userRole != string(user.RoleAdmin) {
			return httputils.HandleUnsuccess(c, "no permission", "", nil, fiber.StatusForbidden)
		}

		c.Locals("id", userUuid)
		c.Locals("role", userRole)
		flog.Info("MIDDLEWARE PASSED IP: ", c.IP())

		return c.Next()
	}
}

func buildAccessJwtTokenByRefreshJwtToken(c *fiber.Ctx, jwtRepo jwtutils.JwtRepository) (string, user.UserRole, error) {
	errLoginBeforeBeProcessed := httputils.HandleUnsuccess(c, "login before be processed", "unauthorized", nil, fiber.StatusUnauthorized)

	jwtRefreshToken := c.Cookies("jwtRefreshToken")
	if jwtRefreshToken == "" {
		return "", user.RoleUser, errLoginBeforeBeProcessed
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
		return "", user.RoleUser, errLoginBeforeBeProcessed
	}

	jwtAccessCookie := httputils.BuildCookie(jwtAccessToken, httputils.JwtAccessToken)
	c.Cookie(jwtAccessCookie)

	return userUuid.String(), userRole, nil
}

func parseUserDataFromClaims(c *fiber.Ctx, claims jwt.MapClaims) (uuid.UUID, user.UserRole, error) {
	errLoginBeforeBeProcessed := httputils.HandleUnsuccess(c, "login before be processed", "unauthorized", nil, fiber.StatusUnauthorized)

	userUuidStr, ok := claims["user_id"].(string)
	if !ok {
		flog.Warnf("User has not string-type user_id in claims")
		return uuid.Nil, user.RoleUser, errLoginBeforeBeProcessed
	}
	userUuid, parseErr := uuid.Parse(userUuidStr)
	if parseErr != nil {
		flog.Warnf("Failed to parse user UUID: %v", parseErr)
		return uuid.Nil, user.RoleUser, errLoginBeforeBeProcessed
	}

	var userRole user.UserRole
	switch claims["role"] {
	case "user":
		userRole = user.RoleUser
	case "admin":
		userRole = user.RoleAdmin
	default:
		flog.Warnf("User has not existed role: %v", claims["role"])
		return uuid.Nil, user.RoleUser, errLoginBeforeBeProcessed
	}

	return userUuid, userRole, nil
}

func maxBodyLimitMiddleware(c *fiber.Ctx) error {
	maxSize := maxBodySize
	if len(c.Body()) > maxSize {
		return httputils.HandleUnsuccess(c, "payload too large", "", nil, fiber.StatusRequestEntityTooLarge)
	}
	return c.Next()
}
