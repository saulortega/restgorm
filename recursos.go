package restgorm

import (
	"fmt"
	"github.com/jinzhu/gorm"
)

type Recurso struct {
	DB  *gorm.DB
	Dir string
	Obj interface{}
}

func (r *Recurso) BD() *gorm.DB {
	if r.DB != nil {
		return r.DB
	}
	return M.DB
}

func recurso(dir string, obj interface{}, otros ...interface{}) *Recurso {
	dir, _ = recursoSinLlave(dir)

	if len(dir) <= 1 {
		panic("No se recibió la dirección del recurso")
	}

	if obj == nil {
		panic("No se recibió el objeto del recurso")
	}

	var R = new(Recurso)
	R.Dir = dir
	R.Obj = obj

	for _, o := range otros {
		switch o.(type) {
		case *gorm.DB:
			R.DB = o.(*gorm.DB)
		default:
			panic(fmt.Sprintf("Parámetro erróneo: %v", o))
		}
	}

	return R
}

func RegistrarRecurso(dir string, obj interface{}, otros ...interface{}) {
	M.RegistrarRecurso(dir, obj, otros...)
}
func (m *Manejador) RegistrarRecurso(dir string, obj interface{}, otros ...interface{}) {
	R := recurso(dir, obj, otros...)
	m.Recursos[R.Dir] = R
}

func RegistrarRecursos(rcss []Recurso) {
	M.RegistrarRecursos(rcss)
}
func (m *Manejador) RegistrarRecursos(rcss []Recurso) {
	for _, r := range rcss {
		if r.DB != nil {
			m.RegistrarRecurso(r.Dir, r.Obj, r.DB)
		} else {
			m.RegistrarRecurso(r.Dir, r.Obj)
		}
	}
}
