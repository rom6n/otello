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

func parseFlightTicketParams(c *fiber.Ctx, allRequired bool, parseTo *flightticket.FlightTicket) (*string, *string, error) {
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
	}

	if cityFrom != "" && cityTo == "" {
		return nil, nil, fmt.Errorf("you must also provide query value 'city-to'")
	}

	if cityFrom == "" && cityTo != "" {
		return nil, nil, fmt.Errorf("you must provide query value 'city-from'")
	}

	if cityViaStr != "" && cityFrom == "" {
		return nil, nil, fmt.Errorf("you also must provide query values 'city-from' and 'city-to'")
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

		if memoryTakeOffUnix > unixArrival {
			return nil, nil, fmt.Errorf("query value 'take-off' must be greater or equal query value 'arrival'")
		}
		parseTo.Arrival = unixArrival
	}

	return arrange, cityVia, nil
}

func (v *FlightTicketHandler) Create() fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx := c.Context()
		unsuccessMessage := "failed to create a flight ticket"

		var flightTicket flightticket.FlightTicket
		_, _, parseErr := parseFlightTicketParams(c, true, &flightTicket)
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

func (v *FlightTicketHandler) Update() fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx := c.Context()
		unsuccessMessage := "failed to update the flight ticket"

		if c.Query("id") == "" {
			return httputils.HandleUnsuccess(c, unsuccessMessage, "query value 'id' is required", nil, fiber.StatusBadRequest)
		}
		var flightTicketFilter flightticket.FlightTicket
		_, _, parseErr := parseFlightTicketParams(c, false, &flightTicketFilter)
		if parseErr != nil {
			return httputils.HandleUnsuccess(c, unsuccessMessage, fmt.Sprintf("%v", parseErr), nil, fiber.StatusBadRequest)
		}

		foundFlightTicket, getErr := v.FlightTicketUsecase.Get(ctx, flightTicketFilter.Uuid)
		if getErr != nil {
			return httputils.HandleUnsuccess(c, unsuccessMessage, fmt.Sprintf("%v", getErr), nil, fiber.StatusInternalServerError)
		}

		_, _, parse2Err := parseFlightTicketParams(c, false, foundFlightTicket)
		if parse2Err != nil {
			return httputils.HandleUnsuccess(c, unsuccessMessage, fmt.Sprintf("%v", parse2Err), nil, fiber.StatusBadRequest)
		}

		err := v.FlightTicketUsecase.Update(ctx, foundFlightTicket)
		if err != nil {
			return httputils.HandleUnsuccess(c, unsuccessMessage, fmt.Sprintf("%v", err), nil, fiber.StatusInternalServerError)
		}

		return httputils.HandleSuccess(c, "successfully updated the flight ticket", nil)
	}
}

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

func (v *FlightTicketHandler) Find() fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx := c.Context()
		unsuccessMessage := "failed to find flight tickets"

		var flightTicketFilter flightticket.FlightTicket
		arrange, cityVia, parseErr := parseFlightTicketParams(c, false, &flightTicketFilter)
		if parseErr != nil {
			return httputils.HandleUnsuccess(c, unsuccessMessage, fmt.Sprintf("%v", parseErr), nil, fiber.StatusBadRequest)
		}

		foundFlightTickets, err := v.FlightTicketUsecase.GetWithParams(ctx, &flightTicketFilter, cityVia, arrange != nil, arrange != nil && *arrange == "asc")
		if err != nil {
			return httputils.HandleUnsuccess(c, unsuccessMessage, fmt.Sprintf("%v", err), nil, fiber.StatusInternalServerError)
		}

		if len(foundFlightTickets) == 0 {
			return httputils.HandleUnsuccess(c, unsuccessMessage, "flight tickets not found", nil, fiber.StatusNotFound)
		}

		return httputils.HandleSuccess(c, "successfully found flight tickets", foundFlightTickets)
	}
}

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
