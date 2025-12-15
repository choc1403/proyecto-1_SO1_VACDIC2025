# Manual de Usuario
---
## [Desarrollo de un módulo de kernel en C y un daemon en Go para el monitoreo de procesos y contenedores en Linux ]
---
## Introducción

Este sistema permite monitorear de forma automática el uso de recursos
(CPU y memoria) de los procesos y contenedores Docker en un sistema Linux.

El sistema detecta contenedores con alto consumo de recursos y toma
acciones correctivas de forma autónoma para mantener la estabilidad del
sistema, mostrando la información de manera visual a través de Grafana.





## Requisitos del Sistema

- Virtualizar un Sistema operativo Linux (Ubuntu 24.04.3 LTS)
- Docker instalado y en ejecución
- Go 1.20 o superior
- GCC y herramientas de compilación del kernel
- SQLite3
- Grafana
- Permisos de superusuario (sudo)






## GUÍA DE INSTALACIÓN (PASO A PASO)

### 3.1 Clonar el repositorio


```bash
git clone https://github.com/choc1403/proyecto-1_SO1_VACDIC2025.git
cd proyecto-1_SO1_VACDIC2025
```


### 3.2 Compilar módulos del kernel


```bash
cd kernel_modules
make
```

### 3.3 Ejecutar el daemon


```bash
cd go-daemon
go run main.go
```



📌 Nota importante:

```md
> El daemon debe ejecutarse con privilegios de superusuario
> para poder interactuar con Docker y con /proc.
````

---

## CÓMO EJECUTAR EL SISTEMA
Una vez iniciado el daemon, el sistema comienza a monitorear
automáticamente los procesos y contenedores del sistema.

No se requiere ninguna acción adicional por parte del usuario.


---

## USO DEL SISTEMA (EJEMPLOS PRÁCTICOS)



### Ejemplo 1: Monitoreo automático

- El sistema detecta contenedores activos
- Se registran métricas de CPU y memoria
- Los datos se almacenan en una base de datos SQLite

### Ejemplo 2: Eliminación automática

- Si un contenedor excede los límites definidos
- El sistema detiene y elimina el contenedor
- La acción queda registrada para auditoría




##  DASHBOARD EN GRAFANA (PARA EL USUARIO FINAL)



El sistema incluye un dashboard en Grafana que permite visualizar:

- Uso de CPU por contenedor
- Uso de memoria por contenedor
- Procesos con mayor consumo
- Historial de eliminaciones

### Acceso al Dashboard

1. Abrir un navegador web
2. Ingresar a: http://localhost:3000
3. Usuario: admin
4. Contraseña: admin
5. Configuración de la base de datos.
**Paso 1** Desde el panel de grafana seleccionar en donde dice *Add your first data source*
![Panel de Grafana](https://github.com/choc1403/proyecto-1_SO1_VACDIC2025/blob/master/dashboard/img/panel_grafana.png)

**Paso 2** Seleccionar SQLITE
![seleccion base de datos](https://github.com/choc1403/proyecto-1_SO1_VACDIC2025/blob/master/dashboard/img/configuracion_bd.png)

**Paso 3** Configuración para la conexión a la Base De Datos.
![Configuracion de BD](https://github.com/choc1403/proyecto-1_SO1_VACDIC2025/blob/master/dashboard/img/config_db.png)
```bash
/var/lib/sqlite/monitor.db

mode=ro&_ignore_check_constraints=1
```
Luego de llenar los campos, dar *Save & test*

6. Seleccionar el dashboard del proyecto
**Paso 1** Desde el panel de grafana seleccionar en el menu de grafana, en donde dice *Home*
![Menu de Grafana](https://github.com/choc1403/proyecto-1_SO1_VACDIC2025/blob/master/dashboard/img/configuracion_dashboard.png)

Y seleccionamos en donde dice *Dashboards*

**Paso 2** Le damos click al boton de *New* y le damos a *Import*
![Dashboards](https://github.com/choc1403/proyecto-1_SO1_VACDIC2025/blob/master/dashboard/img/config_dashboard.png)

**Paso 3** Le damos click a donde dice *Upload dashboard JSON file*, aqui nos vamos a la carpeta de dashboard de nuestro proyecto, y seleccionamos el archivo *dashboard.json*

Luego se procede a dar click a *Load* y ya estaria conectado al Dashboard
---

## ARQUITECTURA DEL SISTEMA 

El sistema está compuesto por tres partes principales:

1. Un módulo del kernel que obtiene información de los procesos
2. Un daemon que analiza la información y toma decisiones
3. Un dashboard que muestra la información al usuario

Estas partes trabajan de forma automática y transparente.


---

##  DIAGRAMAS 

### Diagrama de Flujo (ASCII)



### Flujo General del Sistema

![Arquitectura del proyecto](https://github.com/choc1403/proyecto-1_SO1_VACDIC2025/blob/master/documentacion/manual_tecnico/img/arquitectura.png)

### PANEL DE CONTENEDORES
![Panel de contenedores](https://github.com/choc1403/proyecto-1_SO1_VACDIC2025/blob/master/dashboard/img/panel_contenedores_202041390.png)

### PANEL DE PROCESOS DEL SISTEMA
![Panel de contenedores](https://github.com/choc1403/proyecto-1_SO1_VACDIC2025/blob/master/dashboard/img/panel_contenedores_202041390.png)


##  SOLUCIÓN DE PROBLEMAS


### El daemon no inicia
- Verificar permisos de superusuario
```bash
whoami

```

Resultado esperado: *root*
Si no es root, debes usar sudo.

- Verificar que Docker esté en ejecución

```bash
sudo systemctl status docker

docker ps
```


### No aparecen datos en Grafana
- Verificar que el daemon esté activo

```bash
ps aux | grep so1-daemon

```
- Verificar la base de datos SQLite

```bash
ls -l data/

```