package restgorm

import (
	"errors"
	"fmt"
	"github.com/asaskevich/govalidator"
	"github.com/jinzhu/gorm"
	"github.com/leebenson/conform"
	"github.com/saulortega/restgorm/plbrs"
	"github.com/saulortega/restgorm/rspndr"
	"log"
	"net/http"
	"reflect"
)

//
//
//

func llamarMétodo(mtdo string, obj interface{}, args []interface{}) []interface{} {
	var reflectValues = []reflect.Value{}
	var valueDirect = reflect.ValueOf(obj)
	var argumentos = []reflect.Value{}
	var valores = []interface{}{}
	//var valueIndirect = reflect.Indirect(valueDirect)

	for _, a := range args {
		argumentos = append(argumentos, reflect.ValueOf(a))
	}

	var valueDir = valueDirect.MethodByName(mtdo)
	var valueInd = reflect.Indirect(valueDirect).MethodByName(mtdo)
	if valueDir.Kind() == reflect.Func {
		reflectValues = valueDir.Call(argumentos)
	} else if valueInd.Kind() == reflect.Func {
		reflectValues = valueInd.Call(argumentos)
	}

	//if len(reflectValues) > 0 {
	for _, v := range reflectValues {
		valores = append(valores, v.Interface())
	}
	//}

	return valores
}

func llamarMétodoRetornandoError(mtdo string, obj interface{}, argumentos []interface{}) error {
	var valores = llamarMétodo(mtdo, obj, argumentos)
	if len(valores) > 0 && valores[0] != nil {
		return valores[0].(error)
	}

	return nil
}

func camposOmitir(obj interface{}) []string {
	var valores = llamarMétodo("OmitirCampos", obj, []interface{}{})
	var omitir = []string{}

	if len(valores) > 0 && valores[0] != nil {
		omitir = valores[0].([]string)
	}

	return omitir
}

//
//
//

func Armar(obj interface{}, BD *gorm.DB, r *http.Request) error {
	err := FD.Decode(obj, r.PostForm)
	if err != nil {
		log.Println(err)
		return err
	}

	conform.Strings(obj)

	/*var argumentos = []interface{}{obj, BD, r}
	var valores = llamarMétodo(obj, "Armar", argumentos)
	if len(valores) > 0 && valores[0] != nil {
		err = valores[0].(error)
		if err != nil {
			return err
		}
	}*/

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

/*func AntesDeEditar(obj interface{}, BD *gorm.DB, r *http.Request) error {
	var argumentos = []interface{}{obj, BD, r}
	var valores = llamarMétodo(obj, "AntesDeEditar", argumentos)
	if len(valores) > 0 && valores[0] != nil {
		err := valores[0].(error)
		if err != nil {
			return err
		}
	}

	return nil
}*/

func AntesDeEditar(obj interface{}, BD *gorm.DB, r *http.Request) error {
	var argumentos = []interface{}{obj, BD, r}
	return llamarMétodoRetornandoError("AntesDeEditar", obj, argumentos)
}

func AntesDeCrear(obj interface{}, BD *gorm.DB, r *http.Request) error {
	var argumentos = []interface{}{obj, BD, r}
	return llamarMétodoRetornandoError("AntesDeCrear", obj, argumentos)
}

//
//
//

func Obtener(BD *gorm.DB, w http.ResponseWriter, r *http.Request, obj interface{}) {
	var Obj = reflect.New(reflect.Indirect(reflect.ValueOf(obj)).Type()).Interface()
	var errEncontrar error

	var _, llave = recursoConLlave(r.URL.Path)
	if llave != "" {
		errEncontrar = BD.Where(fmt.Sprintf(`%s = ?`, PK), llave).First(Obj).Error
	}

	rspndr.Obtención(w, llave, errEncontrar, Obj)
}

func Crear(BD *gorm.DB, w http.ResponseWriter, r *http.Request, obj interface{}) {
	var Obj = reflect.New(reflect.Indirect(reflect.ValueOf(obj)).Type()).Interface()
	var errArmar, errVerificar, errCrear error
	var llave string

	var omitir = camposOmitir(Obj)
	omitir = append(omitir, "fecha_eliminación")

	errArmar = Armar(Obj, BD, r)
	if errArmar == nil {
		errVerificar = Verificar(Obj)
		if errVerificar == nil {
			errVerificar = AntesDeCrear(Obj, BD, r)
			if errVerificar == nil {
				errCrear = BD.Omit(omitir...).Create(Obj).Error
				if errCrear == nil {
					llave = llaveDesdeEstructura(Obj)
				}
			}
		}
	}

	rspndr.Creación(w, errArmar, errVerificar, errCrear, llave)
}

func Editar(BD *gorm.DB, w http.ResponseWriter, r *http.Request, obj interface{}) {
	var Obj = reflect.New(reflect.Indirect(reflect.ValueOf(obj)).Type()).Interface()
	var errEncontrar, errArmar, errVerificar, errEditar error

	var omitir = camposOmitir(Obj)
	omitir = append(omitir, "fecha_creación", "fecha_eliminación")

	var llave = r.FormValue("Llave") // Pendiente -------------------------------------------------
	//Obtener llave desde r también y comparar ------------ pendiente ------------

	if len(llave) > 0 {
		errEncontrar = BD.Where(fmt.Sprintf(`%s = ?`, PK), llave).First(Obj).Error //Pendiente cambiar lo del PK --------
		if errEncontrar == nil {
			errArmar = Armar(Obj, BD, r)
			if errArmar == nil {
				errVerificar = Verificar(Obj)
				if errVerificar == nil {
					errVerificar = AntesDeEditar(Obj, BD, r)
					if errVerificar == nil {
						errEditar = BD.Omit(omitir...).Save(Obj).Error //Pendiente lo de las omisiones -------
					}
				}
			}
		}
	}

	rspndr.Edición(w, errEncontrar, errArmar, errVerificar, errEditar, llave)
}

func Listar(BD *gorm.DB, w http.ResponseWriter, r *http.Request, obj interface{}) {
	var Objs = reflect.New(reflect.SliceOf(reflect.Indirect(reflect.ValueOf(obj)).Type())).Interface()
	var Obj = reflect.New(reflect.Indirect(reflect.ValueOf(obj)).Type()).Interface()
	var errListar error

	sql, args := plbrs.Filtros(BD.NewScope(Obj), r)
	if len(args) > 0 {
		errListar = BD.Where(sql, args...).Order("fecha_creación DESC").Find(Objs).Error
	} else {
		errListar = BD.Order("fecha_creación DESC").Find(Objs).Error
	}

	rspndr.Listado(w, Objs, errListar)
}
