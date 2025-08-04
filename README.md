
# Azure DevOps CLI

Esta es una herramienta de línea de comandos (CLI) no oficial para interactuar con Azure DevOps. Simplifica la gestión de recursos de Azure DevOps, como los *Variable Groups*, directamente desde tu terminal.

## Características

  * **Gestión de Variable Groups**:
      * **Crear**: Crea nuevos *Variable Groups* con variables y secretos.
      * **Obtener**: Recupera *Variable Groups* por su nombre en formato JSON.
      * **Actualizar**: Agrega o modifica variables en uno o más *Variable Groups* existentes.
      * **Eliminar**: Borra *Variable Groups* completos o variables específicas dentro de ellos.
  * **Gestión de Pipelines**: Comandos para trabajar con pipelines de Azure DevOps.

## Prerrequisitos

Antes de empezar, necesitas configurar las siguientes variables de entorno para la autenticación:

  * `AZURE_ORG`: El nombre de tu organización de Azure DevOps.
  * `AZURE_PROJECT`: El nombre de tu proyecto de Azure DevOps.
  * `AZURE_PAT`: Tu Token de Acceso Personal (PAT) de Azure DevOps.

## Instalación

### Desde Binarios

Puedes compilar los binarios para diferentes plataformas utilizando el `Makefile` incluido.

1.  Clona el repositorio:

    ```bash
    git clone https://github.com/henryd24/azure_devops_cli.git
    cd azure_devops_cli
    ```

2.  Ejecuta el comando `make`:

    ```bash
    make build
    ```

    Esto generará los binarios en el directorio `dist/` para Linux, macOS y Windows.

### Con Docker

El proyecto incluye un `Dockerfile` para crear un entorno de desarrollo con Go.

## Uso

El comando raíz es `azuredevops`. A partir de ahí, puedes usar los subcomandos `variables` y `pipelines`.

### Ejemplos con `variables`

  * **Crear un nuevo Variable Group**:

    ```bash
    azuredevops variables create --name "MiGrupoDeVariables" --description "Un grupo de ejemplo" --variables "clave1=valor1" "secret:claveSecreta=valorSecreto"
    ```

      * Usa el prefijo `secret:` para crear variables secretas.

  * **Obtener un Variable Group**:

    ```bash
    azuredevops variables get --name "MiGrupoDeVariables"
    ```
    También puedes jugar con expreciones como `MiGrupoDeVariables*` lo que retornara aquellos que contengan dicha estructura, devolviendo no solo aquel que tiene el nombre exacto
    ```bash
    azuredevops variables get --name "MiGrupoDeVariables*"
    ```

  * **Actualizar un Variable Group agregando nuevas variables**:

    ```bash
    azuredevops variables update --name "MiGrupoDeVariables" --variables "clave2=valorNuevo,secret:otraClave=otroSecreto"
    ```
    También podras actualizar varias al tiempo
    ```bash
    azuredevops variables update --name "MiGrupoDeVariables" --name "MiGrupoDeVariables2" --variables "clave2=valorNuevo,secret:otraClave=otroSecreto"
    ```

  * **Eliminar variables específicas de un grupo**:

    ```bash
    azuredevops variables delete --name "MiGrupoDeVariables" --variables "clave1,otraClave"
    ```

  * **Eliminar un Variable Group completo**:

    ```bash
    azuredevops variables delete --name "MiGrupoDeVariables" --yes
    ```

      * Usa `--yes` o `-y` para confirmar la eliminación sin que se te pregunte.

---

### Ejemplos con `pipelines` (Nuevo)

  * **Crear un nuevo pipeline**:

    ```bash
    azuredevops pipelines create --name "MiPipeline" --repo-type "azureReposGit" --repo-name "mi-repo" --branch "main" --yaml-path ".azure-pipelines.yml" --folder "\\" --service-connection "id-conexion"
    ```
    * Los parámetros `--name`, `--repo-type`, `--repo-name`, `--branch` y `--yaml-path` son obligatorios.
    * Puedes especificar la carpeta y la conexión de servicio si lo necesitas.

  * **Obtener un pipeline por nombre**:

    ```bash
    azuredevops pipelines get --name "MiPipelinePrefix*"
    azuredevops pipelines get --name "MiPipeline"
    ```

  * **Actualizar un pipeline existente**:

    ```bash
    azuredevops pipelines update --id 123 --name "NuevoNombre" --yaml-path ".azure-pipelines.yml" --repo-name "mi-repo-actualizado" --service-connection "nuevo-id-conexion"
    ```
    * Puedes actualizar solo los campos que necesites, los demás se mantienen igual.

  * **Eliminar un pipeline**:

    ```bash
    azuredevops pipelines delete --id 123
    ```

---

## Licencia

Este proyecto está bajo la Licencia Pública General de GNU v3.0. Consulta el archivo `LICENSE` para más detalles.

## Contribuciones

Las contribuciones son bienvenidas. Si deseas colaborar, por favor abre un *issue* para discutir tus ideas o envía un *pull request* con tus cambios.
