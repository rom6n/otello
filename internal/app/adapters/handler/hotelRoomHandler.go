package handler

import (
	"fmt"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/rom6n/otello/internal/app/application/usecases/hotelroomusecases"
	"github.com/rom6n/otello/internal/app/domain/hotelroom"
	"github.com/rom6n/otello/internal/utils/httputils"
)

type HotelRoomHandler struct {
	HotelRoomUsecase hotelroomusecases.HotelRoomUsecases
}

func ParseHotelRoomParams(hotelUuidStr, roomsStr, typeStr, amountPeopleStr, valueStr, hotelRoomUuidStr string, allRequired bool, oldHotelRoom *hotelroom.HotelRoom) error {
	if allRequired || hotelRoomUuidStr != "" {
		hotelRoomUuid, parseErr := uuid.Parse(hotelRoomUuidStr)
		if parseErr != nil {
			return fmt.Errorf("failed to parse query value 'id': %v", parseErr)
		}
		oldHotelRoom.Uuid = hotelRoomUuid
	}

	if allRequired || hotelUuidStr != "" {
		hotelUuid, parseErr := uuid.Parse(hotelUuidStr)
		if parseErr != nil {
			return fmt.Errorf("failed to parse query value 'hotel-id': %v", parseErr)
		}
		oldHotelRoom.HotelUuid = hotelUuid
	}

	if allRequired || roomsStr != "" {
		rooms, parseErr := strconv.ParseUint(roomsStr, 0, 32)
		if parseErr != nil {
			return fmt.Errorf("failed to parse query valuer 'rooms': %v", parseErr)
		}
		if rooms == 0 {
			return fmt.Errorf("query value 'rooms' must be greater than zero")
		}
		oldHotelRoom.Rooms = uint32(rooms)
	}

	if allRequired || amountPeopleStr != "" {
		amountPeople, parseErr := strconv.ParseUint(amountPeopleStr, 0, 32)
		if parseErr != nil {
			return fmt.Errorf("failed to parse query value 'amount-people': %v", parseErr)
		}
		if amountPeople == 0 {
			return fmt.Errorf("query value 'amount-people' must be greater than zero")
		}
		oldHotelRoom.AmountPeople = uint32(amountPeople)
	}

	if allRequired || valueStr != "" {
		value, parseErr := strconv.ParseInt(valueStr, 0, 32)
		if parseErr != nil {
			return fmt.Errorf("failed to parse query value 'value': %v", parseErr)
		}
		if value < 0 {
			return fmt.Errorf("query value 'value' must be equal or greater than zero")
		}
		oldHotelRoom.Value = &value
	}

	if allRequired || typeStr != "" {
		var Type string
		switch typeStr {
		case string(hotelroom.Standard):
			Type = string(hotelroom.Standard)
		case string(hotelroom.Lagre):
			Type = string(hotelroom.Lagre)
		case string(hotelroom.Premium):
			Type = string(hotelroom.Premium)
		default:
			return fmt.Errorf("value of query value 'type' is incorrect. pick any of: 'standard', 'large', 'premium'")
		}
		oldHotelRoom.Type = hotelroom.HotelType(Type)
	}

	return nil
}

// @Summary Создать номер отеля (Admin only)
// @Description Создаёт номер отеля с переданными параметрами
// @Tags Номер отеля
// @Accept json
// @Produce json
// @Param hotel-id query string true "ID отеля"
// @Param rooms query int true "Количество комнат"
// @Param type query string true "Тип номера (standard, large, premium)"
// @Param amount-people query int true "Количество человек"
// @Param value query int true "Цена за номер"
// @Success 200 {object} httputils.SuccessResponse{data=hotelroom.HotelRoom}
// @Failure 400 {object} httputils.ErrorResponse
// @Failure 500 {object} httputils.ErrorResponse
// @Router /api/admin/hotel-room/create [post]
func (v *HotelRoomHandler) Create() fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx := c.Context()

		hotelUuidStr := c.Query("hotel-id")
		roomsStr := c.Query("rooms")
		typeStr := c.Query("type")
		amountPeopleStr := c.Query("amount-people")
		valueStr := c.Query("value")

		if hotelUuidStr == "" || roomsStr == "" || typeStr == "" || amountPeopleStr == "" || valueStr == "" {
			return c.Status(fiber.StatusBadRequest).JSON(httputils.Response{
				Success: false,
				Message: "failed to create hotel room",
				Error:   "query values 'hotel-id', 'rooms', 'type', 'amount-people' and 'value' are required",
			})
		}

		var hotelRoom hotelroom.HotelRoom
		hotelRoom.Uuid = uuid.New()

		err := ParseHotelRoomParams(hotelUuidStr, roomsStr, typeStr, amountPeopleStr, valueStr, hotelRoom.Uuid.String(), true, &hotelRoom)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(httputils.Response{
				Success: false,
				Message: "failed to create hotel room",
				Error:   fmt.Sprintf("%v", err),
			})
		}

		if err := v.HotelRoomUsecase.Create(ctx, &hotelRoom); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(httputils.Response{
				Success: false,
				Message: "failed to create hotel room",
				Error:   fmt.Sprintf("%v", err),
			})
		}

		return c.JSON(httputils.Response{
			Success: true,
			Message: "successfully created hotel room",
			Data:    hotelRoom,
		})
	}
}

// @Summary Изменить номер отеля (Admin only)
// @Description Изменяет номер отеля переданными параметрами
// @Tags Номер отеля
// @Accept json
// @Produce json
// @Param id query string true "ID номера отеля"
// @Param hotel-id query string false "Новый ID отеля (необязательно)"
// @Param rooms query int false "Новое количество комнат (необязательно)"
// @Param type query string false "Новый тип номера (standard, large, premium) (необязательно)"
// @Param amount-people query int false "Новое количество человек (необязательно)"
// @Param value query int false "Новая цена за номер (необязательно)"
// @Success 200 {object} httputils.SuccessResponse{data=hotelroom.HotelRoom}
// @Failure 400 {object} httputils.ErrorResponse
// @Failure 500 {object} httputils.ErrorResponse
// @Router /api/admin/hotel-room/update [put]
func (v *HotelRoomHandler) Update() fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx := c.Context()

		hotelRoomUuidStr := c.Query("id")
		hotelUuidStr := c.Query("hotel-id")
		roomsStr := c.Query("rooms")
		typeStr := c.Query("type")
		amountPeopleStr := c.Query("amount-people")
		valueStr := c.Query("value")

		if hotelRoomUuidStr == "" {
			return c.Status(fiber.StatusBadRequest).JSON(httputils.Response{
				Success: false,
				Message: "failed to update hotel room",
				Error:   "query value 'id' is required",
			})
		}

		var foundedHotelRoom hotelroom.HotelRoom

		parseErr1 := ParseHotelRoomParams(hotelUuidStr, roomsStr, typeStr, amountPeopleStr, valueStr, hotelRoomUuidStr, false, &foundedHotelRoom)
		if parseErr1 != nil {
			return c.Status(fiber.StatusBadRequest).JSON(httputils.Response{
				Success: false,
				Message: "failed to update hotel room",
				Error:   fmt.Sprintf("%v", parseErr1),
			})
		}

		gottenHotelRoom, getErr := v.HotelRoomUsecase.Get(ctx, foundedHotelRoom.Uuid)
		foundedHotelRoom = *gottenHotelRoom
		if getErr != nil {
			return c.Status(fiber.StatusNotFound).JSON(httputils.Response{
				Success: false,
				Message: "failed to update hotel room",
				Error:   fmt.Sprintf("failed to get hotel room: %v", getErr),
			})
		}

		parseErr2 := ParseHotelRoomParams(hotelUuidStr, roomsStr, typeStr, amountPeopleStr, valueStr, hotelRoomUuidStr, false, &foundedHotelRoom)
		if parseErr2 != nil {
			return c.Status(fiber.StatusBadRequest).JSON(httputils.Response{
				Success: false,
				Message: "failed to update hotel room",
				Error:   fmt.Sprintf("%v", parseErr2),
			})
		}

		if err := v.HotelRoomUsecase.Update(ctx, &foundedHotelRoom); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(httputils.Response{
				Success: false,
				Message: "failed to update hotel room",
				Error:   fmt.Sprintf("%v", err),
			})
		}

		return c.JSON(httputils.Response{
			Success: true,
			Message: "successfully updated hotel room",
			Data:    foundedHotelRoom,
		})
	}
}

// @Summary Удалить номер отеля (Admin only)
// @Description Удаляет номер отеля по ID
// @Tags Номер отеля
// @Accept json
// @Produce json
// @Param id query string true "ID номера отеля"
// @Success 200 {object} httputils.SuccessResponse
// @Failure 400 {object} httputils.ErrorResponse
// @Failure 500 {object} httputils.ErrorResponse
// @Router /api/admin/hotel-room/delete [delete]
func (v *HotelRoomHandler) Delete() fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx := c.Context()

		hotelRoomUuidStr := c.Query("id")

		if hotelRoomUuidStr == "" {
			return c.Status(fiber.StatusBadRequest).JSON(httputils.Response{
				Success: false,
				Message: "failed to delete hotel room",
				Error:   "query value 'id' is required",
			})
		}

		hotelRoomUuid, parseErr := uuid.Parse(hotelRoomUuidStr)
		if parseErr != nil {
			return c.Status(fiber.StatusBadRequest).JSON(httputils.Response{
				Success: false,
				Message: "failed to delete hotel room",
				Error:   fmt.Sprintf("failed to parse query value 'id': %v", parseErr),
			})
		}

		if err := v.HotelRoomUsecase.Delete(ctx, hotelRoomUuid); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(httputils.Response{
				Success: false,
				Message: "failed to delete hotel room",
				Error:   fmt.Sprintf("%v", err),
			})
		}

		return c.JSON(httputils.Response{
			Success: true,
			Message: "successfully deleted hotel room",
		})
	}
}

// @Summary Найти номер отеля
// @Description Находит номер отеля по фильтрам. Можно искать без фильтров. Нельзя использовать одновременно 'date-to' и 'days'.  Если указать 'arrange', то билеты будут отсортированы по цене в указанном порядке (asc - по возрастанию, desc - по убыванию)
// @Tags Номер отеля
// @Accept json
// @Produce json
// @Param id query string false "ID номера отеля (необязательно)"
// @Param hotel-id query string false "ID отеля (необязательно)"
// @Param date-from query string false "Дата заселения/свободен ли (прим. 2016-10-06, год-месяц-день) (необязательно)"
// @Param date-to query string false "Дата выезда/свободен ли (прим. 2016-10-06, год-месяц-день) (необязательно)"
// @Param days query int false "Количество дней (необязательно)"
// @Param rooms-from query int false "Минимальное количество комнат (необязательно)"
// @Param rooms-to query int false "Максимальное количество комнат (необязательно)"
// @Param type-first query string false "Первый тип (standard, large, premium) (необязательно)"
// @Param type-second query string false "Второй тип (standard, large, premium) (необязательно)"
// @Param type query string false "Единый тип (standard, large, premium) (при использовании перекрывает первый тип) (необязательно)"
// @Param amount-people-from query int false "Минимальное количество человек (необязательно)"
// @Param amount-people-to query int false "Максимальное количество человек (необязательно)"
// @Param value-from query int false "Минимальная цена за номер (необязательно)"
// @Param value-to query int false "Максимальная цена за номер (необязательно)"
// @Param arrange query string false "Упорядочить по цене ('asc' возрастание, 'desc' убывание) (необязательно)"
// @Success 200 {object} httputils.SuccessResponse{data=[]hotelroom.HotelRoom}
// @Failure 400 {object} httputils.ErrorResponse
// @Failure 500 {object} httputils.ErrorResponse
// @Router /api/hotel-room/find [get]
func (v *HotelRoomHandler) Find() fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx := c.Context()

		hotelUuidStr := c.Query("hotel-id")
		hotelRoomUuidStr := c.Query("id")

		dateFromStr := c.Query("date-from")
		dateToStr := c.Query("date-to")

		daysStr := c.Query("days")

		roomsFromStr := c.Query("rooms-from")
		roomsToStr := c.Query("rooms-to")

		typeFirstStr := c.Query("type-first")
		typeSecondStr := c.Query("type-second")
		absoluteType := c.Query("type")

		amountPeopleFromStr := c.Query("amount-people-from")
		amountPeopleToStr := c.Query("amount-people-to")

		valueFromStr := c.Query("value-from")
		valueToStr := c.Query("value-to")

		arrange := c.Query("arrange")

		if absoluteType != "" {
			typeFirstStr = absoluteType
		}

		if dateToStr != "" && daysStr != "" {
			return c.Status(fiber.StatusBadRequest).JSON(httputils.Response{
				Success: false,
				Message: "failed to find hotel room",
				Error:   "you can choose only 'date-from' + 'date-to' or 'date-from' + 'days'",
			})
		}

		if (dateFromStr != "" && (dateToStr == "" && daysStr == "")) || (dateFromStr == "" && (dateToStr != "" || daysStr != "")) {
			return c.Status(fiber.StatusBadRequest).JSON(httputils.Response{
				Success: false,
				Message: "failed to find hotel room",
				Error:   "you must choose 'date-from' + 'date-to' or 'date-from' + 'days'",
			})
		}

		if arrange != "" && (arrange != "asc" && arrange != "desc") {
			return c.Status(fiber.StatusBadRequest).JSON(httputils.Response{
				Success: false,
				Message: "failed to find hotel rooms",
				Error:   fmt.Sprintf("invalid query value 'arrange': %v. supported: 'asc' and 'desc'", arrange),
			})
		}

		var dateFrom *int64
		if dateFromStr != "" {
			dateFr, parseErr := httputils.ParseTimeDate(dateFromStr)
			fmt.Printf("date from: %v\n", dateFr)
			if parseErr != nil {
				return c.Status(fiber.StatusBadRequest).JSON(httputils.Response{
					Success: false,
					Message: "failed to find hotel rooms",
					Error:   fmt.Sprintf("failed to parse query value 'date-from', must match pattern 2016-10-06: %v", parseErr),
				})
			}

			dateFromUnix := dateFr.Unix()
			dateFrom = &dateFromUnix
		}

		var dateTo *int64
		if dateToStr != "" {
			dateToParsed, parseErr := httputils.ParseTimeDate(dateToStr)
			fmt.Printf("date to: %v\n", dateToParsed)
			if parseErr != nil {
				return c.Status(fiber.StatusBadRequest).JSON(httputils.Response{
					Success: false,
					Message: "failed to find hotel rooms",
					Error:   fmt.Sprintf("failed to parse query value 'date-ещ', must match pattern 2016-10-06: %v", parseErr),
				})
			}
			dateToUnix := dateToParsed.Unix()

			if *dateFrom+24*60*60 > dateToUnix {
				return c.Status(fiber.StatusBadRequest).JSON(httputils.Response{
					Success: false,
					Message: "failed to find hotel rooms",
					Error:   fmt.Sprintf("query value 'date-to' must be greater than 'date-from' atleast for 1 day"),
				})
			}
			dateTo = &dateToUnix
		}

		var days *uint64
		if daysStr != "" {
			daysParsed, parseErr := strconv.ParseUint(daysStr, 0, 64)
			fmt.Printf("days: %v\n", daysParsed)
			if parseErr != nil {
				return c.Status(fiber.StatusBadRequest).JSON(httputils.Response{
					Success: false,
					Message: "failed to find hotel rooms",
					Error:   fmt.Sprintf("failed to parse query value 'days': %v", parseErr),
				})
			}
			if daysParsed < 1 {
				return c.Status(fiber.StatusBadRequest).JSON(httputils.Response{
					Success: false,
					Message: "failed to find hotel rooms",
					Error:   fmt.Sprintf("query value 'days' must be greater than zero"),
				})
			}
			days = &daysParsed
		}

		var hotelRoomFirstFilter hotelroom.HotelRoom
		var hotelRoomSecondFilter hotelroom.HotelRoom

		parseFirstFilterErr := ParseHotelRoomParams(hotelUuidStr, roomsFromStr, typeFirstStr, amountPeopleFromStr, valueFromStr, hotelRoomUuidStr, false, &hotelRoomFirstFilter)
		if parseFirstFilterErr != nil {
			return c.Status(fiber.StatusBadRequest).JSON(httputils.Response{
				Success: false,
				Message: "failed to find hotel rooms",
				Error:   fmt.Sprintf("failed to parse first filter: %v", parseFirstFilterErr),
			})
		}

		parseSecondFilterErr := ParseHotelRoomParams(hotelUuidStr, roomsToStr, typeSecondStr, amountPeopleToStr, valueToStr, hotelRoomUuidStr, false, &hotelRoomSecondFilter)
		if parseSecondFilterErr != nil {
			return c.Status(fiber.StatusBadRequest).JSON(httputils.Response{
				Success: false,
				Message: "failed to find hotel rooms",
				Error:   fmt.Sprintf("failed to parse second filter: %v", parseSecondFilterErr),
			})
		}

		if roomsFromStr != "" && roomsToStr != "" && hotelRoomFirstFilter.Rooms > hotelRoomSecondFilter.Rooms ||
			((hotelRoomFirstFilter.Value != nil && hotelRoomSecondFilter.Value != nil) && *hotelRoomFirstFilter.Value > *hotelRoomSecondFilter.Value) ||
			amountPeopleFromStr != "" && amountPeopleToStr != "" && hotelRoomFirstFilter.AmountPeople > hotelRoomSecondFilter.AmountPeople ||
			dateFrom != nil && dateTo != nil && *dateFrom > *dateTo {
			return c.Status(fiber.StatusBadRequest).JSON(httputils.Response{
				Success: false,
				Message: "failed to find hotel rooms",
				Error:   fmt.Sprintf("every value '*-to' must be greater or equal to value '*-from'"),
			})
		}

		foundedHotelRooms, err := v.HotelRoomUsecase.GetWithParams(ctx, &hotelRoomFirstFilter, &hotelRoomSecondFilter, dateFrom, dateTo, days, arrange != "", arrange == "asc")
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(httputils.Response{
				Success: false,
				Message: "failed to find hotel rooms",
				Error:   fmt.Sprintf("%v", err),
			})
		}

		if len(foundedHotelRooms) == 0 {
			return c.Status(fiber.StatusNotFound).JSON(httputils.Response{
				Success: false,
				Message: "hotel rooms not found",
			})
		}

		return c.JSON(httputils.Response{
			Success: true,
			Message: "successfully found hotel rooms",
			Data:    foundedHotelRooms,
		})
	}
}
