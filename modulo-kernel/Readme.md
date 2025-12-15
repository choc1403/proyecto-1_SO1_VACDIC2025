# Ejecución del Kernel
---
Instalación de los recursos de C
```bash
sudo apt install gcc

# ver versión
gcc --version

```
Instalación del uso de MAKEFILE
```bash

sudo apt-get install make
sudo apt-get install build-essential

```
### Verificación de PYTHON instalado

```bash
python3 --version
```

---
## Ejecución del Kernel
**Paso 1.** Antes de compilar nuevamente el proyecto, es recomendable eliminar los archivos generados en compilaciones anteriores.
Para ello, ejecute el siguiente comando en la terminal

``` bash
make clean
```
Este comando se encarga de limpiar todos los archivos temporales y binarios creados previamente.
Una vez finalizada la limpieza, puede proceder a generar los nuevos archivos ejecutando

``` bash
make
```
**Paso 2.** Ejecutar los siguientes comandos para crear los archivos en proc
``` bash
sudo insmod sysinfo.ko # Módulo de Procesos del Sistema
sudo insmod continfo.ko # Módulo de Procesos de Contenedores
```

**Paso 3.** Verificar los `proc` creados
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
**Paso 4.** Ver los logs del kernel.
```bash
sudo dmesg | tail -n 20
```

Para eliminar los modulos
```bash
sudo rmmod continfo
sudo rmmod sysinfo
```
---

## Archivo `sysinfo.c`

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

## Archivo `continfo.c`

### Descripción técnica

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


## Archivo `common.h`

### Descripción técnica

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


