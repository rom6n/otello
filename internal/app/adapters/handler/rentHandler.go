package handler

import (
	"fmt"
	"strconv"

	"github.com/gofiber/fiber/v2"
	flog "github.com/gofiber/fiber/v2/log"
	"github.com/google/uuid"
	"github.com/rom6n/otello/internal/app/application/usecases/rentusecases"
	"github.com/rom6n/otello/internal/app/domain/rent"
	"github.com/rom6n/otello/internal/utils/httputils"
)

type RentHandler struct {
	RentUsecase rentusecases.RentUsecases
}

func (v *RentHandler) Create() fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx := c.Context()

		hotelRoomUuidStr := c.Query("hotel-room-id")
		dateFromStr := c.Query("date-from")
		dateToStr := c.Query("date-to")
		daysStr := c.Query("days")

		userUuidStr := c.Locals("id").(string)
		userUuid, parseUserUuidErr := uuid.Parse(userUuidStr)
		if parseUserUuidErr != nil {
			flog.Warnf("Error parsing user UUID from JWT: %v\n", parseUserUuidErr)
			return c.Status(fiber.StatusInternalServerError).JSON(httputils.Response{
				Success: false,
				Message: "failed to create rent",
				Error:   fmt.Sprintf("failed to parse user uuid %s", parseUserUuidErr),
			})
		}

		if hotelRoomUuidStr == "" {
			return c.Status(fiber.StatusBadRequest).JSON(httputils.Response{
				Success: false,
				Message: "failed to rent",
				Error:   "query value 'hotel-room-id' is required",
			})
		}

		if (dateToStr == "" && daysStr == "") || dateFromStr == "" {
			return c.Status(fiber.StatusBadRequest).JSON(httputils.Response{
				Success: false,
				Message: "failed to rent",
				Error:   "you must choose query values 'date-from' + 'date-to' or 'date-from' + 'days'",
			})
		}

		if dateToStr != "" && daysStr != "" {
			return c.Status(fiber.StatusBadRequest).JSON(httputils.Response{
				Success: false,
				Message: "failed to rent",
				Error:   "you can choose only query values 'date-from' + 'date-to' or 'date-from' + 'days'",
			})
		}

		hotelRoomUuid, parseErr := uuid.Parse(hotelRoomUuidStr)
		if parseErr != nil {
			return c.Status(fiber.StatusBadRequest).JSON(httputils.Response{
				Success: false,
				Message: "failed to rent",
				Error:   fmt.Sprintf("failed to parse query value 'hotel-room-id': %v", parseErr),
			})
		}

		dateFrom, parseErr := strconv.ParseUint(dateFromStr, 0, 64)
		if parseErr != nil {
			return c.Status(fiber.StatusBadRequest).JSON(httputils.Response{
				Success: false,
				Message: "failed to rent",
				Error:   fmt.Sprintf("failed to parse query value 'date-from': %v", parseErr),
			})
		}
		if dateFrom < 0 {
			return c.Status(fiber.StatusBadRequest).JSON(httputils.Response{
				Success: false,
				Message: "failed to rent",
				Error:   fmt.Sprintf("query value 'date-from' must be equal or greater than zero': %v", parseErr),
			})
		}

		var dateTo uint64
		if dateToStr != "" {
			var parseErr error
			dateTo, parseErr = strconv.ParseUint(dateToStr, 0, 64)
			if parseErr != nil {
				return c.Status(fiber.StatusBadRequest).JSON(httputils.Response{
					Success: false,
					Message: "failed to rent",
					Error:   fmt.Sprintf("failed to parse query value 'date-to': %v", parseErr),
				})
			}
			if dateTo < 1 {
				return c.Status(fiber.StatusBadRequest).JSON(httputils.Response{
					Success: false,
					Message: "failed to rent",
					Error:   fmt.Sprintf("query value 'date-to' must be greater than zero': %v", parseErr),
				})
			}
		}

		if daysStr != "" {
			days, parseErr := strconv.ParseUint(daysStr, 0, 64)
			if parseErr != nil {
				return c.Status(fiber.StatusBadRequest).JSON(httputils.Response{
					Success: false,
					Message: "failed to rent",
					Error:   fmt.Sprintf("failed to parse query value 'days': %v", parseErr),
				})
			}
			if days < 1 {
				return c.Status(fiber.StatusBadRequest).JSON(httputils.Response{
					Success: false,
					Message: "failed to rent",
					Error:   fmt.Sprintf("query value 'days' must be greater than zero: %v", parseErr),
				})
			}
			dateTo = dateFrom + uint64(days)*24*60*60
		}

		if dateFrom > dateTo {
			return c.Status(fiber.StatusBadRequest).JSON(httputils.Response{
				Success: false,
				Message: "failed to rent",
				Error:   fmt.Sprintf("query value 'date-from' must be greater than 'date-to'"),
			})
		}

		hotelRoomRent := rent.NewRent(hotelRoomUuid, userUuid, dateFrom, dateTo)

		if err := v.RentUsecase.Create(ctx, hotelRoomRent); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(httputils.Response{
				Success: false,
				Message: "failed to create rent",
				Error:   fmt.Sprintf("%v", err),
			})
		}

		return c.JSON(httputils.Response{
			Success: true,
			Message: "successfully created rent",
			Data:    hotelRoomRent,
		})
	}
}

func (v *RentHandler) Delete() fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx := c.Context()

		rentUuidStr := c.Query("rent-id")

		userRole := c.Locals("role").(string)
		userUuidStr := c.Locals("id").(string)

		userUuid, parseUserUuidErr := uuid.Parse(userUuidStr)
		if parseUserUuidErr != nil {
			flog.Warnf("Error parsing user UUID from JWT: %v\n", parseUserUuidErr)
			return c.Status(fiber.StatusInternalServerError).JSON(httputils.Response{
				Success: false,
				Message: "failed to delete",
				Error:   fmt.Sprintf("failed to parse user uuid from jwt: %v", parseUserUuidErr),
			})
		}

		if rentUuidStr == "" {
			return c.Status(fiber.StatusBadRequest).JSON(httputils.Response{
				Success: false,
				Message: "failed to delete rent",
				Error:   "query value 'rent-id' is required",
			})
		}

		rentUuid, parseErr := uuid.Parse(rentUuidStr)
		if parseErr != nil {
			return c.Status(fiber.StatusBadRequest).JSON(httputils.Response{
				Success: false,
				Message: "failed to delete rent",
				Error:   fmt.Sprintf("failed to parse query value 'rent-id': %v", parseErr),
			})
		}

		if err := v.RentUsecase.Delete(ctx, rentUuid, userUuid, userRole); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(httputils.Response{
				Success: false,
				Message: "failed to delete rent",
				Error:   fmt.Sprintf("%v", err),
			})
		}

		return c.Status(fiber.StatusOK).JSON(httputils.Response{
			Success: true,
			Message: "successfully deleted rent",
		})
	}
}
