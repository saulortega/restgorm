package restgorm

import (
	//"errors"
	"github.com/saulortega/restgorm/rspndr"
	"log"
	"net/http"
)

func Receptor(w http.ResponseWriter, r *http.Request) {
	M.Receptor(w, r)
}
func (m *Manejador) Receptor(w http.ResponseWriter, r *http.Request) {
	r.ParseMultipartForm(32 << 20) //32 MB
	//w.Header().Set("Access-Control-Allow-Origin", "http://localhost")                                //Al menos para pruebas. Probablemente se deba quitar...
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Origin, X-Requested-With, Content-Type, Accept") //necesario para dominio cruzado. Quitar después ...
	w.Header().Set("Access-Control-Allow-Methods", "GET,POST,PUT,DELETE")                            //No necesario para dominio cruzado. Plantearse ponerlo siempre...
	w.Header().Set("Access-Control-Expose-Headers", "X-Msj")                                         //Necesario en CORS para poder ver este header desde axios
	w.Header().Add("Access-Control-Expose-Headers", "X-Llave")
	w.Header().Add("Access-Control-Expose-Headers", "X-Notificaciones")

	log.Println("SOLICITUD "+r.Method+"::", r.URL.Path)

	var exte bool
	var R = new(Recurso)
	var rcsSinLlave, snglr = recursoSinLlave(r.URL.Path)
	var rcsConLlave, llave = recursoConLlave(r.URL.Path)

	switch r.Method {
	case "OPTIONS":
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("GET, POST, PUT, DELETE"))
		return
	case "GET":
		//
		R, exte = m.Recursos[rcsSinLlave]
		if exte {
			if snglr { //No debería ser singular. Si lo es, es porque falta la llave
				responderLlaveNoRecibida(w, r.URL.Path)
				return
			}

			Listar(R.BD(), w, r, R.Obj)
			return
		}

		R, exte = m.Recursos[rcsConLlave]
		if !exte { //Sí debería existir.
			responderRecursoDesconocido(w, r.URL.Path)
			return
		}

		if llave == "" {
			responderLlaveNoRecibida(w, r.URL.Path)
			return
		}

		Obtener(R.BD(), w, r, R.Obj)

		return
	case "POST":
		//Puede venir en cualquiera de estos:
		// /algo
		// /algo/
		// /algo/+
		R, exte = m.Recursos[rcsSinLlave]
		if !exte {
			responderRecursoDesconocido(w, r.URL.Path)
			return
		}

		Crear(R.BD(), w, r, R.Obj)

		return
	case "PUT":
		R, exte = m.Recursos[rcsSinLlave]
		if exte { //No debería existir. Si existe es porque no se recibió la llave
			responderLlaveNoRecibida(w, r.URL.Path)
			return
		}

		R, exte = m.Recursos[rcsConLlave]
		if !exte { //Sí debería existir.
			responderRecursoDesconocido(w, r.URL.Path)
			return
		}

		if llave == "" {
			responderLlaveNoRecibida(w, r.URL.Path)
			return
		}

		Editar(R.BD(), w, r, R.Obj)

		return
	//case "DELETE":
	//
	//Pendiente...
	//
	default:
		responderMétodoNoAdmitido(w, r.Method)
	}
}

//
//
//

//Pasar estas funciones al paquete rspndr ?????????????''' --------------------

func responderRecursoDesconocido(w http.ResponseWriter, rcs string) {
	log.Println("Recurso desconocido: " + rcs)
	w.Header().Set("X-Notificaciones", rspndr.NotificaciónError("Recurso desconocido. [6593]").Base64())
	w.WriteHeader(http.StatusBadRequest)
}

func responderLlaveNoRecibida(w http.ResponseWriter, rcs string) {
	log.Println("No se recibió el identificador del recurso: " + rcs)
	w.Header().Set("X-Notificaciones", rspndr.NotificaciónError("No se recibió el identificador. [8251]").Base64())
	w.WriteHeader(http.StatusBadRequest)
}

func responderMétodoNoAdmitido(w http.ResponseWriter, mtd string) {
	log.Println("Método no admitido:", mtd)
	http.Error(w, "Método no admitido", http.StatusMethodNotAllowed)
}
