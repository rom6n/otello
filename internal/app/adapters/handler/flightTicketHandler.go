package handler

import (
	"fmt"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	flightticketusecases "github.com/rom6n/otello/internal/app/application/usecases/flightticketusecases"
	flightticket "github.com/rom6n/otello/internal/app/domain/flightticket"
)

type FlightTicketHandler struct {
	FlightTicketUsecase flightticketusecases.FlightTicketUsecases
}

func parseFlightTicketParams(c *fiber.Ctx, allRequired bool, parseTo *flightticket.FlightTicket) (*string, *string, error) {
	uuidStr := c.Params("id")
	cityFrom := c.Query("city-from")
	cityTo := c.Query("city-to")
	cityViaStr := c.Query("city-via")
	quantityStr := queryOneOf(c, "quantity", "amount-people")
	valueStr := queryOneOf(c, "value", "price")
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

	if takeOffStr != "" {
		takeOffParsed, parseTakeOffErr := parseTimeZ(takeOffStr)
		if parseTakeOffErr != nil {
			return nil, nil, fmt.Errorf("failed to parse query value 'take-off'. must match pattern '2006-01-02T15:04:05Z': %v", parseTakeOffErr)
		}
		unixTakeOff := takeOffParsed.Unix()
		parseTo.TakeOff = &unixTakeOff
	}

	if arrivalStr != "" {
		arrivalParsed, parseArrivalErr := parseTimeZ(arrivalStr)
		if parseArrivalErr != nil {
			return nil, nil, fmt.Errorf("failed to parse query value 'arrival'. must match pattern '2006-01-02T15:04:05Z': %v", parseArrivalErr)
		}
		parseTo.Arrival = arrivalParsed.Unix()
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
			return handleUnsuccess(c, unsuccessMessage, fmt.Sprintf("%v", parseErr), nil, fiber.StatusBadRequest)
		}

		createdFlightTicket, err := v.FlightTicketUsecase.Create(ctx, &flightTicket)
		if err != nil {
			return handleUnsuccess(c, unsuccessMessage, fmt.Sprintf("%v", err), nil, fiber.StatusInternalServerError)
		}

		return handleSuccess(c, "successfully created a flight ticket", createdFlightTicket)
	}
}

func (v *FlightTicketHandler) Update() fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx := c.Context()
		unsuccessMessage := "failed to update the flight ticket"

		if c.Query("id") == "" {
			return handleUnsuccess(c, unsuccessMessage, "query value 'id' is required", nil, fiber.StatusBadRequest)
		}

		var flightTicketFilter flightticket.FlightTicket
		_, _, parseErr := parseFlightTicketParams(c, false, &flightTicketFilter)
		if parseErr != nil {
			return handleUnsuccess(c, unsuccessMessage, fmt.Sprintf("%v", parseErr), nil, fiber.StatusBadRequest)
		}

		foundFlightTicket, getErr := v.FlightTicketUsecase.Get(ctx, flightTicketFilter.Uuid)
		if getErr != nil {
			return handleUnsuccess(c, unsuccessMessage, fmt.Sprintf("%v", getErr), nil, fiber.StatusInternalServerError)
		}

		_, _, parse2Err := parseFlightTicketParams(c, false, &foundFlightTicket)
		if parse2Err != nil {
			return handleUnsuccess(c, unsuccessMessage, fmt.Sprintf("%v", parse2Err), nil, fiber.StatusBadRequest)
		}

		err := v.FlightTicketUsecase.Update(ctx, &foundFlightTicket)
		if err != nil {
			return handleUnsuccess(c, unsuccessMessage, fmt.Sprintf("%v", err), nil, fiber.StatusInternalServerError)
		}

		return handleSuccess(c, "successfully updated the flight ticket", nil)
	}
}

func (v *FlightTicketHandler) Delete() fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx := c.Context()
		unsuccessMessage := "failed to delete the flight ticket"

		if c.Query("id") == "" {
			return handleUnsuccess(c, unsuccessMessage, "query value 'id' is required", nil, fiber.StatusBadRequest)
		}

		var flightTicketForDeletion flightticket.FlightTicket
		_, _, parseErr := parseFlightTicketParams(c, false, &flightTicketForDeletion)
		if parseErr != nil {
			return handleUnsuccess(c, unsuccessMessage, fmt.Sprintf("%v", parseErr), nil, fiber.StatusBadRequest)
		}

		err := v.FlightTicketUsecase.Delete(ctx, flightTicketForDeletion.Uuid)
		if err != nil {
			return handleUnsuccess(c, unsuccessMessage, fmt.Sprintf("%v", err), nil, fiber.StatusInternalServerError)
		}

		return handleSuccess(c, "successfully deleted the flight ticket", nil)
	}
}

func (v *FlightTicketHandler) Find() fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx := c.Context()
		unsuccessMessage := "failed to find flight tickets"

		var flightTicketFilter flightticket.FlightTicket
		arrange, cityVia, parseErr := parseFlightTicketParams(c, false, &flightTicketFilter)
		if parseErr != nil {
			return handleUnsuccess(c, unsuccessMessage, fmt.Sprintf("%v", parseErr), nil, fiber.StatusBadRequest)
		}

		foundFlightTickets, err := v.FlightTicketUsecase.Find(ctx, flightTicketFilter, arrange, cityVia)
		if err != nil {
			return handleUnsuccess(c, unsuccessMessage, fmt.Sprintf("%v", err), nil, fiber.StatusInternalServerError)
		}

		return handleSuccess(c, "successfully found flight tickets", foundFlightTickets)
	}
}
