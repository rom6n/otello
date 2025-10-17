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

func parseHotelParams(c *fiber.Ctx, allRequired bool, parseTo *hotel.Hotel) (*string, uint32, uint32, error) {
	city := c.Query("city")
	starsStr := c.Query("stars")
	starsFromStr := c.Query("stars-from")
	starsToStr := c.Query("stars-to")
	arrangeStr := c.Query("arrange")
	uuidStr := c.Query("id")
	name := c.Query("name")

	if allRequired && (city == "" || starsStr == "") {
		return nil, 0, 0, fmt.Errorf("query values 'city', 'stars' are required'")
	}

	if allRequired {
		parseTo.Uuid = uuid.New()
	}

	var arrange *string
	if arrangeStr != "" {
		if arrangeStr != "asc" && arrangeStr != "desc" {
			return nil, 0, 0, fmt.Errorf("invalid value of query value 'arrange', must be 'asc' or 'desc'")
		}

		arrange = &arrangeStr
	}

	if uuidStr != "" {
		uuidParsed, parseTicketUuidErr := uuid.Parse(uuidStr)
		if parseTicketUuidErr != nil {
			return nil, 0, 0, fmt.Errorf("failed to parse query value 'id': %v", parseTicketUuidErr)
		}
		parseTo.Uuid = uuidParsed
	}

	if city != "" {
		parseTo.City = city
	}
	if name != "" {
		parseTo.Name = name
	}

	if starsStr != "" && (starsFromStr != "" || starsToStr != "") {
		return nil, 0, 0, fmt.Errorf("you can choose only one of queries variants: 'stars' or 'stars-from' + 'stars-to'")
	}

	if starsStr != "" {
		starsParsed, parseStarsErr := strconv.ParseInt(starsStr, 0, 32)
		if parseStarsErr != nil {
			return nil, 0, 0, fmt.Errorf("failed to parse query value 'stars': %v", parseStarsErr)
		}
		if starsParsed < 1 || starsParsed > 5 {
			return nil, 0, 0, fmt.Errorf("invalid query value 'stars'. supported value in range 1-5")
		}
		parseTo.Stars = int32(starsParsed)
	}

	var starsFrom uint32
	if starsFromStr != "" {
		starsParsed, parseStarsErr := strconv.ParseInt(starsFromStr, 0, 32)
		if parseStarsErr != nil {
			return nil, 0, 0, fmt.Errorf("failed to parse query value 'stars-from': %v", parseStarsErr)
		}
		if starsParsed < 1 || starsParsed > 5 {
			return nil, 0, 0, fmt.Errorf("invalid query value 'stars-from'. supported value in range 1-5")
		}
		starsFrom = uint32(starsParsed)
	}

	var starsTo uint32
	if starsToStr != "" {
		starsParsed, parseStarsErr := strconv.ParseInt(starsToStr, 0, 32)
		if parseStarsErr != nil {
			return nil, 0, 0, fmt.Errorf("failed to parse query value 'stars-to': %v", parseStarsErr)
		}
		if starsParsed < 1 || starsParsed > 5 {
			return nil, 0, 0, fmt.Errorf("invalid query value 'stars-to'. supported value in range 1-5")
		}
		starsTo = uint32(starsParsed)
	}

	if starsFrom > starsTo {
		return nil, 0, 0, fmt.Errorf("query value 'stars-to' must be greater or equal query value 'stars-from'")
	}

	return arrange, starsFrom, starsTo, nil
}

// @Summary Создать отель (Admin only)
// @Description Создаёт отель с переданными параметрами
// @Tags Отель
// @Accept json
// @Produce json
// @Param city query string true "Город"
// @Param stars query int true "Количество звёзд"
// @Param name query string false "Название отеля (необязательно)"
// @Success 200 {object} httputils.SuccessResponse{data=hotel.Hotel}
// @Failure 400 {object} httputils.ErrorResponse
// @Failure 500 {object} httputils.ErrorResponse
// @Router /api/admin/hotel/create [post]
func (v *HotelHandler) Create() fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx := c.Context()
		unsuccessMessage := "failed to create a hotel"

		var parsedHotel hotel.Hotel

		if _, _, _, parseErr := parseHotelParams(c, true, &parsedHotel); parseErr != nil {
			return httputils.HandleUnsuccess(c, unsuccessMessage, fmt.Sprintf("%v", parseErr), nil, fiber.StatusBadRequest)
		}

		if err := v.HotelUsecase.Create(ctx, &parsedHotel); err != nil {
			return httputils.HandleUnsuccess(c, unsuccessMessage, fmt.Sprintf("%v", err), nil, fiber.StatusInternalServerError)
		}

		return httputils.HandleSuccess(c, "successfully created the hotel", &parsedHotel)
	}
}

// @Summary Изменить отель (Admin only)
// @Description Изменяет отель переданными параметрами
// @Tags Отель
// @Accept json
// @Produce json
// @Param id query string true "ID отеля"
// @Param city query string false "Новый город (необязательно)"
// @Param stars query int false "Новое количество звёзд (необязательно)"
// @Param name query string false "Новое название отеля (необязательно)"
// @Success 200 {object} httputils.SuccessResponse{data=hotel.Hotel}
// @Failure 400 {object} httputils.ErrorResponse
// @Failure 500 {object} httputils.ErrorResponse
// @Router /api/admin/hotel/update [put]
func (v *HotelHandler) Update() fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx := c.Context()
		unsuccessMessage := "failed to update the hotel"

		if c.Query("id") == "" {
			return httputils.HandleUnsuccess(c, unsuccessMessage, "query value 'id' is required", nil, fiber.StatusBadRequest)
		}

		var parsedHotel hotel.Hotel
		if _, _, _, parseErr := parseHotelParams(c, false, &parsedHotel); parseErr != nil {
			return httputils.HandleUnsuccess(c, unsuccessMessage, fmt.Sprintf("%v", parseErr), nil, fiber.StatusBadRequest)
		}

		foundHotel, getErr := v.HotelUsecase.Get(ctx, parsedHotel.Uuid)
		if getErr != nil {
			return httputils.HandleUnsuccess(c, unsuccessMessage, "failed to get the hotel", getErr, fiber.StatusInternalServerError)
		}

		if _, _, _, parse2Err := parseHotelParams(c, false, foundHotel); parse2Err != nil {
			return httputils.HandleUnsuccess(c, unsuccessMessage, fmt.Sprintf("%v", parse2Err), nil, fiber.StatusBadRequest)
		}

		if err := v.HotelUsecase.Update(ctx, foundHotel); err != nil {
			return httputils.HandleUnsuccess(c, unsuccessMessage, fmt.Sprintf("%v", err), nil, fiber.StatusInternalServerError)
		}

		return httputils.HandleSuccess(c, "successfully updated the hotel", foundHotel)
	}
}

// @Summary Удалить отель (Admin only)
// @Description Удаляет отель по ID
// @Tags Отель
// @Accept json
// @Produce json
// @Param id query string true "ID отеля"
// @Success 200 {object} httputils.SuccessResponse
// @Failure 400 {object} httputils.ErrorResponse
// @Failure 500 {object} httputils.ErrorResponse
// @Router /api/admin/hotel/delete [delete]
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

// @Summary Найти отель по параметрам
// @Description Находит отель по фильтрам. Можно искать без фильтров. Нельзя использовать одновременно 'stars' и 'stars-from' + 'stars-to'. Если указать 'arrange', то билеты будут отсортированы по цене в указанном порядке (asc - по возрастанию, desc - по убыванию)
// @Tags Отель
// @Accept json
// @Produce json
// @Param id query string false "ID отеля (необязательно)"
// @Param city query string false "Город (необязательно)"
// @Param stars query int false "Количество звёзд (необязательно)"
// @Param stars-from query int false "Минимальное кол-во звезд (необязательно)"
// @Param stars-to query int false "Максимальное кол-во звезд (необязательно)"
// @Param arrange query string false "Упорядочить по звездам ('asc' возрастание, 'desc' убывание) (необязательно)"
// @Success 200 {object} httputils.SuccessResponse{data=[]hotel.Hotel}
// @Failure 400 {object} httputils.ErrorResponse
// @Failure 500 {object} httputils.ErrorResponse
// @Router /api/hotel/find [get]
func (v *HotelHandler) Find() fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx := c.Context()
		unsuccessMessage := "failed to find the hotel"

		var parsedHotel hotel.Hotel

		arrange, starsFrom, starsTo, parseErr := parseHotelParams(c, false, &parsedHotel)
		if parseErr != nil {
			return httputils.HandleUnsuccess(c, unsuccessMessage, fmt.Sprintf("%v", parseErr), nil, fiber.StatusBadRequest)
		}

		foundHotels, err := v.HotelUsecase.GetWithParams(ctx, parsedHotel.City, parsedHotel.Stars, parsedHotel.Uuid, starsFrom, starsTo, arrange != nil, arrange != nil && *arrange == "asc")
		if err != nil {
			return httputils.HandleUnsuccess(c, unsuccessMessage, fmt.Sprintf("%v", parseErr), nil, fiber.StatusInternalServerError)
		}

		if len(foundHotels) == 0 {
			return httputils.HandleUnsuccess(c, unsuccessMessage, "hotels not found", nil, fiber.StatusNotFound)
		}

		return httputils.HandleSuccess(c, "successfully found hotels", foundHotels)
	}
}
