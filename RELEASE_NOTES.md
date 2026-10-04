# Release v0.1.0

Esta versión convierte la CLI en una herramienta interactiva y mucho más completa, y reescribe su núcleo para que sea más fiable en scripts y CI.

## ✨ Novedades

### Modo interactivo
- `azdevops` sin argumentos (o `azdevops interactive`) abre un menú navegable generado a partir de los comandos.
- Todos los comandos preguntan los datos que falten: listas filtrables de Variable Groups, pipelines, ejecuciones, repositorios, conexiones de servicio y grupos de seguridad; los valores secretos se piden sin mostrarse.
- Confirmaciones antes de operaciones destructivas. Se desactiva con `--no-input`, `AZDEVOPS_NO_INPUT` o `CI`.

### Perfiles y autenticación
- `azdevops login` valida las credenciales (incluida la identidad, para no aceptar un PAT inválido en organizaciones con proyectos públicos) y las guarda en un perfil.
- El PAT se guarda en el **llavero del sistema** (Keychain, Credential Manager, Secret Service) cuando está disponible (`--store auto|keyring|file`).
- `azdevops config view | use | delete | path` y flags globales `--org`, `--project`, `--pat`, `--profile`.

### Variable Groups
- `variables list` (tabla con conteo de variables y secretas).
- `variables get --id` y `-o table` con valores secretos enmascarados.
- `variables export` / `variables import` desde `.env` o JSON (`--secret`, `--create`).
- `variables copy`: copia un grupo (también a otro proyecto) pidiendo los valores secretos o con `--secret-value`; `--merge` para fusionar en uno existente.
- `variables diff`: compara dos grupos (dev vs prod, o entre proyectos) con `--exit-code` para CI.

### Pipelines
- `pipelines list` con el último resultado de cada pipeline.
- `pipelines run`: `--name`, `--branch`, `--timeout`, progreso en vivo con la etapa en curso, resumen de tareas fallidas y código de salida 1 si la ejecución falla.
- `pipelines runs`, `pipelines status [--watch]`, `pipelines logs [--failed | --log-id]` y `pipelines cancel`.
- `pipelines update --branch` y `pipelines create --branch` (ahora sí se aplica la rama por defecto).

### Seguridad
- `security list-groups --search` y `--project-only`.
- `security list-members` y `security remove-member`.

### Aprobaciones y environments
- `approvals list | approve | reject` para las aprobaciones de despliegue (con selección múltiple en modo interactivo).
- `environments list | deployments`.

### Work items (Boards)
- `workitems list | get | create | update | delete` (alias `wi`): por defecto lista tus work items abiertos; filtros por tipo, estado, texto, área o WIQL; `create --parent`, `--assigned-to @me`, `update --state --comment` y cambio de estado guiado.

### Conexiones de servicio y agentes
- `service-connections list | get | history | share` (alias `sc`).
- `agents pools | list | enable | disable`: estado de los agentes, trabajo actual y último resultado.

### Navegador
- `azdevops open [project|pipeline|run|variables|workitem|repo|approvals|…]`, con `--print` para solo obtener la URL.

### General
- Salida `-o table|json|yaml|tsv` y `--query` con JMESPath (como en `az`), autocompletado dinámico (nombres de Variable Groups, IDs de pipelines), `--version`, `--debug`.

## 🛠 Mejoras y correcciones
- Cliente HTTP centralizado: se valida el código de respuesta de **todas** las llamadas (antes muchas `GET` ignoraban errores), mensajes de error legibles de la API, detección de PAT inválido (respuesta 203 de Azure DevOps) y reintentos ante 429/5xx.
- Los nombres con espacios o caracteres especiales ahora se codifican correctamente en las URLs.
- `pipelines run --wait` ya no se queda esperando para siempre ni devuelve éxito si falla la consulta de estado.
- `variables set-permissions` aplica a todos los grupos que coinciden con el nombre (antes solo al primero).
- `pipelines update` ya no falla con definiciones sin `process`/`repository`.
- Se corrigieron fugas de conexiones (`defer` dentro de bucles).
- CI con `gofmt`, `go vet`, tests y build; la versión se inyecta en el binario al compilar.

## ⚠️ Cambios de comportamiento
- Los mensajes de estado (✔/✖) se escriben en *stderr*; *stdout* queda solo para datos.
- `pipelines run --wait` devuelve código 1 si la ejecución termina en `failed` o `canceled`.
- Sin terminal, las eliminaciones exigen `--yes` en lugar de leer la confirmación de stdin.
- `variables get` y `pipelines get --name` devuelven error (código 1) si no encuentran resultados.
- `pipelines get-by-id` queda obsoleto: usa `pipelines get --id`.
