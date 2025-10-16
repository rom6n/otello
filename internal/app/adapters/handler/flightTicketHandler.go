package handler

import (
	"fmt"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/rom6n/otello/internal/app/application/usecases/flightticketusecases"
	"github.com/rom6n/otello/internal/app/domain/flightticket"
	"github.com/rom6n/otello/internal/utils/httputils"
)

type FlightTicketHandler struct {
	FlightTicketUsecase flightticketusecases.FlightTicketUsecases
}

func parseFlightTicketParams(c *fiber.Ctx, parseTo *flightticket.FlightTicket, allRequired bool, isUpdate bool) (*string, *string, error) {
	uuidStr := c.Query("id")
	cityFrom := c.Query("city-from")
	cityTo := c.Query("city-to")
	cityViaStr := c.Query("city-via")
	quantityStr := httputils.QueryOneOf(c, "quantity", "amount-people")
	valueStr := httputils.QueryOneOf(c, "value", "price")
	takeOffStr := c.Query("take-off")
	arrivalStr := c.Query("arrival")
	arrangeStr := c.Query("arrange")

	if allRequired && (cityFrom == "" || cityTo == "" || quantityStr == "" || valueStr == "" || takeOffStr == "" || arrivalStr == "") {
		return nil, nil, fmt.Errorf("query values 'city-from', 'city-to', 'quantity', 'value', 'take-off', 'arrival' are required'")
	}

	if allRequired {
		parseTo.Uuid = uuid.New()
		parseTo.Category = flightticket.None
	}

	if (arrivalStr != "" && takeOffStr == "") || (arrivalStr == "" && takeOffStr != "") {
		return nil, nil, fmt.Errorf("you must choose query value 'take-off' and 'arrival'")
	}

	if !isUpdate {
		if cityFrom != "" && cityTo == "" {
			return nil, nil, fmt.Errorf("you must also provide query value 'city-to'")
		}

		if cityFrom == "" && cityTo != "" {
			return nil, nil, fmt.Errorf("you must provide query value 'city-from'")
		}

		if cityViaStr != "" && cityFrom == "" {
			return nil, nil, fmt.Errorf("you also must provide query values 'city-from' and 'city-to'")
		}
	}

	var cityVia *string
	if cityViaStr != "" {
		cityVia = &cityViaStr
	}

	var arrange *string
	if arrangeStr != "" {
		if arrangeStr != "asc" && arrangeStr != "desc" {
			return nil, nil, fmt.Errorf("invalid value of query value 'arrange', must be 'asc' or 'desc'")
		}

		arrange = &arrangeStr
	}
	if uuidStr != "" {
		uuidParsed, parseTicketUuidErr := uuid.Parse(uuidStr)
		if parseTicketUuidErr != nil {
			return nil, nil, fmt.Errorf("failed to parse ticket uuid: %v", parseTicketUuidErr)
		}
		parseTo.Uuid = uuidParsed
	}

	if cityFrom != "" {
		parseTo.CityFrom = cityFrom
	}

	if cityTo != "" {
		parseTo.CityTo = cityTo
	}

	if quantityStr != "" {
		quantityParsed, parseQuantityErr := strconv.ParseUint(quantityStr, 0, 32)
		if parseQuantityErr != nil {
			return nil, nil, fmt.Errorf("failed to parse query value 'quantity': %v", parseQuantityErr)
		}
		if quantityParsed == 0 {
			return nil, nil, fmt.Errorf("query value 'quantity' is must be greater than zero")
		}
		parseTo.Quantity = uint32(quantityParsed)
	}

	if valueStr != "" {
		valueParsed, parseValueErr := strconv.ParseUint(valueStr, 0, 64)
		if parseValueErr != nil {
			return nil, nil, fmt.Errorf("failed to parse query value 'value': %v", parseValueErr)
		}
		uint32Value := uint32(valueParsed)
		parseTo.Value = &uint32Value
	}

	var memoryTakeOffUnix int64
	if takeOffStr != "" {
		takeOffParsed, parseTakeOffErr := httputils.ParseTimeZ(takeOffStr)
		if parseTakeOffErr != nil {
			return nil, nil, fmt.Errorf("failed to parse query value 'take-off'. must match pattern '2006-01-02T15:04:05Z': %v", parseTakeOffErr)
		}
		unixTakeOff := takeOffParsed.Unix()
		memoryTakeOffUnix = unixTakeOff
		parseTo.TakeOff = &unixTakeOff
	}

	if arrivalStr != "" {
		arrivalParsed, parseArrivalErr := httputils.ParseTimeZ(arrivalStr)
		if parseArrivalErr != nil {
			return nil, nil, fmt.Errorf("failed to parse query value 'arrival'. must match pattern '2006-01-02T15:04:05Z': %v", parseArrivalErr)
		}
		unixArrival := arrivalParsed.Unix()

		if memoryTakeOffUnix >= unixArrival {
			return nil, nil, fmt.Errorf("query value 'take-off' must be greater than query value 'arrival'")
		}
		parseTo.Arrival = unixArrival
	}

	return arrange, cityVia, nil
}

// @Summary Создать авиабилет (Admin only)
// @Description Создаёт авиабилет с переданными параметрами
// @Tags Авиабилет
// @Accept json
// @Produce json
// @Param city-from query string true "Из какого города"
// @Param city-to query string true "В какой город"
// @Param quantity query int true "Количество билетов/мест"
// @Param value query int true "Цена за билет"
// @Param take-off query string true "Дата и время взлета (формат: 2006-01-02T15:04:05Z)"
// @Param arrival query string true "Дата и время посадки (формат: 2006-01-02T15:04:06Z)"
// @Success 200 {object} httputils.SuccessResponse{data=flightticket.FlightTicket}
// @Failure 400 {object} httputils.ErrorResponse
// @Failure 500 {object} httputils.ErrorResponse
// @Router /api/admin/flight-ticket/create [post]
func (v *FlightTicketHandler) Create() fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx := c.Context()
		unsuccessMessage := "failed to create a flight ticket"

		var flightTicket flightticket.FlightTicket
		_, _, parseErr := parseFlightTicketParams(c, &flightTicket, true, false)
		if parseErr != nil {
			return httputils.HandleUnsuccess(c, unsuccessMessage, fmt.Sprintf("%v", parseErr), nil, fiber.StatusBadRequest)
		}

		err := v.FlightTicketUsecase.Create(ctx, &flightTicket)
		if err != nil {
			return httputils.HandleUnsuccess(c, unsuccessMessage, fmt.Sprintf("%v", err), nil, fiber.StatusInternalServerError)
		}

		return httputils.HandleSuccess(c, "successfully created a flight ticket", flightTicket)
	}
}

// @Summary Изменить авиабилет (Admin only)
// @Description Изменяет авиабилет переданными параметрами
// @Tags Авиабилет
// @Accept json
// @Produce json
// @Param id query string true "ID авиабилета"
// @Param city-from query string false "Новый из какого города (необязатено)"
// @Param city-to query string false "Новый в какой город (необязатено)"
// @Param quantity query int false "Новое количество билетов/мест (необязатено)"
// @Param value query int false "Новая цена за билет (необязатено)"
// @Param take-off query string false "Новые дата и время взлета (формат: 2006-01-02T15:04:05Z) (необязатено)"
// @Param arrival query string false "Новые дата и время посадки (формат: 2006-01-02T15:04:06Z) (необязатено)"
// @Success 200 {object} httputils.SuccessResponse
// @Failure 400 {object} httputils.ErrorResponse
// @Failure 500 {object} httputils.ErrorResponse
// @Router /api/admin/flight-ticket/update [put]
func (v *FlightTicketHandler) Update() fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx := c.Context()
		unsuccessMessage := "failed to update the flight ticket"

		if c.Query("id") == "" {
			return httputils.HandleUnsuccess(c, unsuccessMessage, "query value 'id' is required", nil, fiber.StatusBadRequest)
		}
		var flightTicketFilter flightticket.FlightTicket
		_, _, parseErr := parseFlightTicketParams(c, &flightTicketFilter, false, true)
		if parseErr != nil {
			return httputils.HandleUnsuccess(c, unsuccessMessage, fmt.Sprintf("%v", parseErr), nil, fiber.StatusBadRequest)
		}

		foundFlightTicket, getErr := v.FlightTicketUsecase.Get(ctx, flightTicketFilter.Uuid)
		if getErr != nil {
			return httputils.HandleUnsuccess(c, unsuccessMessage, fmt.Sprintf("%v", getErr), nil, fiber.StatusInternalServerError)
		}

		_, _, parse2Err := parseFlightTicketParams(c, foundFlightTicket, false, true)
		if parse2Err != nil {
			return httputils.HandleUnsuccess(c, unsuccessMessage, fmt.Sprintf("%v", parse2Err), nil, fiber.StatusBadRequest)
		}

		err := v.FlightTicketUsecase.Update(ctx, foundFlightTicket)
		if err != nil {
			return httputils.HandleUnsuccess(c, unsuccessMessage, fmt.Sprintf("%v", err), nil, fiber.StatusInternalServerError)
		}

		return httputils.HandleSuccess(c, "successfully updated the flight ticket", foundFlightTicket)
	}
}

// @Summary Удалить авиабилет (Admin only)
// @Description Удаляет авиабилет по ID
// @Tags Авиабилет
// @Accept json
// @Produce json
// @Param id query string true "ID авиабилета"
// @Success 200 {object} httputils.SuccessResponse
// @Failure 400 {object} httputils.ErrorResponse
// @Failure 500 {object} httputils.ErrorResponse
// @Router /api/admin/flight-ticket/delete [delete]
func (v *FlightTicketHandler) Delete() fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx := c.Context()
		unsuccessMessage := "failed to delete the flight ticket"

		uuidStr := c.Query("id")

		if c.Query("id") == "" {
			return httputils.HandleUnsuccess(c, unsuccessMessage, "query value 'id' is required", nil, fiber.StatusBadRequest)
		}

		uuidParsed, parseTicketUuidErr := uuid.Parse(uuidStr)
		if parseTicketUuidErr != nil {
			return httputils.HandleUnsuccess(c, unsuccessMessage, fmt.Sprintf("failed to parse ticket uuid: %v", parseTicketUuidErr), nil, fiber.StatusBadRequest)
		}

		err := v.FlightTicketUsecase.Delete(ctx, uuidParsed)
		if err != nil {
			return httputils.HandleUnsuccess(c, unsuccessMessage, fmt.Sprintf("%v", err), nil, fiber.StatusInternalServerError)
		}

		return httputils.HandleSuccess(c, "successfully deleted the flight ticket", nil)
	}
}

// @Summary Найти авиабилет
// @Description Находит авиабилет по фильтрам. Можно использовать 'city-from' и 'city-to' вместе с 'city-via' для поиска билетов с пересадкой в этом городе или без него. Если указать 'arrange', то билеты будут отсортированы по цене в указанном порядке (asc - по возрастанию, desc - по убыванию). Самый быстрый и самый дешевый авиабилеты/рейсы отмечаются категориями 'самый быстрый', 'самый дешевый', исключение если нет прямого пути без пересадок - в таком случае выдается самый быстрый путь
// @Tags Авиабилет
// @Accept json
// @Produce json
// @Param city-from query string true "Из какого города"
// @Param city-via query string false "Через какой город (по выбору)"
// @Param city-to query string false "В какой город (по выбору)"
// @Param quantity query int false "Количество билетов/мест (необязатено)"
// @Param take-off query string false "Дата и время взлета (формат: 2006-01-02T15:04:05Z) (необязатено)"
// @Param arrival query string false "Дата и время посадки (формат: 2006-01-02T15:04:06Z) (необязатено)"
// @Param arrange query string false "Упорядочить по цене ('asc' возрастание, 'desc' убывание) (необязатено)"
// @Success 200 {object} httputils.SuccessResponse
// @Failure 400 {object} httputils.ErrorResponse
// @Failure 500 {object} httputils.ErrorResponse
// @Router /api/flight-ticket/find [get]
func (v *FlightTicketHandler) Find() fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx := c.Context()
		unsuccessMessage := "failed to find flight tickets"

		var flightTicketFilter flightticket.FlightTicket
		arrange, cityVia, parseErr := parseFlightTicketParams(c, &flightTicketFilter, false, false)
		if parseErr != nil {
			return httputils.HandleUnsuccess(c, unsuccessMessage, fmt.Sprintf("%v", parseErr), nil, fiber.StatusBadRequest)
		}

		foundFlightTickets, foundTicketsWithStraightWay, err := v.FlightTicketUsecase.GetWithParams(ctx, &flightTicketFilter, cityVia, arrange != nil, arrange != nil && *arrange == "asc")
		if err != nil {
			return httputils.HandleUnsuccess(c, unsuccessMessage, fmt.Sprintf("%v", err), nil, fiber.StatusInternalServerError)
		}

		if len(foundFlightTickets) == 0 && len(foundTicketsWithStraightWay) == 0 {
			return httputils.HandleUnsuccess(c, unsuccessMessage, "flight tickets not found", nil, fiber.StatusNotFound)
		}

		var valueToReturn interface{}
		if len(foundTicketsWithStraightWay) > 0 {
			valueToReturn = foundTicketsWithStraightWay
		} else {
			valueToReturn = foundFlightTickets
		}

		return httputils.HandleSuccess(c, "successfully found flight tickets", valueToReturn)
	}
}

// @Summary Купить авиабилет
// @Description Покупает авиабилеты по ID и количеству пассажиров. Требуется авторизация
// @Tags Авиабилет
// @Accept json
// @Produce json
// @Param id query string true "ID авиабилета"
// @Param quantity query int true "Количество пассажиров"
// @Success 200 {object} httputils.SuccessResponse{data=flightticket.FlightTicket}
// @Failure 400 {object} httputils.ErrorResponse
// @Failure 500 {object} httputils.ErrorResponse
// @Router /api/flight-ticket/buy [post]
func (v *FlightTicketHandler) Buy() fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx := c.Context()
		unsuccessMessage := "failed to buy flight tickets"

		uuidStr := c.Query("id")
		amountPassengersStr := c.Query("quantity")

		if uuidStr == "" || amountPassengersStr == "" {
			return httputils.HandleUnsuccess(c, unsuccessMessage, "query values 'id' and 'quantity' are required", nil, fiber.StatusBadRequest)
		}

		uuidParsed, parseTicketUuidErr := uuid.Parse(uuidStr)
		if parseTicketUuidErr != nil {
			return httputils.HandleUnsuccess(c, unsuccessMessage, fmt.Sprintf("failed to parse ticket uuid: %v", parseTicketUuidErr), nil, fiber.StatusBadRequest)
		}

		amountPassengersParsed, parseAmountPassengersErr := strconv.ParseUint(amountPassengersStr, 0, 32)
		if parseAmountPassengersErr != nil {
			return httputils.HandleUnsuccess(c, unsuccessMessage, fmt.Sprintf("failed to parse quantity: %v", parseAmountPassengersErr), nil, fiber.StatusBadRequest)
		}

		if amountPassengersParsed < 1 {
			return httputils.HandleUnsuccess(c, unsuccessMessage, "query value 'quantity' must be greater than 0", nil, fiber.StatusBadRequest)
		}

		amountPassengers := uint32(amountPassengersParsed)

		boughtFlightTicket, err := v.FlightTicketUsecase.Buy(ctx, uuidParsed, amountPassengers)
		if err != nil {
			return httputils.HandleUnsuccess(c, unsuccessMessage, fmt.Sprintf("%v", err), nil, fiber.StatusInternalServerError)
		}

		return httputils.HandleSuccess(c, "successfully bought flight ticket", boughtFlightTicket)
	}
}
