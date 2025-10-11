package handler

import (
	"fmt"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	hotelroomusecases "github.com/rom6n/otello/internal/app/application/usecases/hotelroomusecases"
	hotelroom "github.com/rom6n/otello/internal/app/domain/hotelroom"
)

type HotelRoomHandler struct {
	HotelRoomUsecase hotelroomusecases.HotelRoomUsecases
}

func ParseHotelRoomParams(hotelUuidStr, roomsStr, typeStr, amountPeopleStr, valueStr string, allRequired bool, oldHotelRoom *hotelroom.HotelRoom) error {
	if allRequired || hotelUuidStr != "" {
		hotelUuid, parseErr := uuid.Parse(hotelUuidStr)
		if parseErr != nil {
			return fmt.Errorf("failed to parse hotel id: %v", parseErr)
		}
		oldHotelRoom.HotelUuid = hotelUuid
	}

	if allRequired || roomsStr != "" {
		rooms, parseErr := strconv.ParseUint(roomsStr, 0, 32)
		if parseErr != nil {
			return fmt.Errorf("failed to parse rooms: %v", parseErr)
		}
		if rooms == 0 {
			return fmt.Errorf("value 'rooms' must be greater than zero")
		}
		oldHotelRoom.Rooms = uint32(rooms)
	}

	if allRequired || amountPeopleStr != "" {
		amountPeople, parseErr := strconv.ParseUint(amountPeopleStr, 0, 32)
		if parseErr != nil {
			return fmt.Errorf("failed to parse amount of people: %v", parseErr)
		}
		if amountPeople == 0 {
			return fmt.Errorf("value 'amount people' must be greater than zero")
		}
		oldHotelRoom.AmountPeople = uint32(amountPeople)
	}

	if allRequired || valueStr != "" {
		value, parseErr := strconv.ParseInt(valueStr, 0, 32)
		if parseErr != nil {
			return fmt.Errorf("failed to parse value: %v", parseErr)
		}
		if value < 0 {
			return fmt.Errorf("value 'value' must be equal or greater than zero")
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
			return fmt.Errorf("this type does not exist. exists: standard, large, premium")
		}
		oldHotelRoom.Type = hotelroom.HotelType(Type)
	}

	return nil
}

func (v *HotelRoomHandler) Create() fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx := c.Context()

		hotelUuidStr := c.Query("hotel-id")
		roomsStr := c.Query("rooms")
		typeStr := c.Query("type")
		amountPeopleStr := c.Query("amount-people")
		valueStr := c.Query("value")

		if hotelUuidStr == "" || roomsStr == "" || typeStr == "" || amountPeopleStr == "" || valueStr == "" {
			return c.Status(fiber.StatusBadRequest).JSON(Response{
				Success: false,
				Message: "failed to create hotel room",
				Error:   "hotel id, rooms, type, amount people and value are required",
			})
		}

		var hotelRoom hotelroom.HotelRoom
		hotelRoom.Uuid = uuid.New()

		err := ParseHotelRoomParams(hotelUuidStr, roomsStr, typeStr, amountPeopleStr, valueStr, true, &hotelRoom)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(Response{
				Success: false,
				Message: "failed to create hotel room",
				Error:   fmt.Sprintf("%v", err),
			})
		}

		if err := v.HotelRoomUsecase.Create(ctx, &hotelRoom); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(Response{
				Success: false,
				Message: "failed to create hotel room",
				Error:   fmt.Sprintf("%v", err),
			})
		}

		return c.JSON(Response{
			Success: true,
			Message: "successfully created hotel room",
			Data:    hotelRoom,
		})
	}
}

func (v *HotelRoomHandler) Update() fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx := c.Context()

		hotelRoomUuidStr := c.Query("id")
		hotelUuidStr := c.Query("new-hotel-id")
		roomsStr := c.Query("new-rooms")
		typeStr := c.Query("new-type")
		amountPeopleStr := c.Query("new-amount-people")
		valueStr := c.Query("new-value")

		if hotelRoomUuidStr == "" {
			return c.Status(fiber.StatusBadRequest).JSON(Response{
				Success: false,
				Message: "failed to update hotel room",
				Error:   "id (hotel room id) is required",
			})
		}

		hotelRoomUuid, parseErr := uuid.Parse(hotelRoomUuidStr)
		if parseErr != nil {
			return fmt.Errorf("failed to parse id: %v", parseErr)
		}

		foundedHotelRoom, getErr := v.HotelRoomUsecase.Get(ctx, hotelRoomUuid)
		if getErr != nil {
			return c.Status(fiber.StatusNotFound).JSON(Response{
				Success: false,
				Message: "failed to update hotel room",
				Error:   fmt.Sprintf("failed to get hotel room: %v", getErr),
			})
		}

		err := ParseHotelRoomParams(hotelUuidStr, roomsStr, typeStr, amountPeopleStr, valueStr, false, foundedHotelRoom)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(Response{
				Success: false,
				Message: "failed to create hotel room",
				Error:   fmt.Sprintf("%v", err),
			})
		}

		if err := v.HotelRoomUsecase.Update(ctx, foundedHotelRoom); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(Response{
				Success: false,
				Message: "failed to update hotel room",
				Error:   fmt.Sprintf("%v", err),
			})
		}

		return c.JSON(Response{
			Success: true,
			Message: "successfully updated hotel room",
			Data:    foundedHotelRoom,
		})
	}
}

func (v *HotelRoomHandler) Delete() fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx := c.Context()

		hotelRoomUuidStr := c.Query("id")

		if hotelRoomUuidStr == "" {
			return c.Status(fiber.StatusBadRequest).JSON(Response{
				Success: false,
				Message: "failed to delete hotel room",
				Error:   "id is required",
			})
		}

		hotelRoomUuid, parseErr := uuid.Parse(hotelRoomUuidStr)
		if parseErr != nil {
			return c.Status(fiber.StatusBadRequest).JSON(Response{
				Success: false,
				Message: "failed to delete hotel room",
				Error:   fmt.Sprintf("failed to parse id: %v", parseErr),
			})
		}

		if err := v.HotelRoomUsecase.Delete(ctx, hotelRoomUuid); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(Response{
				Success: false,
				Message: "failed to delete hotel room",
				Error:   fmt.Sprintf("%v", err),
			})
		}

		return c.JSON(Response{
			Success: true,
			Message: "successfully deleted hotel room",
		})
	}
}

func (v *HotelRoomHandler) Find() fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx := c.Context()

		hotelUuidStr := c.Query("hotel-id")

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
			return c.Status(fiber.StatusBadRequest).JSON(Response{
				Success: false,
				Message: "failed to find hotel room",
				Error:   "you can choose only 'date-from' + 'date-to' or 'date-from' + 'days'",
			})
		}

		if (dateFromStr != "" && (dateToStr == "" && daysStr == "")) || (dateFromStr == "" && (dateToStr != "" || daysStr != "")) {
			return c.Status(fiber.StatusBadRequest).JSON(Response{
				Success: false,
				Message: "failed to find hotel room",
				Error:   "you must choose 'date-from' + 'date-to' or 'date-from' + 'days'",
			})
		}

		if arrange != "" && (arrange != "asc" && arrange != "desc") {
			return c.Status(fiber.StatusBadRequest).JSON(Response{
				Success: false,
				Message: "failed to find hotel rooms",
				Error:   fmt.Sprintf("invalid query value 'arrange': %v. supported: 'asc' and 'desc'", arrange),
			})
		}

		var dateFrom *uint64
		if dateFromStr != "" {
			dateFr, parseErr := strconv.ParseUint(dateFromStr, 0, 64)
			dateFrom = &dateFr
			if parseErr != nil {
				return c.Status(fiber.StatusBadRequest).JSON(Response{
					Success: false,
					Message: "failed to find hotel rooms",
					Error:   fmt.Errorf("failed to parse query value 'date-from': %v", parseErr),
				})
			}
		}

		var dateTo *uint64
		if dateToStr != "" {
			dateT, parseErr := strconv.ParseUint(dateToStr, 0, 64)
			dateTo = &dateT
			if parseErr != nil {
				return c.Status(fiber.StatusBadRequest).JSON(Response{
					Success: false,
					Message: "failed to find hotel rooms",
					Error:   fmt.Errorf("failed to parse query value 'date-to': %v", parseErr),
				})
			}
			if *dateTo < 1 {
				return c.Status(fiber.StatusBadRequest).JSON(Response{
					Success: false,
					Message: "failed to find hotel rooms",
					Error:   fmt.Errorf("query value 'date-to' must be greater than zero"),
				})
			}
		}

		var days *uint64
		if daysStr != "" {
			day, parseErr := strconv.ParseUint(dateFromStr, 0, 64)
			days = &day
			if parseErr != nil {
				return c.Status(fiber.StatusBadRequest).JSON(Response{
					Success: false,
					Message: "failed to find hotel rooms",
					Error:   fmt.Errorf("failed to parse query value 'days': %v", parseErr),
				})
			}
			if *days < 1 {
				return c.Status(fiber.StatusBadRequest).JSON(Response{
					Success: false,
					Message: "failed to find hotel rooms",
					Error:   fmt.Errorf("query value 'days' must be greater than zero"),
				})
			}
		}

		var hotelRoomFirstFilter hotelroom.HotelRoom
		var hotelRoomSecondFilter hotelroom.HotelRoom

		parseFirstFilterErr := ParseHotelRoomParams(hotelUuidStr, roomsFromStr, typeFirstStr, amountPeopleFromStr, valueFromStr, false, &hotelRoomFirstFilter)
		if parseFirstFilterErr != nil {
			return c.Status(fiber.StatusBadRequest).JSON(Response{
				Success: false,
				Message: "failed to find hotel rooms",
				Error:   fmt.Sprintf("failed to parse first filter: %v", parseFirstFilterErr),
			})
		}

		parseSecondFilterErr := ParseHotelRoomParams(hotelUuidStr, roomsToStr, typeSecondStr, amountPeopleToStr, valueToStr, false, &hotelRoomSecondFilter)
		if parseSecondFilterErr != nil {
			return c.Status(fiber.StatusBadRequest).JSON(Response{
				Success: false,
				Message: "failed to find hotel rooms",
				Error:   fmt.Sprintf("failed to parse second filter: %v", parseSecondFilterErr),
			})
		}

		if roomsFromStr != "" && roomsToStr != "" && hotelRoomFirstFilter.Rooms > hotelRoomSecondFilter.Rooms ||
			((hotelRoomFirstFilter.Value != nil && hotelRoomSecondFilter.Value != nil) && *hotelRoomFirstFilter.Value > *hotelRoomSecondFilter.Value) ||
			amountPeopleFromStr != "" && amountPeopleToStr != "" && hotelRoomFirstFilter.AmountPeople > hotelRoomSecondFilter.AmountPeople ||
			dateFrom != nil && dateTo != nil && *dateFrom > *dateTo {
			return c.Status(fiber.StatusBadRequest).JSON(Response{
				Success: false,
				Message: "failed to find hotel rooms",
				Error:   fmt.Sprintf("every value '*-to' must be greater or equal to value '*-from'"),
			})
		}

		foundedHotelRooms, err := v.HotelRoomUsecase.GetWithParams(ctx, &hotelRoomFirstFilter, &hotelRoomSecondFilter, dateFrom, dateTo, days, arrange != "", arrange == "asc")
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(Response{
				Success: false,
				Message: "failed to find hotel rooms",
				Error:   fmt.Sprintf("usecase error: %v", err),
			})
		}

		if len(foundedHotelRooms) == 0 {
			return c.Status(fiber.StatusNotFound).JSON(Response{
				Success: false,
				Message: "hotel rooms not found",
			})
		}

		return c.JSON(Response{
			Success: true,
			Message: "successfully found hotel rooms",
			Data:    foundedHotelRooms,
		})
	}
}
