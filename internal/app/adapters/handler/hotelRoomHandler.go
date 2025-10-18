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

func parseBothFindHotelRoomParams(dateStr, roomsStr, typeStr, amountPeopleStr, valueStr string, parseTo *hotelroom.FindHotelRoomDTO) error {
	if dateStr != "" {
		dateFr, parseDateErr := httputils.ParseTimeDate(dateStr)
		fmt.Printf("date from: %v\n", dateFr)
		if parseDateErr != nil {
			return fmt.Errorf("failed to parse 'date-*' query value, must match pattern 2016-10-06: %v", parseDateErr)
		}
		dateFromUnix := dateFr.Unix()
		parseTo.Date = &dateFromUnix
	}

	if roomsStr != "" {
		rooms, parseErr := strconv.ParseUint(roomsStr, 0, 32)
		if parseErr != nil {
			return fmt.Errorf("failed to parse 'rooms-*' query value: %v", parseErr)
		}
		if rooms == 0 {
			return fmt.Errorf("every query value 'rooms-*' must be greater than zero")
		}
		parseTo.Rooms = uint32(rooms)
	}

	if typeStr != "" {
		parsedType, parseTypeErr := parseHotelRoomType(typeStr)
		if parseTypeErr != nil {
			return fmt.Errorf("failed to parse 'type-*' query value: %v", parseTypeErr)
		}
		parseTo.Type = parsedType
	}

	if amountPeopleStr != "" {
		amountPeople, parseAmountPeopleErr := strconv.ParseUint(amountPeopleStr, 0, 32)
		if parseAmountPeopleErr != nil {
			return fmt.Errorf("failed to parse 'amount-people-*' query value: %v", parseAmountPeopleErr)
		}
		if amountPeople == 0 {
			return fmt.Errorf("every query value 'amount-people-*' must be greater than zero")
		}
		parseTo.AmountPeople = uint32(amountPeople)
	}

	if valueStr != "" {
		value, parseValueErr := strconv.ParseInt(valueStr, 0, 64)
		if parseValueErr != nil {
			return fmt.Errorf("failed to parse 'value-*' query value: %v", parseValueErr)
		}
		if value < 0 {
			return fmt.Errorf("every query value 'value-*' must be greater than zero")
		}
		parseTo.Value = &value
	}

	return nil
}

func isFindDataCorrect(dateFromStr, dateToStr, daysStr, arrange string, firstFilter, secondFilter *hotelroom.FindHotelRoomDTO) error {
	if dateToStr != "" && daysStr != "" {
		return fmt.Errorf("you can choose only 'date-from' + 'date-to' or 'date-from' + 'days'")
	}

	if (dateFromStr != "" && (dateToStr == "" && daysStr == "")) || (dateFromStr == "" && (dateToStr != "" || daysStr != "")) {
		return fmt.Errorf("you must choose 'date-from' + 'date-to' or 'date-from' + 'days'")
	}

	if arrange != "" && (arrange != "asc" && arrange != "desc") {
		return fmt.Errorf("invalid query value 'arrange'. supported: 'asc' and 'desc'")
	}

	if secondFilter.Date != nil && *firstFilter.Date+24*60*60 > *secondFilter.Date {
		return fmt.Errorf("query value 'date-to' must be greater than 'date-from' atleast for 1 day")
	}

	if firstFilter.Rooms != 0 && secondFilter.Rooms != 0 && firstFilter.Rooms > secondFilter.Rooms ||
		((firstFilter.Value != nil && secondFilter.Value != nil) && *firstFilter.Value > *secondFilter.Value) ||
		firstFilter.AmountPeople != 0 && secondFilter.AmountPeople != 0 && firstFilter.AmountPeople > secondFilter.AmountPeople ||
		firstFilter.Date != nil && secondFilter.Date != nil && *firstFilter.Date > *secondFilter.Date {
		return fmt.Errorf("every value '*-to' must be greater or equal to value '*-from'")
	}

	return nil
}

func ParseHotelRoomFindRequest(c *fiber.Ctx, parseTo *hotelroom.FindHotelRoomFilterDTO) error {
	hotelUuidStr := c.Query("hotel-id")
	hotelRoomUuidStr := c.Query("id")

	dateFromStr := c.Query("date-from")
	dateToStr := c.Query("date-to")
	daysStr := c.Query("days")

	roomsFromStr := c.Query("rooms-from")
	roomsToStr := c.Query("rooms-to")

	typeFirstStr := c.Query("type-first")
	typeSecondStr := c.Query("type-second")
	absoluteTypeStr := c.Query("type")

	amountPeopleFromStr := c.Query("amount-people-from")
	amountPeopleToStr := c.Query("amount-people-to")

	valueFromStr := c.Query("value-from")
	valueToStr := c.Query("value-to")

	arrange := c.Query("arrange")

	var firstFilter hotelroom.FindHotelRoomDTO
	var secondFilter hotelroom.FindHotelRoomDTO
	if parseFromErr := parseBothFindHotelRoomParams(dateFromStr, roomsFromStr, typeFirstStr, amountPeopleFromStr, valueFromStr, &firstFilter); parseFromErr != nil {
		return parseFromErr
	}
	if parseToErr := parseBothFindHotelRoomParams(dateToStr, roomsToStr, typeSecondStr, amountPeopleToStr, valueToStr, &secondFilter); parseToErr != nil {
		return parseToErr
	}

	if dataErr := isFindDataCorrect(dateFromStr, dateToStr, daysStr, arrange, &firstFilter, &secondFilter); dataErr != nil {
		return dataErr
	}

	parseTo.Arrange = arrange
	parseTo.RoomsFrom = firstFilter.Rooms
	parseTo.RoomsTo = secondFilter.Rooms
	parseTo.AmountPeopleFrom = firstFilter.AmountPeople
	parseTo.AmountPeopleTo = secondFilter.AmountPeople
	parseTo.ValueFrom = firstFilter.Value
	parseTo.ValueTo = secondFilter.Value
	parseTo.DateFrom = firstFilter.Date
	parseTo.TypeFirst = string(firstFilter.Type)
	parseTo.TypeSecond = string(secondFilter.Type)

	if daysStr != "" {
		daysParsed, parseErr := strconv.ParseUint(daysStr, 0, 64)
		fmt.Printf("days: %v\n", daysParsed)
		if parseErr != nil {
			return fmt.Errorf("failed to parse query value 'days': %v", parseErr)
		}
		if daysParsed < 1 {
			return fmt.Errorf("query value 'days' must be greater than zero")
		}
		newDate := *firstFilter.Date + int64(daysParsed*24*60*60)
		parseTo.DateTo = &newDate
	} else {
		parseTo.DateTo = secondFilter.Date
	}

	if hotelRoomUuidStr != "" {
		hotelRoomUuid, parseErr := uuid.Parse(hotelRoomUuidStr)
		if parseErr != nil {
			return fmt.Errorf("failed to parse query value 'id': %v", parseErr)
		}
		parseTo.Uuid = hotelRoomUuid
	}

	if hotelUuidStr != "" {
		hotelUuid, parseErr := uuid.Parse(hotelUuidStr)
		if parseErr != nil {
			return fmt.Errorf("failed to parse query value 'hotel-id': %v", parseErr)
		}
		parseTo.HotelUuid = hotelUuid
	}

	if absoluteTypeStr != "" {
		absoluteType, parseTypeErr := parseHotelRoomType(absoluteTypeStr)
		if parseTypeErr != nil {
			return fmt.Errorf("failed to parse query value 'type': %v", parseTypeErr)
		}
		parseTo.TypeFirst = string(absoluteType)
	}

	return nil
}

func parseHotelRoomCreateRequest(c *fiber.Ctx, parseTo *hotelroom.HotelRoom) error {
	hotelUuidStr := c.Query("hotel-id")
	roomsStr := c.Query("rooms")
	typeStr := c.Query("type")
	amountPeopleStr := c.Query("amount-people")
	valueStr := c.Query("value")

	if hotelUuidStr == "" || roomsStr == "" || typeStr == "" || amountPeopleStr == "" || valueStr == "" {
		return fmt.Errorf("query values 'hotel-id', 'rooms', 'type', 'amount-people' and 'value' are required")
	}

	hotelUuid, parseUuidErr := uuid.Parse(hotelUuidStr)
	if parseUuidErr != nil {
		return fmt.Errorf("failed to parse query value 'hotel-id': %v", parseUuidErr)
	}

	rooms, parseRoomsErr := strconv.ParseUint(roomsStr, 0, 32)
	if parseRoomsErr != nil {
		return fmt.Errorf("failed to parse query value 'rooms': %v", parseRoomsErr)
	}
	if rooms == 0 {
		return fmt.Errorf("every query value 'rooms' must be greater than zero")
	}

	parsedType, parseTypeErr := parseHotelRoomType(typeStr)
	if parseTypeErr != nil {
		return fmt.Errorf("failed to parse query value 'type': %v", parseTypeErr)
	}

	amountPeople, parseAmountPeopleErr := strconv.ParseUint(amountPeopleStr, 0, 32)
	if parseAmountPeopleErr != nil {
		return fmt.Errorf("failed to parse query value 'value': %v", parseAmountPeopleErr)
	}
	if amountPeople == 0 {
		return fmt.Errorf("every query value 'amount-people' must be greater than zero")
	}

	value, parseValueErr := strconv.ParseInt(valueStr, 0, 64)
	if parseValueErr != nil {
		return fmt.Errorf("failed to parse query value 'value': %v", parseValueErr)
	}
	if value < 0 {
		return fmt.Errorf("every query value 'value' must be greater than zero")
	}

	parseTo.Uuid = uuid.New()
	parseTo.HotelUuid = hotelUuid
	parseTo.Rooms = uint32(rooms)
	parseTo.Type = parsedType
	parseTo.AmountPeople = uint32(amountPeople)
	parseTo.Value = &value

	return nil
}

func parseHotelRoomUpdateRequest(c *fiber.Ctx, parseTo *hotelroom.HotelRoom) error {
	hotelUuidStr := c.Query("hotel-id")
	roomsStr := c.Query("rooms")
	typeStr := c.Query("type")
	amountPeopleStr := c.Query("amount-people")
	valueStr := c.Query("value")

	if hotelUuidStr != "" {
		hotelUuid, parseUuidErr := uuid.Parse(hotelUuidStr)
		if parseUuidErr != nil {
			return fmt.Errorf("failed to parse query value 'hotel-id': %v", parseUuidErr)
		}
		parseTo.HotelUuid = hotelUuid
	}

	if roomsStr != "" {
		rooms, parseRoomsErr := strconv.ParseUint(roomsStr, 0, 32)
		if parseRoomsErr != nil {
			return fmt.Errorf("failed to parse query value 'rooms': %v", parseRoomsErr)
		}
		if rooms == 0 {
			return fmt.Errorf("every query value 'rooms' must be greater than zero")
		}
		parseTo.Rooms = uint32(rooms)
	}

	if typeStr != "" {
		parsedType, parseTypeErr := parseHotelRoomType(typeStr)
		if parseTypeErr != nil {
			return fmt.Errorf("failed to parse query value 'type': %v", parseTypeErr)
		}
		parseTo.Type = parsedType
	}

	if amountPeopleStr != "" {
		amountPeople, parseAmountPeopleErr := strconv.ParseUint(amountPeopleStr, 0, 32)
		if parseAmountPeopleErr != nil {
			return fmt.Errorf("failed to parse query value 'value': %v", parseAmountPeopleErr)
		}
		if amountPeople == 0 {
			return fmt.Errorf("every query value 'amount-people' must be greater than zero")
		}
		parseTo.AmountPeople = uint32(amountPeople)
	}

	if valueStr != "" {
		value, parseValueErr := strconv.ParseInt(valueStr, 0, 64)
		if parseValueErr != nil {
			return fmt.Errorf("failed to parse query value 'value': %v", parseValueErr)
		}
		if value < 0 {
			return fmt.Errorf("every query value 'value' must be greater than zero")
		}
		parseTo.Value = &value
	}

	return nil
}

func parseHotelRoomType(typeStr string) (hotelroom.HotelType, error) {
	var Type hotelroom.HotelType
	switch typeStr {
	case string(hotelroom.Standard):
		Type = hotelroom.Standard
	case string(hotelroom.Lagre):
		Type = hotelroom.Lagre
	case string(hotelroom.Premium):
		Type = hotelroom.Premium
	default:
		return hotelroom.Standard, fmt.Errorf("value of query value 'type' is incorrect. pick any of: 'standard', 'large', 'premium'")
	}

	return Type, nil
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

		unsuccessMessage := "failed to create hotel room"

		var hotelRoom hotelroom.HotelRoom
		if parseErr := parseHotelRoomCreateRequest(c, &hotelRoom); parseErr != nil {
			return httputils.HandleUnsuccess(c, unsuccessMessage, fmt.Sprintf("%v", parseErr), nil, fiber.StatusBadRequest)
		}

		if err := v.HotelRoomUsecase.Create(ctx, &hotelRoom); err != nil {
			return httputils.HandleUnsuccess(c, unsuccessMessage, fmt.Sprintf("%v", err), nil, fiber.StatusInternalServerError)
		}

		return httputils.HandleSuccess(c, "successfully created hotel room", &hotelRoom)
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

		unsuccessMessage := "failed to update the hotel room"

		hotelRoomUuidStr := c.Query("id")
		if hotelRoomUuidStr == "" {
			return httputils.HandleUnsuccess(c, unsuccessMessage, "query value 'id' is required", nil, fiber.StatusBadRequest)
		}

		roomUuid, parseUuidErr := uuid.Parse(hotelRoomUuidStr)
		if parseUuidErr != nil {
			return fmt.Errorf("failed to parse query value 'id': %v", parseUuidErr)
		}

		foundHotelRoom, getErr := v.HotelRoomUsecase.Get(ctx, roomUuid)
		if getErr != nil {
			return httputils.HandleUnsuccess(c, unsuccessMessage, fmt.Sprintf("%v", getErr), nil, fiber.StatusInternalServerError)
		}

		parseErr := parseHotelRoomUpdateRequest(c, foundHotelRoom)
		if parseErr != nil {
			return httputils.HandleUnsuccess(c, unsuccessMessage, fmt.Sprintf("%v", parseErr), nil, fiber.StatusBadRequest)
		}

		if err := v.HotelRoomUsecase.Update(ctx, foundHotelRoom); err != nil {
			return httputils.HandleUnsuccess(c, unsuccessMessage, fmt.Sprintf("%v", parseErr), nil, fiber.StatusInternalServerError)
		}

		return httputils.HandleSuccess(c, "successfully updated the hotel room", &foundHotelRoom)
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

		unsuccessMessage := "failed to delete the hotel room"

		hotelRoomUuidStr := c.Query("id")

		if hotelRoomUuidStr == "" {
			return httputils.HandleUnsuccess(c, unsuccessMessage, "query value 'id' is required", nil, fiber.StatusBadRequest)
		}

		hotelRoomUuid, parseErr := uuid.Parse(hotelRoomUuidStr)
		if parseErr != nil {
			return httputils.HandleUnsuccess(c, unsuccessMessage, fmt.Sprintf("failed to parse query valie 'id': %v", parseErr), nil, fiber.StatusInternalServerError)
		}

		if err := v.HotelRoomUsecase.Delete(ctx, hotelRoomUuid); err != nil {
			return httputils.HandleUnsuccess(c, unsuccessMessage, fmt.Sprintf("%v", parseErr), nil, fiber.StatusInternalServerError)
		}

		return httputils.HandleSuccess(c, "successfully deleted the hotel room", nil)
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

		unsuccessMessage := "failed to find hotel rooms"

		var hotelRoomFilter hotelroom.FindHotelRoomFilterDTO
		if parseErr := ParseHotelRoomFindRequest(c, &hotelRoomFilter); parseErr != nil {
			return httputils.HandleUnsuccess(c, unsuccessMessage, fmt.Sprintf("%v", parseErr), nil, fiber.StatusBadRequest)
		}

		foundHotelRooms, err := v.HotelRoomUsecase.GetWithParams(ctx, &hotelRoomFilter)
		if err != nil {
			return httputils.HandleUnsuccess(c, unsuccessMessage, fmt.Sprintf("%v", err), nil, fiber.StatusInternalServerError)
		}

		if len(foundHotelRooms) == 0 {
			return httputils.HandleUnsuccess(c, unsuccessMessage, "hotel rooms not found", nil, fiber.StatusNotFound)
		}

		return httputils.HandleSuccess(c, "successfully found hotel rooms", foundHotelRooms)
	}
}
