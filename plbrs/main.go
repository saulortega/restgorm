package plbrs

import (
	"fmt"
	"github.com/jinzhu/gorm"
	"net/http"
	"reflect"
	"regexp"
	"strings"
)

func CamposBúsquedaTexto(scope *gorm.Scope) []string {
	var columnas = []string{}
	for _, sf := range scope.GetStructFields() {
		if sf.IsIgnored || sf.DBName == "llave" || sf.DBName == "fecha_creación" || sf.DBName == "fecha_modificación" || sf.DBName == "fecha_eliminación" {
			continue
		}
		if !(sf.Struct.Type.Kind() == reflect.String || (sf.Struct.Type.Kind() == reflect.Ptr && sf.Struct.Type.String() == "*string")) {
			continue
		}
		if sf.Tag.Get("buscar") == "-" {
			continue
		}
		if regexp.MustCompile(`^llave_`).MatchString(sf.DBName) {
			continue
		}
		columnas = append(columnas, sf.DBName)
	}

	return columnas
}

//
//
//

func PalabrasExactasBuscar(t string) []string {
	var texto = NormalizarTextoBúsqueda(t)
	if texto == "" {
		return []string{}
	}

	return strings.Split(texto, " ")
}

func PalabrasFiltradasBuscar(t string) []string {
	var palabras = PalabrasExactasBuscar(t)

	var palabrasFiltradas = []string{}
	for _, p := range palabras {
		if len(p) <= 2 || p == "las" || p == "los" || p == "les" || p == "una" || p == "por" {
			continue
		}
		palabrasFiltradas = append(palabrasFiltradas, p)
	}

	return palabrasFiltradas
}

func SQLBuscarTextoExacto(scope *gorm.Scope, r *http.Request) (string, []interface{}) {
	var texto = NormalizarTextoBúsqueda(r.FormValue("buscar"))
	var palabras = []string{}
	if texto != "" {
		palabras = append(palabras, texto)
	}

	return SQLBuscar(CamposBúsquedaTexto(scope), palabras)
}

func SQLBuscarPalabrasExactas(scope *gorm.Scope, r *http.Request) (string, []interface{}) {
	return SQLBuscar(CamposBúsquedaTexto(scope), PalabrasExactasBuscar(r.FormValue("buscar")))
}

func SQLBuscarPalabrasFiltradas(scope *gorm.Scope, r *http.Request) (string, []interface{}) {
	return SQLBuscar(CamposBúsquedaTexto(scope), PalabrasFiltradasBuscar(r.FormValue("buscar")))
}

func SQLBuscar(campos, palabras []string) (string, []interface{}) {
	if len(campos) == 0 || len(palabras) == 0 {
		return "", []interface{}{}
	}

	var ors = []string{}
	var parmtrs = []interface{}{}
	for _, c := range campos {
		for _, p := range palabras {
			ors = append(ors, fmt.Sprintf(`%s ILIKE ?`, NormalizarColumnaSQL(c)))
			parmtrs = append(parmtrs, interface{}(fmt.Sprintf(`%%%s%%`, p)))
		}
	}

	var condición = fmt.Sprintf(`(%s)`, strings.Join(ors, " OR "))

	return condición, parmtrs
}

func Filtros(scope *gorm.Scope, r *http.Request) (string, []interface{}) {
	//Pendiente crear condiciones para ver si la búsqueda es de texto exacto, palabras filtradas, etc...
	//Por ahora, como debe ser por defecto:
	return SQLBuscarPalabrasFiltradas(scope, r)
}

//
//
//

func NormalizarColumnaSQL(col string) string {
	var sql string
	sql = fmt.Sprintf("REPLACE(%s, 'Á', 'A')", col)
	sql = fmt.Sprintf("REPLACE(%s, 'É', 'E')", sql)
	sql = fmt.Sprintf("REPLACE(%s, 'Í', 'I')", sql)
	sql = fmt.Sprintf("REPLACE(%s, 'Ó', 'O')", sql)
	sql = fmt.Sprintf("REPLACE(%s, 'Ú', 'U')", sql)
	sql = fmt.Sprintf("REPLACE(%s, 'Ü', 'U')", sql)
	sql = fmt.Sprintf("REPLACE(%s, 'á', 'a')", sql)
	sql = fmt.Sprintf("REPLACE(%s, 'é', 'e')", sql)
	sql = fmt.Sprintf("REPLACE(%s, 'í', 'i')", sql)
	sql = fmt.Sprintf("REPLACE(%s, 'ó', 'o')", sql)
	sql = fmt.Sprintf("REPLACE(%s, 'ú', 'u')", sql)
	sql = fmt.Sprintf("REPLACE(%s, 'ü', 'u')", sql)
	sql = fmt.Sprintf("REPLACE(%s, 'Ñ', 'ñ')", sql)
	sql = fmt.Sprintf("REGEXP_REPLACE(%s, '\\s+', ' ')", sql)
	sql = fmt.Sprintf("LOWER(%s)", sql)
	return sql
}

func NormalizarTextoBúsqueda(t string) string {
	t = strings.ToLower(t)
	t = strings.TrimSpace(t)
	t = regexp.MustCompile(`\s+`).ReplaceAllString(t, " ")
	t = regexp.MustCompile("á").ReplaceAllString(t, "a")
	t = regexp.MustCompile("é").ReplaceAllString(t, "e")
	t = regexp.MustCompile("í").ReplaceAllString(t, "i")
	t = regexp.MustCompile("ó").ReplaceAllString(t, "o")
	t = regexp.MustCompile("ú").ReplaceAllString(t, "u")
	t = regexp.MustCompile("ü").ReplaceAllString(t, "u")
	t = regexp.MustCompile("à").ReplaceAllString(t, "a")
	t = regexp.MustCompile("è").ReplaceAllString(t, "e")
	t = regexp.MustCompile("ì").ReplaceAllString(t, "i")
	t = regexp.MustCompile("ò").ReplaceAllString(t, "o")
	t = regexp.MustCompile("ù").ReplaceAllString(t, "u")
	return t
}
