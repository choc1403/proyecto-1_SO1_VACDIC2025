# Manual Técnico
---
## [Desarrollo de un módulo de kernel en C y un daemon en Go para el monitoreo de procesos y contenedores en Linux ]
---
## 1. Introducción

Este documento describe el diseño técnico, la arquitectura y el funcionamiento
interno del sistema de monitoreo de procesos y contenedores desarrollado
mediante módulos de kernel en C y un daemon en Go para sistemas Linux.

## 2. Alcance del sistema

El sistema permite recolectar métricas de procesos directamente desde el kernel,
exponerlas mediante el sistema de archivos /proc y analizarlas en un daemon
en espacio de usuario, el cual toma decisiones automatizadas para la
estabilización del sistema.

## 3. Arquitectura general

El sistema se compone de los siguientes elementos:

- Módulos de kernel en C (sensores de bajo nivel)
- Interfaz /proc para comunicación kernel–usuario
- Daemon en Go (procesamiento y toma de decisiones)
- Base de datos SQLite
- Sistema de visualización con Grafana


## 4. Requisitos del sistema

- Virtualizar un Sistema operativo Linux (Ubuntu 24.04.3 LTS)
- Compilador GCC
- Go 1.20 o superior
- Docker
- SQLite
- Grafana
- Permisos de superusuario

## 5. Estructura del proyecto

```text
proyecto-1
├── modulo-kernel/      # Módulos de kernel en C
├── go-daemon/              # Daemon en Go
├── bash/             # Scripts de automatización
├── crodocumentacion/                # Documentación del proyecto
├── dashboard/             # Dashboards
└── README.md
```

## 6. Descripción de componentes
   - **Módulos de kernel**
   
### Archivo `common.h`
El archivo `common.h` define un conjunto de **constantes, inclusiones y funciones auxiliares compartidas** utilizadas por los módulos de kernel del proyecto. Su objetivo principal es **centralizar la lógica común** relacionada con la obtención de métricas del sistema y de procesos, evitando duplicación de código y facilitando el mantenimiento.

Este archivo es incluido por los distintos módulos de kernel encargados del monitoreo de procesos generales y de contenedores, y proporciona utilidades para acceder de forma segura a estructuras internas del kernel de Linux.

---

### Funcionalidades principales

El archivo `common.h` provee las siguientes funcionalidades:

#### 1. Definición de nombres de interfaces `/proc`

```c
#define PROC_NAME_SYS  "sysinfo_so1_202041390"
#define PROC_NAME_CONT "continfo_so1_202041390"
```

Estas constantes definen los nombres de los archivos virtuales que los módulos de kernel crearán dentro del sistema de archivos `/proc`.
Cada archivo expone información distinta:

* **Procesos generales del sistema**
* **Procesos asociados a contenedores**

Esto permite una separación clara de responsabilidades y facilita el consumo de datos desde el espacio de usuario.

---

#### 2. Obtención de información de memoria del sistema

```c
static void get_meminfo_kb(unsigned long *total_kb, unsigned long *free_kb)
```

Esta función auxiliar obtiene información global de memoria del sistema utilizando la estructura `sysinfo` del kernel.
Devuelve:

* Memoria total del sistema (en KB)
* Memoria libre del sistema (en KB)

Es utilizada para calcular porcentajes de uso de memoria y métricas agregadas.

---

#### 3. Cálculo de memoria virtual y residente de un proceso

```c
static void get_mem_from_mm(struct mm_struct *mm,
                            unsigned long *vsz_kb,
                            unsigned long *rss_kb)
```

Esta función extrae métricas de memoria de un proceso a partir de su estructura `mm_struct`:

* **VSZ (Virtual Size)**: tamaño total de memoria virtual del proceso.
* **RSS (Resident Set Size)**: memoria física realmente utilizada.

Las métricas se expresan en kilobytes y se calculan a partir del número de páginas asignadas al proceso.

---

#### 4. Lectura de la línea de comandos de un proceso

```c
static int read_task_cmdline(struct task_struct *task,
                             char *buf,
                             int bufsize)
```

Esta función obtiene la línea de comandos (`cmdline`) asociada a un proceso:

* Accede al espacio de memoria del proceso mediante `access_process_vm`.
* Lee los argumentos originales con los que fue ejecutado el proceso.
* Convierte los separadores nulos (`\\0`) en espacios para facilitar su visualización.

La lectura se realiza de forma segura utilizando mecanismos de sincronización del kernel como `rcu_read_lock`.

---

#### 5. Cálculo de porcentajes de uso

```c
static unsigned long percent_of_x100(unsigned long part_kb,
                                     unsigned long total_kb)
```

Función auxiliar para calcular porcentajes con precisión entera, retornando el resultado multiplicado por 100 (por ejemplo, 12.34% → 1234).
Este enfoque evita el uso de números de punto flotante dentro del kernel.

---

### Importancia dentro del sistema

El archivo `common.h` cumple un rol clave en el proyecto:

* Centraliza funciones reutilizables entre módulos de kernel.
* Facilita el acceso seguro a métricas internas del kernel.
* Mejora la legibilidad y mantenibilidad del código.
* Reduce la complejidad de los módulos principales al abstraer lógica común.

Gracias a este archivo, los módulos de kernel pueden enfocarse exclusivamente en la recolección y exposición de información, delegando los cálculos y accesos auxiliares a una capa compartida.

---

### Archivo `continfo.c`



El archivo `continfo.c` implementa un **módulo de kernel en C encargado de monitorear procesos asociados a contenedores** en un sistema Linux. Este módulo actúa como un **sensor de bajo nivel**, accediendo directamente a las estructuras internas del kernel para recolectar métricas detalladas de procesos que aparentan pertenecer a entornos contenerizados.

La información recolectada es expuesta al espacio de usuario mediante un **archivo virtual en el sistema de archivos `/proc`**, permitiendo que aplicaciones en espacio de usuario (como el daemon en Go) consuman los datos de forma eficiente y estructurada.

---

### Funcionalidades principales

#### 1. Identificación de procesos asociados a contenedores

```c
static bool is_container_task(char *cmdline)
```

Esta función auxiliar determina si un proceso puede considerarse parte de un contenedor, analizando su línea de comandos (`cmdline`).
El criterio de clasificación se basa en la presencia de palabras clave comúnmente asociadas a entornos contenerizados, tales como:

* `docker`
* `container`
* `runc`
* `busybox`

Este enfoque permite una detección heurística de procesos de contenedores sin depender directamente de APIs específicas de Docker.

---

#### 2. Recolección de métricas de memoria del sistema

Antes de analizar procesos individuales, el módulo obtiene información global de memoria:

* Memoria total del sistema
* Memoria libre
* Memoria utilizada

Estos valores se utilizan como referencia para calcular porcentajes de consumo por proceso.

---

#### 3. Recorrido de procesos del sistema

```c
for_each_process(task)
```

El módulo itera sobre todos los procesos activos del sistema utilizando las primitivas del kernel.
Durante este recorrido:

* Se obtiene el tiempo de CPU consumido por el proceso (en jiffies).
* Se recupera el estado actual del proceso.
* Se analiza su línea de comandos para determinar si pertenece a un contenedor.

El acceso a la lista de procesos se realiza dentro de una sección protegida con `rcu_read_lock`, garantizando la seguridad de concurrencia.

---

#### 4. Obtención de métricas por proceso

Para cada proceso identificado como contenedor, el módulo recolecta:

* **PID** del proceso.
* **Nombre del proceso** (`comm`).
* **Línea de comandos completa**.
* **Memoria virtual (VSZ)** en KB.
* **Memoria residente (RSS)** en KB.
* **Porcentaje de uso de memoria**, calculado sin el uso de punto flotante.
* **Tiempo de CPU acumulado** (user + system) en jiffies.
* **Estado del proceso**.

Estas métricas permiten un análisis detallado del impacto de cada contenedor sobre los recursos del sistema.

---

#### 5. Exposición de información vía `/proc`

```c
/proc/continfo_so1_202041390
```

El módulo crea un archivo virtual en `/proc` que, al ser leído, genera dinámicamente una salida estructurada en formato **JSON-like**, facilitando su parseo por aplicaciones en espacio de usuario.

Ejemplo conceptual de la salida:

```json
{
  "mem_total_kb": 16384000,
  "mem_free_kb": 8200000,
  "mem_used_kb": 8184000,
  "containers": [
    {
      "pid": 1234,
      "name": "docker",
      "cmdline": "docker run ...",
      "vsz_kb": 204800,
      "rss_kb": 102400,
      "mem_pct": "1.25",
      "proc_jiffies": 34567,
      "state": "R"
    }
  ]
}
```

---

#### 6. Uso de la interfaz `seq_file`

El módulo utiliza la interfaz `seq_file` del kernel para:

* Manejar lecturas de archivos `/proc` de forma eficiente.
* Evitar problemas de desbordamiento de buffers.
* Facilitar la generación progresiva del contenido.

---

#### 7. Inicialización y liberación del módulo

```c
static int __init cont_init(void)
static void __exit cont_exit(void)
```

* En la inicialización, se registra el archivo `/proc` y se notifica mediante `pr_info`.
* En la salida, se elimina la entrada de `/proc`, asegurando una correcta liberación de recursos.

---

### Importancia dentro del sistema

El archivo `continfo.c` cumple un rol fundamental dentro del proyecto:

* Proporciona visibilidad de bajo nivel sobre procesos de contenedores.
* Sirve como fuente primaria de métricas para el daemon en Go.
* Permite la toma de decisiones automatizadas basadas en datos reales del kernel.
* Facilita la integración con sistemas de monitoreo y visualización como Grafana.

Este módulo constituye la base del sistema de monitoreo proactivo de contenedores, permitiendo combinar observabilidad, automatización y estabilidad del sistema.

---

### Archivo `sysinfo.c`

### Descripción técnica

El archivo `sysinfo.c` implementa un **módulo de kernel en C encargado de monitorear los procesos generales del sistema Linux**, sin aplicar filtros por contenedores. Este módulo proporciona una **visión global del estado de los procesos activos**, permitiendo analizar el consumo de recursos a nivel de sistema.

Al igual que el módulo de contenedores, la información recolectada se expone mediante un **archivo virtual en el sistema de archivos `/proc`**, el cual es consumido por aplicaciones en espacio de usuario, como el daemon desarrollado en Go.

---

### Funcionalidades principales

#### 1. Recolección de métricas globales de memoria

Al iniciar la lectura del archivo `/proc`, el módulo obtiene información general del estado de la memoria del sistema:

* Memoria total del sistema (KB)
* Memoria libre (KB)
* Memoria utilizada (KB)

Estas métricas sirven como contexto para interpretar el consumo de memoria de cada proceso individual.

---

#### 2. Recorrido de todos los procesos del sistema

```c
for_each_process(task)
```

El módulo recorre la lista completa de procesos activos del sistema utilizando las primitivas internas del kernel.
Durante este recorrido, se accede de forma segura a las estructuras del kernel mediante un bloqueo `rcu_read_lock`, garantizando la consistencia de los datos mientras se realiza la lectura.

---

#### 3. Obtención de métricas por proceso

Para cada proceso del sistema, el módulo recolecta las siguientes métricas:

* **PID** del proceso.
* **Nombre del proceso** (`comm`).
* **Línea de comandos completa** utilizada para su ejecución.
* **Memoria virtual (VSZ)** en KB.
* **Memoria residente (RSS)** en KB.
* **Porcentaje de uso de memoria**, calculado sin el uso de punto flotante.
* **Tiempo de CPU acumulado** en jiffies (modo usuario + modo kernel).
* **Estado actual del proceso** (ejecución, espera, detenido, etc.).

Este conjunto de datos permite realizar análisis detallados sobre el comportamiento y el impacto de cada proceso en el sistema.

---

#### 4. Cálculo de porcentajes sin punto flotante

El cálculo del porcentaje de uso de memoria se realiza mediante aritmética entera, devolviendo el valor con dos decimales simulados.
Este enfoque es necesario debido a las restricciones del entorno del kernel, donde el uso de operaciones de punto flotante no está permitido.

---

#### 5. Exposición de información mediante `/proc`

```c
/proc/sysinfo_so1_202041390
```

El módulo crea un archivo virtual en `/proc` que, al ser leído, genera dinámicamente una salida estructurada en formato **JSON-like**, facilitando su parseo desde el espacio de usuario.

Ejemplo conceptual de salida:

```json
{
  "mem_total_kb": 16384000,
  "mem_free_kb": 8200000,
  "mem_used_kb": 8184000,
  "processes": [
    {
      "pid": 1,
      "name": "systemd",
      "cmdline": "/sbin/init",
      "vsz_kb": 120000,
      "rss_kb": 8000,
      "mem_pct": "0.73",
      "proc_jiffies": 152345,
      "state": "S"
    }
  ]
}
```

---

#### 6. Uso de la interfaz `seq_file`

El módulo hace uso de la interfaz `seq_file` del kernel para:

* Manejar lecturas eficientes de archivos `/proc`.
* Evitar problemas de memoria y desbordamientos.
* Facilitar la generación secuencial de grandes volúmenes de información.

---

#### 7. Inicialización y liberación del módulo

```c
static int __init sys_init(void)
static void __exit sys_exit(void)
```

* Durante la inicialización, se crea la entrada correspondiente en `/proc` y se registra el evento en el log del kernel.
* Durante la salida, se elimina correctamente la entrada del sistema de archivos virtual, garantizando la liberación adecuada de recursos.

---

### Importancia dentro del sistema

El archivo `sysinfo.c` cumple un rol esencial dentro del proyecto:

* Proporciona una vista completa del estado de los procesos del sistema.
* Sirve como base comparativa frente a los procesos asociados a contenedores.
* Permite detectar procesos anómalos o de alto consumo.
* Alimenta al daemon en Go con información confiable directamente desde el kernel.

Este módulo complementa al módulo de contenedores, ofreciendo una perspectiva integral del consumo de recursos y permitiendo una gestión más efectiva del sistema.

---

   - **Daemon en Go**
   

### Package `functions`

**Monitorización de CPU, recursos del sistema y gestión inteligente de contenedores Docker en Linux**

Este paquete forma parte del *daemon de monitorización* del proyecto **so1-daemon**, cuyo objetivo es leer métricas del sistema, calcular el uso de recursos por proceso y aplicar políticas automatizadas (como eliminar contenedores) cuando se superan umbrales críticos.

El paquete contiene lógica para:

* Leer métricas del sistema desde `/proc`.
* Calcular uso de CPU por proceso.
* Mapear procesos ↔ contenedores Docker.
* Registrar métricas en la base de datos.
* Decidir si finalizar contenedores que consumen demasiados recursos.

---

### Estructura del paquete

### 1. `functions.cpu.go`

Archivo encargado de la **lectura de tiempos de CPU** y del **cálculo del porcentaje de CPU usado por proceso**.

---

### `func ReadTotalJiffies() (uint64, error)`

Lee el tiempo total acumulado de CPU del sistema desde `/proc/stat`.

**Proceso:**

1. Lee la primera línea del archivo `/proc/stat` (que inicia con `cpu`).
2. Ignora el primer token (`cpu`).
3. Suma todos los valores numéricos restantes (tiempos en modo usuario, sistema, idle, etc.).
4. Retorna el total en **jiffies** (unidades internas del kernel).

**Ejemplo de línea:**

```
cpu 100000 0 50000 200000 0 0 1000 0
```

---

###  `func ReadProcPidTime(pid int) (uint64, error)`

Obtiene el tiempo de CPU consumido por un proceso específico leyendo el archivo:

```
/proc/<pid>/stat
```

**Campos relevantes:**

* **Campo 14** → `utime`: tiempo en modo usuario
* **Campo 15** → `stime`: tiempo en modo kernel

La función extrae ambos y retorna:

```
utime + stime
```

**Nota:** Maneja correctamente el nombre del comando encerrado en paréntesis, el cual puede contener espacios.

---

###  `func CalcCpuPercent(pid int, curProcTime, curTotal uint64, curTs time.Time) float64`

Calcula el **porcentaje de uso de CPU** para un PID.

El cálculo se basa en el delta entre mediciones consecutivas:

```
CPU% = (ΔTiempoProceso / ΔTiempoTotalSistema) * 100
```

**Lógica clave:**

* Usa un `mutex` para evitar condiciones de carrera al leer/escribir muestras previas.
* Si no existe muestra previa → retorna `0.0`.
* Si existe, calcula las diferencias y el porcentaje.

Este proceso permite un monitoreo continuo y preciso por proceso.

---

### 2. `functions.docker.go`

Contiene la lógica para correlacionar **procesos del sistema** con sus correspondientes **contenedores Docker**.

---

###  `func GetDockerPidMap() map[int]DockerInfo`

Genera un mapa donde:

* **Clave** → PID del proceso en el host
* **Valor** → Información del contenedor (`DockerInfo`)

**Pasos:**

1. Obtiene todos los contenedores activos:

   ```
   docker ps -q
   ```
2. Para cada contenedor ejecuta:

   ```
   docker inspect --format "{{.State.Pid}} {{.Id}} {{.Config.Image}} {{.Name}}"
   ```
3. Construye un mapa:

   ```go
   map[pid]DockerInfo
   ```

Esto permite relacionar cualquier PID detectado durante la monitorización con su contenedor correspondiente.

---

### 3. `functions.logic.go`

El archivo más importante: **la lógica de decisión del daemon**.
Orquesta lectura de métricas, cálculo de uso de recursos y decide si debe terminar contenedores.

---

###  `func ProcessOnce() error`

Ejecuta un ciclo completo del daemon.

1. **Lee métricas globales del sistema** desde `PROC_SYS`.
2. Limpia JSON y hace `Unmarshal` a `ProcSys`.
3. Inserta las métricas en la base de datos.
4. **Lee métricas de procesos** desde `PROC_CONT`.
5. Llama a `DecideAndAct()` para aplicar políticas de gestión.

---

##  `func DecideAndAct(containers []ProcProcess)`

Implementa la política de gestión de recursos del sistema.

### 1. Asociar procesos con contenedores Docker

* Llama a `GetDockerPidMap()`.
* Clasifica procesos como:

  * Contenedores Docker.
  * Procesos no Docker (placeholder con nombre del comando).

### 2. Clasificación por prioridad (heurística)

Basada en el nombre del contenedor o de la imagen:

* Contenedor **High Priority** → contiene `high`, `cpu`, `mem`.
* Contenedor **Low Priority** → cualquier otro.

Se mantienen mínimos definidos por las constantes:

```
MIN_HIGH_CONTAINERS
MIN_LOW_CONTAINERS
```

### 3. Registro y cálculo de métricas

Para cada proceso:

* Calcula CPU% con las funciones del archivo cpu.go.
* Calcula uso de memoria.
* Inserta la información en la base de datos.

### 4. Política de eliminación ("kill switch")

Un contenedor puede ser eliminado si:

* CPU% > `CPU_THRESHOLD` **o**
* Mem% > `MEM_THRESHOLD`

**PERO** se evita eliminar si:

* No es un contenedor Docker.
* Es grafana (se protege explícitamente).
* Violenta los mínimos de contenedores High/Low.

Si se decide eliminar un contenedor:

1. Se registra en el log.
2. Se ejecuta:

   ```
   docker rm -f <container>
   ```
3. Se registra la eliminación en la base de datos.
4. Se actualizan contadores de clasificaciones.

---

###  Conclusión

El paquete `functions` implementa:
- Monitorización de recursos del sistema
- Cálculo de uso de CPU por proceso
- Correlación proceso ↔ contenedor Docker
- Registro de métricas históricas
- Sistema automático de control y terminación de contenedores
- Políticas de seguridad y prioridades

Es un bloque esencial del daemon encargado de mantener la **estabilidad del sistema** mediante decisiones inteligentes basadas en umbrales de recursos.

---
##  Package `utils`

El paquete **utils** proporciona funciones auxiliares esenciales para el daemon, enfocadas en:

* Limpieza y normalización de datos (especialmente JSON).
* Lectura segura de archivos del sistema (como los de `/proc`).
* Parsing de porcentajes de memoria.
* Ejecución de comandos del sistema *(si deseas, puedo agregar esta sección si tu package lo usa)*.

Estas utilidades permiten que otras partes del sistema (como el módulo de CPU, lógica o monitoreo de contenedores) trabajen con datos limpios, seguros y en un formato consistente.

---

### Funciones principales

##  `var TrailingCommaRe = regexp.MustCompile(",\\s*([\\]\\}])")`

Expresión regular utilizada para detectar **comas finales no válidas en JSON**.

Ejemplo de JSON no estándar:

```json
{
    "key": "value",
}
```

Este tipo de comas no son válidas en JSON estricto, por lo que deben eliminarse antes de hacer `json.Unmarshal`.
La expresión regular identifica casos como:

* `"value", }`
* `"item1", ]`

Y permite sanearlos correctamente.

---

## `func SanitizeJSON(b []byte) []byte`

Limpia y normaliza JSON **malformado** eliminando comas finales inválidas.

### Proceso:

1. Recibe un `[]byte` con contenido JSON.
2. Aplica la expresión regular `TrailingCommaRe`.
3. Reemplaza secuencias como `", }"` → `"}"`.
4. Devuelve un JSON válido que puede parsearse con seguridad mediante `json.Unmarshal`.

### ¿Por qué es útil?

Muchos sistemas generan JSON con trailing commas, lo que rompe el parsing.
Esta función garantiza robustez en el daemon ante este tipo de errores.

---

##  `func ReadProcFile(path string) ([]byte, error)`

Lee archivos del sistema, especialmente los ubicados en:

```
/proc
```

Estos archivos contienen información del kernel, CPU, contenedores montados, etc.

### Proceso:

1. Abre el archivo indicado en `path`.
2. Usa `defer f.Close()` para cerrar el archivo correctamente.
3. Lee su contenido con un límite de:

   ```
   10 MB (10<<20)
   ```

   Esto protege el daemon si un archivo es inesperadamente grande.
4. Retorna un slice de bytes con su contenido o un error.

### Ventajas:

* Seguro ante archivos grandes.
* Ideal para la lectura repetitiva que hace el daemon.
* Funciona de forma uniforme para cualquier archivo del `/proc`.

---

##  `func ParseMemPct(s string) (float64, error)`

Convierte una cadena que representa un porcentaje de memoria a un número `float64`.

### Proceso:

1. Limpia espacios con `strings.TrimSpace`.
2. Si la cadena queda vacía → retorna `0.0`.
3. Convierte usando:

   ```go
   strconv.ParseFloat(s, 64)
   ```

### Uso típico:

En la lógica del daemon, esta función se usa para procesar valores provenientes de archivos como `/proc/cont`, donde la memoria puede venir como:

```
"45.3"
"89"
" 12.7 "
```

La función garantiza que siempre retorne un valor numérico válido.

---

###  Conclusión

El paquete `utils` proporciona funciones esenciales para:

- Manipulación segura de JSON no estándar
- Lectura confiable de archivos del sistema Linux
- Procesamiento de números provenientes de texto
- Robustez total para los módulos que dependen de datos externos

Es un componente fundamental para la estabilidad del sistema de monitorización, garantizando que los datos siempre sean válidos, limpios y utilizables.


---



   - **Base de datos**
   La base de datos esta construido de las siguientes entidades.
   ```sql
   CREATE TABLE IF NOT EXISTS containers (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  container_id TEXT,
  pid INTEGER,
  image TEXT,
  cpu_pct REAL,
  mem_pct REAL,
  ts INTEGER
);

CREATE TABLE IF NOT EXISTS deletions (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  container_id TEXT,
  reason TEXT,
  ts INTEGER
);

CREATE TABLE IF NOT EXISTS sys_metrics (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  mem_total_kb INTEGER,
  mem_free_kb INTEGER,
  mem_used_kb INTEGER,
  ts INTEGER
);

CREATE TABLE IF NOT EXISTS process_count (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  total INTEGER,
  ts INTEGER
);

   ```
   

## 7. Interfaces del sistema

La comunicación entre el kernel y el espacio de usuario se realiza mediante
archivos virtuales en /proc:

- /proc/continfo_so1_20201390
- /proc/sysinfo_so1_202041390

## 8. Flujo de funcionamiento
![Arquitectura del proyecto](https://github.com/choc1403/proyecto-1_SO1_VACDIC2025/blob/master/documentacion/manual_tecnico/img/arquitectura.png)

## 9. Instalación y compilación

### Compilación del módulo de kernel
**Paso 1.** Situarse en la carpeta `modulo-kernel`
```bash
cd modulo-kernel 
```
**Paso 2.** Antes de compilar nuevamente el proyecto, es recomendable eliminar los archivos generados en compilaciones anteriores.
Para ello, ejecute el siguiente comando en la terminal

``` bash
make clean
```
Este comando se encarga de limpiar todos los archivos temporales y binarios creados previamente.
Una vez finalizada la limpieza, puede proceder a generar los nuevos archivos ejecutando

``` bash
make
```
**Paso 3.** Ejecutar los siguientes comandos para crear los archivos en proc
``` bash
sudo insmod sysinfo.ko # Módulo de Procesos del Sistema
sudo insmod continfo.ko # Módulo de Procesos de Contenedores
```

**Paso 4.** Verificar los `proc` creados
```bash
cat /proc/continfo_so1_202041390
```
Este comando muestra todos los procesos de los contenedores, el resultado debe ser el siguiente.
```bash
{
"mem_total_kb": 3960492,
"mem_free_kb": 2198276,
"mem_used_kb": 1762216,
"containers": [
{ "pid": 1420, "name": "containerd", "cmdline": "/usr/bin/containerd ", "vsz_kb": 2391788, "rss_kb": 38184, "mem_pct": "60.39", "proc_jiffies": 358000000, "state": "S" },
{ "pid": 1532, "name": "dockerd", "cmdline": "/usr/bin/dockerd -H fd:// --containerd=/run/containerd/containerd.sock ", "vsz_kb": 2494152, "rss_kb": 68032, "mem_pct": "62.97", "proc_jiffies": 903000000, "state": "S" },
{ "pid": 22071, "name": "containerd-shim", "cmdline": "/usr/bin/containerd-shim-runc-v2 -namespace moby -id 3fa836846de1a1950cd414ffb4936a50aac4200f8e20c7030c38e433020cd388 -address /run/containerd/containerd.sock ", "vsz_kb": 1233836, "rss_kb": 11016, "mem_pct": "31.15", "proc_jiffies": 7000000, "state": "S" },
{ "pid": 23775, "name": "containerd-shim", "cmdline": "/usr/bin/containerd-shim-runc-v2 -namespace moby -id 2be1de5f5fafb89991873dd415bd11547f7b7123ef7284c84ef56f4c27b7197a -address /run/containerd/containerd.sock ", "vsz_kb": 1233580, "rss_kb": 10968, "mem_pct": "31.14", "proc_jiffies": 11000000, "state": "S" },
]
}
```
```bash
cat /proc/sysinfo_so1_202041390
```
Este comando muestra todos los procesos del sistema, el resultado debe ser el siguiente.

```bash
{
  "mem_total_kb": 3960492,
  "mem_free_kb": 1748972,
  "mem_used_kb": 2211520,
  "processes": [
    { "pid": 1, "name": "systemd", "cmdline": "/sbin/init splash ", "vsz_kb": 23284, "rss_kb": 14276, "mem_pct": "0.58", "proc_jiffies": 7035000000, "state": "S" },
    { "pid": 2, "name": "kthreadd", "cmdline": "", "vsz_kb": 0, "rss_kb": 0, "mem_pct": "0.00", "proc_jiffies": 75000000, "state": "S" },
    { "pid": 3, "name": "pool_workqueue_", "cmdline": "", "vsz_kb": 0, "rss_kb": 0, "mem_pct": "0.00", "proc_jiffies": 0, "state": "S" },
    { "pid": 4, "name": "kworker/R-rcu_g", "cmdline": "", "vsz_kb": 0, "rss_kb": 0, "mem_pct": "0.00", "proc_jiffies": 0, "state": "I" },
    { "pid": 5, "name": "kworker/R-sync_", "cmdline": "", "vsz_kb": 0, "rss_kb": 0, "mem_pct": "0.00", "proc_jiffies": 0, "state": "I" },
    { "pid": 6, "name": "kworker/R-kvfre", "cmdline": "", "vsz_kb": 0, "rss_kb": 0, "mem_pct": "0.00", "proc_jiffies": 0, "state": "I" },
    { "pid": 7, "name": "kworker/R-slub_", "cmdline": "", "vsz_kb": 0, "rss_kb": 0, "mem_pct": "0.00", "proc_jiffies": 0, "state": "I" },
    { "pid": 8, "name": "kworker/R-netns", "cmdline": "", "vsz_kb": 0, "rss_kb": 0, "mem_pct": "0.00", "proc_jiffies": 0, "state": "I" },
    { "pid": 9, "name": "kworker/0:0", "cmdline": "", "vsz_kb": 0, "rss_kb": 0, "mem_pct": "0.00", "proc_jiffies": 1000000, "state": "I" },
    { "pid": 10, "name": "kworker/0:1", "cmdline": "", "vsz_kb": 0, "rss_kb": 0, "mem_pct": "0.00", "proc_jiffies": 187000000, "state": "I" },
]
}
```

### Ejecución del sistema
**Paso 1.** Situarse en la carpeta `go-daemon` y ejecutar los siguientes comandos.

```bash
go run main.go
```
Este comando es para la verificación que todo el flujo del `daemon` en go este funcionando correctamente.

Para pausar todo el procedimiento del deamon es ejecutar `CTRL+C`

**Paso 2.** Ahora crearemos el ejecutable de nuestro daemon

```bash
sudo go build -o /usr/local/bin/mydaemon main.go
```
**Paso 3.** Crear un servicio con relacion a nuestro daemon.
```bash
sudo nano /etc/systemd/system/mydaemon.service
```
El comando abre (o crea, si no existe) el archivo `mydaemon.service` dentro de la carpeta de configuración de systemd, para que puedas editar la definición del servicio mydaemon.

De nustro archivo creado, lo siguiente es poner lo siguiente.
```bash
[Unit]
Description=Mi daemon en Go
After=network.target

[Service]
ExecStart=/usr/local/bin/mydaemon
Restart=always

[Install]
WantedBy=multi-user.target
```
**Paso 4.** Activar y arrancar el daemon
```bash
sudo systemctl daemon-reload        # recargar systemd
sudo systemctl enable --now mydaemon
```
Estado del servicio
```bash
systemctl status mydaemon
```
Logs en journald:
```bash
journalctl -u mydaemon -f
```

Logs en el archivo que configuramos (/var/log/mydaemon.log):
```bash
tail -f /var/log/mydaemon.log
```
Detener y Deshabilitar el servicio
```bash
sudo systemctl stop mydaemon.service
sudo systemctl disable mydaemon.service
```

Eliminar el archivo de configuracion.
```bash
sudo rm /etc/systemd/system/mydaemon.service
sudo systemctl daemon-reload
```

## 11. Configuración y umbrales
Los umbrales de CPU y memoria se definen como constantes dentro del daemon
en Go y pueden ajustarse según las necesidades del sistema.

Para la modificación de estos ubicarse en la carpeta `go-daemon` y en la carpeta `var_const` editar el archivo `const.go`

```bash
// Umbrales
	CPU_THRESHOLD = 20.0 // %
	MEM_THRESHOLD = 20.0 // %
// Mínimos
	MIN_LOW_CONTAINERS  = 3
	MIN_HIGH_CONTAINERS = 2
```

## 12. Conclusiones
El sistema demuestra la integración efectiva entre programación a bajo nivel
y aplicaciones de usuario, permitiendo una gestión proactiva y automatizada
de recursos en entornos Linux.
