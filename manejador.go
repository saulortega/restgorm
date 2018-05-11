package restgorm

import (
	"github.com/jinzhu/gorm"
)

var M *Manejador

type Manejador struct {
	DB       *gorm.DB
	Recursos map[string]*Recurso
}

func manejador() *Manejador {
	var m = new(Manejador)
	m.Recursos = make(map[string]*Recurso)
	return m
}

func RegistrarBD(bd *gorm.DB) {
	M.RegistrarBD(bd)
}
func (m *Manejador) RegistrarBD(bd *gorm.DB) {
	m.DB = bd
}
