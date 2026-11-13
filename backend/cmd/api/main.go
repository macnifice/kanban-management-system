// Raíz de composición: aquí se cablean las capas (infraestructura → aplicación → interfaz).
package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"strings"

	appid "kanban/internal/application/identity"
	appkb "kanban/internal/application/kanban"
	"kanban/internal/infrastructure/mysql"
	"kanban/internal/infrastructure/security"
	"kanban/internal/interfaces/httpapi"
)

func main() {
	dsn := os.Getenv("DB_DSN")
	if dsn == "" {
		log.Fatal("DB_DSN no definido")
	}
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		log.Print("AVISO: JWT_SECRET no definido, usando un secreto de desarrollo")
		secret = "dev-secret-change-me"
	}

	db, err := mysql.Connect(dsn)
	if err != nil {
		log.Fatalf("no se pudo conectar a MySQL: %v", err)
	}
	if err := mysql.Migrate(dsn); err != nil {
		log.Fatalf("migraciones: %v", err)
	}

	// infraestructura
	uow := mysql.NewUnitOfWork(db)
	readModel := mysql.NewReadModel(db)
	users := mysql.NewUserRepository(db)
	tokens := security.NewJWTIssuer(secret)

	// aplicación
	srv := &httpapi.Server{
		Identity:       appid.NewService(users, security.BcryptHasher{}, tokens),
		Tokens:         tokens,
		Boards:         appkb.NewBoardService(uow),
		Cards:          appkb.NewCardService(uow),
		Labels:         appkb.NewLabelService(uow),
		Queries:        appkb.NewQueryService(readModel),
		Ping:           func(ctx context.Context) error { return db.PingContext(ctx) },
		AllowedOrigins: strings.Split(os.Getenv("ALLOWED_ORIGINS"), ","),
	}

	log.Printf("API escuchando en :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, srv.Routes()))
}
