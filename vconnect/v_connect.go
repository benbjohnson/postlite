package main

import (
	"database/sql"
	"fmt"

	vertigo "github.com/vertica/vertica-sql-go"
)

var _ = vertigo.VerticaContext.Deadline

func main() {
	conn, err := sql.Open("vertica", "vertica://dbadmin:@localhost:15433/default")
	fmt.Println(err)

	rows, err := conn.Query("SELECT * FROM v_monitor.cpu_usage LIMIT 5")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		var nodeName string
		var startTime string
		var endTime string
		var avgCPU float64

		rows.Scan(&nodeName, &startTime, &endTime, &avgCPU)

		fmt.Println(nodeName)
	}

}
