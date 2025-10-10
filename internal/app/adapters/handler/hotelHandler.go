package handler

import (
	"fmt"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	hotelusecases "github.com/rom6n/otello/internal/app/application/usecases/hotelUsecases"
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
				Error:   "name, city and stars are required",
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

		hotel := hotel.NewHotel(name, city, int32(stars))

		if err := v.HotelUsecase.Create(ctx, hotel); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(Response{
				Success: false,
				Message: "failed to create hotel",
				Error:   fmt.Sprintf("%v", err),
			})
		}

		return c.JSON(Response{
			Success: true,
			Message: "successfully created hotel",
			Data:    hotel,
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
				Error:   "id is required",
			})
		}

		hotelUuid, parseErr := uuid.Parse(hotelUuidStr)
		if parseErr != nil {
			return c.Status(fiber.StatusBadRequest).JSON(Response{
				Success: false,
				Message: "failed to update hotel",
				Error:   fmt.Sprintf("failed to parse id: %v", parseErr),
			})
		}

		var newStars int64
		var parseIntErr error
		if newStarsStr != "" {
			newStars, parseIntErr = strconv.ParseInt(newStarsStr, 0, 32)
			if parseIntErr != nil {
				return c.Status(fiber.StatusBadRequest).JSON(Response{
					Success: false,
					Message: "failed to update hotel",
					Error:   fmt.Sprintf("failed to parse new amount of stars: %v", parseIntErr),
				})
			}
		}

		foundedHotel, getErr := v.HotelUsecase.Get(ctx, hotelUuid)
		if getErr != nil {
			return c.Status(fiber.StatusNotFound).JSON(Response{
				Success: false,
				Message: "failed to update hotel",
				Error:   fmt.Sprintf("failed to get hotel: %v", parseErr),
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
				Error:   "id is required",
			})
		}

		hotelUuid, parseErr := uuid.Parse(hotelUuidStr)
		if parseErr != nil {
			return c.Status(fiber.StatusBadRequest).JSON(Response{
				Success: false,
				Message: "failed to delete hotel",
				Error:   fmt.Sprintf("failed to parse id: %v", parseErr),
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
