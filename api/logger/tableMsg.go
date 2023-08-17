package logger

import (
	"fmt"

	"github.com/fatih/color"
	cloudController "github.com/kubesimplify/ksctl/api/resources/controllers/cloud"
	"github.com/rodaine/table"
)

func (logger *Logger) Table(data []cloudController.AllClusterData) {
	headerFmt := color.New(color.FgGreen, color.Underline).SprintfFunc()
	columnFmt := color.New(color.FgYellow).SprintfFunc()

	tbl := table.New("Name", "Provider", "Nodes", "Type", "K8s")
	tbl.WithHeaderFormatter(headerFmt).WithFirstColumnFormatter(columnFmt)

	for _, row := range data {
		node := ""
		if row.Type == "ha" {
			node = fmt.Sprintf("cp: %d\nwp: %d\nds: %d\nlb: 1", row.NoCP, row.NoWP, row.NoDS)
		} else {
			node = fmt.Sprintf("wp: %d", row.NoMgt)
		}
		tbl.AddRow(row.Name, row.Provider+"("+row.Region+")", node, row.Type, row.K8sDistro)
	}

	tbl.Print()
}
