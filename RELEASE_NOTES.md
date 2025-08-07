# Release v0.0.1-beta

Este es el primer lanzamiento beta de una herramienta CLI para interactuar con Azure DevOps.

El objetivo de esta versión inicial es proporcionar funcionalidades para la gestión de **Grupos de Variables** y **Pipelines como Código** desde la línea de comandos.

## ✨ Funcionalidades

### Gestión de Grupos de Variables (`variables`)

- **`create`**: Permite la creación de nuevos *Variable Groups*, incluyendo variables normales y secretas.
- **`get`**: Recupera la información de un *Variable Group* por su nombre en formato JSON.
- **`update`**: Agrega o modifica variables en grupos de variables existentes.
- **`delete`**: Elimina *Variable Groups* completos o variables específicas dentro de un grupo. Incluye un sistema de confirmación para operaciones destructivas.

### Gestión de Pipelines (`pipelines`)

- **`create`**: Crea un pipeline en Azure DevOps a partir de un archivo YAML ubicado en un repositorio de **Azure Repos** o **GitHub**.
- **`get`**: Busca y muestra los detalles de un pipeline a partir de su nombre.
- **`update`**: Modifica un pipeline existente, permitiendo cambiar su nombre, la ruta al archivo YAML, el repositorio de origen o la conexión de servicio (Service Connection).
- **`delete`**: Elimina un pipeline por su ID. Esta función maneja la eliminación previa de las retenciones (retention leases) para asegurar que el borrado se complete.

## 🚀 Cómo Empezar

1.  Descargar el binario correspondiente a su sistema operativo desde los **Assets** de este release.
2.  Configurar las siguientes variables de entorno para la autenticación:
    ```bash
    export AZURE_ORG="su-organizacion"
    export AZURE_PROJECT="su-proyecto"
    export AZURE_PAT="su-personal-access-token"
    ```
3.  Ejecutar `azdevops --help` para ver la lista de comandos disponibles.

