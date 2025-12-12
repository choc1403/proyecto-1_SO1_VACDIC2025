

# Package `functions`

**Monitorización de CPU, recursos del sistema y gestión inteligente de contenedores Docker en Linux**

Este paquete forma parte del *daemon de monitorización* del proyecto **so1-daemon**, cuyo objetivo es leer métricas del sistema, calcular el uso de recursos por proceso y aplicar políticas automatizadas (como eliminar contenedores) cuando se superan umbrales críticos.

El paquete contiene lógica para:

* Leer métricas del sistema desde `/proc`.
* Calcular uso de CPU por proceso.
* Mapear procesos ↔ contenedores Docker.
* Registrar métricas en la base de datos.
* Decidir si finalizar contenedores que consumen demasiados recursos.

---

# Estructura del paquete

## 1. `functions.cpu.go`

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

## 2. `functions.docker.go`

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

## 3. `functions.logic.go`

El archivo más importante: **la lógica de decisión del daemon**.
Orquesta lectura de métricas, cálculo de uso de recursos y decide si debe terminar contenedores.

---

##  `func ProcessOnce() error`

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

#  Conclusión

El paquete `functions` implementa:
- Monitorización de recursos del sistema
- Cálculo de uso de CPU por proceso
- Correlación proceso ↔ contenedor Docker
- Registro de métricas históricas
- Sistema automático de control y terminación de contenedores
- Políticas de seguridad y prioridades

Es un bloque esencial del daemon encargado de mantener la **estabilidad del sistema** mediante decisiones inteligentes basadas en umbrales de recursos.

---


