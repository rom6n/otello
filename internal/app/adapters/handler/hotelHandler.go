package handler

import (
	"fmt"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/rom6n/otello/internal/app/application/usecases/hotelusecases"
	"github.com/rom6n/otello/internal/app/domain/hotel"
	"github.com/rom6n/otello/internal/utils/httputils"
)

type HotelHandler struct {
	HotelUsecase hotelusecases.HotelUsecases
}

func parseHotelParams(c *fiber.Ctx, allRequired bool, parseTo *hotel.Hotel) (*string, error) {
	city := c.Query("city")
	starsStr := c.Query("stars")
	arrangeStr := c.Query("arrange")
	uuidStr := c.Query("id")
	name := c.Query("name")

	if allRequired && (city == "" || starsStr == "" || name == "") {
		return nil, fmt.Errorf("query values 'city', 'stars', 'name' are required'")
	}

	if allRequired {
		parseTo.Uuid = uuid.New()
	}

	var arrange *string
	if arrangeStr != "" {
		if arrangeStr != "asc" && arrangeStr != "desc" {
			return nil, fmt.Errorf("invalid value of query value 'arrange', must be 'asc' or 'desc'")
		}

		arrange = &arrangeStr
	}

	if uuidStr != "" {
		uuidParsed, parseTicketUuidErr := uuid.Parse(uuidStr)
		if parseTicketUuidErr != nil {
			return nil, fmt.Errorf("failed to parse query value 'id': %v", parseTicketUuidErr)
		}
		parseTo.Uuid = uuidParsed
	}

	if city != "" {
		parseTo.City = city
	}
	if name != "" {
		parseTo.Name = name
	}

	if starsStr != "" {
		starsParsed, parseStarsErr := strconv.ParseInt(starsStr, 0, 32)
		if parseStarsErr != nil {
			return nil, fmt.Errorf("failed to parse query value 'stars': %v", parseStarsErr)
		}
		if starsParsed < 1 || starsParsed > 5 {
			return nil, fmt.Errorf("invalid query value 'stars'. supported value in range 1-5")
		}
		parseTo.Stars = int32(starsParsed)
	}

	return arrange, nil
}

func (v *HotelHandler) Create() fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx := c.Context()
		unsuccessMessage := "failed to create a hotel"

		var parsedHotel hotel.Hotel

		if _, parseErr := parseHotelParams(c, true, &parsedHotel); parseErr != nil {
			return httputils.HandleUnsuccess(c, unsuccessMessage, fmt.Sprintf("%v", parseErr), nil, fiber.StatusBadRequest)
		}

		if err := v.HotelUsecase.Create(ctx, &parsedHotel); err != nil {
			return httputils.HandleUnsuccess(c, unsuccessMessage, fmt.Sprintf("%v", err), nil, fiber.StatusInternalServerError)
		}

		return httputils.HandleSuccess(c, "successfully created the hotel", &parsedHotel)
	}
}

func (v *HotelHandler) Update() fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx := c.Context()
		unsuccessMessage := "failed to update the hotel"

		if c.Query("id") == "" {
			return httputils.HandleUnsuccess(c, unsuccessMessage, "query value 'id' is required", nil, fiber.StatusBadRequest)
		}

		var parsedHotel hotel.Hotel
		if _, parseErr := parseHotelParams(c, false, &parsedHotel); parseErr != nil {
			return httputils.HandleUnsuccess(c, unsuccessMessage, fmt.Sprintf("%v", parseErr), nil, fiber.StatusBadRequest)
		}

		foundHotel, getErr := v.HotelUsecase.Get(ctx, parsedHotel.Uuid)
		if getErr != nil {
			return httputils.HandleUnsuccess(c, unsuccessMessage, "failed to get the hotel", getErr, fiber.StatusInternalServerError)
		}

		if _, parse2Err := parseHotelParams(c, false, foundHotel); parse2Err != nil {
			return httputils.HandleUnsuccess(c, unsuccessMessage, fmt.Sprintf("%v", parse2Err), nil, fiber.StatusBadRequest)
		}

		if err := v.HotelUsecase.Update(ctx, foundHotel); err != nil {
			return httputils.HandleUnsuccess(c, unsuccessMessage, fmt.Sprintf("%v", err), nil, fiber.StatusInternalServerError)
		}

		return httputils.HandleSuccess(c, "successfully updated the hotel", foundHotel)
	}
}

func (v *HotelHandler) Delete() fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx := c.Context()
		unsuccessMessage := "failed to delete the hotel"

		if c.Query("id") == "" {
			return httputils.HandleUnsuccess(c, unsuccessMessage, "query value 'id' is required", nil, fiber.StatusBadRequest)
		}

		hotelUuidStr := c.Query("id")

		if hotelUuidStr == "" {
			return httputils.HandleUnsuccess(c, unsuccessMessage, "query value 'id' is required", nil, fiber.StatusBadRequest)
		}

		hotelUuid, parseErr := uuid.Parse(hotelUuidStr)
		if parseErr != nil {
			return httputils.HandleUnsuccess(c, unsuccessMessage, fmt.Sprintf("failed to parse query value 'id': %v", parseErr), nil, fiber.StatusBadRequest)
		}

		if err := v.HotelUsecase.Delete(ctx, hotelUuid); err != nil {
			return httputils.HandleUnsuccess(c, unsuccessMessage, fmt.Sprintf("%v", err), nil, fiber.StatusInternalServerError)
		}

		return httputils.HandleSuccess(c, "successfully deleted the hotel", nil)
	}
}

func (v *HotelHandler) Find() fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx := c.Context()
		unsuccessMessage := "failed to find the hotel"

		var parsedHotel hotel.Hotel

		arrange, parseErr := parseHotelParams(c, false, &parsedHotel)
		if parseErr != nil {
			return httputils.HandleUnsuccess(c, unsuccessMessage, fmt.Sprintf("%v", parseErr), nil, fiber.StatusBadRequest)
		}

		foundHotels, err := v.HotelUsecase.GetWithParams(ctx, parsedHotel.City, parsedHotel.Stars, parsedHotel.Uuid, arrange != nil, arrange != nil && *arrange == "asc")
		if err != nil {
			return httputils.HandleUnsuccess(c, unsuccessMessage, fmt.Sprintf("%v", parseErr), nil, fiber.StatusInternalServerError)
		}

		if len(foundHotels) == 0 {
			return httputils.HandleUnsuccess(c, unsuccessMessage, "hotels not found", nil, fiber.StatusNotFound)
		}

		return httputils.HandleSuccess(c, "successfully found hotels", foundHotels)
	}
}
