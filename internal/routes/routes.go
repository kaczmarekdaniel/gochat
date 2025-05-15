package routes

import (
	"net/http"

	"github.com/kaczmarekdaniel/gochat/internal/app"
	"github.com/kaczmarekdaniel/gochat/internal/middleware"
)

func SetupRoutes(app *app.Application) {

	standardMiddleware := []func(http.HandlerFunc) http.HandlerFunc{
		middleware.WithCORS,
	}

	// authMiddleware := []func(http.HandlerFunc) http.HandlerFunc{
	// 	middleware.WithCORS,
	// }

	http.HandleFunc("/login", app.AuthHandler.HandleLogin)

	http.HandleFunc("/create-user", app.UserHandler.HandleCreateUser)

	// dev endpoints
	http.HandleFunc("/join-room", middleware.Chain(app.RoomHandler.HandleJoinRoom, standardMiddleware...))
	http.HandleFunc("/leave-room", middleware.Chain(app.RoomHandler.HandleLeaveRoom, standardMiddleware...))
	http.HandleFunc("/user-rooms", middleware.Chain(app.RoomHandler.HandleUserRooms, standardMiddleware...))
	// http.HandleFunc("/rooms", withMiddleware(app.RoomHandler.HandleRooms))
	// http.HandleFunc("/init", withMiddleware(app.MessageHandler.HandleGetAllMesssages))

}
