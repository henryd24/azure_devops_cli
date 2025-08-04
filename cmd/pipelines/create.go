package pipeline

import (
	"log"

	"azuredevops/azdevops"
	"azuredevops/azdevops/pipeline"
	"azuredevops/cmd"

	"github.com/spf13/cobra"
)

var createPipelineCmd = &cobra.Command{
	Use:   "create",
	Short: "Crea un nuevo pipeline a partir de un archivo YAML en un repositorio.",
	Run: func(cmd *cobra.Command, args []string) {
		name, _ := cmd.Flags().GetString("name")
		repoType, _ := cmd.Flags().GetString("repo-type")
		repoName, _ := cmd.Flags().GetString("repo-name")
		branch, _ := cmd.Flags().GetString("branch")
		yamlPath, _ := cmd.Flags().GetString("yaml-path")
		folder, _ := cmd.Flags().GetString("folder")
		serviceConnection, _ := cmd.Flags().GetString("service-connection")

		if name == "" || repoType == "" || repoName == "" || branch == "" || yamlPath == "" {
			log.Fatal("Los parámetros --name, --repo-type, --repo-name, --branch y --yaml-path son obligatorios.")
		}

		client := azdevops.GetClientFromEnv()
		if _, err := pipeline.CreatePipeline(client, name, repoType, repoName, branch, yamlPath, folder, serviceConnection); err != nil {
			log.Fatalf("Error al crear el pipeline: %v", err)
		}
	},
}

func init() {
	createPipelineCmd.Flags().StringP("name", "n", "", "Nombre del pipeline (obligatorio)")
	createPipelineCmd.Flags().StringP("repo-type", "t", "azureReposGit", "Tipo de repositorio ('azureReposGit' o 'gitHub')")
	createPipelineCmd.Flags().StringP("repo-name", "r", "", "Nombre del repositorio. Para GitHub, usar 'organizacion/repositorio' (obligatorio)")
	createPipelineCmd.Flags().StringP("branch", "b", "main", "Rama del repositorio")
	createPipelineCmd.Flags().StringP("yaml-path", "p", "", "Ruta al archivo YAML del pipeline (obligatorio)")
	createPipelineCmd.Flags().String("folder", "\\", "Carpeta para organizar el pipeline (opcional)")
	createPipelineCmd.Flags().StringP("service-connection", "s", "", "ID de la Conexión de Servicio para repositorios privados (opcional)")
	cmd.Pipelines.AddCommand(createPipelineCmd)
}
