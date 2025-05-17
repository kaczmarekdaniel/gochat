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

	http.HandleFunc("/login", middleware.Chain(app.AuthHandler.HandleLogin, standardMiddleware...))

	http.HandleFunc("/create-user", app.UserHandler.HandleCreateUser)

	// dev endpoints
	http.HandleFunc("/join-room", middleware.Chain(app.RoomHandler.HandleJoinRoom, standardMiddleware...))
	http.HandleFunc("/leave-room", middleware.Chain(app.RoomHandler.HandleLeaveRoom, standardMiddleware...))
	http.HandleFunc("/user-rooms", middleware.Chain(app.RoomHandler.HandleUserRooms, standardMiddleware...))
	http.HandleFunc("/rooms", middleware.Chain(app.RoomHandler.HandleRooms, standardMiddleware...))
	// http.HandleFunc("/init", withMiddleware(app.MessageHandler.HandleGetMesssages))

}
