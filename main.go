package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"reflect"
	"strings"

	"github.com/xwb1989/sqlparser"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Println("<program> <sql file>")
		return
	}
	cont, err := ioutil.ReadFile(os.Args[1])
	if err != nil {
		panic(err)
	}
	stmt, err := sqlparser.Parse(string(cont))
	if err != nil {
		panic(err)
		// Do something with the err
	}
	fmt.Println(stmt)

	fmt.Println(reflect.TypeOf(stmt))

	// Otherwise do something with stmt
	switch stmt := stmt.(type) {
	case *sqlparser.DDL:
		var primaryKey string
		// Go through table spec
		for _, ind := range stmt.TableSpec.Indexes {
			switch ind.Info.Type {
			case "primary key":
				primaryKey = ind.Columns[0].Column.String()
			case "key":
			}
		}

		fmt.Println(stmt.TableSpec.Indexes[0].Columns[0])
		fmt.Printf("type %s struct {\n", stmt.NewName.Name.String())
		for _, col := range stmt.TableSpec.Columns {
			// gorm tags
			var gormTags strings.Builder
			gormTags.WriteString("type:" + col.Type.Type)
			if col.Type.Length != nil {
				gormTags.WriteString("(" + string(col.Type.Length.Val) + ")")
			}
			gormTags.WriteRune(';')
			gormTags.WriteString("column:" + col.Name.String())
			if col.Type.Autoincrement {
				gormTags.WriteString(";AUTO_INCREMENT")
			}
			if col.Type.NotNull {
				gormTags.WriteString(";not null")
			}
			if col.Type.Default != nil {
				gormTags.WriteString(";default:" + string(col.Type.Default.Val))
			}
			if col.Name.String() == primaryKey {
				gormTags.WriteString(";PRIMARY_KEY")
			}

			// Go values
			goName := strings.Title(col.Name.String())
			var goType string
			switch col.Type.Type {
			case "varchar":
				goType = "string"
			case "text":
				goType = "string"
			case "int":
				goType = "int64"
			case "tinyint":
				goType = "int"
			case "double":
				goType = "float64"
			case "date":
				goType = "time.Time"
			case "blob":
				goType = "[]byte"
			}
			fmt.Printf("\t%s %s `gorm:\"%s\"`\n", goName, goType,
				gormTags.String())
		}
		fmt.Println("}")
	}
}
