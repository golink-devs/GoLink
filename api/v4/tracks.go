package v4

import (
	"github.com/gofiber/fiber/v2"
	"github.com/golink-devs/golink/internal/sources"
)

func RegisterTrackRoutes(router fiber.Router, registry *sources.Registry) {
	router.Get("/loadtracks", LoadTracks(registry))
	router.Get("/decodetrack", DecodeTrack())
	router.Post("/decodetracks", DecodeTracks())
}

func LoadTracks(registry *sources.Registry) fiber.Handler {
	return func(c *fiber.Ctx) error {
		identifier := c.Query("identifier")
		if identifier == "" {
			return c.Status(400).JSON(fiber.Map{
				"loadType": sources.LoadTypeError,
				"data": fiber.Map{
					"message":  "identifier is required",
					"severity": "common",
				},
			})
		}

		result, err := registry.Resolve(c.Context(), identifier)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{
				"loadType": sources.LoadTypeError,
				"data": fiber.Map{
					"message":  err.Error(),
					"severity": "common",
				},
			})
		}

		return c.JSON(result)
	}
}

func DecodeTrack() fiber.Handler {
	return func(c *fiber.Ctx) error {
		encoded := c.Query("encodedTrack")
		if encoded == "" {
			return c.Status(400).SendString("encodedTrack query param required")
		}

		info, err := sources.DecodeTrack(encoded)
		if err != nil {
			return c.Status(400).SendString(err.Error())
		}

		return c.JSON(info)
	}
}

func DecodeTracks() fiber.Handler {
	return func(c *fiber.Ctx) error {
		var encoded []string
		if err := c.BodyParser(&encoded); err != nil {
			return c.Status(400).SendString(err.Error())
		}

		tracks := make([]sources.Track, 0)
		for _, e := range encoded {
			info, err := sources.DecodeTrack(e)
			if err != nil {
				continue
			}
			tracks = append(tracks, sources.Track{
				Encoded: e,
				Info:    *info,
			})
		}

		return c.JSON(tracks)
	}
}
