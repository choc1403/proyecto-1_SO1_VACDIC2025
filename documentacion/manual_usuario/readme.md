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
5. Seleccionar el dashboard del proyecto


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





##  SOLUCIÓN DE PROBLEMAS


## Solución de Problemas

### El daemon no inicia
- Verificar permisos de superusuario
- Verificar que Docker esté en ejecución

### No aparecen datos en Grafana
- Verificar que el daemon esté activo
- Verificar la base de datos SQLite


