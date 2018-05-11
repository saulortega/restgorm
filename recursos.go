package restgorm

import (
	"fmt"
	"github.com/jinzhu/gorm"
)

type Recurso struct {
	DB  *gorm.DB //Nulo para usar la general
	Dir string
	Obj interface{} // Necesario ???????????????????????
	//Type reflect.Type
	//Mtds []string //Nada para POST, GET, PUT, DELETE
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
	//R.Type = reflect.Indirect(reflect.ValueOf(obj)).Type()
	//R.Mtds = make([]string, 0)

	for _, o := range otros {
		switch o.(type) {
		case *gorm.DB:
			R.DB = o.(*gorm.DB)
		//case []string:
		//	R.Mtds = o.([]string)
		default:
			panic(fmt.Sprintf("Parámetro erróneo: %v", o))
		}
	}

	return R
}

//
//
//

func (r *Recurso) BD() *gorm.DB {
	if r.DB != nil {
		return r.DB
	}
	return M.DB
}

//
//
//

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
		/*mtds := 0
		if r.Mtds != nil {
			mtds = len(r.Mtds)
		}

		if r.BD != nil && mtds > 0 {
			m.RegistrarRecurso(r.Dir, r.Obj, r.BD, r.Mtds)
		} else if r.BD != nil && mtds == 0 {
			m.RegistrarRecurso(r.Dir, r.Obj, r.BD)
		} else if r.BD == nil && mtds > 0 {
			m.RegistrarRecurso(r.Dir, r.Obj, r.Mtds)
		} else {
			m.RegistrarRecurso(r.Dir, r.Obj)
		}*/

		if r.DB != nil {
			m.RegistrarRecurso(r.Dir, r.Obj, r.DB)
		} else {
			m.RegistrarRecurso(r.Dir, r.Obj)
		}
	}
}
