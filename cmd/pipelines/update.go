package pipeline

import (
	"log"

	"azuredevops/azdevops"
	"azuredevops/azdevops/pipeline"
	"azuredevops/cmd"

	"github.com/spf13/cobra"
)

var updatePipelineCmd = &cobra.Command{
	Use:   "update",
	Short: "Actualiza un pipeline existente.",
	Run: func(cmd *cobra.Command, args []string) {
		pipelineID, _ := cmd.Flags().GetInt("id")
		newName, _ := cmd.Flags().GetString("new-name")
		yamlPath, _ := cmd.Flags().GetString("yaml-path")
		repoName, _ := cmd.Flags().GetString("repo-name")
		serviceConnection, _ := cmd.Flags().GetString("service-connection")

		if pipelineID == 0 {
			log.Fatal("Debes proporcionar el ID del pipeline con --id.")
		}
		if newName == "" && yamlPath == "" && repoName == "" && serviceConnection == "" {
			log.Fatal("Debes proporcionar al menos un valor para actualizar.")
		}

		client := azdevops.GetClientFromEnv()
		if _, err := pipeline.UpdateBuildDefinition(client, pipelineID, newName, yamlPath, repoName, serviceConnection); err != nil {
			log.Fatalf("Error al actualizar el pipeline: %v", err)
		}
	},
}

func init() {
	updatePipelineCmd.Flags().IntP("id", "i", 0, "ID del pipeline a actualizar (obligatorio)")
	updatePipelineCmd.Flags().String("new-name", "", "El nuevo nombre para el pipeline (opcional)")
	updatePipelineCmd.Flags().String("yaml-path", "", "La nueva ruta al archivo YAML (opcional)")
	updatePipelineCmd.Flags().String("repo-name", "", "El nuevo nombre del repositorio en formato 'org/repo' (opcional)")
	updatePipelineCmd.Flags().String("service-connection", "", "El nuevo ID de la conexión de servicio (opcional)")
	cmd.Pipelines.AddCommand(updatePipelineCmd)
}
