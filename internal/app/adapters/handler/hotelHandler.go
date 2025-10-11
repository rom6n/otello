package handler

import (
	"fmt"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	hotelusecases "github.com/rom6n/otello/internal/app/application/usecases/hotelusecases"
	"github.com/rom6n/otello/internal/app/domain/hotel"
)

type HotelHandler struct {
	HotelUsecase hotelusecases.HotelUsecases
}

func (v *HotelHandler) Create() fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx := c.Context()

		name := c.Query("name")
		city := c.Query("city")
		starsStr := c.Query("stars")

		if name == "" || city == "" || starsStr == "" {
			return c.Status(fiber.StatusBadRequest).JSON(Response{
				Success: false,
				Message: "failed to create hotel",
				Error:   "query values 'name', 'city' and 'stars' are required",
			})
		}

		stars, parseErr := strconv.ParseInt(starsStr, 0, 32)
		if parseErr != nil {
			return c.Status(fiber.StatusBadRequest).JSON(Response{
				Success: false,
				Message: "failed to create hotel",
				Error:   fmt.Sprintf("%v", parseErr),
			})
		}

		if stars < 1 || stars > 5 {
			return c.Status(fiber.StatusBadRequest).JSON(Response{
				Success: false,
				Message: "failed to create hotel",
				Error:   "invalid query value 'stars'. supported value in range 1-5",
			})
		}

		newHotel := hotel.NewHotel(name, city, int32(stars))

		if err := v.HotelUsecase.Create(ctx, newHotel); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(Response{
				Success: false,
				Message: "failed to create hotel",
				Error:   fmt.Sprintf("%v", err),
			})
		}

		return c.JSON(Response{
			Success: true,
			Message: "successfully created hotel",
			Data:    newHotel,
		})
	}
}

func (v *HotelHandler) Update() fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx := c.Context()

		hotelUuidStr := c.Query("id")
		newName := c.Query("new-name")
		newCity := c.Query("new-city")
		newStarsStr := c.Query("new-stars")

		if hotelUuidStr == "" {
			return c.Status(fiber.StatusBadRequest).JSON(Response{
				Success: false,
				Message: "failed to update hotel",
				Error:   "query value 'id' is required",
			})
		}

		hotelUuid, parseErr := uuid.Parse(hotelUuidStr)
		if parseErr != nil {
			return c.Status(fiber.StatusBadRequest).JSON(Response{
				Success: false,
				Message: "failed to update hotel",
				Error:   fmt.Sprintf("failed to parse query value 'id': %v", parseErr),
			})
		}

		var newStars int64
		if newStarsStr != "" {
			var parseIntErr error
			newStars, parseIntErr = strconv.ParseInt(newStarsStr, 0, 32)
			if parseIntErr != nil {
				return c.Status(fiber.StatusBadRequest).JSON(Response{
					Success: false,
					Message: "failed to update hotel",
					Error:   fmt.Sprintf("failed to parse query value 'new-stars': %v", parseIntErr),
				})
			}

			if newStars < 1 || newStars > 5 {
				return c.Status(fiber.StatusBadRequest).JSON(Response{
					Success: false,
					Message: "failed to update hotel",
					Error:   "invalid query value 'stars'. supported value in range 1-5",
				})
			}
		}

		foundedHotel, getErr := v.HotelUsecase.Get(ctx, hotelUuid)
		if getErr != nil {
			return c.Status(fiber.StatusNotFound).JSON(Response{
				Success: false,
				Message: "failed to update hotel",
				Error:   fmt.Sprintf("failed to update hotel: %v", parseErr),
			})
		}

		if newName != "" {
			foundedHotel.Name = newName
		}
		if newCity != "" {
			foundedHotel.City = newCity
		}
		if newStarsStr != "" {
			foundedHotel.Stars = int32(newStars)
		}

		if err := v.HotelUsecase.Update(ctx, foundedHotel); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(Response{
				Success: false,
				Message: "failed to update hotel",
				Error:   fmt.Sprintf("%v", err),
			})
		}

		return c.JSON(Response{
			Success: true,
			Message: "successfully updated hotel",
			Data:    foundedHotel,
		})
	}
}

func (v *HotelHandler) Delete() fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx := c.Context()

		hotelUuidStr := c.Query("id")

		if hotelUuidStr == "" {
			return c.Status(fiber.StatusBadRequest).JSON(Response{
				Success: false,
				Message: "failed to delete hotel",
				Error:   "query value 'id' is required",
			})
		}

		hotelUuid, parseErr := uuid.Parse(hotelUuidStr)
		if parseErr != nil {
			return c.Status(fiber.StatusBadRequest).JSON(Response{
				Success: false,
				Message: "failed to delete hotel",
				Error:   fmt.Sprintf("failed to parse query value 'id': %v", parseErr),
			})
		}

		if err := v.HotelUsecase.Delete(ctx, hotelUuid); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(Response{
				Success: false,
				Message: "failed to delete hotel",
				Error:   fmt.Sprintf("%v", err),
			})
		}

		return c.JSON(Response{
			Success: true,
			Message: "successfully deleted hotel",
		})
	}
}

func (v *HotelHandler) Find() fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx := c.Context()

		city := c.Query("city")
		starsStr := c.Query("stars")
		arrange := c.Query("arrange")

		if arrange != "" && (arrange != "asc" && arrange != "desc") {
			return c.Status(fiber.StatusBadRequest).JSON(Response{
				Success: false,
				Message: "failed to find hotels",
				Error:   fmt.Sprintf("invalid query value 'arrange': %v. supported values: 'asc' and 'desc'", arrange),
			})
		}

		var stars int64
		if starsStr != "" {
			var parseErr error
			stars, parseErr = strconv.ParseInt(starsStr, 0, 32)
			if parseErr != nil {
				return c.Status(fiber.StatusBadRequest).JSON(Response{
					Success: false,
					Message: "failed to find hotels",
					Error:   fmt.Sprintf("failed to parse query value 'stars': %v", parseErr),
				})
			}
			if stars < 1 || stars > 5 {
				return c.Status(fiber.StatusBadRequest).JSON(Response{
					Success: false,
					Message: "failed to find hotels",
					Error:   "invalid query value 'stars'. supported value in range 1-5",
				})
			}
		}

		foundedHotels, err := v.HotelUsecase.GetWithParams(ctx, city, int32(stars), arrange != "", arrange == "asc")
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(Response{
				Success: false,
				Message: "failed to find hotels",
				Error:   fmt.Sprintf("%v", err),
			})
		}

		if len(foundedHotels) == 0 {
			return c.Status(fiber.StatusNotFound).JSON(Response{
				Success: false,
				Message: "hotels not found",
			})
		}

		return c.JSON(Response{
			Success: true,
			Message: "successfully found hotels",
			Data:    foundedHotels,
		})
	}
}
