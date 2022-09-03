package main

import (
	"github.com/adamluzsi/frameless/adapters/memory"
	app "github.com/adamluzsi/poc-app-with-cascade-transaction"
	"github.com/adamluzsi/poc-app-with-cascade-transaction/ext/int/httpapi"
	"log"
	"net/http"
)

func main() {
	mem := memory.NewMemory()
	storage := memory.NewStorage[app.Entity, string](mem)

	uc := app.UseCase{
		Service1: app.SomeService{
			EntityRepository: storage,
		},
		Service2: app.SomeService{
			EntityRepository: storage,
		},
		Service3: app.FlakyService{},
	}

	handler := httpapi.NewHandler(uc)

	if err := http.ListenAndServe(":8080", handler); err != nil {
		log.Fatal("FATAL", err.Error())
	}
}
