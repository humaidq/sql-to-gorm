package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/xwb1989/sqlparser"
)

type SQLTable struct {
	Name string
	Cols []SQLColumn
}

type SQLColumn struct {
	Name, Type    string
	IsPrimaryKey  bool
	Length        string
	EnumValues    []string
	AutoIncrement bool
	NotNull       bool
	Default       string
}

func (t *SQLTable) ToGorm() string {
	var str strings.Builder
	str.WriteString(fmt.Sprintf("type %s struct {\n", t.Name))
	for _, col := range t.Cols {
		str.WriteRune('\t')
		// Go variable name and type
		str.WriteString(strings.Title(col.Name))

		var goType string
		switch col.Type {
		case "varchar", "text", "enum":
			goType = "string"
		case "int":
			goType = "int64"
		case "tinyint":
			goType = "int"
		case "double":
			goType = "float64"
		case "date", "datetime":
			goType = "time.Time"
		case "blob":
			goType = "[]byte"
		}
		str.WriteString(" " + goType)
		str.WriteString(" `gorm:\"")

		str.WriteString("type:" + col.Type)

		// Bracketed type metadata
		if len(col.EnumValues) > 0 {
			str.WriteRune('(')
			for i, en := range col.EnumValues {
				str.WriteString(en)
				if i != len(col.EnumValues)-1 {
					str.WriteRune(',')
				}
			}
			str.WriteRune(')')
		} else if len(col.Length) > 0 {
			str.WriteString("(" + col.Length + ")")
		}

		str.WriteString(";column:" + col.Name)
		if col.AutoIncrement {
			str.WriteString(";AUTO_INCREMENT")
		}
		if col.NotNull {
			str.WriteString(";not null")
		}
		if len(col.Default) > 0 {
			str.WriteString(";default:" + col.Default)
		}
		if col.IsPrimaryKey {
			str.WriteString(";PRIMARY_KEY")
		}

		// close variable tag
		str.WriteString("\"`\n")
	}
	str.WriteString("}")
	return str.String()
}

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
	}

	// Otherwise do something with stmt
	switch stmt := stmt.(type) {
	case *sqlparser.DDL:
		var table SQLTable

		var primaryKey string
		// Go through table spec
		for _, ind := range stmt.TableSpec.Indexes {
			switch ind.Info.Type {
			case "primary key":
				primaryKey = ind.Columns[0].Column.String()
			case "key":
			}
		}

		table.Name = stmt.NewName.Name.String()
		for _, col := range stmt.TableSpec.Columns {
			var scol SQLColumn

			scol.Name = col.Name.String()
			scol.Type = col.Type.Type
			scol.EnumValues = col.Type.EnumValues
			if col.Type.Length != nil {
				scol.Length = string(col.Type.Length.Val)
			}
			scol.AutoIncrement = bool(col.Type.Autoincrement)
			scol.NotNull = bool(col.Type.NotNull)
			if col.Type.Default != nil {
				scol.Default = string(col.Type.Default.Val)
			}
			scol.IsPrimaryKey = (col.Name.String() == primaryKey)

			table.Cols = append(table.Cols, scol)
		}
		fmt.Println(table.ToGorm())
	}
}
