package restgorm

import (
	"errors"
	"fmt"
	"github.com/asaskevich/govalidator"
	"github.com/go-playground/form"
	"github.com/jinzhu/gorm"
	"github.com/leebenson/conform"
	"github.com/saulortega/restgorm/rspndr"
	"log"
	"net/http"
	"reflect"
	"strings"
)

var (
	PK = "id"

	fd *form.Decoder
)

//
//
//

var Llave = struct {
	BD  string
	Obj string
	//¿También desde web? O con json
}{
	"id",
	"ID",
}

//
//
//

func Armar(obj interface{}, r *http.Request) error {
	err := fd.Decode(obj, r.PostForm)
	if err != nil {
		log.Println(err)
		return err
	}

	conform.Strings(obj)

	return nil
}

func Verificar(obj interface{}) error {
	vld, err := govalidator.ValidateStruct(obj)
	if !vld && err == nil {
		err = errors.New("Error. Revise los datos e intente nuevamente.")
	} else if vld && err != nil {
		log.Println("Error inesperado [5036]: ", err) //Esto no debería ocurrir nunca...
	}

	return err
}

//
//
//

func Obtener(BD *gorm.DB, w http.ResponseWriter, r *http.Request, obj interface{}) {
	var Obj = reflect.New(reflect.Indirect(reflect.ValueOf(obj)).Type()).Interface()
	var errEncontrar error

	var llave = llaveDesdeURL(r)
	if llave != "" {
		errEncontrar = BD.Where(fmt.Sprintf(`%s = ?`, PK), llave).First(Obj).Error
	}

	rspndr.Obtención(w, llave, errEncontrar, Obj)
}

func Crear(BD *gorm.DB, w http.ResponseWriter, r *http.Request, obj interface{}) {
	var Obj = reflect.New(reflect.Indirect(reflect.ValueOf(obj)).Type()).Interface()
	var errArmar, errVerificar, errCrear error
	var llave string

	errArmar = Armar(Obj, r)
	if errArmar == nil {
		errVerificar = Verificar(Obj)
		if errVerificar == nil {
			errCrear = BD.Create(Obj).Error
			if errCrear == nil {
				llave = llaveDesdeEstructura(Obj)
			}
		}
	}

	rspndr.Creación(w, errArmar, errVerificar, errCrear, llave)
}

func Editar(BD *gorm.DB, w http.ResponseWriter, r *http.Request, obj interface{}) {
	var Obj = reflect.New(reflect.Indirect(reflect.ValueOf(obj)).Type()).Interface()
	var errEncontrar, errArmar, errVerificar, errEditar error

	var llave = r.FormValue("Llave") // Pendiente -------------------------------------------------

	if len(llave) > 0 {
		errEncontrar = BD.Where(fmt.Sprintf(`%s = ?`, PK), llave).First(Obj).Error //Pendiente cambiar lo del PK --------
		if errEncontrar == nil {
			errArmar = Armar(Obj, r)
			if errArmar == nil {
				errVerificar = Verificar(Obj)
				if errVerificar == nil {
					errEditar = BD.Omit("fecha_creación", "fecha_eliminación").Save(Obj).Error //Pendiente lo de las omisiones -------
				}
			}
		}
	}

	rspndr.Edición(w, errEncontrar, errArmar, errVerificar, errEditar, llave)
}

func Listar(BD *gorm.DB, w http.ResponseWriter, r *http.Request, obj interface{}) {
	var Objs = reflect.New(reflect.SliceOf(reflect.Indirect(reflect.ValueOf(obj)).Type())).Interface()
	var errListar error

	errListar = BD.Find(Objs).Error

	rspndr.Listado(w, Objs, errListar)

	/*var errListar error

	sql, args := común.Búsqueda(BD.NewScope(&Persona{}), r)
	if len(args) > 0 {
		BD = BD.Where(sql, args...)
	}
	errListar = BD.Find(p).Error
	*/
}

//
//
//

func llaveDesdeURL(r *http.Request) string {
	var llave string
	ped := strings.Split(r.URL.Path, "/")
	ll := strings.TrimSpace(ped[len(ped)-1])
	if len(ll) == 6 {
		llave = ll
	}

	return llave
}

func llaveDesdeEstructura(obj interface{}) string {
	var L = reflect.Indirect(reflect.ValueOf(obj)).FieldByName("Llave").String() //Sólo funciona con String, ojo... cambiar a interface y hacer assertion ..
	return L
}
