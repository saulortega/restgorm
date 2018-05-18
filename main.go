package restgorm

import (
	"github.com/go-playground/form"
)

var (
	PK = "id" // Cambiar esto -------- pendiente ------------

	FD *form.Decoder
)

var Llave = struct {
	BD  string
	Obj string
	//¿También desde web? O con json
}{
	"id",
	"ID",
} // esto cambia ------------------------ debe ir en M ----------- seguir aquí ---------------------

func init() {
	FD = form.NewDecoder()
	M = manejador()
}
