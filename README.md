# Azure DevOps CLI

Herramienta de línea de comandos (CLI) no oficial para interactuar con Azure DevOps. Simplifica la gestión de *Variable Groups*, *Pipelines*, aprobaciones, *work items*, conexiones de servicio, agentes y *grupos de seguridad* desde tu terminal, tanto de forma **interactiva** (menús y asistentes) como en **scripts y CI**.

## Características

* **Modo interactivo**: ejecuta `azdevops` sin argumentos para navegar por menús. Además, cualquier comando al que le falte un dato te lo pregunta (con listas filtrables de Variable Groups, pipelines, repositorios, ejecuciones…).
* **Perfiles**: guarda organización, proyecto y PAT con `azdevops login` (el PAT va al **llavero del sistema** cuando está disponible) y cambia entre varios con `azdevops config use`.
* **Variable Groups**: listar, consultar, crear, actualizar, eliminar (grupos o variables), asignar permisos, **exportar/importar** desde `.env` o JSON, **copiar** (incluso a otro proyecto) y **comparar** (`diff`).
* **Pipelines**: listar con su último resultado, crear, actualizar, eliminar, **ejecutar con progreso en vivo**, ver ejecuciones, estado por etapas, **logs** (incluyendo solo las tareas fallidas) y cancelar.
* **Aprobaciones y environments**: ver aprobaciones de despliegue pendientes y aprobarlas o rechazarlas; listar environments y su historial de despliegues.
* **Work items (Boards)**: listar los tuyos, consultar, crear, cambiar estado, comentar, reasignar y eliminar.
* **Conexiones de servicio y agentes**: listar, ver qué pipelines usan una conexión, compartirla con otros proyectos; ver pools y agentes (en línea, trabajo actual) y habilitarlos o deshabilitarlos.
* **`azdevops open`**: abre en el navegador el proyecto, un pipeline, una ejecución, un work item, etc.
* **Seguridad**: listar y buscar grupos, ver miembros, agregar y quitar miembros.
* **Salida** en tabla, JSON, YAML o TSV (`-o`), con filtros **JMESPath** (`--query`, igual que en `az`). Los mensajes de estado van a *stderr* y los datos a *stdout*.
* **Autocompletado** de comandos, flags, nombres de Variable Groups e IDs de pipelines.
* Cliente HTTP robusto: errores legibles de la API, detección de PAT inválido/expirado, reintentos ante *throttling* (429) y errores 5xx, y `--debug` para ver cada petición.

## Instalación

### Desde las releases

Descarga el binario para tu plataforma desde la página de *Releases* del repositorio y colócalo en tu `PATH`.

### Desde el código

```bash
git clone https://github.com/henryd24/azure_devops_cli.git
cd azure_devops_cli
make install           # instala azdevops en $GOPATH/bin
make build VERSION=v0.1.0   # o genera binarios para Linux, macOS y Windows en dist/
```

## Autenticación

La forma más cómoda es guardar un perfil:

```bash
azdevops login                      # asistente: organización, PAT y selección de proyecto
azdevops login --profile cliente2   # un segundo perfil
azdevops config view                # lista perfiles (PAT enmascarado) y la configuración efectiva
azdevops config use cliente2        # cambia el perfil por defecto
```

El perfil se guarda en `~/.config/azdevops/config.json` con permisos `0600`. El PAT se guarda en el **llavero del sistema** (Keychain en macOS, Credential Manager en Windows, Secret Service/GNOME Keyring en Linux) si está disponible; si no, en ese archivo. Puedes forzarlo con `--store keyring` o `--store file`.

También puedes usar variables de entorno (ideal para CI) o flags:

| Dato         | Flag        | Variable de entorno |
|--------------|-------------|---------------------|
| Organización | `--org`     | `AZURE_ORG`         |
| Proyecto     | `--project` | `AZURE_PROJECT`     |
| PAT          | `--pat`     | `AZURE_PAT`         |
| Perfil       | `--profile` | `AZDEVOPS_PROFILE`  |

Precedencia: flags › perfil indicado con `--profile` › variables de entorno › perfil por defecto.

## Modo interactivo

```bash
azdevops              # abre el menú (también: azdevops interactive)
azdevops pipelines run          # te deja elegir el pipeline, la rama, parámetros y si esperar
azdevops variables update       # eliges los grupos y agregas variables una a una (las secretas no se muestran)
```

Los prompts solo aparecen en una terminal. Se desactivan con `--no-input`, con `AZDEVOPS_NO_INPUT=1` o cuando existe la variable `CI`; en ese caso, si falta un dato obligatorio el comando falla con un mensaje claro.

## Uso

### Variable Groups (`variables`, alias `vg`)

```bash
azdevops variables list                                  # tabla con todos los grupos
azdevops variables list --filter "app-*" -o json
azdevops variables get --name "MiGrupo"                  # JSON (admite comodines: "MiGrupo*")
azdevops variables get --id 42 -o table                  # tabla de variables (secretas enmascaradas)

azdevops variables create --name MiGrupo -d "Descripción" -v clave1=valor1 -v secret:token=abc
azdevops variables update --name MiGrupo --name OtroGrupo -v "clave2=nuevo,secret:otra=xyz"
azdevops variables delete --name MiGrupo --variables clave1,otra
azdevops variables delete --name MiGrupo --yes           # elimina el grupo completo

azdevops variables set-permissions --variable MiGrupo --user ana@empresa.com --group Devs --role Reader

azdevops variables export --name MiGrupo > .env                    # o --format json --file vars.json
azdevops variables import --name MiGrupo --file .env --secret DB_PASSWORD
azdevops variables import --name NuevoGrupo --file vars.json --create

azdevops variables copy --from app-dev --to app-qa
azdevops variables copy --from app-dev --to app-dev --to-project OtroProyecto --secret-value DB_PASS=xxx
azdevops variables diff app-dev app-prod                       # solo lo que cambia
azdevops variables diff app-prod app-prod --project-b Otro --all
azdevops variables diff app-dev app-prod --exit-code           # código 1 si hay diferencias (CI)
```

* Usa el prefijo `secret:` para crear variables secretas.
* Los grupos de seguridad se indican sin el prefijo `[Proyecto]\`.
* La API no devuelve el valor de las variables secretas: `export` las deja comentadas, `copy` te pide su valor (o usa `--secret-value`) y `diff` solo compara si existen.

### Pipelines (`pipelines`, alias `pl`)

```bash
azdevops pipelines list                                   # nombre, carpeta y último resultado
azdevops pipelines get --name "MiPipeline*"
azdevops pipelines get --id 123

azdevops pipelines create --name MiPipeline --repo-name mi-repo --yaml-path azure-pipelines.yml --branch main
azdevops pipelines create -n MiPipeline -t gitHub -r org/repo -p .azure/ci.yml -s <id-conexion>
azdevops pipelines update --id 123 --new-name NuevoNombre --branch develop
azdevops pipelines delete --id 123 --yes

# Ejecutar (por --id o --name) y esperar mostrando la etapa en curso
azdevops pipelines run --name MiPipeline --branch feature/x --wait --timeout 30m
azdevops pipelines run --id 123 --param imageTag=1.2.3 --var deployEnv=staging --var secret:apiKey=xxx

azdevops pipelines runs --id 123 --top 10                 # ejecuciones recientes
azdevops pipelines status --build-id 4567                 # estado + etapas/jobs
azdevops pipelines status --build-id 4567 --watch         # seguirla hasta que termine
azdevops pipelines logs --build-id 4567 --failed          # logs de las tareas que fallaron
azdevops pipelines logs --build-id 4567 --log-id 12
azdevops pipelines cancel --build-id 4567
```

Con `--wait`, el comando termina con código **1** si la ejecución falla o se cancela y muestra las tareas con error, lo que permite usarlo como paso de CI.

### Aprobaciones y environments

```bash
azdevops approvals list                                   # pendientes
azdevops approvals list --state all -o json
azdevops approvals approve --id <id> --comment "OK para prod"
azdevops approvals approve                                # elegir de la lista
azdevops approvals reject --id <id> --comment "Falta validar" --yes

azdevops environments list
azdevops environments deployments --name produccion --top 10
```

### Work items (`workitems`, alias `wi`)

```bash
azdevops wi list                                          # tus work items abiertos
azdevops wi list --type Bug --state Active --assigned-to any
azdevops wi list --search "login" --all-states
azdevops wi list --wiql "SELECT [System.Id] FROM WorkItems WHERE [System.Tags] CONTAINS 'urgente'"
azdevops wi get 123
azdevops wi create --type Bug --title "Falla el login" --assigned-to @me --tags "frontend; urgente"
azdevops wi create --type Task --title "Escribir tests" --parent 123
azdevops wi update 123 --state Active --comment "Empiezo con esto"
azdevops wi update                                        # eliges el work item y qué cambiar
azdevops wi delete 123 --yes                              # a la papelera (recuperable)
```

### Conexiones de servicio (`service-connections`, alias `sc`) y agentes

```bash
azdevops sc list --type azurerm
azdevops sc get --name DockerHub
azdevops sc history --name DockerHub                      # qué pipelines la usaron
azdevops sc share --name DockerHub --with-project OtroProyecto

azdevops agents pools
azdevops agents list --pool MiPool                        # estado, versión, trabajo actual y último
azdevops agents disable --pool MiPool --agent build-01 --yes
azdevops agents enable --pool MiPool --agent build-01
```

### Abrir en el navegador

```bash
azdevops open                       # el proyecto
azdevops open pipeline MiPipeline
azdevops open run 4567
azdevops open variables app-dev
azdevops open workitem 42 --print   # solo imprime la URL
```

### Seguridad (`security`)

```bash
azdevops security list-groups --search devs -o table
azdevops security list-groups --project-only
azdevops security search-group --name "MiGrupo"
azdevops security list-members --group "MiGrupo"
azdevops security add-member --target-group Destino --target-group Destino2 --user ana@empresa.com --group MiGrupo
azdevops security remove-member --target-group Destino --user ana@empresa.com --yes
```

## Scripting

* `-o table|json|yaml|tsv`: los comandos de consulta (`get`, `search-group`, `list-groups`) devuelven JSON por defecto; los listados devuelven tabla.
* `--query` (`-q`) aplica una consulta [JMESPath](https://jmespath.org) sobre los datos, igual que `az`:

  ```bash
  azdevops pipelines list -q "[?latestBuild.result=='failed'].{id:id, nombre:name}" -o yaml
  azdevops variables list -q "[].[id,name]" -o tsv | while IFS=$'\t' read id name; do ...; done
  ID=$(azdevops wi create --type Task --title "x" -q id -o tsv)
  ```
* Los datos van a *stdout* y los mensajes (✔, !, ✖) a *stderr*: `azdevops variables get -n MiGrupo | jq '.[0].variables'`.
* Códigos de salida: `0` éxito, `1` error, `130` cancelado por el usuario.
* `--debug` (o `AZDEVOPS_DEBUG=1`) muestra cada petición HTTP con su código y duración.
* `AZDEVOPS_BASE_URL` / `AZDEVOPS_VSSPS_URL` permiten apuntar a otro host (por ejemplo, para pruebas).

## Autocompletado

```bash
# bash
source <(azdevops completion bash)
# zsh
azdevops completion zsh > "${fpath[1]}/_azdevops"
# PowerShell
azdevops completion powershell | Out-String | Invoke-Expression
```

Completa también nombres de Variable Groups (`--name`) e IDs de pipelines (`--id`) consultando tu proyecto.

## Desarrollo

```bash
make test    # go test ./...
make vet
```

Estructura: `azdevops/` contiene el cliente y las llamadas a la API (sin E/S de consola), `cmd/` los comandos de cobra, `internal/ui` los prompts y el formateo (tablas, JSON/YAML/TSV, JMESPath), e `internal/config` los perfiles y el llavero. El menú interactivo se genera a partir del árbol de comandos, así que los comandos nuevos aparecen en él automáticamente.

## Licencia

Este proyecto está bajo la Licencia Pública General de GNU v3.0. Consulta el archivo `LICENSE` para más detalles.

## Contribuciones

Las contribuciones son bienvenidas. Si deseas colaborar, abre un *issue* para discutir tus ideas o envía un *pull request* con tus cambios.
