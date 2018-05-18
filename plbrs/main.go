package plbrs

import (
	"errors"
	"fmt"
	"github.com/jinzhu/gorm"
	"log"
	"net/http"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

//Pendiente búsqueda exacta por campos exactos, como llave, llave_xxxx, numéricos, etc...

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
		if regexp.MustCompile(`^llave_`).MatchString(sf.DBName) { //Esto debe cambiar...
			continue
		}
		columnas = append(columnas, sf.DBName)
	}

	return columnas
}

func CamposBúsquedaExacta(scope *gorm.Scope) []*gorm.StructField {
	var campos = []*gorm.StructField{}
	for _, sf := range scope.GetStructFields() {
		if sf.IsIgnored || sf.Tag.Get("buscar") == "-" || sf.DBName == "fecha_creación" || sf.DBName == "fecha_modificación" || sf.DBName == "fecha_eliminación" {
			continue
		}

		kind := sf.Struct.Type.Kind()
		if kind == reflect.Int || kind == reflect.Int8 || kind == reflect.Int16 || kind == reflect.Int32 || kind == reflect.Int64 || kind == reflect.Uint || kind == reflect.Uint8 || kind == reflect.Uint16 || kind == reflect.Uint32 || kind == reflect.Uint64 || kind == reflect.Bool || sf.IsPrimaryKey || sf.IsForeignKey {
			campos = append(campos, sf)
		} else if sf.DBName == "llave" || regexp.MustCompile(`^llave_`).MatchString(sf.DBName) { //Esto debe cambiar, debe ser dinámico..
			campos = append(campos, sf)
		}
	}

	return campos
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

func SQLBuscarCamposExactos(scope *gorm.Scope, r *http.Request) (string, []interface{}) {
	return SQLFiltrar(CamposBúsquedaExacta(scope), r)
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

func SQLFiltrar(campos []*gorm.StructField, r *http.Request) (string, []interface{}) {
	var errores = []error{}
	var ands = []string{}
	var parmtrs = []interface{}{}

	for _, sf := range campos {
		if len(r.Form[sf.DBName]) == 0 {
			continue
		}

		ors := []string{}
		kind := sf.Struct.Type.Kind()

		for _, v := range r.Form[sf.DBName] {
			val := strings.TrimSpace(v)
			if len(val) == 0 {
				continue
			}

			switch kind {
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				i, er := strconv.ParseInt(val, 10, bitSizeByKind(kind))
				if er != nil {
					errores = append(errores, er)
					continue
				}

				ors = append(ors, fmt.Sprintf(`%s = ?`, sf.DBName))
				parmtrs = append(parmtrs, interface{}(i))
			case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
				ui, er := strconv.ParseUint(val, 10, bitSizeByKind(kind))
				if er != nil {
					errores = append(errores, er)
					continue
				}

				ors = append(ors, fmt.Sprintf(`%s = ?`, sf.DBName))
				parmtrs = append(parmtrs, interface{}(ui))
			case reflect.Bool: //No debería ser necesario evaluarlo en for. Debe venir un único valor, lo contrario no tiene mucho sentido
				b, er := strconv.ParseBool(val)
				if er != nil {
					errores = append(errores, er)
					continue
				}

				not := ""
				if !b {
					not = "NOT "
				}
				ors = append(ors, fmt.Sprintf(`%s%s`, not, sf.DBName))
			case reflect.String: //Sin normalizar porque se asume que deben ser valores exactos, como las llaves
				ors = append(ors, fmt.Sprintf(`%s = ?`, sf.DBName))
				parmtrs = append(parmtrs, interface{}(val))
			default:
				errores = append(errores, errors.New(fmt.Sprintf(`"tipo «%v» desconocido"`, kind)))
				continue
			}
		}
		if len(ors) > 0 {
			ands = append(ands, fmt.Sprintf(`(%s)`, strings.Join(ors, " OR ")))
		}
	}

	if len(errores) > 0 {
		log.Println("Errores al procesar campos de búsqueda:")
		for _, e := range errores {
			log.Println("\t", e)
		}
	}

	var condición string
	if len(ands) > 0 {
		condición = fmt.Sprintf(`(%s)`, strings.Join(ands, " AND "))
	}

	return condición, parmtrs
}

func Filtros(scope *gorm.Scope, r *http.Request) (string, []interface{}) {
	var prmtrs = []interface{}{}
	var sql string

	sql1, prmtrs1 := SQLBuscarPalabrasFiltradas(scope, r) //Por ahora, como debe ser por defecto
	sql2, prmtrs2 := SQLBuscarCamposExactos(scope, r)
	if len(sql1) > 0 && len(sql2) > 0 {
		sql = fmt.Sprintf(`(%s OR %s)`, sql1, sql2)
		prmtrs = append(prmtrs, prmtrs1...)
		prmtrs = append(prmtrs, prmtrs2...)
	} else if len(sql1) > 0 && len(sql2) == 0 {
		prmtrs = prmtrs1
		sql = sql1
	} else if len(sql1) == 0 && len(sql2) > 0 {
		prmtrs = prmtrs2
		sql = sql2
	}

	return sql, prmtrs
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

func bitSizeByKind(kind reflect.Kind) int {
	var bitSize = 64
	if kind == reflect.Int8 || kind == reflect.Uint8 {
		bitSize = 8
	} else if kind == reflect.Int16 || kind == reflect.Uint16 {
		bitSize = 16
	} else if kind == reflect.Int32 || kind == reflect.Uint32 {
		bitSize = 32
	}
	return bitSize
}
