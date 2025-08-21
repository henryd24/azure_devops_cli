package pipeline

import (
	"log"
	"strings"

	"azuredevops/azdevops"
	"azuredevops/azdevops/pipeline"
	"azuredevops/cmd"
	"azuredevops/models"
	"github.com/spf13/cobra"
)

var runPipelineCmd = &cobra.Command{
	Use:   "run",
	Short: "Ejecuta un pipeline y opcionalmente espera a que finalice.",
	Run: func(cmd *cobra.Command, args []string) {
		pipelineID, _ := cmd.Flags().GetInt("id")
		wait, _ := cmd.Flags().GetBool("wait")
		paramsSlice, _ := cmd.Flags().GetStringSlice("param")
		varsSlice, _ := cmd.Flags().GetStringSlice("var")

		if pipelineID == 0 {
			log.Fatal("Debes proporcionar el ID del pipeline con --id.")
		}

		paramsMap := make(map[string]string)
		for _, p := range paramsSlice {
			parts := strings.SplitN(p, "=", 2)
			if len(parts) == 2 {
				paramsMap[parts[0]] = parts[1]
			}
		}

		varsMap := make(map[string]models.BuildVariable)
		for _, v := range varsSlice {
			parts := strings.SplitN(v, "=", 2)
			if len(parts) == 2 {
				key := parts[0]
				value := parts[1]
				isSecret := false

				if strings.HasPrefix(key, "secret:") {
					isSecret = true
					key = strings.TrimPrefix(key, "secret:")
				}

				varsMap[key] = models.BuildVariable{
					Value:    value,
					IsSecret: isSecret,
				}
			}
		}
		client := azdevops.GetClientFromEnv()
		if _, err := pipeline.RunPipeline(client, pipelineID, wait, paramsMap, varsMap); err != nil {
			log.Fatalf("Error al ejecutar el pipeline: %v", err)
		}
	},
}

func init() {
	runPipelineCmd.Flags().IntP("id", "i", 0, "ID del pipeline a ejecutar (obligatorio)")
	runPipelineCmd.Flags().BoolP("wait", "w", false, "Esperar a que el pipeline finalice (opcional)")
	runPipelineCmd.Flags().StringSliceP("param", "p", nil, "Parámetros del pipeline en formato: clave=valor (se puede repetir)")
	runPipelineCmd.Flags().StringSliceP("var", "v", nil, "Variables de ejecución en formato: clave=valor o secret:clave=valor")
	cmd.Pipelines.AddCommand(runPipelineCmd)
}
