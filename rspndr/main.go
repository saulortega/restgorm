package rspndr

import (
	"encoding/json"
	"errors"
	"github.com/jinzhu/gorm"
	"log"
	"net/http"
	"reflect"
	"strings"
)

func Obtención(w http.ResponseWriter, llave string, errEncontrar error, obj interface{}) {
	err := responderEdiObt(w, llave, errEncontrar)
	if err != nil {
		return //Respondido en responderEdiObt
	}

	objJSON, err := json.Marshal(obj)
	if err != nil {
		responderErrorInterno(w)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("X-Llave", llave)
	w.WriteHeader(http.StatusOK)
	w.Write(objJSON)
}

func Creación(w http.ResponseWriter, errArmar error, errVerificar error, errCrear error, llave string) {
	err := responderEdiCre(w, errArmar, errVerificar, errCrear)
	if err != nil {
		return //Respondido en responderEdiCre
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("X-Llave", llave)
	w.WriteHeader(http.StatusCreated)
}

func Edición(w http.ResponseWriter, errEncontrar error, errArmar error, errVerificar error, errEditar error, llave string) {
	err := responderEdiObt(w, llave, errEncontrar)
	if err != nil {
		return //Respondido en responderEdiObt
	}

	err = responderEdiCre(w, errArmar, errVerificar, errEditar)
	if err != nil {
		return //Respondido en responderEdiCre
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("X-Llave", llave)
	w.WriteHeader(http.StatusOK)
}

func Listado(w http.ResponseWriter, obj interface{}, errListar error) {
	if errListar != nil {
		if errListar == gorm.ErrRecordNotFound {
			responderCorrectoSinContenido(w)
		} else {
			responderErrorInterno(w)
		}
		log.Println(errListar)
		return
	}

	if reflect.Indirect(reflect.ValueOf(obj)).Len() == 0 {
		responderCorrectoSinContenido(w)
		w.Write([]byte{'[', ']'})
		return
	}

	oJSON, err := json.Marshal(obj)
	if err != nil {
		responderErrorInterno(w)
		log.Println(err)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	// pendiente agregar encabezados de cantidades -------------------------------------
	w.WriteHeader(http.StatusOK)
	w.Write(oJSON)
}

//
//
//

func responderEdiObt(w http.ResponseWriter, llave string, errEncontrar error) error {
	if llave == "" {
		msj := "No se recibió el identificador o es erróneo. [2681]"
		responderErrorSolicitud(w, msj)
		return errors.New(msj)
	}

	if errEncontrar != nil {
		if errEncontrar == gorm.ErrRecordNotFound {
			responderErrorNoEncontrado(w, "Identificador erróneo. Intente nuevamente.")
		} else {
			responderErrorInterno(w)
		}
		log.Println(errEncontrar)
		return errEncontrar
	}

	return nil
}

func responderEdiCre(w http.ResponseWriter, errArmar error, errVerificar error, errEdiCre error) error {
	if errArmar != nil {
		responderErrorSolicitud(w, "los datos recibidos son incorrectos")
		log.Println(errArmar)
		return errArmar
	}

	if errVerificar != nil {
		responderErrorSolicitud(w, strings.Split(errVerificar.Error(), "; ")...)
		log.Println(errVerificar)
		return errVerificar
	}

	if errEdiCre != nil {
		responderErrorInterno(w)
		log.Println(errEdiCre)
		return errEdiCre
	}

	return nil
}

//
//
//

func responderErrorNoEncontrado(w http.ResponseWriter, msj ...string) {
	w.WriteHeader(http.StatusNotFound)
	if len(msj) > 0 {
		w.Header().Set("X-Notificaciones", NotificaciónError(msj...).Base64())
		_, err := w.Write([]byte(strings.Join(msj, "; ")))
		if err != nil {
			log.Println(err)
		}
	}
}

func responderErrorSolicitud(w http.ResponseWriter, msj ...string) {
	if len(msj) == 0 {
		msj = append(msj, "Solicitud errónea. Intente nuevamente. [3294]")
	}
	w.Header().Set("X-Notificaciones", NotificaciónError(msj...).Base64())
	w.WriteHeader(http.StatusBadRequest)
	_, err := w.Write([]byte(strings.Join(msj, "; ")))
	if err != nil {
		log.Println(err)
	}
}

func responderErrorInterno(w http.ResponseWriter) {
	msj := "Ocurrió un error. Intente nuevamente. [4419]"
	w.Header().Set("X-Notificaciones", NotificaciónError(msj).Base64())
	w.WriteHeader(http.StatusInternalServerError)
	_, err := w.Write([]byte(msj))
	if err != nil {
		log.Println(err)
	}
}

func responderCorrectoSinContenido(w http.ResponseWriter, msj ...string) {
	if len(msj) == 0 {
		msj = append(msj, "No se encontraron registros.")
	}
	w.Header().Set("X-Notificaciones", NotificaciónInformación(msj...).Base64())
	w.WriteHeader(http.StatusNoContent) //204, no hay contenido
}
