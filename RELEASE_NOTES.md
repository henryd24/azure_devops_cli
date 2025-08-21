# Release v0.0.2-beta

Esta versión expande significativamente las capacidades de la CLI, introduciendo la ejecución de pipelines y una suite completa para la gestión de permisos y grupos de seguridad. Se ha puesto un foco especial en la optimización de las llamadas a la API para un rendimiento eficiente.

## ✨ Funcionalidades Nuevas

### Ejecución de Pipelines (`pipelines run`)

Se ha añadido un nuevo comando `run` al subcomando `pipelines` que permite:
- **Iniciar un pipeline** por su ID de definición.
- **Esperar su finalización** con el flag `--wait` para flujos de trabajo síncronos.
- **Pasar parámetros** de plantilla YAML usando el flag `--param`.
- **Inyectar variables** en tiempo de ejecución con el flag `--var`, con soporte para variables secretas usando el prefijo `secret:`.

### Gestión de Seguridad y Permisos (`security`)

Se introduce un nuevo conjunto de comandos de alto nivel para gestionar la seguridad:
- **`security list-groups`**: Lista todos los grupos de seguridad. Incluye un flag `--search` para filtrar los resultados y maneja la paginación automáticamente para obtener la lista completa.
- **`security search-group`**: Busca un grupo específico por nombre de forma eficiente, ideal para scripts.
- **`security add-member`**: Agrega uno o más usuarios y/o grupos a uno o más grupos de seguridad en una sola operación.

### Permisos para Variable Groups (`variables set-permissions`)

Se añade un comando para gestionar los roles en los *Variable Groups*:
- **`variables set-permissions`**: Asigna los roles de **Reader**, **User**, o **Administrator** a múltiples usuarios y/o grupos en múltiples *Variable Groups* de forma masiva.
