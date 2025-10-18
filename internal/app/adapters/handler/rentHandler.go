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

func isCreateRentRequestCorrect(hotelRoomUuidStr, daysStr, dateToStr, dateFromStr string) error {
	if hotelRoomUuidStr == "" {
		return fmt.Errorf("query value 'id' is required")
	}

	if dateToStr != "" && daysStr != "" {
		return fmt.Errorf("you can choose only query values 'date-from' + 'date-to' or 'date-from' + 'days'")
	}

	if (dateToStr == "" && daysStr == "") || dateFromStr == "" {
		return fmt.Errorf("you must choose query values 'date-from' + 'date-to' or 'date-from' + 'days'")
	}

	return nil
}

func parseRentCreateRequest(c *fiber.Ctx, parseTo *rent.Rent) error {
	hotelRoomUuidStr := c.Query("id")
	dateFromStr := c.Query("date-from")
	dateToStr := c.Query("date-to")
	daysStr := c.Query("days")

	if err := isCreateRentRequestCorrect(hotelRoomUuidStr, daysStr, dateToStr, dateFromStr); err != nil {
		return err
	}

	userUuidStr := c.Locals("id").(string)
	userUuid, parseUserUuidErr := uuid.Parse(userUuidStr)
	if parseUserUuidErr != nil {
		flog.Warnf("Error parsing user UUID from JWT: %v\n", parseUserUuidErr)
		return fmt.Errorf("failed to parse user uuid from jwt: %v", parseUserUuidErr)
	}

	hotelRoomUuid, parseUuidErr := uuid.Parse(hotelRoomUuidStr)
	if parseUuidErr != nil {
		return fmt.Errorf("failed to parse query value 'hotel-room-id': %v", parseUuidErr)
	}

	dateFromParsed, parseTimeErr := httputils.ParseTimeDate(dateFromStr)
	if parseTimeErr != nil {
		return fmt.Errorf("failed to parse query value 'date-from', must match pattern 2016-10-06:  %v", parseTimeErr)
	}
	dateFrom := dateFromParsed.Unix()

	var dateTo int64
	if dateToStr != "" {
		dateToParsed, parseErr := httputils.ParseTimeDate(dateToStr)
		if parseErr != nil {
			return fmt.Errorf("failed to parse query value 'date-to', must match pattern 2016-10-06: %v", parseErr)
		}
		if dateFrom+24*60*60 > dateToParsed.Unix() {
			return fmt.Errorf("query value 'date-to' must be greater than 'date-from' atleast for 1 day")
		}
		dateTo = dateToParsed.Unix()
	}

	if daysStr != "" {
		days, parseErr := strconv.ParseUint(daysStr, 0, 64)
		if parseErr != nil {
			return fmt.Errorf("failed to parse query value 'days': %v", parseErr)
		}
		if days < 1 {
			return fmt.Errorf("query value 'days' must be greater than zero: %v", parseErr)
		}
		dateTo = dateFrom + int64(days)*24*60*60
	}

	if dateFrom > dateTo {
		return fmt.Errorf("query value 'date-from' must be greater than 'date-to'")
	}

	parseTo.Uuid = uuid.New()
	parseTo.RoomUuid = hotelRoomUuid
	parseTo.RenterUuid = userUuid
	parseTo.DateFrom = dateFrom
	parseTo.DateTo = dateTo

	return nil
}

func parseRentDeleteRequest(c *fiber.Ctx, parseTo *rent.DeleteDTO) error {
	rentUuidStr := c.Query("rent-id")

	userRole := c.Locals("role").(string)
	userUuidStr := c.Locals("id").(string)

	userUuid, parseUserUuidErr := uuid.Parse(userUuidStr)
	if parseUserUuidErr != nil {
		flog.Warnf("Error parsing user UUID from JWT: %v\n", parseUserUuidErr)
		return fmt.Errorf("failed to parse user uuid from jwt: %v", parseUserUuidErr)
	}

	if rentUuidStr == "" {
		return fmt.Errorf("query value 'rent-id' is required")
	}

	rentUuid, parseErr := uuid.Parse(rentUuidStr)
	if parseErr != nil {
		return fmt.Errorf("failed to parse query value 'rent-id': %v", parseErr)
	}

	parseTo.UserUuid = userUuid
	parseTo.UserRole = userRole
	parseTo.Uuid = rentUuid

	return nil
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

		unsuccessMessage := "failed to rent"

		var hotelRoomRent rent.Rent
		if err := parseRentCreateRequest(c, &hotelRoomRent); err != nil {
			return httputils.HandleUnsuccess(c, unsuccessMessage, fmt.Sprintf("%v", err), nil, fiber.StatusBadRequest)
		}

		if err := v.RentUsecase.Create(ctx, &hotelRoomRent); err != nil {
			return httputils.HandleUnsuccess(c, unsuccessMessage, fmt.Sprintf("%v", err), nil, fiber.StatusInternalServerError)
		}

		return httputils.HandleSuccess(c, "successfully rented the hotel room", &hotelRoomRent)
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

		unsuccessMessage := "failed to delete the rent"

		var deleteDto rent.DeleteDTO
		if err := parseRentDeleteRequest(c, &deleteDto); err != nil {
			return httputils.HandleUnsuccess(c, unsuccessMessage, fmt.Sprintf("%v", err), nil, fiber.StatusBadRequest)
		}

		if err := v.RentUsecase.Delete(ctx, deleteDto); err != nil {
			return httputils.HandleUnsuccess(c, unsuccessMessage, fmt.Sprintf("%v", err), nil, fiber.StatusInternalServerError)
		}

		return httputils.HandleSuccess(c, "successfully deleted the rent", nil)
	}
}
