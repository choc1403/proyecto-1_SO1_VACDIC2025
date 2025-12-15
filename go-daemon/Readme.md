
## Archivo `main.go`
El archivo `main.go` constituye el **punto de entrada del daemon desarrollado en Go**, el cual actúa como el **componente central de control y orquestación del sistema de monitoreo**. Este daemon se ejecuta de forma continua en segundo plano y coordina la interacción entre los módulos de kernel, la base de datos, los mecanismos de automatización y las herramientas de visualización.

Su función principal es **inicializar el entorno**, **recolectar métricas periódicamente**, **persistir información relevante** y **ejecutar acciones correctivas automáticas** para mantener la estabilidad del sistema.

---

### Funcionalidades principales

#### 1. Inicialización del sistema de logging

```go
log.SetFlags(log.LstdFlags | log.Lshortfile)
```

El daemon configura un sistema de logging básico que incluye:

* Marca de tiempo.
* Archivo y línea de código.

Esto facilita la depuración y el seguimiento de eventos durante la ejecución continua del servicio.

---

#### 2. Inicialización de Grafana

```go
utils.StartGrafana()
```

El daemon intenta iniciar automáticamente el servicio de Grafana al momento de arrancar.
En caso de fallo, el sistema continúa su ejecución y registra una advertencia, evitando la interrupción completa del servicio.

---

#### 3. Inicialización de la base de datos SQLite

```go
database.InitDB()
```

Se inicializa la base de datos SQLite utilizada para almacenar métricas históricas recolectadas por el daemon.
Este almacenamiento persistente permite:

* Análisis histórico del rendimiento.
* Integración con dashboards de Grafana.
* Auditoría del comportamiento del sistema a lo largo del tiempo.

La ruta de la base de datos es gestionada mediante constantes globales.

---

#### 4. Preparación del entorno de ejecución

El daemon verifica y crea, si es necesario:

* El directorio de almacenamiento de datos (`./data`).
* Las estructuras requeridas para la ejecución de scripts y persistencia de información.

Esta validación garantiza que el entorno esté listo antes de iniciar el monitoreo continuo.

---

#### 5. Automatización mediante Cron

```go
utils.CreateCron()
```

El daemon configura un **cronjob automático** encargado de ejecutar scripts que generan contenedores Docker de prueba a intervalos regulares.
Este mecanismo permite:

* Simular escenarios de carga constante.
* Evaluar la capacidad del sistema para detectar y corregir condiciones de sobreconsumo de recursos.

---

#### 6. Carga dinámica de módulos del kernel

```go
utils.LoadModules()
```

El daemon se encarga de cargar dinámicamente los módulos de kernel desarrollados en C, los cuales exponen métricas de procesos y contenedores mediante `/proc`.
Este enfoque asegura que el sistema sea completamente funcional desde el arranque del daemon.

---

#### 7. Manejo de señales del sistema

```go
signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
```

El daemon implementa un manejo explícito de señales del sistema operativo, permitiendo:

* Finalización controlada (`SIGINT`, `SIGTERM`).
* Liberación adecuada de recursos.
* Eliminación de cronjobs y contenedores activos.

Este diseño evita estados inconsistentes al detener el servicio.

---

#### 8. Bucle principal de monitoreo

```go
ticker := time.NewTicker(20 * time.Second)
```

El daemon ejecuta un ciclo periódico cada 20 segundos, en el cual:

1. Lee la información expuesta por los módulos de kernel en `/proc`.
2. Procesa y analiza las métricas recolectadas.
3. Evalúa el consumo de recursos frente a umbrales definidos.
4. Toma decisiones automatizadas cuando se detectan condiciones críticas.
5. Persiste los datos en la base de datos.

Antes de iniciar el bucle, se realiza una **primera medición base**, utilizada como referencia para cálculos posteriores (por ejemplo, uso de CPU).

---

#### 9. Apagado controlado y limpieza de recursos

Al recibir una señal de finalización, el daemon ejecuta una secuencia de limpieza que incluye:

* Eliminación del cronjob creado.
* Detención y eliminación de contenedores generados durante la simulación.
* Registro del cierre correcto del servicio.

Esto garantiza que el sistema quede en un estado limpio y estable tras la salida del daemon.

---

### Importancia dentro del sistema

El archivo `main.go` representa el **cerebro del sistema de monitoreo**, ya que:

* Coordina todos los componentes del proyecto.
* Controla el ciclo de vida completo del sistema.
* Transforma métricas de bajo nivel en acciones concretas de estabilización.
* Sirve como puente entre el kernel de Linux y las herramientas de visualización y análisis.

Sin este componente, los módulos de kernel funcionarían de forma aislada, sin capacidad de análisis, persistencia ni automatización.

---
## Ejecución del sistema
**Paso 1.** Ejecutar el main de go

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

