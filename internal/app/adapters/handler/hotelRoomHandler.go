package handler

import (
	"fmt"
	"log"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	hotelroomusecases "github.com/rom6n/otello/internal/app/application/usecases/hotelroomusecases"
	hotelroom "github.com/rom6n/otello/internal/app/domain/hotelroom"
)

type HotelRoomHandler struct {
	HotelRoomUsecase hotelroomusecases.HotelRoomUsecases
}

func ParseHotelRoomParams(hotelUuidStr, dateFromStr, dateToStr, daysStr, roomsStr, typeStr, amountPeopleStr, valueStr string, allRequired bool, oldHotelRoom *hotelroom.HotelRoom) error {
	if allRequired || hotelUuidStr != "" {
		hotelUuid, parseErr := uuid.Parse(hotelUuidStr)
		if parseErr != nil {
			return fmt.Errorf("failed to parse hotel id: %v", parseErr)
		}
		oldHotelRoom.HotelUuid = hotelUuid
	}

	if allRequired || dateFromStr != "" {
		dateFrom, parseErr := strconv.ParseInt(dateFromStr, 0, 64)
		if parseErr != nil {
			return fmt.Errorf("failed to parse date from: %v", parseErr)
		}
		if dateFrom < 0 {
			return fmt.Errorf("value 'date to' must be equal or greater than zero")
		}
		oldHotelRoom.DateFrom = &dateFrom
	}

	if allRequired || dateToStr != "" {
		dateTo, parseErr := strconv.ParseUint(dateToStr, 0, 64)
		if parseErr != nil {
			return fmt.Errorf("failed to parse date to: %v", parseErr)
		}
		if dateTo == 0 {
			return fmt.Errorf("value 'date 'to' must be greater than zero")
		}
		if int64(dateTo) < *oldHotelRoom.DateFrom {
			return fmt.Errorf("value 'date 'to' must be greater than value 'date from'")
		}
		oldHotelRoom.DateTo = dateTo
	}

	if allRequired || daysStr != "" {
		days, parseErr := strconv.ParseUint(daysStr, 0, 32)
		if parseErr != nil {
			return fmt.Errorf("failed to parse days: %v", parseErr)
		}
		if days == 0 {
			return fmt.Errorf("value 'days' must be greater than zero")
		}
		oldHotelRoom.Days = uint32(days)
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
		dateFromStr := c.Query("date-from")
		dateToStr := c.Query("date-to")
		daysStr := c.Query("days")
		roomsStr := c.Query("rooms")
		typeStr := c.Query("type")
		amountPeopleStr := c.Query("amount-people")
		valueStr := c.Query("value")

		if hotelUuidStr == "" || dateFromStr == "" || dateToStr == "" || daysStr == "" || roomsStr == "" || typeStr == "" || amountPeopleStr == "" || valueStr == "" {
			return c.Status(fiber.StatusBadRequest).JSON(Response{
				Success: false,
				Message: "failed to create hotel room",
				Error:   "hotel id, date from, date to, days, rooms, type, amount people and value are required",
			})
		}

		var hotelRoom hotelroom.HotelRoom
		hotelRoom.Uuid = uuid.New()

		err := ParseHotelRoomParams(hotelUuidStr, dateFromStr, dateToStr, daysStr, roomsStr, typeStr, amountPeopleStr, valueStr, true, &hotelRoom)
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
		dateFromStr := c.Query("new-date-from")
		dateToStr := c.Query("new-date-to")
		daysStr := c.Query("new-days")
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

		err := ParseHotelRoomParams(hotelUuidStr, dateFromStr, dateToStr, daysStr, roomsStr, typeStr, amountPeopleStr, valueStr, false, foundedHotelRoom)
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

		amountPeopleFromStr := c.Query("amount-people-from")
		amountPeopleToStr := c.Query("amount-people-to")

		valueFromStr := c.Query("value-from")
		valueToStr := c.Query("value-to")

		arrange := c.Query("arrange")

		if arrange != "" && (arrange != "asc" && arrange != "desc") {
			return c.Status(fiber.StatusBadRequest).JSON(Response{
				Success: false,
				Message: "failed to find hotel rooms",
				Error:   fmt.Sprintf("invalid arrange value: %v. supported: 'asc' and 'desc'", arrange),
			})
		}

		var hotelRoomFirstFilter hotelroom.HotelRoom
		var hotelRoomSecondFilter hotelroom.HotelRoom

		parseFirstFilterErr := ParseHotelRoomParams(hotelUuidStr, dateFromStr, dateToStr, daysStr, roomsFromStr, typeFirstStr, amountPeopleFromStr, valueFromStr, false, &hotelRoomFirstFilter)
		if parseFirstFilterErr != nil {
			return c.Status(fiber.StatusBadRequest).JSON(Response{
				Success: false,
				Message: "failed to find hotel rooms",
				Error:   fmt.Sprintf("failed to parse first filter: %v", parseFirstFilterErr),
			})
		}

		parseSecondFilterErr := ParseHotelRoomParams(hotelUuidStr, dateFromStr, dateToStr, daysStr, roomsToStr, typeSecondStr, amountPeopleToStr, valueToStr, false, &hotelRoomSecondFilter)
		if parseSecondFilterErr != nil {
			return c.Status(fiber.StatusBadRequest).JSON(Response{
				Success: false,
				Message: "failed to find hotel rooms",
				Error:   fmt.Sprintf("failed to parse second filter: %v", parseSecondFilterErr),
			})
		}

		log.Print(hotelRoomFirstFilter.AmountPeople)
		log.Print(hotelRoomSecondFilter.AmountPeople)

		if hotelRoomFirstFilter.Rooms > hotelRoomSecondFilter.Rooms ||
			((hotelRoomFirstFilter.Value != nil && hotelRoomSecondFilter.Value != nil) && *hotelRoomFirstFilter.Value > *hotelRoomSecondFilter.Value) ||
			hotelRoomFirstFilter.AmountPeople > hotelRoomSecondFilter.AmountPeople {
			return c.Status(fiber.StatusBadRequest).JSON(Response{
				Success: false,
				Message: "failed to find hotel rooms",
				Error:   fmt.Sprintf("every value '*-to' must be greater or equal to value '*-from'"),
			})
		}
		foundedHotelRooms, err := v.HotelRoomUsecase.GetWithParams(ctx, &hotelRoomFirstFilter, &hotelRoomSecondFilter, arrange != "", arrange == "asc")
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(Response{
				Success: false,
				Message: "failed to find hotel rooms",
				Error:   fmt.Sprintf("%v", err),
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
