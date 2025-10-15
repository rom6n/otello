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

// @Summary Забронировать номер отеля
// @Description Бронирует номер отеля на заданные даты. Дата выезда считается свободной для другого бронирования. 'days' нельзя использовать вместе с 'date-to'. Требуется авторизация
// @Tags Номер отеля
// @Accept json
// @Produce json
// @Param id query string true "ID номера отеля"
// @Param date-from query string true "Дата заселения (прим. 2016-10-06, год-месяц-день)"
// @Param date-to query string false "Дата выезда (прим. 2016-10-06, год-месяц-день) (по выбору)"
// @Param days query int false "Количество дней (по выбору)"
// @Success 200 {object} httputils.SuccessResponse{data=rent.Rent}
// @Failure 400 {object} httputils.ErrorResponse
// @Failure 500 {object} httputils.ErrorResponse
// @Router /api/hotel-room/rent [post]
func (v *RentHandler) Create() fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx := c.Context()

		hotelRoomUuidStr := c.Query("id")
		dateFromStr := c.Query("date-from")
		dateToStr := c.Query("date-to")
		daysStr := c.Query("days")

		userUuidStr := c.Locals("id").(string)
		userUuid, parseUserUuidErr := uuid.Parse(userUuidStr)
		if parseUserUuidErr != nil {
			flog.Warnf("Error parsing user UUID from JWT: %v\n", parseUserUuidErr)
			return c.Status(fiber.StatusInternalServerError).JSON(httputils.Response{
				Success: false,
				Message: "failed to rent",
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

		dateFromParsed, parseErr := httputils.ParseTimeDate(dateFromStr)
		if parseErr != nil {
			return c.Status(fiber.StatusBadRequest).JSON(httputils.Response{
				Success: false,
				Message: "failed to rent",
				Error:   fmt.Sprintf("failed to parse query value 'date-from', must match pattern 2016-10-06:  %v", parseErr),
			})
		}
		dateFrom := dateFromParsed.Unix()

		var dateTo int64
		if dateToStr != "" {
			dateToParsed, parseErr := httputils.ParseTimeDate(dateToStr)
			if parseErr != nil {
				return c.Status(fiber.StatusBadRequest).JSON(httputils.Response{
					Success: false,
					Message: "failed to rent",
					Error:   fmt.Sprintf("failed to parse query value 'date-to', must match pattern 2016-10-06: %v", parseErr),
				})
			}
			if dateFrom+24*60*60 > dateToParsed.Unix() {
				return c.Status(fiber.StatusBadRequest).JSON(httputils.Response{
					Success: false,
					Message: "failed to rent",
					Error:   fmt.Sprintf("query value 'date-to' must be greater than 'date-from' atleast for 1 day"),
				})
			}
			dateTo = dateToParsed.Unix()
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
			dateTo = dateFrom + int64(days)*24*60*60
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
			Message: "successfully rented the hotel room",
			Data:    hotelRoomRent,
		})
	}
}

// @Summary Отменить бронь номера отеля
// @Description Отменяет бронь номер отеля по ID брони. Админ может удалить бронь у любого пользователя. Требуется авторизация
// @Tags Номер отеля
// @Accept json
// @Produce json
// @Param rent-id query string true "ID брони"
// @Success 200 {object} httputils.SuccessResponse
// @Failure 400 {object} httputils.ErrorResponse
// @Failure 500 {object} httputils.ErrorResponse
// @Router /api/hotel-room/unrent [post]
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
				Message: "failed to unrent",
				Error:   fmt.Sprintf("failed to parse user uuid from jwt: %v", parseUserUuidErr),
			})
		}

		if rentUuidStr == "" {
			return c.Status(fiber.StatusBadRequest).JSON(httputils.Response{
				Success: false,
				Message: "failed to unrent",
				Error:   "query value 'rent-id' is required",
			})
		}

		rentUuid, parseErr := uuid.Parse(rentUuidStr)
		if parseErr != nil {
			return c.Status(fiber.StatusBadRequest).JSON(httputils.Response{
				Success: false,
				Message: "failed to unrent",
				Error:   fmt.Sprintf("failed to parse query value 'rent-id': %v", parseErr),
			})
		}

		if err := v.RentUsecase.Delete(ctx, rentUuid, userUuid, userRole); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(httputils.Response{
				Success: false,
				Message: "failed to unrent",
				Error:   fmt.Sprintf("%v", err),
			})
		}

		return c.Status(fiber.StatusOK).JSON(httputils.Response{
			Success: true,
			Message: "successfully deleted rent",
		})
	}
}
