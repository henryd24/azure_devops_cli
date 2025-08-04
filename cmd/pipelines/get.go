package pipeline

import (
	"encoding/json"
	"fmt"
	"log"

	"azuredevops/azdevops"
	"azuredevops/azdevops/pipeline"
	"azuredevops/cmd"

	"github.com/spf13/cobra"
)

var getPipelineCmd = &cobra.Command{
	Use:   "get",
	Short: "Obtiene uno o más pipelines por su nombre.",
	Run: func(cmd *cobra.Command, args []string) {
		name, _ := cmd.Flags().GetString("name")
		id, _ := cmd.Flags().GetInt("id")

		if name == "" && id == 0 {
			log.Fatal("Debes proporcionar el nombre del pipeline con --name o el ID con --id.")
		}

		client := azdevops.GetClientFromEnv()
		var pipelineData interface{}
		var err error
		if id != 0 {
			pipelineData, err = pipeline.GetBuildDefinitionByID(client, id)
		} else {
			pipelineData, err = pipeline.GetBuildDefinitionByName(client, name)
		}
		if err != nil {
			log.Fatalf("Error al obtener el pipeline: %v", err)
		}
		jsonData, err := json.MarshalIndent(pipelineData, "", "  ")
		if err != nil {
			log.Fatalf("Error al convertir a JSON: %v", err)
		}
		fmt.Println(string(jsonData))
	},
}

var getPipelineByIDCmd = &cobra.Command{
	Use:   "get-by-id",
	Short: "Obtiene un pipeline por su ID.",
	Run: func(cmd *cobra.Command, args []string) {
		id, _ := cmd.Flags().GetInt("id")
		if id == 0 {
			log.Fatal("Debes proporcionar el ID del pipeline con --id.")
		}

		client := azdevops.GetClientFromEnv()
		pipeline, err := pipeline.GetBuildDefinitionByID(client, id)
		if err != nil {
			log.Fatalf("Error al obtener el pipeline: %v", err)
		}

		jsonData, err := json.MarshalIndent(pipeline, "", "  ")
		if err != nil {
			log.Fatalf("Error al convertir a JSON: %v", err)
		}
		fmt.Println(string(jsonData))
	},
}

func init() {
	getPipelineCmd.Flags().StringP("name", "n", "", "Nombre del pipeline a buscar (obligatorio)")
	getPipelineCmd.Flags().IntP("id", "i", 0, "ID del pipeline a buscar (opcional)")
	cmd.Pipelines.AddCommand(getPipelineCmd)
	getPipelineByIDCmd.Flags().IntP("id", "i", 0, "ID del pipeline a buscar (obligatorio)")
	cmd.Pipelines.AddCommand(getPipelineByIDCmd)
}
