package parser

import (
	"encoding/json"
	"fmt"
	"log"
)

func PrintAST(ss []Stmt) {
	for _, s := range ss {
		switch st := s.(type) {
		case *Print, *Expression, *Var:
			if expr, err := json.MarshalIndent(st, "", "  "); err != nil {
				log.Fatalf("JSON Marshaling Failed: %s", err)
			} else {
				fmt.Printf("%s\n", expr)
			}
		}
	}
}