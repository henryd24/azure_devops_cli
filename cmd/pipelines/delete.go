package pipeline

import (
	"fmt"
	"log"

	"azuredevops/azdevops"
	"azuredevops/azdevops/pipeline"
	"azuredevops/cmd"

	"github.com/spf13/cobra"
)

var deletePipelineCmd = &cobra.Command{
	Use:   "delete",
	Short: "Elimina un pipeline y sus retenciones.",
	Run: func(cmd *cobra.Command, args []string) {
		pipelineID, err := cmd.Flags().GetInt("id")
		if err != nil || pipelineID == 0 {
			log.Fatal("Debes proporcionar un ID de pipeline válido con --id.")
		}

		autoconfirmation, _ := cmd.Flags().GetBool("yes")
		if !autoconfirmation {
			fmt.Printf("Estás a punto de eliminar el pipeline con ID %d y todas sus retenciones.\n", pipelineID)
			fmt.Print("¿Estás seguro de que deseas continuar? (s/n): ")
			var confirmation string
			fmt.Scanln(&confirmation)
			if confirmation != "s" && confirmation != "S" {
				fmt.Println("Operación cancelada.")
				return
			}
		}

		client := azdevops.GetClientFromEnv()
		if err := pipeline.DeletePipeline(client, pipelineID, autoconfirmation); err != nil {
			log.Fatalf("Error al eliminar el pipeline: %v", err)
		}
	},
}

func init() {
	deletePipelineCmd.Flags().IntP("id", "i", 0, "ID del pipeline a eliminar (obligatorio)")
	deletePipelineCmd.Flags().BoolP("yes", "y", false, "Confirmar eliminación sin preguntar")
	cmd.Pipelines.AddCommand(deletePipelineCmd)
}
