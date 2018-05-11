package restgorm

import (
	"fmt"
	//"log"
	"reflect"
	"regexp"
	"strings"
)

func llaveDesdeEstructura(obj interface{}) string {
	var L = reflect.Indirect(reflect.ValueOf(obj)).FieldByName("Llave").Interface()
	fmt.Println(" --------------- llaveDesdeEstructura --------------------------------------- L ::::::::::::::::::: ")
	fmt.Println(L)
	return fmt.Sprintf(`%v`, L)
}

//Asumiendo el último parámetro después del último slash como la llave:
func recursoConLlave(dir string) (string, string) {
	var rcs, llave string
	rcs = strings.TrimSpace(dir)

	if regexp.MustCompile(`/(\+)?$`).MatchString(rcs) {
		rcs = strings.TrimRight(rcs, "/+")
	} else {
		var pdzs = strings.Split(rcs, "/")
		rcs = strings.Join(pdzs[:len(pdzs)-1], "/")
		llave = pdzs[len(pdzs)-1]
		if llave == "+" {
			llave = ""
		}
	}

	return rcs, llave
}

//Asumiendo que no hay llave.
//Si termina en «/» o «/+», es singular (no se debe buscar con llave).
func recursoSinLlave(dir string) (string, bool) {
	var rcs = strings.TrimSpace(dir)
	var snglr = regexp.MustCompile(`/(\+)?$`).MatchString(rcs)
	if snglr {
		rcs = strings.TrimRight(rcs, "/+")
	}

	return rcs, snglr
}

/*

/recursos/ciencias 			plural GET  // a futuro también DELETE ????
/recursos/ciencias/ 		singular POST  --------------------------------- en este caso no hay duda ---------
/recursos/ciencias/sociales singular GET PUT DELETE  // OJO: "sociales" es una llave, no un subrrecurso...

*/
